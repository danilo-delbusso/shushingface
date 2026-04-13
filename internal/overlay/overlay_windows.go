//go:build windows

package overlay

import (
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"sync/atomic"
	"syscall"
	"unsafe"

	"codeberg.org/dbus/shushingface/internal/platform"
	"codeberg.org/dbus/shushingface/internal/win32"
)

func capability() platform.Capability { return platform.Supported() }

// Window-style and message constants used only by the overlay.
const (
	wsPopup        = 0x80000000
	wsExTransparent = 0x00000020 // mouse events fall through
	wsExTopmost    = 0x00000008
	wsExNoActivate = 0x08000000
	wsExToolWindow = 0x00000080

	swHide           = 0
	swShowNoActivate = 4

	wmPaint   = 0x000F
	wmTimer   = 0x0113
	wmDestroy = 0x0002

	overlayWidth   = 96
	overlayHeight  = 22
	overlayCorner  = 8  // window corner radius (rounded rectangle)
	overlayMargin  = 12 // distance above the work-area bottom (taskbar)

	swpNoActivate = 0x0010
	swpShowWindow = 0x0040
	hwndTopmost   = ^uintptr(0) // (HWND)-1

	// Visual tuning.
	numBars       = 7
	barWidth      = 2
	barGap        = 3
	barRadius     = 1
	barMinHeight  = 2 // px — idle bars never collapse to 0
	barMaxHeight  = 14
	levelGain     = 8  // amplification applied to RMS before mapping to bar height
	smoothingStep = 0.25 // 0..1 — fraction of the gap between current and target heights closed each frame
	idleAmplitude = 1.5  // px peak of the resting wave when the mic is silent
	idleFreq      = 0.012 // rad / frame — slow gentle ripple
	timerID       = 1
	timerInterval = 33 // ms (~30 fps)
)

var (
	u = win32.User32()
	g = win32.GDI32()

	procRegisterClassExW = u.NewProc("RegisterClassExW")
	procCreateWindowExW  = u.NewProc("CreateWindowExW")
	procDefWindowProcW   = u.NewProc("DefWindowProcW")
	procShowWindow       = u.NewProc("ShowWindow")
	procDestroyWindow    = u.NewProc("DestroyWindow")
	procSetWindowPos     = u.NewProc("SetWindowPos")
	procSetWindowRgn     = u.NewProc("SetWindowRgn")
	procInvalidateRect   = u.NewProc("InvalidateRect")
	procBeginPaint       = u.NewProc("BeginPaint")
	procEndPaint         = u.NewProc("EndPaint")
	procFillRect         = u.NewProc("FillRect")
	procLoadCursorW      = u.NewProc("LoadCursorW")
	procSetTimer         = u.NewProc("SetTimer")
	procKillTimer        = u.NewProc("KillTimer")

	procCreateSolidBrush   = g.NewProc("CreateSolidBrush")
	procCreateRoundRectRgn = g.NewProc("CreateRoundRectRgn")
	procDeleteObject       = g.NewProc("DeleteObject")
	procRoundRect          = g.NewProc("RoundRect")
	procEllipse            = g.NewProc("Ellipse")
	procSelectObject       = g.NewProc("SelectObject")
	procGetStockObject     = g.NewProc("GetStockObject")
)

// Stock object indexes.
const (
	nullPen = 8
)

type wndclassex struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type paintStruct struct {
	Hdc         uintptr
	FErase      int32
	RcPaint     win32.RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

type overlay struct {
	threadID uint32
	hwnd     uintptr
	bgBrush  uintptr // black
	fgBrush  uintptr // bar fill
	cmds     chan command
}

type cmdKind int

const (
	cmdShow cmdKind = iota
	cmdHide
	cmdClose
)

type command struct {
	kind cmdKind
	resp chan error
}

const wmOverlayWake = win32.WMApp

// activeBars holds the moving target heights that paintOverlay reads.
// Audio thread writes targets via push; paint thread does the visible
// per-frame interpolation in paintOverlay using its own non-shared
// state. Accessed via atomics so we don't lock in the audio callback.
type barHistory struct {
	targets [numBars]uint32 // float32 bits, target heights in [0,1]
}

func (b *barHistory) push(level float32) {
	if level < 0 {
		level = 0
	}
	if level > 1 {
		level = 1
	}
	// Shift the target series left by one and push the freshest sample
	// into the rightmost slot. The visible bars then chase these targets
	// each render frame, so even a single audio update produces several
	// frames of smooth motion across the bar bank.
	for i := 0; i < numBars-1; i++ {
		atomic.StoreUint32(&b.targets[i], atomic.LoadUint32(&b.targets[i+1]))
	}
	atomic.StoreUint32(&b.targets[numBars-1], math32Bits(level))
}

func (b *barHistory) snapshot() [numBars]float32 {
	var out [numBars]float32
	for i := range b.targets {
		out[i] = math32FromBits(atomic.LoadUint32(&b.targets[i]))
	}
	return out
}

// Tiny helpers to avoid importing math.Float32bits / Float32frombits at
// the top of every reference.
func math32Bits(f float32) uint32     { return *(*uint32)(unsafe.Pointer(&f)) }
func math32FromBits(u uint32) float32 { return *(*float32)(unsafe.Pointer(&u)) }

// Single global state — only one overlay window is ever shown at once.
var (
	activeBars      barHistory
	smoothedHeights [numBars]float32 // per-frame interpolated heights, paint-thread only
	idlePhase       float32          // monotonically increasing radians for the resting wave
	currentMode     uint32           // overlay.Mode read atomically by paint thread
	processingPhase float32          // monotonically increasing radians for the loader
)

// Loader visuals.
const (
	loaderDots   = 3
	loaderRadius = 2
	loaderSpace  = 6
	loaderSpeed  = 0.16 // rad / frame — full pulse cycle takes ~40 frames (~1.3 s)
)

// New creates the overlay; spawns a dedicated message-loop goroutine.
func New() Overlay {
	ov := &overlay{cmds: make(chan command, 8)}
	ready := make(chan error, 1)
	go ov.loop(ready)
	if err := <-ready; err != nil {
		slog.Warn("overlay init failed", "error", err)
		return stub{}
	}
	return ov
}

func (o *overlay) Show(text string, opacity float64) error {
	// text and opacity ignored: the new overlay is always the bar widget.
	resp := make(chan error, 1)
	o.cmds <- command{kind: cmdShow, resp: resp}
	o.wake()
	return <-resp
}

func (o *overlay) Hide() error {
	resp := make(chan error, 1)
	o.cmds <- command{kind: cmdHide, resp: resp}
	o.wake()
	return <-resp
}

func (o *overlay) Close() error {
	resp := make(chan error, 1)
	o.cmds <- command{kind: cmdClose, resp: resp}
	o.wake()
	return <-resp
}

func (o *overlay) SetLevel(level float32) { activeBars.push(level) }

func (o *overlay) SetMode(mode Mode) { atomic.StoreUint32(&currentMode, uint32(mode)) }

func (o *overlay) wake() {
	win32.PostThreadMessageW(o.threadID, wmOverlayWake, 0, 0)
}

var (
	classRegistered bool
	classNamePtr    *uint16
)

// loop is pinned to one OS thread (Win32 windows live on the thread that
// created them).
func (o *overlay) loop(ready chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	o.threadID = win32.GetCurrentThreadID()

	if err := o.ensureClass(); err != nil {
		ready <- err
		return
	}

	hInst := win32.GetModuleHandleW()
	hwnd, _, err := procCreateWindowExW.Call(
		wsExTransparent|wsExTopmost|wsExNoActivate|wsExToolWindow,
		uintptr(unsafe.Pointer(classNamePtr)),
		0,
		wsPopup,
		0, 0, overlayWidth, overlayHeight,
		0, 0, hInst, 0,
	)
	if hwnd == 0 {
		ready <- fmt.Errorf("CreateWindowEx: %w", err)
		return
	}
	o.hwnd = hwnd

	// Clip the window to a rounded rectangle. Pixels outside the region
	// are never painted by the system, giving us crisp rounded corners
	// without needing per-pixel alpha.
	rgn, _, _ := procCreateRoundRectRgn.Call(
		0, 0, overlayWidth+1, overlayHeight+1,
		overlayCorner*2, overlayCorner*2,
	)
	procSetWindowRgn.Call(o.hwnd, rgn, 1)

	// Brushes: solid black background, pure white bars.
	bg, _, _ := procCreateSolidBrush.Call(rgb(0, 0, 0))
	fg, _, _ := procCreateSolidBrush.Call(rgb(255, 255, 255))
	o.bgBrush = bg
	o.fgBrush = fg

	ready <- nil

	var msg win32.MSG
	for {
		select {
		case c := <-o.cmds:
			if o.handleCommand(c) {
				return
			}
			continue
		default:
		}
		if win32.GetMessageW(&msg) <= 0 {
			return
		}
		win32.TranslateMessage(&msg)
		win32.DispatchMessageW(&msg)
	}
}

func (o *overlay) handleCommand(c command) (done bool) {
	switch c.kind {
	case cmdShow:
		c.resp <- o.doShow()
	case cmdHide:
		c.resp <- o.doHide()
	case cmdClose:
		if o.hwnd != 0 {
			procKillTimer.Call(o.hwnd, timerID)
			procDestroyWindow.Call(o.hwnd)
			o.hwnd = 0
		}
		if o.bgBrush != 0 {
			procDeleteObject.Call(o.bgBrush)
			o.bgBrush = 0
		}
		if o.fgBrush != 0 {
			procDeleteObject.Call(o.fgBrush)
			o.fgBrush = 0
		}
		c.resp <- nil
		return true
	}
	return false
}

func (o *overlay) ensureClass() error {
	if classRegistered {
		return nil
	}
	name, err := syscall.UTF16PtrFromString("ShushingfaceOverlay")
	if err != nil {
		return err
	}
	classNamePtr = name

	hInst := win32.GetModuleHandleW()
	cursor, _, _ := procLoadCursorW.Call(0, 32512) // IDC_ARROW

	wc := wndclassex{
		lpfnWndProc:   syscall.NewCallback(wndProc),
		hInstance:     hInst,
		hCursor:       cursor,
		lpszClassName: name,
	}
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	r, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 {
		return fmt.Errorf("RegisterClassEx: %w", err)
	}
	classRegistered = true
	return nil
}

func wndProc(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	switch msg {
	case wmPaint:
		paintOverlay(hwnd)
		return 0
	case wmTimer:
		// 30 Hz: invalidate so wmPaint runs.
		procInvalidateRect.Call(hwnd, 0, 0)
		return 0
	case wmDestroy:
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wparam, lparam)
	return r
}

// activeOverlayBrushes is read by paintOverlay (called on the message thread).
var activeOverlayBrushes struct {
	bg uintptr
	fg uintptr
}

func paintOverlay(hwnd uintptr) {
	var ps paintStruct
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

	var rc win32.RECT
	win32.GetWindowRect(hwnd, &rc)
	rc.Right -= rc.Left
	rc.Bottom -= rc.Top
	rc.Left, rc.Top = 0, 0

	// Fill black background. The window region clips to an ellipse so
	// only the inside of the ellipse actually paints.
	if activeOverlayBrushes.bg != 0 {
		procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), activeOverlayBrushes.bg)
	}

	if activeOverlayBrushes.fg == 0 {
		return
	}
	nullPenObj, _, _ := procGetStockObject.Call(nullPen)
	oldPen, _, _ := procSelectObject.Call(hdc, nullPenObj)
	oldBrush, _, _ := procSelectObject.Call(hdc, activeOverlayBrushes.fg)
	defer procSelectObject.Call(hdc, oldPen)
	defer procSelectObject.Call(hdc, oldBrush)

	switch Mode(atomic.LoadUint32(&currentMode)) {
	case ModeProcessing:
		paintLoader(hdc, &rc)
	default:
		paintBars(hdc, &rc)
	}
}

func paintBars(hdc uintptr, rc *win32.RECT) {
	// Advance the resting-wave phase so quiet moments still ripple.
	idlePhase += idleFreq * float32(numBars)

	targets := activeBars.snapshot()
	totalW := int32(numBars*barWidth + (numBars-1)*barGap)
	startX := (rc.Right - totalW) / 2
	centerY := rc.Bottom / 2

	for i, target := range targets {
		// Interpolate the visible height toward this bar's target so a
		// single audio update animates over several frames rather than
		// snapping.
		amplified := target * levelGain
		if amplified > 1 {
			amplified = 1
		}
		targetH := float32(barMinHeight) + amplified*float32(barMaxHeight-barMinHeight)
		idle := float32(math.Sin(float64(idlePhase+float32(i)*0.9))) * idleAmplitude
		targetH += idle
		smoothedHeights[i] += (targetH - smoothedHeights[i]) * smoothingStep

		h := int32(smoothedHeights[i] + 0.5)
		if h < barMinHeight {
			h = barMinHeight
		}
		if h > barMaxHeight {
			h = barMaxHeight
		}
		x := startX + int32(i)*int32(barWidth+barGap)
		y := centerY - h/2
		procRoundRect.Call(
			hdc,
			uintptr(x), uintptr(y),
			uintptr(x+barWidth), uintptr(y+h),
			barRadius*2, barRadius*2,
		)
	}
}

// paintLoader draws three dots whose alpha-equivalent (here: drawn vs
// shrunken-to-1px) chases a sine wave so the row reads as a "typing
// indicator" — small, calm, obvious that work is in progress.
func paintLoader(hdc uintptr, rc *win32.RECT) {
	processingPhase += loaderSpeed
	totalW := int32(loaderDots*loaderRadius*2 + (loaderDots-1)*loaderSpace)
	startX := (rc.Right - totalW) / 2
	centerY := rc.Bottom / 2
	for i := 0; i < loaderDots; i++ {
		// Phase-shift each dot by 1/3 cycle so the bump travels.
		phase := float64(processingPhase) - float64(i)*(2*math.Pi/float64(loaderDots))
		// Map [-1,1] to [0,1] so dots breathe between minimum and full.
		amp := (math.Sin(phase) + 1) / 2
		// Scale radius between half and full size.
		r := int32(float64(loaderRadius)*0.5 + amp*float64(loaderRadius)*0.5)
		if r < 1 {
			r = 1
		}
		cx := startX + int32(i)*(loaderRadius*2+loaderSpace) + loaderRadius
		cy := centerY
		procEllipse.Call(
			hdc,
			uintptr(cx-r), uintptr(cy-r),
			uintptr(cx+r), uintptr(cy+r),
		)
	}
}

func (o *overlay) doShow() error {
	if o.hwnd == 0 {
		return nil
	}
	x, y := positionAboveTaskbar(overlayWidth, overlayHeight, overlayMargin)
	procSetWindowPos.Call(o.hwnd, hwndTopmost,
		uintptr(x), uintptr(y),
		uintptr(int32(overlayWidth)), uintptr(int32(overlayHeight)),
		swpNoActivate|swpShowWindow)

	activeOverlayBrushes.bg = o.bgBrush
	activeOverlayBrushes.fg = o.fgBrush

	// Reset bar history + smoothed render state so the new session starts
	// from idle, default back to recording mode, then begin the redraw timer.
	activeBars = barHistory{}
	smoothedHeights = [numBars]float32{}
	idlePhase = 0
	processingPhase = 0
	atomic.StoreUint32(&currentMode, uint32(ModeRecording))
	procSetTimer.Call(o.hwnd, timerID, timerInterval, 0)

	procInvalidateRect.Call(o.hwnd, 0, 1)
	procShowWindow.Call(o.hwnd, swShowNoActivate)
	return nil
}

func (o *overlay) doHide() error {
	if o.hwnd == 0 {
		return nil
	}
	procKillTimer.Call(o.hwnd, timerID)
	procShowWindow.Call(o.hwnd, swHide)
	return nil
}

// positionAboveTaskbar returns the top-left screen coords for an overlay
// of the given size, centered horizontally and pinned just above the
// taskbar of the active monitor (the one containing the foreground
// window). Falls back to the primary monitor's work area if there is
// no foreground window.
func positionAboveTaskbar(w, h, margin int32) (int32, int32) {
	hmon := win32.MonitorFromWindow(win32.GetForegroundWindow(), win32.MonitorDefaultToNearest)
	var mi win32.MONITORINFO
	if !win32.GetMonitorInfo(hmon, &mi) {
		// No monitor info — pick a safe corner. Better than crashing.
		return 100, 100
	}
	work := mi.RcWork
	cx := work.Left + (work.Right-work.Left)/2
	return cx - w/2, work.Bottom - h - margin
}

func rgb(r, g, b byte) uintptr {
	return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16
}

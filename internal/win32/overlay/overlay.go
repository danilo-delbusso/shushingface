// Package winoverlay is the Win32 implementation of the floating recording
// indicator. It owns the entire window lifecycle (window class, message
// pump, GDI brushes), the bar/loader animation state, and monitor-aware
// positioning. The parent internal/overlay package wraps this in a thin
// build-tagged adapter that satisfies overlay.Overlay.
package winoverlay

import (
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"sync/atomic"
	"syscall"
	"unsafe"

	"codeberg.org/dbus/shushingface/internal/win32"
)

// --- window-style and message constants used only here ----------------------

const (
	wsPopup         = 0x80000000
	wsExTransparent = 0x00000020 // mouse events fall through
	wsExTopmost     = 0x00000008
	wsExNoActivate  = 0x08000000
	wsExToolWindow  = 0x00000080

	swHide           = 0
	swShowNoActivate = 4

	wmPaint   = 0x000F
	wmTimer   = 0x0113
	wmDestroy = 0x0002

	overlayWidth  = 96
	overlayHeight = 22
	overlayCorner = 8  // window corner radius (rounded rectangle)
	overlayMargin = 12 // distance above the work-area bottom (taskbar)

	swpNoActivate = 0x0010
	swpShowWindow = 0x0040
	hwndTopmost   = ^uintptr(0) // (HWND)-1

	// Visual tuning.
	numBars       = 7
	barWidth      = 4
	barGap        = 3
	barRadius     = 2
	barMinHeight  = 3 // px — idle bars never collapse to 0
	barMaxHeight  = 16
	levelGain     = 8    // amplification applied to RMS before mapping to bar height
	smoothingStep = 0.25 // 0..1 — fraction of the gap closed each frame
	idleAmplitude = 1.5  // px peak of the resting wave when the mic is silent
	idleFreq      = 0.012
	timerID       = 1
	timerInterval = 33 // ms (~30 fps)

	// Loader (processing-mode) visuals.
	loaderDots   = 3
	loaderRadius = 4
	loaderSpace  = 8
	loaderSpeed  = 0.16

	wmOverlayWake = win32.WMApp
)

// Stock object indexes.
const nullPen = 8

// --- proc handles -----------------------------------------------------------

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

// --- public API -------------------------------------------------------------

// Window is the live Win32 overlay. There is at most one shown at a time.
type Window struct {
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

// New creates the overlay window on a dedicated message-loop goroutine.
// Returns the window once that thread is initialised and ready to accept
// commands, or an error if class registration / window creation fails.
func New() (*Window, error) {
	w := &Window{cmds: make(chan command, 8)}
	ready := make(chan error, 1)
	go w.loop(ready)
	if err := <-ready; err != nil {
		return nil, err
	}
	return w, nil
}

// Show pops the overlay above the taskbar of the active monitor and
// kicks off the bar / loader animation timer.
func (w *Window) Show() error {
	resp := make(chan error, 1)
	w.cmds <- command{kind: cmdShow, resp: resp}
	w.wake()
	return <-resp
}

// Hide stops the animation timer and hides the window. Idempotent.
func (w *Window) Hide() error {
	resp := make(chan error, 1)
	w.cmds <- command{kind: cmdHide, resp: resp}
	w.wake()
	return <-resp
}

// Close destroys the window and frees GDI brushes; the window cannot
// be reused after this returns.
func (w *Window) Close() error {
	resp := make(chan error, 1)
	w.cmds <- command{kind: cmdClose, resp: resp}
	w.wake()
	return <-resp
}

// SetLevel pushes a fresh microphone amplitude in [0,1] for the bar
// animation. Cheap, non-blocking, safe from any goroutine.
func (w *Window) SetLevel(level float32) { activeBars.push(level) }

// SetProcessing toggles between the live recording bars (false) and the
// "transcribing" loader (true). Cheap and non-blocking.
func (w *Window) SetProcessing(processing bool) {
	var v uint32
	if processing {
		v = 1
	}
	atomic.StoreUint32(&currentMode, v)
}

func (w *Window) wake() {
	win32.PostThreadMessageW(w.threadID, wmOverlayWake, 0, 0)
}

// --- per-frame animation state (paint thread) -------------------------------

// barHistory holds the moving target heights that paintBars reads. Audio
// thread writes targets via push; paint thread does the visible per-frame
// interpolation in paintBars using its own non-shared state. Accessed via
// atomics so we don't lock in the audio callback.
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

func math32Bits(f float32) uint32     { return *(*uint32)(unsafe.Pointer(&f)) }
func math32FromBits(u uint32) float32 { return *(*float32)(unsafe.Pointer(&u)) }

// Single global state — only one overlay window is ever shown at once.
var (
	activeBars      barHistory
	smoothedHeights [numBars]float32 // per-frame interpolated heights, paint-thread only
	idlePhase       float32
	processingPhase float32
	currentMode     uint32 // 0 = bars, 1 = loader; atomic
	classRegistered bool
	classNamePtr    *uint16

	// activeOverlayBrushes is read by paintOverlay (called on the message thread).
	activeOverlayBrushes struct {
		bg uintptr
		fg uintptr
	}
)

// --- message loop -----------------------------------------------------------

// loop is pinned to one OS thread (Win32 windows live on the thread that
// created them). All command handling and painting happens here.
func (w *Window) loop(ready chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	w.threadID = win32.GetCurrentThreadID()

	if err := ensureClass(); err != nil {
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
	w.hwnd = hwnd

	// Clip the window to a rounded rectangle. Pixels outside the region
	// are never painted — crisp rounded corners without per-pixel alpha.
	rgn, _, _ := procCreateRoundRectRgn.Call(
		0, 0, overlayWidth+1, overlayHeight+1,
		overlayCorner*2, overlayCorner*2,
	)
	procSetWindowRgn.Call(w.hwnd, rgn, 1)

	bg, _, _ := procCreateSolidBrush.Call(rgb(0, 0, 0))
	fg, _, _ := procCreateSolidBrush.Call(rgb(255, 255, 255))
	w.bgBrush = bg
	w.fgBrush = fg

	ready <- nil

	var msg win32.MSG
	for {
		select {
		case c := <-w.cmds:
			if w.handleCommand(c) {
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

func (w *Window) handleCommand(c command) (done bool) {
	switch c.kind {
	case cmdShow:
		c.resp <- w.doShow()
	case cmdHide:
		c.resp <- w.doHide()
	case cmdClose:
		if w.hwnd != 0 {
			procKillTimer.Call(w.hwnd, timerID)
			procDestroyWindow.Call(w.hwnd)
			w.hwnd = 0
		}
		if w.bgBrush != 0 {
			procDeleteObject.Call(w.bgBrush)
			w.bgBrush = 0
		}
		if w.fgBrush != 0 {
			procDeleteObject.Call(w.fgBrush)
			w.fgBrush = 0
		}
		c.resp <- nil
		return true
	}
	return false
}

func ensureClass() error {
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

// --- painting ---------------------------------------------------------------

func paintOverlay(hwnd uintptr) {
	var ps paintStruct
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

	var rc win32.RECT
	win32.GetWindowRect(hwnd, &rc)
	rc.Right -= rc.Left
	rc.Bottom -= rc.Top
	rc.Left, rc.Top = 0, 0

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

	if atomic.LoadUint32(&currentMode) == 1 {
		paintLoader(hdc, &rc)
	} else {
		paintBars(hdc, &rc)
	}
}

func paintBars(hdc uintptr, rc *win32.RECT) {
	idlePhase += idleFreq * float32(numBars)

	targets := activeBars.snapshot()
	totalW := int32(numBars*barWidth + (numBars-1)*barGap)
	startX := (rc.Right - totalW) / 2
	centerY := rc.Bottom / 2

	for i, target := range targets {
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

// paintLoader draws three dots whose radii chase a sine wave so the row
// reads as a "typing indicator" — small, calm, obvious that work is in
// progress.
func paintLoader(hdc uintptr, rc *win32.RECT) {
	processingPhase += loaderSpeed
	totalW := int32(loaderDots*loaderRadius*2 + (loaderDots-1)*loaderSpace)
	startX := (rc.Right - totalW) / 2
	centerY := rc.Bottom / 2
	for i := 0; i < loaderDots; i++ {
		phase := float64(processingPhase) - float64(i)*(2*math.Pi/float64(loaderDots))
		amp := (math.Sin(phase) + 1) / 2
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

func (w *Window) doShow() error {
	if w.hwnd == 0 {
		return nil
	}
	x, y := positionAboveTaskbar(overlayWidth, overlayHeight, overlayMargin)
	procSetWindowPos.Call(w.hwnd, hwndTopmost,
		uintptr(x), uintptr(y),
		uintptr(int32(overlayWidth)), uintptr(int32(overlayHeight)),
		swpNoActivate|swpShowWindow)

	activeOverlayBrushes.bg = w.bgBrush
	activeOverlayBrushes.fg = w.fgBrush

	// Reset animation state so each session starts from idle, default to
	// the bars view, then begin the redraw timer.
	activeBars = barHistory{}
	smoothedHeights = [numBars]float32{}
	idlePhase = 0
	processingPhase = 0
	atomic.StoreUint32(&currentMode, 0)
	procSetTimer.Call(w.hwnd, timerID, timerInterval, 0)

	procInvalidateRect.Call(w.hwnd, 0, 1)
	procShowWindow.Call(w.hwnd, swShowNoActivate)
	return nil
}

func (w *Window) doHide() error {
	if w.hwnd == 0 {
		return nil
	}
	procKillTimer.Call(w.hwnd, timerID)
	procShowWindow.Call(w.hwnd, swHide)
	return nil
}

// positionAboveTaskbar returns the top-left screen coords for a window
// of the given size, centered horizontally and pinned just above the
// taskbar of the active monitor (the one containing the foreground
// window). Falls back to a safe corner if monitor info fails.
func positionAboveTaskbar(w, h, margin int32) (int32, int32) {
	hmon := win32.MonitorFromWindow(win32.GetForegroundWindow(), win32.MonitorDefaultToNearest)
	var mi win32.MONITORINFO
	if !win32.GetMonitorInfo(hmon, &mi) {
		slog.Debug("winoverlay: GetMonitorInfo failed, falling back to corner")
		return 100, 100
	}
	work := mi.RcWork
	cx := work.Left + (work.Right-work.Left)/2
	return cx - w/2, work.Bottom - h - margin
}

func rgb(r, g, b byte) uintptr {
	return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16
}

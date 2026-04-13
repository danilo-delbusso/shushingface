//go:build windows

package overlay

import (
	"fmt"
	"log/slog"
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

	overlayWidth  = 240
	overlayHeight = 60
	overlayMargin = 12 // distance above the bottom edge of the active window

	swpNoActivate = 0x0010
	swpShowWindow = 0x0040
	hwndTopmost   = ^uintptr(0) // (HWND)-1

	// Visual tuning.
	numBars       = 9
	barWidth      = 4
	barGap        = 4
	barRadius     = 2
	barMinHeight  = 4 // px — idle bars never collapse to 0
	barMaxHeight  = 36
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

	procCreateSolidBrush  = g.NewProc("CreateSolidBrush")
	procCreateEllipticRgn = g.NewProc("CreateEllipticRgn")
	procDeleteObject      = g.NewProc("DeleteObject")
	procRoundRect         = g.NewProc("RoundRect")
	procSelectObject      = g.NewProc("SelectObject")
	procGetStockObject    = g.NewProc("GetStockObject")
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

// activeBars holds the smoothed amplitude history that paintOverlay
// reads. Accessed via atomics from the audio side and the paint side.
type barHistory struct {
	bars [numBars]uint32 // each holds a float32 bits via atomic.StoreUint32
}

func (b *barHistory) push(level float32) {
	if level < 0 {
		level = 0
	}
	if level > 1 {
		level = 1
	}
	// Shift left, push newest into the rightmost slot. Cheap because
	// numBars is small.
	for i := 0; i < numBars-1; i++ {
		atomic.StoreUint32(&b.bars[i], atomic.LoadUint32(&b.bars[i+1]))
	}
	atomic.StoreUint32(&b.bars[numBars-1], math32Bits(level))
}

func (b *barHistory) snapshot() [numBars]float32 {
	var out [numBars]float32
	for i := range b.bars {
		out[i] = math32FromBits(atomic.LoadUint32(&b.bars[i]))
	}
	return out
}

// Tiny helpers to avoid importing math.Float32bits / Float32frombits at
// the top of every reference.
func math32Bits(f float32) uint32     { return *(*uint32)(unsafe.Pointer(&f)) }
func math32FromBits(u uint32) float32 { return *(*float32)(unsafe.Pointer(&u)) }

// Single global state — only one overlay window is ever shown at once.
var activeBars barHistory

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

	// Clip the window to an ellipse via SetWindowRgn. The system never
	// paints pixels outside this region — the result is a true ellipse
	// silhouette, no per-pixel alpha needed.
	rgn, _, _ := procCreateEllipticRgn.Call(0, 0, overlayWidth, overlayHeight)
	procSetWindowRgn.Call(o.hwnd, rgn, 1)

	// Brushes: black background, near-white bars.
	bg, _, _ := procCreateSolidBrush.Call(rgb(0, 0, 0))
	fg, _, _ := procCreateSolidBrush.Call(rgb(220, 220, 220))
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

	// Draw the bars. Use null pen + foreground brush so RoundRect fills
	// without an outline.
	if activeOverlayBrushes.fg != 0 {
		nullPenObj, _, _ := procGetStockObject.Call(nullPen)
		oldPen, _, _ := procSelectObject.Call(hdc, nullPenObj)
		oldBrush, _, _ := procSelectObject.Call(hdc, activeOverlayBrushes.fg)
		defer procSelectObject.Call(hdc, oldPen)
		defer procSelectObject.Call(hdc, oldBrush)

		bars := activeBars.snapshot()
		totalW := int32(numBars*barWidth + (numBars-1)*barGap)
		startX := (rc.Right - totalW) / 2
		centerY := rc.Bottom / 2
		for i, lvl := range bars {
			h := int32(float32(barMinHeight) + lvl*float32(barMaxHeight-barMinHeight)*4)
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
}

func (o *overlay) doShow() error {
	if o.hwnd == 0 {
		return nil
	}
	x, y := positionUnderActive(overlayWidth, overlayHeight, overlayMargin)
	procSetWindowPos.Call(o.hwnd, hwndTopmost,
		uintptr(x), uintptr(y),
		uintptr(int32(overlayWidth)), uintptr(int32(overlayHeight)),
		swpNoActivate|swpShowWindow)

	activeOverlayBrushes.bg = o.bgBrush
	activeOverlayBrushes.fg = o.fgBrush

	// Reset bar history so the new session starts from idle, then begin
	// the redraw timer.
	activeBars = barHistory{}
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

func positionUnderActive(w, h, margin int32) (int32, int32) {
	hwnd := win32.GetForegroundWindow()
	if hwnd == 0 {
		return 100, 100
	}
	var rc win32.RECT
	win32.GetWindowRect(hwnd, &rc)
	cx := rc.Left + (rc.Right-rc.Left)/2
	return cx - w/2, rc.Bottom - h - margin
}

func rgb(r, g, b byte) uintptr {
	return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16
}

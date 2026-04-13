//go:build windows

package overlay

import (
	"fmt"
	"log/slog"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	wsPopup        = 0x80000000
	wsExLayered    = 0x00080000
	wsExTransparent = 0x00000020
	wsExTopmost    = 0x00000008
	wsExNoActivate = 0x08000000
	wsExToolWindow = 0x00000080

	swHide        = 0
	swShowNoActivate = 4

	lwaAlpha = 0x00000002

	wmPaint   = 0x000F
	wmDestroy = 0x0002
	wmClose   = 0x0010
	wmApp     = 0x8000
	wmShow    = wmApp + 1
	wmHide    = wmApp + 2

	overlayWidth  = 180
	overlayHeight = 36
	overlayMargin = 24 // distance above the bottom edge of the active window

	dtCenter   = 0x00000001
	dtVCenter  = 0x00000004
	dtSingleLine = 0x00000020

	srcCopy = 0x00CC0020
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procRegisterClassExW       = user32.NewProc("RegisterClassExW")
	procCreateWindowExW        = user32.NewProc("CreateWindowExW")
	procDefWindowProcW         = user32.NewProc("DefWindowProcW")
	procShowWindow             = user32.NewProc("ShowWindow")
	procDestroyWindow          = user32.NewProc("DestroyWindow")
	procSetLayeredWindowAttrs  = user32.NewProc("SetLayeredWindowAttributes")
	procSetWindowPos           = user32.NewProc("SetWindowPos")
	procGetForegroundWindow    = user32.NewProc("GetForegroundWindow")
	procGetWindowRect          = user32.NewProc("GetWindowRect")
	procInvalidateRect         = user32.NewProc("InvalidateRect")
	procBeginPaint             = user32.NewProc("BeginPaint")
	procEndPaint               = user32.NewProc("EndPaint")
	procFillRect               = user32.NewProc("FillRect")
	procDrawTextW              = user32.NewProc("DrawTextW")
	procGetMessageW            = user32.NewProc("GetMessageW")
	procDispatchMessageW       = user32.NewProc("DispatchMessageW")
	procTranslateMessage       = user32.NewProc("TranslateMessage")
	procPostThreadMessage      = user32.NewProc("PostThreadMessageW")
	procPostMessageW           = user32.NewProc("PostMessageW")
	procLoadCursorW            = user32.NewProc("LoadCursorW")

	procCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procSetBkMode              = gdi32.NewProc("SetBkMode")
	procSetTextColor           = gdi32.NewProc("SetTextColor")

	procGetCurrentThreadID     = kernel32.NewProc("GetCurrentThreadId")
	procGetModuleHandleW       = kernel32.NewProc("GetModuleHandleW")
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

type rect struct {
	Left, Top, Right, Bottom int32
}

type point struct{ X, Y int32 }

type msg struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       point
	LPrivate uint32
}

type paintStruct struct {
	Hdc         uintptr
	FErase      int32
	RcPaint     rect
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

type overlay struct {
	threadID uint32
	hwnd     uintptr
	bg       uintptr // brush
	cmds     chan command
	done     chan struct{}
	current  string
}

type cmdKind int

const (
	cmdShow cmdKind = iota
	cmdHide
	cmdClose
)

type command struct {
	kind    cmdKind
	text    string
	opacity float64
	resp    chan error
}

// New creates the overlay; spawns a dedicated message-loop goroutine.
func New() Overlay {
	ov := &overlay{
		cmds: make(chan command, 8),
		done: make(chan struct{}),
	}
	ready := make(chan error, 1)
	go ov.loop(ready)
	if err := <-ready; err != nil {
		slog.Warn("overlay init failed", "error", err)
		return stub{}
	}
	return ov
}

type stub struct{}

func (stub) Show(string, float64) error { return nil }
func (stub) Hide() error                { return nil }
func (stub) Close() error               { return nil }

func (o *overlay) Show(text string, opacity float64) error {
	resp := make(chan error, 1)
	o.cmds <- command{kind: cmdShow, text: text, opacity: opacity, resp: resp}
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

func (o *overlay) wake() {
	procPostThreadMessage.Call(uintptr(o.threadID), wmApp, 0, 0)
}

var classRegistered bool
var classNamePtr *uint16

func (o *overlay) loop(ready chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	tid, _, _ := procGetCurrentThreadID.Call()
	o.threadID = uint32(tid)

	if err := o.ensureClass(); err != nil {
		ready <- err
		return
	}

	hInst, _, _ := procGetModuleHandleW.Call(0)
	hwnd, _, err := procCreateWindowExW.Call(
		wsExLayered|wsExTransparent|wsExTopmost|wsExNoActivate|wsExToolWindow,
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

	// Solid red background brush; layered alpha makes it translucent.
	brush, _, _ := procCreateSolidBrush.Call(rgb(220, 38, 38))
	o.bg = brush

	ready <- nil

	var m msg
	for {
		select {
		case c := <-o.cmds:
			switch c.kind {
			case cmdShow:
				err := o.doShow(c.text, c.opacity)
				c.resp <- err
			case cmdHide:
				err := o.doHide()
				c.resp <- err
			case cmdClose:
				if o.hwnd != 0 {
					procDestroyWindow.Call(o.hwnd)
					o.hwnd = 0
				}
				if o.bg != 0 {
					procDeleteObject.Call(o.bg)
					o.bg = 0
				}
				c.resp <- nil
				return
			}
			continue
		default:
		}

		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			return
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
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

	hInst, _, _ := procGetModuleHandleW.Call(0)
	cursor, _, _ := procLoadCursorW.Call(0, 32512) // IDC_ARROW

	wc := wndclassex{
		cbSize:        uint32(unsafe.Sizeof(wndclassex{})),
		lpfnWndProc:   syscall.NewCallback(wndProc),
		hInstance:     hInst,
		hCursor:       cursor,
		hbrBackground: 0,
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
	case wmDestroy:
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wparam, lparam)
	return r
}

// activeOverlayState is read by paintOverlay (called on the message thread).
var activeOverlayState struct {
	text  string
	brush uintptr
}

func paintOverlay(hwnd uintptr) {
	var ps paintStruct
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

	var rc rect
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	rc.Right -= rc.Left
	rc.Bottom -= rc.Top
	rc.Left, rc.Top = 0, 0

	if activeOverlayState.brush != 0 {
		procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), activeOverlayState.brush)
	}

	procSetBkMode.Call(hdc, 1) // TRANSPARENT
	procSetTextColor.Call(hdc, rgb(255, 255, 255))

	text := activeOverlayState.text
	if text == "" {
		text = "Recording"
	}
	utf16, _ := syscall.UTF16FromString(text)
	negOne := int32(-1)
	procDrawTextW.Call(
		hdc,
		uintptr(unsafe.Pointer(&utf16[0])),
		uintptr(negOne),
		uintptr(unsafe.Pointer(&rc)),
		dtCenter|dtVCenter|dtSingleLine,
	)
}

func (o *overlay) doShow(text string, opacity float64) error {
	if o.hwnd == 0 {
		return nil
	}
	if opacity <= 0 {
		opacity = 0.4
	}
	if opacity > 1 {
		opacity = 1
	}
	alpha := byte(opacity * 255)
	procSetLayeredWindowAttrs.Call(o.hwnd, 0, uintptr(alpha), lwaAlpha)

	x, y := positionUnderActive(overlayWidth, overlayHeight, overlayMargin)
	const swpNoActivate = 0x0010
	const swpShowWindow = 0x0040
	procSetWindowPos.Call(o.hwnd, ^uintptr(0) /* HWND_TOPMOST = -1 */, uintptr(x), uintptr(y), uintptr(int32(overlayWidth)), uintptr(int32(overlayHeight)), swpNoActivate|swpShowWindow)

	activeOverlayState.text = text
	activeOverlayState.brush = o.bg
	procInvalidateRect.Call(o.hwnd, 0, 1)
	procShowWindow.Call(o.hwnd, swShowNoActivate)
	return nil
}

func (o *overlay) doHide() error {
	if o.hwnd == 0 {
		return nil
	}
	procShowWindow.Call(o.hwnd, swHide)
	return nil
}

func positionUnderActive(w, h, margin int32) (int32, int32) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return 100, 100
	}
	var rc rect
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	cx := rc.Left + (rc.Right-rc.Left)/2
	x := cx - w/2
	y := rc.Bottom - h - margin
	return x, y
}

func rgb(r, g, b byte) uintptr {
	return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16
}

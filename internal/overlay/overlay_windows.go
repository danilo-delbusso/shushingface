//go:build windows

package overlay

import (
	"fmt"
	"log/slog"
	"runtime"
	"syscall"
	"unsafe"

	"codeberg.org/dbus/shushingface/internal/platform"
	"codeberg.org/dbus/shushingface/internal/win32"
)

func capability() platform.Capability { return platform.Supported() }

// Window-style and message constants used only by the overlay. Shared
// messages like WM_QUIT live in the win32 package.
const (
	wsPopup         = 0x80000000
	wsExLayered     = 0x00080000
	wsExTransparent = 0x00000020
	wsExTopmost     = 0x00000008
	wsExNoActivate  = 0x08000000
	wsExToolWindow  = 0x00000080

	swHide           = 0
	swShowNoActivate = 4

	lwaAlpha = 0x00000002

	wmPaint   = 0x000F
	wmDestroy = 0x0002

	overlayWidth  = 180
	overlayHeight = 36
	overlayMargin = 8 // distance above the bottom edge of the active window

	dtCenter     = 0x00000001
	dtVCenter    = 0x00000004
	dtSingleLine = 0x00000020

	swpNoActivate = 0x0010
	swpShowWindow = 0x0040
	hwndTopmost   = ^uintptr(0) // (HWND)-1
)

// Overlay-specific window/paint procs. Everything that more than one
// package needs (GetForegroundWindow, GetWindowRect, GetMessage, etc.)
// lives in internal/win32; these are the leftovers that are genuinely
// overlay-only.
var (
	u = win32.User32()
	g = win32.GDI32()

	procRegisterClassExW     = u.NewProc("RegisterClassExW")
	procCreateWindowExW      = u.NewProc("CreateWindowExW")
	procDefWindowProcW       = u.NewProc("DefWindowProcW")
	procShowWindow           = u.NewProc("ShowWindow")
	procDestroyWindow        = u.NewProc("DestroyWindow")
	procSetLayeredWindowAttr = u.NewProc("SetLayeredWindowAttributes")
	procSetWindowPos         = u.NewProc("SetWindowPos")
	procInvalidateRect       = u.NewProc("InvalidateRect")
	procBeginPaint           = u.NewProc("BeginPaint")
	procEndPaint             = u.NewProc("EndPaint")
	procFillRect             = u.NewProc("FillRect")
	procDrawTextW            = u.NewProc("DrawTextW")
	procLoadCursorW          = u.NewProc("LoadCursorW")

	procCreateSolidBrush = g.NewProc("CreateSolidBrush")
	procDeleteObject     = g.NewProc("DeleteObject")
	procSetBkMode        = g.NewProc("SetBkMode")
	procSetTextColor     = g.NewProc("SetTextColor")
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
	bg       uintptr // brush
	cmds     chan command
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

// Wake-message code used to poke the overlay's message-loop out of
// GetMessageW when a new command is ready on the cmds chan.
const wmOverlayWake = win32.WMApp

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
	win32.PostThreadMessageW(o.threadID, wmOverlayWake, 0, 0)
}

var (
	classRegistered bool
	classNamePtr    *uint16
)

// loop is pinned to one OS thread (required for Win32 windows/message
// loops) and owns the overlay HWND. Commands from Show/Hide/Close are
// marshalled in via o.cmds + PostThreadMessage.
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

	// Solid red brush; the layered-window alpha makes it translucent.
	brush, _, _ := procCreateSolidBrush.Call(rgb(220, 38, 38))
	o.bg = brush

	ready <- nil

	var msg win32.MSG
	for {
		// Non-blocking: drain any pending commands before sleeping in GetMessageW.
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

// handleCommand returns true when the overlay should tear down and exit.
func (o *overlay) handleCommand(c command) (done bool) {
	switch c.kind {
	case cmdShow:
		c.resp <- o.doShow(c.text, c.opacity)
	case cmdHide:
		c.resp <- o.doHide()
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

	var rc win32.RECT
	win32.GetWindowRect(hwnd, &rc)
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
	procSetLayeredWindowAttr.Call(o.hwnd, 0, uintptr(alpha), lwaAlpha)

	x, y := positionUnderActive(overlayWidth, overlayHeight, overlayMargin)
	procSetWindowPos.Call(o.hwnd, hwndTopmost,
		uintptr(x), uintptr(y),
		uintptr(int32(overlayWidth)), uintptr(int32(overlayHeight)),
		swpNoActivate|swpShowWindow)

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

// positionUnderActive returns the top-left screen coordinates where the
// overlay should sit given the current foreground window. Falls back to a
// fixed corner if no window has focus.
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

// rgb packs 8-bit RGB components into a Win32 COLORREF (0x00BBGGRR).
func rgb(r, g, b byte) uintptr {
	return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16
}

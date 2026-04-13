//go:build windows

package paste

import (
	"fmt"
	"log/slog"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002

	inputKeyboard = 1

	keyeventfKeyup   = 0x0002
	keyeventfUnicode = 0x0004

	vkControl = 0x11
	vkV       = 0x56
)

var (
	user32            = windows.NewLazySystemDLL("user32.dll")
	kernel32          = windows.NewLazySystemDLL("kernel32.dll")
	procOpenClipboard = user32.NewProc("OpenClipboard")
	procEmptyClipboard = user32.NewProc("EmptyClipboard")
	procSetClipboardData = user32.NewProc("SetClipboardData")
	procCloseClipboard = user32.NewProc("CloseClipboard")
	procSendInput      = user32.NewProc("SendInput")
	procGlobalAlloc    = kernel32.NewProc("GlobalAlloc")
	procGlobalLock     = kernel32.NewProc("GlobalLock")
	procGlobalUnlock   = kernel32.NewProc("GlobalUnlock")
	procRtlMoveMemory  = kernel32.NewProc("RtlMoveMemory")
)

type keyboardInput struct {
	typ     uint32
	_pad    uint32
	wVk     uint16
	wScan   uint16
	dwFlags uint32
	time    uint32
	extra   uintptr
	_pad2   [8]byte
}

// Available reports whether auto-paste is supported on this platform.
func Available() bool { return true }

// InstallHint returns an empty string; nothing to install on Windows.
func InstallHint() string { return "" }

// Type places text on the clipboard and simulates Ctrl+V.
func Type(text string) error {
	if err := setClipboard(text); err != nil {
		return fmt.Errorf("set clipboard: %w", err)
	}
	time.Sleep(50 * time.Millisecond)
	if err := sendCtrlV(); err != nil {
		slog.Warn("sendinput failed", "error", err)
		return fmt.Errorf("send ctrl+v: %w", err)
	}
	return nil
}

func setClipboard(text string) error {
	utf16, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}
	size := uintptr(len(utf16) * 2)

	h, _, _ := procGlobalAlloc.Call(gmemMoveable, size)
	if h == 0 {
		return fmt.Errorf("GlobalAlloc failed")
	}

	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		return fmt.Errorf("GlobalLock failed")
	}
	procRtlMoveMemory.Call(ptr, uintptr(unsafe.Pointer(&utf16[0])), size)
	procGlobalUnlock.Call(h)

	r, _, _ := procOpenClipboard.Call(0)
	if r == 0 {
		return fmt.Errorf("OpenClipboard failed")
	}
	defer procCloseClipboard.Call()

	procEmptyClipboard.Call()
	if r, _, _ := procSetClipboardData.Call(cfUnicodeText, h); r == 0 {
		return fmt.Errorf("SetClipboardData failed")
	}
	return nil
}

func sendCtrlV() error {
	inputs := [4]keyboardInput{
		{typ: inputKeyboard, wVk: vkControl},
		{typ: inputKeyboard, wVk: vkV},
		{typ: inputKeyboard, wVk: vkV, dwFlags: keyeventfKeyup},
		{typ: inputKeyboard, wVk: vkControl, dwFlags: keyeventfKeyup},
	}
	ret, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
	if ret != uintptr(len(inputs)) {
		return err
	}
	return nil
}

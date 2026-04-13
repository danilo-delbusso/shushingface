//go:build windows

package win32

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	procOpenClipboard    = user32.NewProc("OpenClipboard")
	procCloseClipboard   = user32.NewProc("CloseClipboard")
	procEmptyClipboard   = user32.NewProc("EmptyClipboard")
	procSetClipboardData = user32.NewProc("SetClipboardData")

	procGlobalAlloc   = kernel32.NewProc("GlobalAlloc")
	procGlobalLock    = kernel32.NewProc("GlobalLock")
	procGlobalUnlock  = kernel32.NewProc("GlobalUnlock")
	procRtlMoveMemory = kernel32.NewProc("RtlMoveMemory")
)

// Clipboard format and GlobalAlloc flag constants.
const (
	CFUnicodeText = 13
	GMemMoveable  = 0x0002
)

// SetClipboardUnicodeText replaces the clipboard with the given text in
// CF_UNICODETEXT format. Handles OpenClipboard / EmptyClipboard /
// SetClipboardData / CloseClipboard as one unit.
func SetClipboardUnicodeText(text string) error {
	utf16, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}
	size := uintptr(len(utf16) * 2)

	h, _, _ := procGlobalAlloc.Call(GMemMoveable, size)
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
	if r, _, _ := procSetClipboardData.Call(CFUnicodeText, h); r == 0 {
		return fmt.Errorf("SetClipboardData failed")
	}
	return nil
}

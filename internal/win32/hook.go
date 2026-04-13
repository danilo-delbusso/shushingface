//go:build windows

package win32

import "syscall"

var (
	procSetWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procGetKeyState         = user32.NewProc("GetKeyState")
)

// Hook type and virtual-key constants used by callers of this package.
const (
	WHKeyboardLL = 13

	WMKeyDown    = 0x0100
	WMKeyUp      = 0x0101
	WMSysKeyDown = 0x0104
	WMSysKeyUp   = 0x0105

	VKLControl = 0xA2
	VKRControl = 0xA3
	VKLAlt     = 0xA4
	VKRAlt     = 0xA5
	VKLShift   = 0xA0
	VKRShift   = 0xA1
	VKLWin     = 0x5B
	VKRWin     = 0x5C
)

// SetWindowsHookExW installs a system-wide or thread-specific hook procedure.
// callback must be a syscall.NewCallback-wrapped function matching the
// expected Win32 signature for the given idHook. Returns the HHOOK or 0
// on failure, in which case err contains the syscall errno.
func SetWindowsHookExW(idHook int, callback uintptr, hInst uintptr, threadID uint32) (uintptr, error) {
	hook, _, err := procSetWindowsHookExW.Call(uintptr(idHook), callback, hInst, uintptr(threadID))
	if hook == 0 {
		return 0, err
	}
	return hook, nil
}

// UnhookWindowsHookEx removes a hook installed by SetWindowsHookExW.
func UnhookWindowsHookEx(hook uintptr) {
	procUnhookWindowsHookEx.Call(hook)
}

// CallNextHookEx must be called from inside a hook procedure to pass the
// event to the next hook in the chain.
func CallNextHookEx(nCode int32, wParam, lParam uintptr) uintptr {
	r, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return r
}

// GetKeyState returns the state of the given virtual key. The high bit of
// the return value is set while the key is held.
func GetKeyState(vk uint32) int16 {
	r, _, _ := procGetKeyState.Call(uintptr(vk))
	return int16(r)
}

// IsKeyDown is the common use of GetKeyState — true while vk is held.
func IsKeyDown(vk uint32) bool { return GetKeyState(vk) < 0 }

// NewCallback wraps syscall.NewCallback for callers that already import
// this package and want to avoid a separate syscall import.
func NewCallback(fn any) uintptr { return syscall.NewCallback(fn) }

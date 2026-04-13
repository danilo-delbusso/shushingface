//go:build windows

package win32

import "unsafe"

var (
	procGetForegroundWindow       = user32.NewProc("GetForegroundWindow")
	procGetWindowRect             = user32.NewProc("GetWindowRect")
	procGetWindowThreadProcessID  = user32.NewProc("GetWindowThreadProcessId")
	procQueryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")
	procGetCurrentThreadID        = kernel32.NewProc("GetCurrentThreadId")
	procGetModuleHandleW          = kernel32.NewProc("GetModuleHandleW")
)

// GetForegroundWindow returns the HWND of the window with keyboard focus,
// or 0 if no window is active.
func GetForegroundWindow() uintptr {
	hwnd, _, _ := procGetForegroundWindow.Call()
	return hwnd
}

// GetWindowRect fills rc with the screen-relative bounds of the given window.
func GetWindowRect(hwnd uintptr, rc *RECT) {
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(rc)))
}

// GetWindowThreadProcessID returns the process ID that owns hwnd.
func GetWindowThreadProcessID(hwnd uintptr) uint32 {
	var pid uint32
	procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	return pid
}

// QueryFullProcessImageName retrieves the full path of the executable
// image for the specified process handle. size is the buffer length in
// WCHARs; on return the actual number of characters written is returned.
func QueryFullProcessImageName(hProcess uintptr, flags uint32, buf *uint16, size *uint32) bool {
	r, _, _ := procQueryFullProcessImageName.Call(
		hProcess, uintptr(flags),
		uintptr(unsafe.Pointer(buf)),
		uintptr(unsafe.Pointer(size)),
	)
	return r != 0
}

// GetCurrentThreadID returns the current OS thread ID.
func GetCurrentThreadID() uint32 {
	tid, _, _ := procGetCurrentThreadID.Call()
	return uint32(tid)
}

// GetModuleHandleW returns the module handle for the process's own binary
// when called with nil.
func GetModuleHandleW() uintptr {
	h, _, _ := procGetModuleHandleW.Call(0)
	return h
}

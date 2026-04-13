//go:build windows

package win32

import "unsafe"

var (
	procGetMessageW       = user32.NewProc("GetMessageW")
	procTranslateMessage  = user32.NewProc("TranslateMessage")
	procDispatchMessageW  = user32.NewProc("DispatchMessageW")
	procPostThreadMessage = user32.NewProc("PostThreadMessageW")
	procPostMessageW      = user32.NewProc("PostMessageW")
)

// Application-defined thread-message codes picked out of WM_APP range.
// Callers can add their own above WMApp; these are the ones we currently use.
const (
	WMApp  = 0x8000
	WMQuit = 0x0012
)

// GetMessageW blocks until a message is retrieved from the calling thread's
// queue. Returns the BOOL-like result (0 on WM_QUIT, -1 on error, positive
// otherwise) and fills msg.
func GetMessageW(msg *MSG) int32 {
	r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(msg)), 0, 0, 0)
	return int32(r)
}

// TranslateMessage processes virtual-key → char conversion for a retrieved message.
func TranslateMessage(msg *MSG) {
	procTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
}

// DispatchMessageW dispatches the message to its window procedure.
func DispatchMessageW(msg *MSG) {
	procDispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
}

// PostThreadMessageW posts a message to the given OS thread's message queue.
// Commonly used to wake a thread blocked in GetMessageW.
func PostThreadMessageW(threadID uint32, code uint32, wParam, lParam uintptr) {
	procPostThreadMessage.Call(uintptr(threadID), uintptr(code), wParam, lParam)
}

// PostMessageW posts a message to the queue of the thread that owns hwnd.
func PostMessageW(hwnd uintptr, code uint32, wParam, lParam uintptr) {
	procPostMessageW.Call(hwnd, uintptr(code), wParam, lParam)
}

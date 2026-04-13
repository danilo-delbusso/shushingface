//go:build windows

package win32

import "unsafe"

var procSendInput = user32.NewProc("SendInput")

// INPUT type and SendInput flag constants.
const (
	InputKeyboard = 1

	KeyEventFKeyup   = 0x0002
	KeyEventFUnicode = 0x0004

	VKControl = 0x11
	VKV       = 0x56
)

// SendInput dispatches a batch of keyboard INPUT events. Returns the number
// of events successfully injected and the syscall errno for the shortfall.
func SendInput(inputs []KEYBDINPUT) (uint32, error) {
	if len(inputs) == 0 {
		return 0, nil
	}
	n, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
	if n != uintptr(len(inputs)) {
		return uint32(n), err
	}
	return uint32(n), nil
}

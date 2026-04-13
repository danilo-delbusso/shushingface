//go:build windows

package win32

// RECT mirrors the Win32 RECT structure (LONG fields are 32-bit signed).
type RECT struct {
	Left, Top, Right, Bottom int32
}

// POINT mirrors the Win32 POINT structure.
type POINT struct {
	X, Y int32
}

// MSG mirrors the Win32 MSG structure. Size and field layout match the
// 64-bit ABI (with padding after Message and Time).
type MSG struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       POINT
	LPrivate uint32
}

// KBDLLHOOKSTRUCT is passed to WH_KEYBOARD_LL hook callbacks via lParam.
type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

// KEYBDINPUT is the keyboard variant of the INPUT union passed to SendInput.
// Type field must be InputKeyboard. The trailing padding is the size delta
// between the keyboard variant and the full INPUT union on 64-bit Windows.
type KEYBDINPUT struct {
	Type    uint32
	_pad    uint32
	WVk     uint16
	WScan   uint16
	DwFlags uint32
	Time    uint32
	Extra   uintptr
	_pad2   [8]byte
}

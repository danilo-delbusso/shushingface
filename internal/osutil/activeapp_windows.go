//go:build windows

package osutil

import (
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"

	"codeberg.org/dbus/shushingface/internal/platform"
	"codeberg.org/dbus/shushingface/internal/win32"
)

func activeAppCapability() platform.Capability { return platform.Supported() }

// GetActiveWindowName returns the base name of the executable that owns the
// foreground window, or an empty string on failure.
func GetActiveWindowName() string {
	hwnd := win32.GetForegroundWindow()
	if hwnd == 0 {
		return ""
	}
	pid := win32.GetWindowThreadProcessID(hwnd)
	if pid == 0 {
		return ""
	}

	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)

	var buf [windows.MAX_PATH]uint16
	size := uint32(len(buf))
	if !win32.QueryFullProcessImageName(uintptr(h), 0, &buf[0], &size) {
		return ""
	}
	return filepath.Base(syscall.UTF16ToString(buf[:size]))
}

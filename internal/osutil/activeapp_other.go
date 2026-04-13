//go:build !linux && !windows

package osutil

// GetActiveWindowName returns the name of the currently focused application.
// TODO: implement via NSWorkspace on macOS.
func GetActiveWindowName() string {
	return ""
}

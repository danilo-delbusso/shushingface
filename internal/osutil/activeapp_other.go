//go:build !linux

package osutil

// GetActiveWindowName returns the name of the currently focused application.
// TODO: implement via NSWorkspace (macOS) or GetForegroundWindow (Windows).
func GetActiveWindowName() string {
	return ""
}

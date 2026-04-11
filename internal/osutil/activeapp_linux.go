//go:build linux

package osutil

// GetActiveWindowName returns the name of the currently focused application.
// On Wayland this is not available without compositor-specific APIs.
func GetActiveWindowName() string {
	return ""
}

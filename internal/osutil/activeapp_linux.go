//go:build linux

package osutil

import "codeberg.org/dbus/shushingface/internal/platform"

func activeAppCapability() platform.Capability {
	return platform.Unsupported("Active-window detection on Linux requires compositor-specific APIs we don't implement yet.")
}

// GetActiveWindowName returns "" on Linux — both Wayland and X11 would
// need protocol-specific integrations we haven't done.
func GetActiveWindowName() string { return "" }

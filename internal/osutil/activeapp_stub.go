//go:build !linux && !windows

package osutil

import "codeberg.org/dbus/shushingface/internal/platform"

func activeAppCapability() platform.Capability {
	return platform.Unsupported("Active-window detection is not implemented on this platform.")
}

// GetActiveWindowName returns "" on unsupported platforms.
func GetActiveWindowName() string { return "" }

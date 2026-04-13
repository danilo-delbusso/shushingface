//go:build !linux

package indicator

import "codeberg.org/dbus/shushingface/internal/platform"

func capability() platform.Capability {
	return platform.Unsupported("A system-tray indicator is not implemented on this platform yet.")
}

// Start is a no-op on unsupported platforms.
func Start(onActivate func()) {}

// SetRecording is a no-op on unsupported platforms.
func SetRecording(recording bool) {}

// Stop is a no-op on unsupported platforms.
func Stop() {}

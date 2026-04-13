//go:build !linux && !windows

package paste

import "codeberg.org/dbus/shushingface/internal/platform"

func capability() platform.Capability {
	return platform.Unsupported("Auto-paste is not implemented on this platform yet.")
}

// Type is a no-op on unsupported platforms.
func Type(text string) error { return nil }

//go:build !windows

package overlay

import "codeberg.org/dbus/shushingface/internal/platform"

func capability() platform.Capability {
	return platform.Unsupported("The recording overlay is only available on Windows today.")
}

// New returns a no-op overlay on platforms that don't yet have an implementation.
func New() Overlay { return stub{} }

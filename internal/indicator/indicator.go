// Package indicator manages a system-tray / panel icon that reflects
// recording state. Only Linux (StatusNotifierItem via D-Bus) is supported
// today; other platforms get a stub that reports Unsupported.
package indicator

import "codeberg.org/dbus/shushingface/internal/platform"

// Capability reports whether the panel indicator is available on this OS.
func Capability() platform.Capability { return capability() }

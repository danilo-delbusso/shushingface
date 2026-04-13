// Package osutil exposes small host-OS utilities that don't fit elsewhere
// — currently just active-window introspection used to tag transcripts
// with the app the user was speaking into.
package osutil

import "codeberg.org/dbus/shushingface/internal/platform"

// ActiveAppCapability reports whether GetActiveWindowName returns a real
// value on this platform. Linux under Wayland, for example, has no
// compositor-agnostic API for this and falls back to "".
func ActiveAppCapability() platform.Capability { return activeAppCapability() }

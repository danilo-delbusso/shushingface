// Package paste simulates typing text into the currently focused
// application by placing it on the clipboard and sending the platform
// paste shortcut. Each OS has its own implementation file; unsupported
// platforms report Unsupported and Type becomes a no-op.
package paste

import "codeberg.org/dbus/shushingface/internal/platform"

// Capability reports whether auto-paste works on this platform right now.
// On Linux this depends on whether the required helper (wtype / xdotool)
// is on PATH — so the Reason field may guide the user to install one.
func Capability() platform.Capability { return capability() }

// Available is a thin bool view of Capability() kept for callers that
// only care whether paste is currently wired up.
func Available() bool { return Capability().Supported }

// InstallHint returns the Capability reason when unsupported, or "" when
// paste is available. Convenience for UI banners.
func InstallHint() string {
	if c := Capability(); !c.Supported {
		return c.Reason
	}
	return ""
}

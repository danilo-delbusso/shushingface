//go:build !linux

package indicator

// Start is a no-op on non-Linux platforms.
// TODO: implement via NSStatusItem (macOS) or NotifyIcon (Windows).
func Start(onActivate func()) {}

// SetRecording is a no-op on non-Linux platforms.
func SetRecording(recording bool) {}

// Stop is a no-op on non-Linux platforms.
func Stop() {}

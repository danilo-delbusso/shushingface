// Package notify shows desktop notifications. Each supported OS has its
// own implementation file (notify_linux.go, notify_windows.go). Platforms
// without an implementation fall through to notify_stub.go which reports
// Unsupported and makes every call a no-op.
package notify

import "codeberg.org/dbus/shushingface/internal/platform"

// Capability reports whether desktop notifications work on this platform.
// Wired into App.GetCapabilities() so the UI can grey out the toggle when
// Supported is false.
func Capability() platform.Capability { return capability() }

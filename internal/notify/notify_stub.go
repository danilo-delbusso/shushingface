//go:build !linux && !windows

package notify

import "codeberg.org/dbus/shushingface/internal/platform"

func capability() platform.Capability {
	return platform.Unsupported("Desktop notifications are not implemented on this platform yet.")
}

// RecordingStarted is a no-op on unsupported platforms.
func RecordingStarted() {}

// RecordingProcessing is a no-op on unsupported platforms.
func RecordingProcessing() {}

// Error is a no-op on unsupported platforms.
func Error(title, body string) {}

// RecordingDone is a no-op on unsupported platforms.
func RecordingDone() {}

//go:build !linux

package notify

// RecordingStarted is a no-op on non-Linux platforms.
// TODO: implement via NSUserNotificationCenter (macOS) or toast notifications (Windows).
func RecordingStarted() {}

// RecordingProcessing is a no-op on non-Linux platforms.
func RecordingProcessing() {}

// Error is a no-op on non-Linux platforms.
func Error(title, body string) {}

// RecordingDone is a no-op on non-Linux platforms.
func RecordingDone() {}

//go:build !linux

package notify

// RecordingStarted is a no-op on non-Linux platforms.
// TODO: implement via NSUserNotificationCenter (macOS) or toast notifications (Windows).
func RecordingStarted() {}

// RecordingProcessing is a no-op on non-Linux platforms.
func RecordingProcessing() {}

// RecordingDone is a no-op on non-Linux platforms.
func RecordingDone() {}

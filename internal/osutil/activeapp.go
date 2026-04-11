package osutil

// GetActiveWindowName returns the name of the currently focused window or application.
// This is used to provide context for the transcription history when dictation is triggered.
func GetActiveWindowName() string {
	// TODO: Implement platform-specific syscalls
	// (e.g., Windows API, macOS AppleScript/Objective-C, Linux X11/Wayland)
	return "Unknown Application"
}

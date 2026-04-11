//go:build !linux

package paste

// Available reports whether auto-paste is supported on this platform.
func Available() bool { return false }

// InstallHint returns an empty string on unsupported platforms.
func InstallHint() string { return "" }

// Type is not yet implemented on this platform.
func Type(text string) error { return nil }

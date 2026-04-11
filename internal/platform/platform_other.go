//go:build !linux

package platform

import "runtime"

// Detect returns platform info for non-Linux systems.
func Detect() Info {
	return Info{
		OS:             runtime.GOOS,
		DisplayServer:  Native,
		Desktop:        "",
		PackageManager: UnknownPM,
	}
}

// InstallCmd is a no-op on non-Linux.
func InstallCmd(pkg string) string { return "" }

// HasCommand checks if a CLI tool is available on PATH.
func HasCommand(name string) bool { return false }

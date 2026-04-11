//go:build !linux

package osutil

// InstallSafeX11ErrorHandler is a no-op on non-Linux platforms.
func InstallSafeX11ErrorHandler() {}

//go:build windows

package ipc

import "fmt"

// IsRunning is not yet implemented on Windows.
func IsRunning() bool { return false }

// Send is not yet implemented on Windows.
func Send(command string) error {
	return fmt.Errorf("IPC not yet supported on Windows")
}

// SendToggle is not yet implemented on Windows.
func SendToggle() error { return Send("TOGGLE") }

// SendShow is not yet implemented on Windows.
func SendShow() error { return Send("SHOW") }

// Listen is not yet implemented on Windows.
func Listen(handler func(command string)) (func(), error) {
	return func() {}, nil
}

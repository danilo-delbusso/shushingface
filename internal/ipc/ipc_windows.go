//go:build windows

package ipc

import "fmt"

// SendToggle is not yet implemented on Windows.
// TODO: implement via named pipes (\\.\pipe\sussurro)
func SendToggle() error {
	return fmt.Errorf("IPC not yet supported on Windows")
}

// Listen is not yet implemented on Windows.
func Listen(onToggle func()) (func(), error) {
	return func() {}, nil
}

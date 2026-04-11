//go:build !windows

package ipc

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
)

func socketPath() string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "shushingface.sock")
}

// IsRunning checks if another instance is listening on the socket.
func IsRunning() bool {
	conn, err := net.Dial("unix", socketPath())
	if err != nil {
		return false
	}
	conn.Write([]byte("PING"))
	conn.Close()
	return true
}

// Send sends a command to the running instance.
func Send(command string) error {
	conn, err := net.Dial("unix", socketPath())
	if err != nil {
		return fmt.Errorf("shushingface is not running: %w", err)
	}
	defer conn.Close()
	_, err = conn.Write([]byte(command))
	return err
}

// SendToggle sends a toggle-recording signal.
func SendToggle() error { return Send("TOGGLE") }

// SendShow tells the running instance to show its window.
func SendShow() error { return Send("SHOW") }

// Listen starts a Unix socket server that dispatches commands.
// Returns a cleanup function to close the listener.
func Listen(handler func(command string)) (func(), error) {
	path := socketPath()
	os.Remove(path)

	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", path, err)
	}

	go func() {
		buf := make([]byte, 32)
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			n, _ := conn.Read(buf)
			cmd := string(buf[:n])
			slog.Info("received IPC command", "command", cmd)
			handler(cmd)
			conn.Close()
		}
	}()

	cleanup := func() {
		ln.Close()
		os.Remove(path)
	}
	return cleanup, nil
}

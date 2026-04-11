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
	return filepath.Join(dir, "sussurro.sock")
}

// SendToggle connects to the running sussurro instance and sends a toggle signal.
func SendToggle() error {
	conn, err := net.Dial("unix", socketPath())
	if err != nil {
		return fmt.Errorf("sussurro is not running: %w", err)
	}
	defer conn.Close()
	_, err = conn.Write([]byte("TOGGLE"))
	return err
}

// Listen starts a Unix socket server that calls onToggle whenever a toggle
// signal is received. Returns a cleanup function to close the listener.
func Listen(onToggle func()) (func(), error) {
	path := socketPath()
	os.Remove(path)

	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", path, err)
	}

	go func() {
		buf := make([]byte, 16)
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			n, _ := conn.Read(buf)
			if string(buf[:n]) == "TOGGLE" {
				slog.Info("received IPC toggle signal")
				onToggle()
			}
			conn.Close()
		}
	}()

	cleanup := func() {
		ln.Close()
		os.Remove(path)
	}
	return cleanup, nil
}

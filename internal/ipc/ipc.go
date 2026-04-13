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
	defer closeWarn(conn, "ping connection")
	if _, err := conn.Write([]byte("PING")); err != nil {
		return false
	}
	return true
}

// Send sends a command to the running instance.
func Send(command string) error {
	conn, err := net.Dial("unix", socketPath())
	if err != nil {
		return fmt.Errorf("shushingface is not running: %w", err)
	}
	defer closeWarn(conn, "send connection")
	_, err = conn.Write([]byte(command))
	return err
}

// closeWarn closes c and logs a warning on failure. Use in defer for io.Closer.
func closeWarn(c interface{ Close() error }, what string) {
	if err := c.Close(); err != nil {
		slog.Warn("close failed", "what", what, "error", err)
	}
}

// SendToggle sends a toggle-recording signal.
func SendToggle() error { return Send("TOGGLE") }

// SendShow tells the running instance to show its window.
func SendShow() error { return Send("SHOW") }

// Listen starts a Unix socket server that dispatches commands.
// Returns a cleanup function to close the listener.
func Listen(handler func(command string)) (func(), error) {
	path := socketPath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		slog.Warn("failed to remove stale socket", "path", path, "error", err)
	}

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
			n, err := conn.Read(buf)
			if err != nil {
				slog.Warn("IPC read failed", "error", err)
				closeWarn(conn, "accepted connection")
				continue
			}
			cmd := string(buf[:n])
			slog.Info("received IPC command", "command", cmd)
			handler(cmd)
			closeWarn(conn, "accepted connection")
		}
	}()

	cleanup := func() {
		closeWarn(ln, "listener")
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			slog.Warn("failed to remove socket", "path", path, "error", err)
		}
	}
	return cleanup, nil
}

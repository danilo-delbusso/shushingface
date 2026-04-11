//go:build linux

package paste

import (
	"log/slog"
	"os/exec"
	"time"

	"codeberg.org/dbus/shushingface/internal/platform"
)

// Available reports whether the required typing tool is installed.
func Available() bool {
	return platform.HasCommand(toolName())
}

// InstallHint returns the install command for the missing tool.
func InstallHint() string {
	if Available() {
		return ""
	}
	return platform.InstallCmd(toolName())
}

func toolName() string {
	if platform.Detect().DisplayServer == platform.Wayland {
		return "wtype"
	}
	return "xdotool"
}

// Type simulates typing text into the currently focused application.
func Type(text string) error {
	time.Sleep(100 * time.Millisecond)

	if platform.Detect().DisplayServer == platform.Wayland {
		return waylandType(text)
	}
	return x11Type(text)
}

func waylandType(text string) error {
	cmd := exec.Command("wtype", "--", text)
	if err := cmd.Run(); err != nil {
		slog.Warn("wtype failed", "error", err)
		return err
	}
	return nil
}

func x11Type(text string) error {
	cmd := exec.Command("xdotool", "type", "--clearmodifiers", "--", text)
	if err := cmd.Run(); err != nil {
		slog.Warn("xdotool failed", "error", err)
		return err
	}
	return nil
}

//go:build linux

package paste

import (
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"codeberg.org/dbus/shushingface/internal/platform"
)

func capability() platform.Capability {
	tool := toolName()
	if platform.HasCommand(tool) {
		return platform.Supported()
	}
	return platform.Unsupported(fmt.Sprintf(
		"%q is required for auto-paste on %s. Install with: %s",
		tool, platform.Detect().DisplayServer, platform.InstallCmd(tool),
	))
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

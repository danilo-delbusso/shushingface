//go:build linux

package paste

import (
	"log/slog"
	"os"
	"os/exec"
	"time"
)

// Type simulates typing the text into the currently focused application.
// On Wayland, uses wtype to inject keystrokes directly (no clipboard pollution).
// On X11, uses xdotool.
func Type(text string) error {
	// Small delay to let the user's focus settle after the app processes
	time.Sleep(100 * time.Millisecond)

	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return waylandType(text)
	}
	return x11Type(text)
}

func waylandType(text string) error {
	cmd := exec.Command("wtype", "--", text)
	if err := cmd.Run(); err != nil {
		slog.Warn("wtype not available — install wtype to enable auto-paste", "error", err)
		return err
	}
	return nil
}

func x11Type(text string) error {
	cmd := exec.Command("xdotool", "type", "--clearmodifiers", "--", text)
	if err := cmd.Run(); err != nil {
		slog.Warn("xdotool not available — install xdotool to enable auto-paste", "error", err)
		return err
	}
	return nil
}

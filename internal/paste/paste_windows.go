//go:build windows

package paste

import (
	"fmt"
	"log/slog"
	"time"

	"codeberg.org/dbus/shushingface/internal/platform"
	"codeberg.org/dbus/shushingface/internal/win32"
)

func capability() platform.Capability { return platform.Supported() }

// Type places text on the clipboard and simulates Ctrl+V.
func Type(text string) error {
	if err := win32.SetClipboardUnicodeText(text); err != nil {
		return fmt.Errorf("set clipboard: %w", err)
	}
	time.Sleep(50 * time.Millisecond)
	if err := sendCtrlV(); err != nil {
		slog.Warn("sendinput failed", "error", err)
		return fmt.Errorf("send ctrl+v: %w", err)
	}
	return nil
}

func sendCtrlV() error {
	inputs := []win32.KEYBDINPUT{
		{Type: win32.InputKeyboard, WVk: win32.VKControl},
		{Type: win32.InputKeyboard, WVk: win32.VKV},
		{Type: win32.InputKeyboard, WVk: win32.VKV, DwFlags: win32.KeyEventFKeyup},
		{Type: win32.InputKeyboard, WVk: win32.VKControl, DwFlags: win32.KeyEventFKeyup},
	}
	_, err := win32.SendInput(inputs)
	return err
}

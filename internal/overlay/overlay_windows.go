//go:build windows

package overlay

import (
	"log/slog"

	"codeberg.org/dbus/shushingface/internal/platform"
	winoverlay "codeberg.org/dbus/shushingface/internal/win32/overlay"
)

func capability() platform.Capability { return platform.Supported() }

// New returns the Win32 overlay or a no-op stub if the underlying window
// fails to come up. All real work lives in internal/win32/overlay.
func New() Overlay {
	w, err := winoverlay.New()
	if err != nil {
		slog.Warn("overlay init failed", "error", err)
		return stub{}
	}
	return &winAdapter{w: w}
}

// winAdapter bridges the concrete *winoverlay.Window to the Overlay
// interface — converts the Mode enum, swallows the unused Show args.
type winAdapter struct {
	w *winoverlay.Window
}

func (a *winAdapter) Show(_ string, _ float64) error { return a.w.Show() }
func (a *winAdapter) Hide() error                    { return a.w.Hide() }
func (a *winAdapter) Close() error                   { return a.w.Close() }
func (a *winAdapter) SetLevel(level float32)         { a.w.SetLevel(level) }
func (a *winAdapter) SetMode(mode Mode)              { a.w.SetProcessing(mode == ModeProcessing) }

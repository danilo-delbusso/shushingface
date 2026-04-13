// Package overlay shows a small translucent "Recording" indicator pinned
// above the focused window. Each supported OS has its own implementation
// file; unsupported platforms get a stub Overlay whose methods are no-ops.
package overlay

import "codeberg.org/dbus/shushingface/internal/platform"

// Overlay shows a small floating recording indicator above the active window.
type Overlay interface {
	Show(text string, opacity float64) error
	Hide() error
	Close() error
	// SetLevel feeds the overlay a fresh microphone amplitude in [0,1] so
	// the indicator can animate in response to incoming audio. Cheap and
	// non-blocking; safe to call from any goroutine.
	SetLevel(level float32)
}

// Capability reports whether the overlay is drawable on this platform.
func Capability() platform.Capability { return capability() }

// stub is a no-op Overlay. Used on unsupported platforms and as a
// fallback when the real implementation fails to initialise.
type stub struct{}

func (stub) Show(string, float64) error { return nil }
func (stub) Hide() error                { return nil }
func (stub) Close() error               { return nil }
func (stub) SetLevel(float32)           {}

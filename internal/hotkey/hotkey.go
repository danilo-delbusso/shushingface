package hotkey

import (
	"errors"
	"fmt"
	"strings"
)

// ErrUnsupported is returned when the platform can't register global hotkeys.
var ErrUnsupported = errors.New("global hotkeys not supported on this platform")

// ErrConflict is returned when another process already owns the hotkey.
var ErrConflict = errors.New("hotkey is already registered by another application")

// ErrInvalidSpec is returned for malformed shortcut strings.
var ErrInvalidSpec = errors.New("invalid shortcut")

// Modifier is a bit flag describing which modifier keys are required.
type Modifier uint32

const (
	ModCtrl  Modifier = 1 << 0
	ModShift Modifier = 1 << 1
	ModAlt   Modifier = 1 << 2
	ModSuper Modifier = 1 << 3 // Win / Cmd
)

// Spec identifies a normalized hotkey.
type Spec struct {
	Mods Modifier
	Key  string // canonical key name, e.g. "B", "F5", "Space"
}

// Capabilities describes what the platform hotkey backend supports.
type Capabilities struct {
	Supported     bool   `json:"supported"`
	ConflictCheck bool   `json:"conflictCheck"`
	Reason        string `json:"reason,omitempty"`
}

// Manager registers global hotkeys and forwards events.
type Manager interface {
	Register(id string, spec Spec) error
	Unregister(id string) error
	Events() <-chan string
	Close() error
}

// ParseSpec parses a "Ctrl+Shift+B"-style string into a Spec.
func ParseSpec(s string) (Spec, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Spec{}, ErrInvalidSpec
	}
	parts := strings.Split(s, "+")
	var spec Spec
	for i, raw := range parts {
		p := strings.ToLower(strings.TrimSpace(raw))
		isLast := i == len(parts)-1
		switch p {
		case "ctrl", "control":
			spec.Mods |= ModCtrl
		case "shift":
			spec.Mods |= ModShift
		case "alt", "option":
			spec.Mods |= ModAlt
		case "super", "win", "cmd", "meta":
			spec.Mods |= ModSuper
		default:
			if !isLast {
				return Spec{}, fmt.Errorf("%w: unknown modifier %q", ErrInvalidSpec, raw)
			}
			if strings.TrimSpace(raw) == "" {
				return Spec{}, fmt.Errorf("%w: missing key", ErrInvalidSpec)
			}
			spec.Key = canonKey(raw)
		}
	}
	if spec.Key == "" {
		return Spec{}, fmt.Errorf("%w: missing key", ErrInvalidSpec)
	}
	if spec.Mods == 0 {
		return Spec{}, fmt.Errorf("%w: at least one modifier required", ErrInvalidSpec)
	}
	return spec, nil
}

// FormatSpec renders a Spec back into a "Ctrl+Shift+B"-style string.
func FormatSpec(spec Spec) string {
	var parts []string
	if spec.Mods&ModCtrl != 0 {
		parts = append(parts, "Ctrl")
	}
	if spec.Mods&ModAlt != 0 {
		parts = append(parts, "Alt")
	}
	if spec.Mods&ModShift != 0 {
		parts = append(parts, "Shift")
	}
	if spec.Mods&ModSuper != 0 {
		parts = append(parts, "Super")
	}
	if spec.Key != "" {
		parts = append(parts, spec.Key)
	}
	return strings.Join(parts, "+")
}

func canonKey(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) == 1 {
		return strings.ToUpper(raw)
	}
	// Title-case multi-letter names (F5, Space, Enter, ArrowLeft...).
	return strings.ToUpper(raw[:1]) + strings.ToLower(raw[1:])
}

//go:build windows

package hotkey

import (
	"errors"
	"fmt"

	"codeberg.org/dbus/shushingface/internal/platform"
	winhotkey "codeberg.org/dbus/shushingface/internal/win32/hotkey"
)

func capability() platform.Capability { return platform.Supported() }

// New returns a Manager backed by the Win32 implementation in
// internal/win32/hotkey.
func New() Manager {
	return &winAdapter{
		w:      winhotkey.New(),
		events: make(chan Event, 16),
		stop:   make(chan struct{}),
	}
}

// winAdapter bridges the concrete *winhotkey.Manager to the Manager
// interface — translates spec / mode / event types and pumps events.
type winAdapter struct {
	w        *winhotkey.Manager
	events   chan Event
	stop     chan struct{}
	pumpOnce bool
}

func (a *winAdapter) Register(name string, spec Spec, mode Mode) error {
	vk, ok := winhotkey.VKFromKey(spec.Key)
	if !ok {
		return fmt.Errorf("%w: unsupported key %q", ErrInvalidSpec, spec.Key)
	}
	mods := winMods(spec.Mods)
	if mods == 0 {
		return fmt.Errorf("%w: at least one modifier required", ErrInvalidSpec)
	}
	if err := a.w.Register(name, winhotkey.Spec{Mods: mods, VK: vk}, winMode(mode)); err != nil {
		if errors.Is(err, winhotkey.ErrConflict) {
			return ErrConflict
		}
		return err
	}
	a.ensurePump()
	return nil
}

func (a *winAdapter) Unregister(name string) error { return a.w.Unregister(name) }
func (a *winAdapter) Events() <-chan Event         { return a.events }

func (a *winAdapter) Close() error {
	close(a.stop)
	return a.w.Close()
}

// ensurePump spawns a single goroutine that translates winhotkey events
// to hotkey events. Started lazily on the first registration so empty
// managers do not leak a goroutine.
func (a *winAdapter) ensurePump() {
	if a.pumpOnce {
		return
	}
	a.pumpOnce = true
	src := a.w.Events()
	go func() {
		for {
			select {
			case <-a.stop:
				return
			case ev, ok := <-src:
				if !ok {
					return
				}
				a.events <- Event{Name: ev.Name, Type: eventType(ev.Type)}
			}
		}
	}()
}

func winMods(m Modifier) uint32 {
	var out uint32
	if m&ModCtrl != 0 {
		out |= winhotkey.ModCtrl
	}
	if m&ModAlt != 0 {
		out |= winhotkey.ModAlt
	}
	if m&ModShift != 0 {
		out |= winhotkey.ModShift
	}
	if m&ModSuper != 0 {
		out |= winhotkey.ModWin
	}
	return out
}

func winMode(m Mode) winhotkey.Mode {
	if m == ModePushToTalk {
		return winhotkey.ModePushToTalk
	}
	return winhotkey.ModeToggle
}

func eventType(t winhotkey.EventType) EventType {
	switch t {
	case winhotkey.Press:
		return Press
	case winhotkey.Release:
		return Release
	default:
		return Trigger
	}
}

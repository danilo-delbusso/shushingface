//go:build !windows

package hotkey

import "codeberg.org/dbus/shushingface/internal/platform"

func capability() platform.Capability {
	return platform.Unsupported("Global hotkey registration is only available on Windows today. Bind a shortcut from your desktop settings instead.")
}

// New returns a manager that refuses registration on unsupported platforms.
// Callers should check Capability() before presenting a recorder UI.
func New() Manager {
	return &stub{events: make(chan Event)}
}

type stub struct {
	events chan Event
}

func (s *stub) Register(string, Spec, Mode) error { return ErrUnsupported }
func (s *stub) Unregister(string) error           { return nil }
func (s *stub) Events() <-chan Event              { return s.events }
func (s *stub) Close() error                      { close(s.events); return nil }

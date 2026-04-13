//go:build !windows

package hotkey

// New returns a manager that refuses registration on unsupported platforms.
// Callers should check Detect() before presenting a recorder UI.
func New() Manager {
	return &stub{events: make(chan string)}
}

// Detect returns platform capabilities for hotkey registration.
func Detect() Capabilities {
	return Capabilities{
		Supported: false,
		Reason:    "global hotkey registration not available; bind from your desktop settings",
	}
}

type stub struct {
	events chan string
}

func (s *stub) Register(string, Spec) error { return ErrUnsupported }
func (s *stub) Unregister(string) error     { return nil }
func (s *stub) Events() <-chan string       { return s.events }
func (s *stub) Close() error                { close(s.events); return nil }

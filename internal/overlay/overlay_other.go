//go:build !windows

package overlay

// New returns a no-op overlay on platforms that don't yet have an implementation.
func New() Overlay { return stub{} }

type stub struct{}

func (stub) Show(string, float64) error { return nil }
func (stub) Hide() error                { return nil }
func (stub) Close() error               { return nil }

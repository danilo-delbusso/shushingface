package overlay

// Overlay shows a small floating recording indicator above the active window.
type Overlay interface {
	Show(text string, opacity float64) error
	Hide() error
	Close() error
}

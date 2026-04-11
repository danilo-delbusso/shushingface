//go:build !linux || !systray

package desktop

// TrayManager is a no-op on non-Linux platforms.
type TrayManager struct{}

// NewTrayManager returns a no-op tray manager.
func NewTrayManager(_ *App) *TrayManager {
	return &TrayManager{}
}

// Run is a no-op on non-Linux platforms.
func (t *TrayManager) Run() {}

func shutdownTray() {}

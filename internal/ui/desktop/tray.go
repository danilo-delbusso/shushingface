//go:build linux && systray

package desktop

import (
	_ "embed"

	"github.com/getlantern/systray"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed icon.png
var trayIcon []byte

// TrayManager handles the system tray lifecycle.
type TrayManager struct {
	app *App
}

// NewTrayManager creates a new tray manager tied to the Wails app.
func NewTrayManager(app *App) *TrayManager {
	return &TrayManager{app: app}
}

// Run starts the system tray. This blocks, so call it in a goroutine.
func (t *TrayManager) Run() {
	systray.Run(t.onReady, t.onExit)
}

func (t *TrayManager) onReady() {
	systray.SetIcon(trayIcon)
	systray.SetTitle("Sussurro")
	systray.SetTooltip("Sussurro - Speech Transcription")

	mShow := systray.AddMenuItem("Show Window", "Show the main window")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit Sussurro completely")

	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				if t.app.ctx != nil {
					wailsRuntime.WindowShow(t.app.ctx)
				}
			case <-mQuit.ClickedCh:
				if t.app.ctx != nil {
					wailsRuntime.Quit(t.app.ctx)
				}
				systray.Quit()
				return
			}
		}
	}()
}

func (t *TrayManager) onExit() {}

func shutdownTray() {
	systray.Quit()
}

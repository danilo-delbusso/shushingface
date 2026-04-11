package desktop

import (
	"context"
	"fmt"
	"log/slog"

	"codeberg.org/dbus/sussurro/internal/ai/factory"
	"codeberg.org/dbus/sussurro/internal/config"
	"codeberg.org/dbus/sussurro/internal/core"
	"codeberg.org/dbus/sussurro/internal/history"
	"codeberg.org/dbus/sussurro/internal/osutil"
)

// App struct is the Wails application bridge.
type App struct {
	ctx     context.Context
	engine  *core.Engine
	cfg     *config.Settings
	history *history.Manager
}

// NewApp creates a new desktop application controller with injected dependencies.
func NewApp(engine *core.Engine, cfg *config.Settings, hist *history.Manager) *App {
	return &App{
		engine:  engine,
		cfg:     cfg,
		history: hist,
	}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods (like events/dialogs).
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	go NewTrayManager(a).Run()

	// Global hotkey registration is disabled for now.
	// golang.design/x/hotkey uses XGrabKey on Linux which triggers a fatal
	// X11 BadAccess error if the key combo is already grabbed, killing the
	// entire process. Needs a safer registration approach before re-enabling.
	// See: https://github.com/nicxvan/systray/issues/1
}

// ProcessResult is the data transfer object for the frontend.
type ProcessResult struct {
	Transcript string `json:"transcript"`
	Refined    string `json:"refined"`
	Error      string `json:"error,omitempty"`
}

// Shutdown is called when the app is closing.
func (a *App) Shutdown(_ context.Context) {
	shutdownTray()
}

// StartRecording triggers the engine to start capturing audio.
func (a *App) StartRecording() error {
	return a.engine.StartRecording()
}

// StopAndProcess stops recording, processes the audio, and saves the result.
func (a *App) StopAndProcess() ProcessResult {
	// Capture active window before processing
	activeApp := osutil.GetActiveWindowName()

	transcript, refined, err := a.engine.StopAndProcess(a.ctx)
	if err != nil {
		slog.Error("StopAndProcess failed", "error", err)
		return ProcessResult{Error: err.Error()}
	}

	if a.cfg.EnableHistory && a.history != nil && transcript != "" {
		_, histErr := a.history.Insert(transcript, refined, activeApp)
		if histErr != nil {
			slog.Error("failed to insert history", "error", histErr)
		}
	}

	return ProcessResult{
		Transcript: transcript,
		Refined:    refined,
	}
}

// GetSettings returns the current application settings.
func (a *App) GetSettings() *config.Settings {
	return a.cfg
}

// SaveSettings updates the application settings.
func (a *App) SaveSettings(newSettings config.Settings) error {
	// Hot-reload the AI processor first so we don't persist broken config
	newProcessor, err := factory.NewFromConfig(&newSettings)
	if err != nil {
		return fmt.Errorf("failed to reload AI processor: %w", err)
	}

	if err := config.Save(&newSettings); err != nil {
		return err
	}

	a.engine.SetProcessor(newProcessor)
	*a.cfg = newSettings
	return nil
}

// GetHistory retrieves the transcription history.
func (a *App) GetHistory(limit, offset int) ([]history.Record, error) {
	if a.history == nil {
		return nil, fmt.Errorf("history is disabled")
	}
	return a.history.GetHistory(limit, offset)
}

// ClearHistory wipes all local history data.
func (a *App) ClearHistory() error {
	if a.history == nil {
		return fmt.Errorf("history is disabled")
	}
	return a.history.ClearHistory()
}

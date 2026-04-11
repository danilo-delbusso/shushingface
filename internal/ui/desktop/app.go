package desktop

import (
	"context"
	"fmt"

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
}

// ProcessResult is the data transfer object for the frontend.
type ProcessResult struct {
	Transcript string `json:"transcript"`
	Refined    string `json:"refined"`
	Error      string `json:"error,omitempty"`
}

// StartRecording triggers the engine to start capturing audio.
func (a *App) StartRecording() error {
	return a.engine.StartRecording()
}

// StopAndProcess stops recording, processes the audio, and saves the result.
func (a *App) StopAndProcess() ProcessResult {
	// Capture active window before processing
	activeApp := osutil.GetActiveWindowName()

	transcript, refined, err := a.engine.StopAndProcess(context.Background())
	if err != nil {
		return ProcessResult{Error: err.Error()}
	}

	if a.cfg.EnableHistory && a.history != nil && transcript != "" {
		_, histErr := a.history.Insert(transcript, refined, activeApp)
		if histErr != nil {
			fmt.Printf("Failed to insert history: %v\n", histErr)
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
	err := config.Save(&newSettings)
	if err == nil {
		// Update the local reference
		*a.cfg = newSettings

		// Hot-reload the AI Processor based on new settings
		newProcessor, factoryErr := factory.NewFromConfig(&newSettings)
		if factoryErr != nil {
			return fmt.Errorf("failed to reload AI processor: %v", factoryErr)
		}
		a.engine.SetProcessor(newProcessor)
	}
	return err
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

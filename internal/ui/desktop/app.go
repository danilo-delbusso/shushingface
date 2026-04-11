package desktop

import (
	"context"
	"fmt"
	"log/slog"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"codeberg.org/dbus/sussurro/internal/ai/factory"
	"codeberg.org/dbus/sussurro/internal/config"
	"codeberg.org/dbus/sussurro/internal/core"
	"codeberg.org/dbus/sussurro/internal/history"
	"codeberg.org/dbus/sussurro/internal/indicator"
	"codeberg.org/dbus/sussurro/internal/ipc"
	"codeberg.org/dbus/sussurro/internal/notify"
	"codeberg.org/dbus/sussurro/internal/osutil"
)

// App struct is the Wails application bridge.
type App struct {
	ctx      context.Context
	engine   *core.Engine
	cfg      *config.Settings
	history  *history.Manager
	cleanIPC func()
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
	if a.cfg.EnableIndicator {
		indicator.Start()
	}

	// IPC listener: allows `sussurro --toggle` to trigger recording
	cleanup, err := ipc.Listen(func() {
		wailsRuntime.EventsEmit(a.ctx, "hotkey-toggle")
	})
	if err != nil {
		slog.Warn("failed to start IPC listener", "error", err)
	} else {
		a.cleanIPC = cleanup
	}
}

// ProcessResult is the data transfer object for the frontend.
type ProcessResult struct {
	Transcript string `json:"transcript"`
	Refined    string `json:"refined"`
	Error      string `json:"error,omitempty"`
}

// Shutdown is called when the app is closing.
func (a *App) Shutdown(_ context.Context) {
	if a.cleanIPC != nil {
		a.cleanIPC()
	}
	indicator.Stop()
	shutdownTray()
}

// StartRecording triggers the engine to start capturing audio.
func (a *App) StartRecording() error {
	err := a.engine.StartRecording()
	if err == nil {
		if a.cfg.EnableNotifications {
			notify.RecordingStarted()
		}
		indicator.SetRecording(true)
	}
	return err
}

// StopAndProcess stops recording, processes the audio, and saves the result.
func (a *App) StopAndProcess() ProcessResult {
	if a.cfg.EnableNotifications {
		notify.RecordingProcessing()
	}
	activeApp := osutil.GetActiveWindowName()

	transcript, refined, err := a.engine.StopAndProcess(a.ctx)
	if a.cfg.EnableNotifications {
		notify.RecordingDone()
	}
	indicator.SetRecording(false)
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
	if newSettings.SystemPrompt != "" {
		a.engine.SetSystemPrompt(newSettings.SystemPrompt)
	}

	// Hot-toggle indicator
	if newSettings.EnableIndicator && !a.cfg.EnableIndicator {
		indicator.Start()
	} else if !newSettings.EnableIndicator && a.cfg.EnableIndicator {
		indicator.Stop()
	}

	*a.cfg = newSettings
	return nil
}

// TestPrompt runs the refinement model against sample text with a given prompt.
func (a *App) TestPrompt(sampleText, systemPrompt string) ProcessResult {
	proc := a.engine.GetProcessor()
	refined, err := proc.Refine(a.ctx, sampleText, systemPrompt)
	if err != nil {
		return ProcessResult{Error: err.Error()}
	}
	return ProcessResult{
		Transcript: sampleText,
		Refined:    refined,
	}
}

// GetDefaultPrompt returns the built-in default system prompt.
func (a *App) GetDefaultPrompt() string {
	return config.DefaultSystemPrompt
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

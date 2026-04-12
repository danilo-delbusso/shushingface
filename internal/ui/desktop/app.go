package desktop

import (
	"context"
	"fmt"
	"log/slog"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"codeberg.org/dbus/shushingface/internal/ai"
	"codeberg.org/dbus/shushingface/internal/ai/factory"
	"codeberg.org/dbus/shushingface/internal/config"
	"codeberg.org/dbus/shushingface/internal/core"
	"codeberg.org/dbus/shushingface/internal/history"
	"codeberg.org/dbus/shushingface/internal/indicator"
	"codeberg.org/dbus/shushingface/internal/ipc"
	"codeberg.org/dbus/shushingface/internal/notify"
	"codeberg.org/dbus/shushingface/internal/osutil"
	"codeberg.org/dbus/shushingface/internal/paste"
	"codeberg.org/dbus/shushingface/internal/platform"
)

type App struct {
	ctx      context.Context
	engine   *core.Engine
	cfg      *config.Settings
	history  *history.Manager
	cleanIPC func()
}

func NewApp(engine *core.Engine, cfg *config.Settings, hist *history.Manager) *App {
	return &App{engine: engine, cfg: cfg, history: hist}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	if a.cfg.EnableIndicator {
		indicator.Start(func() { wailsRuntime.WindowShow(a.ctx) })
	}

	cleanup, err := ipc.Listen(func(cmd string) {
		switch cmd {
		case "TOGGLE":
			wailsRuntime.EventsEmit(a.ctx, "hotkey-toggle")
		case "SHOW":
			wailsRuntime.WindowShow(a.ctx)
		case "QUIT":
			wailsRuntime.Quit(a.ctx)
		}
	})
	if err != nil {
		slog.Warn("failed to start IPC listener", "error", err)
	} else {
		a.cleanIPC = cleanup
	}
}

type ProcessResult struct {
	Transcript string `json:"transcript"`
	Refined    string `json:"refined"`
	Error      string `json:"error,omitempty"`
}

func (a *App) Shutdown(_ context.Context) {
	if a.cleanIPC != nil {
		a.cleanIPC()
	}
	indicator.Stop()
}

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

func (a *App) StopAndProcess() ProcessResult {
	if a.cfg.EnableNotifications {
		notify.RecordingProcessing()
	}
	activeApp := osutil.GetActiveWindowName()

	tOpts := ai.TranscribeOptions{Language: a.cfg.TranscriptionLanguage}
	rOpts := a.buildRefineOptions(activeApp)
	slog.Info("using refinement profile", "id", rOpts.SystemPrompt[:min(30, len(rOpts.SystemPrompt))], "examples", len(rOpts.Examples))

	// Build per-profile refiner override if the profile specifies a different connection/model.
	var refinerOverride ai.Refiner
	if p := a.cfg.ActiveProfile(); p != nil && (p.ConnectionID != "" || p.Model != "") {
		r, err := factory.BuildRefiner(a.cfg, p.ConnectionID, p.Model)
		if err != nil {
			slog.Warn("failed to build profile refiner, using default", "error", err)
		} else {
			refinerOverride = r
		}
	}

	transcript, refined, err := a.engine.StopAndProcess(a.ctx, tOpts, rOpts, refinerOverride)
	if a.cfg.EnableNotifications {
		notify.RecordingDone()
	}
	indicator.SetRecording(false)
	if err != nil {
		slog.Error("StopAndProcess failed", "error", err)
		if a.cfg.EnableNotifications {
			notify.Error("Transcription failed", err.Error())
		}
		// Record the failure in history
		if a.cfg.EnableHistory && a.history != nil {
			a.history.Insert("", "", activeApp, err.Error())
		}
		return ProcessResult{Error: err.Error()}
	}

	if a.cfg.EnableHistory && a.history != nil && transcript != "" {
		if _, histErr := a.history.Insert(transcript, refined, activeApp, ""); histErr != nil {
			slog.Error("failed to insert history", "error", histErr)
		}
	}

	if a.cfg.AutoPaste && refined != "" {
		if err := paste.Type(refined); err != nil {
			slog.Warn("auto-paste failed", "error", err)
			if a.cfg.EnableNotifications {
				notify.Error("Auto-paste failed", err.Error())
			}
		}
	}

	return ProcessResult{Transcript: transcript, Refined: refined}
}

func (a *App) GetPlatform() platform.Info {
	return platform.Detect()
}

func (a *App) GetSettings() *config.Settings { return a.cfg }

func (a *App) SaveSettings(newSettings config.Settings) error {
	pair, err := factory.NewFromConfig(&newSettings)
	if err != nil {
		return fmt.Errorf("failed to reload AI processors: %w", err)
	}

	if err := config.Save(&newSettings); err != nil {
		return err
	}

	a.engine.SetTranscriber(pair.Transcriber)
	a.engine.SetRefiner(pair.Refiner)

	if newSettings.EnableIndicator && !a.cfg.EnableIndicator {
		indicator.Start(func() { wailsRuntime.WindowShow(a.ctx) })
	} else if !newSettings.EnableIndicator && a.cfg.EnableIndicator {
		indicator.Stop()
	}

	*a.cfg = newSettings
	return nil
}

type PasteStatus struct {
	Available  bool   `json:"available"`
	InstallCmd string `json:"installCmd"`
}

func (a *App) GetPasteStatus() PasteStatus {
	return PasteStatus{
		Available:  paste.Available(),
		InstallCmd: paste.InstallHint(),
	}
}

func (a *App) GetDefaultProfiles() []config.RefinementProfile {
	return config.DefaultProfiles()
}

func (a *App) GetDefaultSettings() *config.Settings {
	return config.DefaultSettings()
}

func (a *App) ListProviders() []ai.ProviderInfo {
	return ai.ListProviders()
}

// ListModelsForConnection fetches available models for a specific connection.
func (a *App) ListModelsForConnection(connectionID string) ([]ai.ModelInfo, error) {
	conn := a.cfg.GetConnection(connectionID)
	if conn == nil {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}
	provider, err := ai.GetProvider(conn.ProviderID)
	if err != nil {
		return nil, err
	}
	if conn.APIKey == "" {
		return nil, fmt.Errorf("API key not configured for %s", conn.Name)
	}
	return provider.ListModels(a.ctx, conn.APIKey, conn.BaseURL)
}

func (a *App) GetDefaultBuiltInRules() string {
	return config.DefaultBuiltInRules()
}

func (a *App) GetHistory(limit, offset int) ([]history.Record, error) {
	if a.history == nil {
		return nil, fmt.Errorf("history is disabled")
	}
	return a.history.GetHistory(limit, offset)
}

func (a *App) ClearHistory() error {
	if a.history == nil {
		return fmt.Errorf("history is disabled")
	}
	return a.history.ClearHistory()
}

// DeleteAllData clears history and resets settings to factory defaults.
func (a *App) DeleteAllData() error {
	if a.history != nil {
		if err := a.history.ClearHistory(); err != nil {
			return fmt.Errorf("failed to clear history: %w", err)
		}
	}
	defaults := config.DefaultSettings()
	if err := config.Save(defaults); err != nil {
		return fmt.Errorf("failed to reset config: %w", err)
	}

	pair, err := factory.NewFromConfig(defaults)
	if err != nil {
		return fmt.Errorf("failed to reload AI processors: %w", err)
	}
	a.engine.SetTranscriber(pair.Transcriber)
	a.engine.SetRefiner(pair.Refiner)
	*a.cfg = *defaults
	return nil
}

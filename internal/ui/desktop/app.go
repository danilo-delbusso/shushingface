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

	opts := a.buildRefineOptions(activeApp)
	slog.Info("using refinement profile", "id", opts.SystemPrompt[:min(30, len(opts.SystemPrompt))], "examples", len(opts.Examples))

	transcript, refined, err := a.engine.StopAndProcess(a.ctx, opts)
	if a.cfg.EnableNotifications {
		notify.RecordingDone()
	}
	indicator.SetRecording(false)
	if err != nil {
		slog.Error("StopAndProcess failed", "error", err)
		return ProcessResult{Error: err.Error()}
	}

	if a.cfg.EnableHistory && a.history != nil && transcript != "" {
		if _, histErr := a.history.Insert(transcript, refined, activeApp); histErr != nil {
			slog.Error("failed to insert history", "error", histErr)
		}
	}

	// Auto-paste refined text into focused app (no clipboard pollution)
	if a.cfg.AutoPaste && refined != "" {
		if err := paste.Type(refined); err != nil {
			slog.Warn("auto-paste failed", "error", err)
		}
	}

	return ProcessResult{Transcript: transcript, Refined: refined}
}

// buildRefineOptions assembles RefineOptions from the active profile,
// active app context, and recent history.
func (a *App) buildRefineOptions(activeApp string) ai.RefineOptions {
	opts := ai.RefineOptions{}

	if p := a.cfg.ActiveProfile(); p != nil {
		opts.SystemPrompt = p.Prompt
		opts.Sampling = ai.SamplingParams{
			Temperature: p.Temperature,
			TopP:        p.TopP,
		}
		// Static few-shot examples from the profile.
		for _, ex := range p.Examples {
			opts.Examples = append(opts.Examples, ai.FewShotPair{
				Input:  ex.Input,
				Output: ex.Output,
			})
		}
	}

	// Inject active app context when available.
	if activeApp != "" {
		opts.Context = activeApp
	}

	// Append recent history as dynamic few-shot examples so the model
	// calibrates to the user's personal style over time.
	if a.history != nil {
		if records, err := a.history.GetHistory(2, 0); err == nil {
			for _, r := range records {
				if r.RawTranscript != "" && r.RefinedMessage != "" {
					opts.Examples = append(opts.Examples, ai.FewShotPair{
						Input:  r.RawTranscript,
						Output: r.RefinedMessage,
					})
				}
			}
		}
	}

	return opts
}

func (a *App) GetPlatform() platform.Info {
	return platform.Detect()
}

func (a *App) GetSettings() *config.Settings { return a.cfg }

func (a *App) SaveSettings(newSettings config.Settings) error {
	newProcessor, err := factory.NewFromConfig(&newSettings)
	if err != nil {
		return fmt.Errorf("failed to reload AI processor: %w", err)
	}

	if err := config.Save(&newSettings); err != nil {
		return err
	}

	a.engine.SetProcessor(newProcessor)

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

func (a *App) TestPrompt(sampleText, systemPrompt string) ProcessResult {
	proc := a.engine.GetProcessor()
	opts := ai.RefineOptions{SystemPrompt: systemPrompt}
	refined, err := proc.Refine(a.ctx, sampleText, opts)
	if err != nil {
		return ProcessResult{Error: err.Error()}
	}
	return ProcessResult{Transcript: sampleText, Refined: refined}
}

func (a *App) GetDefaultProfiles() []config.RefinementProfile {
	return config.DefaultProfiles(config.DefaultSettings().TranscriptionModel)
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

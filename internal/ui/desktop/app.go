package desktop

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

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
	"codeberg.org/dbus/shushingface/internal/secrets"
	"codeberg.org/dbus/shushingface/internal/update"
	"codeberg.org/dbus/shushingface/internal/version"
)

type App struct {
	ctx      context.Context
	engine   core.Engine
	mu       sync.RWMutex
	cfg      *config.Settings
	secrets  secrets.Store
	history  history.Store
	cleanIPC func()
}

// Caller must hold at least a.mu.RLock.
func (a *App) snapshotConfig() config.Settings {
	s := *a.cfg
	s.Connections = make([]config.Connection, len(a.cfg.Connections))
	copy(s.Connections, a.cfg.Connections)
	s.RefinementProfiles = make([]config.RefinementProfile, len(a.cfg.RefinementProfiles))
	copy(s.RefinementProfiles, a.cfg.RefinementProfiles)
	return s
}

func NewApp(engine core.Engine, cfg *config.Settings, secretStore secrets.Store, hist history.Store) *App {
	return &App{engine: engine, cfg: cfg, secrets: secretStore, history: hist}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	a.mu.RLock()
	enableIndicator := a.cfg.EnableIndicator
	checkForUpdates := a.cfg.CheckForUpdates
	a.mu.RUnlock()

	if enableIndicator {
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

	// Check for updates in background
	if checkForUpdates {
		go func() {
			select {
			case <-time.After(5 * time.Second):
			case <-ctx.Done():
				return
			}
			rel, err := update.Check(ctx, version.Version())
			if err != nil {
				slog.Debug("update check failed", "error", err)
				return
			}
			if rel != nil {
				slog.Info("update available", "version", rel.TagName)
				wailsRuntime.EventsEmit(ctx, "update-available", map[string]string{
					"version": rel.TagName,
					"url":     rel.HTMLURL,
				})
			}
		}()
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
		a.mu.RLock()
		doNotify := a.cfg.EnableNotifications
		a.mu.RUnlock()
		if doNotify {
			notify.RecordingStarted()
		}
		indicator.SetRecording(true)
	}
	return err
}

func (a *App) StopAndProcess() ProcessResult {
	a.mu.RLock()
	cfg := a.snapshotConfig()
	a.mu.RUnlock()

	if cfg.EnableNotifications {
		notify.RecordingProcessing()
	}
	activeApp := osutil.GetActiveWindowName()

	tOpts := ai.TranscribeOptions{Language: cfg.TranscriptionLanguage}
	rOpts := a.buildRefineOptions(&cfg, activeApp)
	slog.Info("using refinement profile", "id", rOpts.SystemPrompt[:min(30, len(rOpts.SystemPrompt))], "examples", len(rOpts.Examples))

	// Build per-profile refiner override if the profile specifies a different connection/model.
	var refinerOverride ai.Refiner
	if p := cfg.ActiveProfile(); p != nil && (p.ConnectionID != "" || p.Model != "") {
		r, err := factory.BuildRefiner(&cfg, p.ConnectionID, p.Model)
		if err != nil {
			slog.Warn("failed to build profile refiner, using default", "error", err)
		} else {
			refinerOverride = r
		}
	}

	transcript, refined, err := a.engine.StopAndProcess(a.ctx, tOpts, rOpts, refinerOverride)
	if cfg.EnableNotifications {
		notify.RecordingDone()
	}
	indicator.SetRecording(false)
	if err != nil {
		slog.Error("StopAndProcess failed", "error", err)
		if cfg.EnableNotifications {
			notify.Error("Transcription failed", err.Error())
		}
		// Record the failure in history
		if cfg.EnableHistory {
			if _, histErr := a.history.Insert("", "", activeApp, err.Error()); histErr != nil {
				slog.Error("failed to insert error into history", "error", histErr)
			}
		}
		return ProcessResult{Error: err.Error()}
	}

	if cfg.EnableHistory && transcript != "" {
		if _, histErr := a.history.Insert(transcript, refined, activeApp, ""); histErr != nil {
			slog.Error("failed to insert history", "error", histErr)
		}
	}

	if cfg.AutoPaste && refined != "" {
		if err := paste.Type(refined); err != nil {
			slog.Warn("auto-paste failed", "error", err)
			if cfg.EnableNotifications {
				notify.Error("Auto-paste failed", err.Error())
			}
		}
	}

	return ProcessResult{Transcript: transcript, Refined: refined}
}

func (a *App) GetVersion() string {
	return version.Version()
}

func (a *App) IsSecretStorageSecure() bool {
	return a.secrets.IsSecure()
}

func (a *App) GetPlatform() platform.Info {
	return platform.Detect()
}

func (a *App) GetSettings() *config.Settings {
	a.mu.RLock()
	s := a.snapshotConfig()
	a.mu.RUnlock()

	// Hydrate API keys on the copy — the in-memory cfg stays stripped.
	config.HydrateAPIKeys(s.Connections, a.secrets.Get)
	return &s
}

func (a *App) SaveSettings(newSettings config.Settings) error {
	// Snapshot old config to detect deleted connections and indicator changes.
	a.mu.RLock()
	oldConns := make([]config.Connection, len(a.cfg.Connections))
	copy(oldConns, a.cfg.Connections)
	oldIndicator := a.cfg.EnableIndicator
	a.mu.RUnlock()

	// Clean up secrets for deleted connections
	newIDs := make(map[string]bool)
	for _, conn := range newSettings.Connections {
		newIDs[conn.ID] = true
	}
	for _, conn := range oldConns {
		if !newIDs[conn.ID] {
			if err := a.secrets.Delete("apikey:" + conn.ID); err != nil {
				slog.Warn("failed to delete secret for removed connection", "connection", conn.ID, "error", err)
			}
		}
	}

	// Store API keys in the secret store
	for i := range newSettings.Connections {
		conn := &newSettings.Connections[i]
		if conn.APIKey != "" {
			if err := a.secrets.Set("apikey:"+conn.ID, conn.APIKey); err != nil {
				slog.Warn("failed to store API key in secret store", "connection", conn.ID, "error", err)
			}
		}
	}

	// Build processors with full keys (before stripping).
	// This validates the config before committing any state changes.
	pair, err := factory.NewFromConfig(&newSettings)
	if err != nil {
		return fmt.Errorf("failed to reload AI processors: %w", err)
	}

	// Strip API keys from the on-disk config when keyring is available.
	// Deep-copy connections to avoid mutating newSettings.
	if a.secrets.IsSecure() {
		stripped := newSettings
		stripped.Connections = make([]config.Connection, len(newSettings.Connections))
		copy(stripped.Connections, newSettings.Connections)
		for i := range stripped.Connections {
			stripped.Connections[i].APIKey = ""
		}
		if err := config.Save(&stripped); err != nil {
			return err
		}
	} else {
		if err := config.Save(&newSettings); err != nil {
			return err
		}
	}

	// Commit state: update engine and config atomically.
	a.engine.SetTranscriber(pair.Transcriber)
	a.engine.SetRefiner(pair.Refiner)

	if newSettings.EnableIndicator && !oldIndicator {
		indicator.Start(func() { wailsRuntime.WindowShow(a.ctx) })
	} else if !newSettings.EnableIndicator && oldIndicator {
		indicator.Stop()
	}

	a.mu.Lock()
	*a.cfg = newSettings
	a.mu.Unlock()
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

func (a *App) ListProviders() []ai.ProviderInfo {
	return ai.ListProviders()
}

func (a *App) ListModelsForConnection(connectionID string) ([]ai.ModelInfo, error) {
	a.mu.RLock()
	orig := a.cfg.GetConnection(connectionID)
	if orig == nil {
		a.mu.RUnlock()
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}
	conn := *orig // copy so we can hydrate outside the lock
	a.mu.RUnlock()

	if conn.APIKey == "" {
		if key, err := a.secrets.Get("apikey:" + conn.ID); err == nil {
			conn.APIKey = key
		}
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

func (a *App) GetLogPath() string {
	p, _ := config.GetLogPath()
	return p
}

func (a *App) GetRecentLogs(lines int) (string, error) {
	logPath, err := config.GetLogPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		return "", err
	}
	// Return last N lines
	all := strings.Split(string(data), "\n")
	if len(all) > lines {
		all = all[len(all)-lines:]
	}
	return strings.Join(all, "\n"), nil
}

func (a *App) GetDefaultBuiltInRules() string {
	return config.DefaultBuiltInRules()
}

func (a *App) GetHistory(limit, offset int) ([]history.Record, error) {
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}
	return a.history.GetHistory(limit, offset)
}

func (a *App) ClearHistory() error {
	return a.history.Clear()
}

func (a *App) DeleteAllData() error {
	if err := a.history.Clear(); err != nil {
		return fmt.Errorf("failed to clear history: %w", err)
	}

	// Snapshot connections for secret cleanup.
	a.mu.RLock()
	oldConns := make([]config.Connection, len(a.cfg.Connections))
	copy(oldConns, a.cfg.Connections)
	a.mu.RUnlock()

	for _, conn := range oldConns {
		if err := a.secrets.Delete("apikey:" + conn.ID); err != nil {
			slog.Warn("failed to delete secret during reset", "connection", conn.ID, "error", err)
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

	a.mu.Lock()
	*a.cfg = *defaults
	a.mu.Unlock()
	return nil
}

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
	"codeberg.org/dbus/shushingface/internal/audio"
	"codeberg.org/dbus/shushingface/internal/config"
	"codeberg.org/dbus/shushingface/internal/core"
	"codeberg.org/dbus/shushingface/internal/history"
	"codeberg.org/dbus/shushingface/internal/hotkey"
	"codeberg.org/dbus/shushingface/internal/indicator"
	"codeberg.org/dbus/shushingface/internal/ipc"
	"codeberg.org/dbus/shushingface/internal/notify"
	"codeberg.org/dbus/shushingface/internal/osutil"
	"codeberg.org/dbus/shushingface/internal/overlay"
	"codeberg.org/dbus/shushingface/internal/paste"
	"codeberg.org/dbus/shushingface/internal/platform"
	"codeberg.org/dbus/shushingface/internal/secrets"
	"codeberg.org/dbus/shushingface/internal/update"
	"codeberg.org/dbus/shushingface/internal/version"
)

type App struct {
	ctx      context.Context
	engine   core.Engine
	recorder audio.Recorder
	mu       sync.RWMutex
	cfg      *config.Settings
	secrets  secrets.Store
	history  history.Store
	cleanIPC func()
	hotkey   hotkey.Manager
	overlay  overlay.Overlay

	// levelStop closes the level-pump goroutine started in StartRecording.
	// Replaced (and the previous one closed) on every start.
	levelStop chan struct{}
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

func NewApp(engine core.Engine, recorder audio.Recorder, cfg *config.Settings, secretStore secrets.Store, hist history.Store) *App {
	return &App{engine: engine, recorder: recorder, cfg: cfg, secrets: secretStore, history: hist}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.overlay = overlay.New()

	a.mu.RLock()
	enableIndicator := a.cfg.EnableIndicator
	checkForUpdates := a.cfg.CheckForUpdates
	shortcutSpec := a.cfg.Shortcut
	a.mu.RUnlock()

	a.mu.RLock()
	deviceID := a.cfg.InputDeviceID
	a.mu.RUnlock()
	if deviceID != "" {
		if err := a.recorder.SetDevice(deviceID); err != nil {
			slog.Warn("failed to apply saved input device, falling back to default", "id", deviceID, "error", err)
		}
	}

	if hotkey.Capability().Supported {
		a.hotkey = hotkey.New()
		go func() {
			for ev := range a.hotkey.Events() {
				if ev.Name != "toggle" {
					continue
				}
				switch ev.Type {
				case hotkey.Trigger:
					wailsRuntime.EventsEmit(a.ctx, "hotkey-toggle")
				case hotkey.Press:
					wailsRuntime.EventsEmit(a.ctx, "hotkey-press")
				case hotkey.Release:
					wailsRuntime.EventsEmit(a.ctx, "hotkey-release")
				}
			}
		}()
		if shortcutSpec != "" {
			if err := a.registerShortcut(shortcutSpec); err != nil {
				slog.Warn("failed to register shortcut at startup", "spec", shortcutSpec, "error", err)
			}
		}
	}

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
	if a.hotkey != nil {
		if err := a.hotkey.Close(); err != nil {
			slog.Warn("hotkey close failed", "error", err)
		}
	}
	if a.overlay != nil {
		if err := a.overlay.Close(); err != nil {
			slog.Warn("overlay close failed", "error", err)
		}
	}
	indicator.Stop()
}

// registerShortcut parses spec and registers the toggle hotkey using the
// recording mode currently in cfg. Caller must not hold a.mu.
func (a *App) registerShortcut(spec string) error {
	if a.hotkey == nil {
		return hotkey.ErrUnsupported
	}
	parsed, err := hotkey.ParseSpec(spec)
	if err != nil {
		return err
	}
	a.mu.RLock()
	mode := a.cfg.RecordingMode
	a.mu.RUnlock()
	return a.hotkey.Register("toggle", parsed, parseMode(mode))
}

func parseMode(s string) hotkey.Mode {
	if s == "push_to_talk" {
		return hotkey.ModePushToTalk
	}
	return hotkey.ModeToggle
}

// GetShortcut returns the saved shortcut string (may be "").
func (a *App) GetShortcut() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.Shortcut
}

// SetShortcut validates, registers, and persists a new shortcut.
func (a *App) SetShortcut(spec string) error {
	slog.Info("SetShortcut called", "spec", spec)
	parsed, err := hotkey.ParseSpec(spec)
	if err != nil {
		slog.Warn("SetShortcut parse failed", "spec", spec, "error", err)
		return err
	}
	canonical := hotkey.FormatSpec(parsed)

	a.mu.RLock()
	mode := parseMode(a.cfg.RecordingMode)
	a.mu.RUnlock()

	if a.hotkey != nil {
		if err := a.hotkey.Register("toggle", parsed, mode); err != nil {
			slog.Warn("SetShortcut register failed", "canonical", canonical, "error", err)
			return err
		}
	} else {
		slog.Warn("SetShortcut: a.hotkey is nil")
	}

	a.mu.Lock()
	a.cfg.Shortcut = canonical
	snapshot := a.snapshotConfig()
	a.mu.Unlock()
	if err := config.Save(&snapshot); err != nil {
		slog.Error("SetShortcut config.Save failed", "error", err)
		return err
	}
	slog.Info("SetShortcut persisted", "canonical", canonical)
	return nil
}

// ClearShortcut removes the registered shortcut.
func (a *App) ClearShortcut() error {
	if a.hotkey != nil {
		if err := a.hotkey.Unregister("toggle"); err != nil {
			slog.Warn("unregister failed", "error", err)
		}
	}
	a.mu.Lock()
	a.cfg.Shortcut = ""
	snapshot := a.snapshotConfig()
	a.mu.Unlock()
	return config.Save(&snapshot)
}

// HotkeyCapabilities reports whether in-app shortcut binding is available.
// Kept as a separate method (not just GetCapabilities().Hotkey) because the
// frontend shortcut recorder binds to it directly.
func (a *App) HotkeyCapabilities() platform.Capability {
	return hotkey.Capability()
}

// Capabilities is the single bundle the frontend reads to decide which
// platform-dependent toggles to show / grey out. Add a new field here when
// a new feature gains a Capability() — there's no reason these should each
// round-trip through their own Wails binding.
type Capabilities struct {
	Hotkey          platform.Capability `json:"hotkey"`
	Paste           platform.Capability `json:"paste"`
	Notifications   platform.Capability `json:"notifications"`
	TrayIndicator   platform.Capability `json:"trayIndicator"`
	Overlay         platform.Capability `json:"overlay"`
	ActiveWindowTag platform.Capability `json:"activeWindowTag"`
}

// GetCapabilities returns every platform feature's Capability in one call.
func (a *App) GetCapabilities() Capabilities {
	return Capabilities{
		Hotkey:          hotkey.Capability(),
		Paste:           paste.Capability(),
		Notifications:   notify.Capability(),
		TrayIndicator:   indicator.Capability(),
		Overlay:         overlay.Capability(),
		ActiveWindowTag: osutil.ActiveAppCapability(),
	}
}

func (a *App) StartRecording() error {
	err := a.engine.StartRecording()
	if err == nil {
		a.mu.RLock()
		doNotify := a.cfg.EnableNotifications
		showOverlay := a.cfg.OverlayEnabled
		opacity := a.cfg.OverlayOpacity
		a.mu.RUnlock()
		if doNotify {
			notify.RecordingStarted()
		}
		indicator.SetRecording(true)
		if showOverlay && a.overlay != nil {
			if err := a.overlay.Show("Recording...", opacity); err != nil {
				slog.Warn("overlay show failed", "error", err)
			}
			a.startLevelPump()
		}
	}
	return err
}

// startLevelPump forwards live mic amplitudes from the engine to the
// overlay until stopLevelPump is called. Idempotent: cancels any prior
// pump first.
func (a *App) startLevelPump() {
	a.stopLevelPump()
	stop := make(chan struct{})
	a.levelStop = stop
	levels := a.engine.Level()
	go func() {
		for {
			select {
			case <-stop:
				return
			case v, ok := <-levels:
				if !ok {
					return
				}
				a.overlay.SetLevel(v)
			}
		}
	}()
}

func (a *App) stopLevelPump() {
	if a.levelStop != nil {
		close(a.levelStop)
		a.levelStop = nil
	}
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

	// Switch the overlay from "recording bars" to "processing loader" the
	// instant the user asks to stop — capture is over, but transcription
	// + refinement may take seconds and we want visible feedback that
	// something is still happening. We drop the level pump (no more mic
	// data is incoming) but keep the window open until processing ends.
	indicator.SetRecording(false)
	a.stopLevelPump()
	if a.overlay != nil {
		a.overlay.SetMode(overlay.ModeProcessing)
	}

	transcript, refined, err := a.engine.StopAndProcess(a.ctx, tOpts, rOpts, refinerOverride)
	if a.overlay != nil {
		if hErr := a.overlay.Hide(); hErr != nil {
			slog.Warn("overlay hide failed", "error", hErr)
		}
	}
	if cfg.EnableNotifications {
		notify.RecordingDone()
	}
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
	oldDeviceID := a.cfg.InputDeviceID
	oldMode := a.cfg.RecordingMode
	oldOverlayEnabled := a.cfg.OverlayEnabled
	currentShortcut := a.cfg.Shortcut
	a.mu.RUnlock()

	// Shortcut is owned by SetShortcut/ClearShortcut; ignore whatever the
	// frontend sent so a stale settings snapshot can't clobber it.
	newSettings.Shortcut = currentShortcut

	if newSettings.InputDeviceID != oldDeviceID {
		if err := a.recorder.SetDevice(newSettings.InputDeviceID); err != nil {
			return fmt.Errorf("switch input device: %w", err)
		}
	}

	// If the recording mode changed and a shortcut is bound, re-register it
	// in the new mode so the keyboard hook / RegisterHotKey reflects the choice.
	if newSettings.RecordingMode != oldMode && currentShortcut != "" && a.hotkey != nil {
		if err := a.hotkey.Unregister("toggle"); err != nil {
			slog.Warn("hotkey unregister during mode change failed", "error", err)
		}
		// Apply the new mode before re-registering by writing cfg here too.
		a.mu.Lock()
		a.cfg.RecordingMode = newSettings.RecordingMode
		a.mu.Unlock()
		if err := a.registerShortcut(currentShortcut); err != nil {
			slog.Warn("hotkey re-register after mode change failed", "error", err)
		}
	}

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

	// If the user turned the overlay off while it's currently visible,
	// dismiss it immediately — otherwise it lingers until the next Show call.
	if oldOverlayEnabled && !newSettings.OverlayEnabled && a.overlay != nil {
		if err := a.overlay.Hide(); err != nil {
			slog.Warn("overlay hide on toggle-off failed", "error", err)
		}
	}

	// Apply debug-log toggle live; next slog call picks up the new level.
	config.ApplyLogLevel(newSettings.DebugLogging)

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

// ListInputDevices returns the available capture devices.
func (a *App) ListInputDevices() ([]audio.DeviceInfo, error) {
	return a.recorder.ListDevices()
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

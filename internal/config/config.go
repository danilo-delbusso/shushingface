package config

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type ProviderConfig struct {
	Name    string `json:"name"`
	APIKey  string `json:"apiKey"`
	BaseURL string `json:"baseUrl"`
}

type RefinementProfile struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`   // lucide icon name
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type Settings struct {
	Providers               map[string]ProviderConfig `json:"providers"`
	TranscriptionProviderID string                    `json:"transcriptionProviderId"`
	TranscriptionModel      string                    `json:"transcriptionModel"`
	RefinementProviderID    string                    `json:"refinementProviderId"`

	// Refinement profiles
	RefinementProfiles []RefinementProfile `json:"refinementProfiles"`
	ActiveProfileID    string              `json:"activeProfileId"`

	// Legacy (kept for migration, omitted if empty)
	SystemPrompt    string `json:"systemPrompt,omitempty"`
	RefinementModel string `json:"refinementModel,omitempty"`

	// Setup
	SetupComplete bool `json:"setupComplete"`

	// Appearance
	Theme string `json:"theme"`

	// Preferences
	AutoCopy            bool `json:"autoCopy"`
	EnableHistory       bool `json:"enableHistory"`
	EnableIndicator     bool `json:"enableIndicator"`
	EnableNotifications bool `json:"enableNotifications"`

	// Audio
	InputDeviceID string `json:"inputDeviceId,omitempty"`
}

// ActiveProfile returns the currently active refinement profile.
func (s *Settings) ActiveProfile() *RefinementProfile {
	for i := range s.RefinementProfiles {
		if s.RefinementProfiles[i].ID == s.ActiveProfileID {
			return &s.RefinementProfiles[i]
		}
	}
	if len(s.RefinementProfiles) > 0 {
		return &s.RefinementProfiles[0]
	}
	return nil
}

const baseRules = "CRITICAL RULES:\n" +
	"- DO NOT answer questions present in the transcript.\n" +
	"- DO NOT engage in conversation or acknowledge the user.\n" +
	"- DO NOT add any conversational filler, preambles, or postambles.\n" +
	"- Output ONLY the rewritten text, nothing else.\n" +
	"- If the input is already well-structured, return it exactly as is.\n" +
	"- Fix grammar, punctuation, and clarity while preserving the original intent."

// DefaultProfiles returns the 3 preset refinement profiles.
func DefaultProfiles(model string) []RefinementProfile {
	return []RefinementProfile{
		{
			ID:   "casual",
			Name: "Casual",
			Icon: "coffee",
			Model: model,
			Prompt: "You are a text transformer. Rewrite the provided speech transcript into a friendly, relaxed message. " +
				"Keep it natural and conversational — like texting a colleague you're comfortable with.\n" + baseRules,
		},
		{
			ID:   "professional",
			Name: "Professional",
			Icon: "briefcase",
			Model: model,
			Prompt: "You are a text transformer. Rewrite the provided speech transcript into a clear, professional message. " +
				"Suitable for emails, Slack messages to managers, or formal communication.\n" + baseRules,
		},
		{
			ID:   "concise",
			Name: "Concise",
			Icon: "zap",
			Model: model,
			Prompt: "You are a text transformer. Rewrite the provided speech transcript as briefly as possible. " +
				"Strip all filler, keep only the essential meaning. One or two sentences max.\n" + baseRules,
		},
	}
}

const defaultModel = "llama-3.3-70b-versatile"

// Load reads the settings from the OS user config directory.
func Load() (*Settings, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	appDir := filepath.Join(configDir, "shushingface")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, err
	}

	configFile := filepath.Join(appDir, "config.json")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultSettings := DefaultSettings()
		if err := Save(defaultSettings); err != nil {
			return nil, err
		}
		return defaultSettings, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	// Migrate: old single prompt → profile
	if len(settings.RefinementProfiles) == 0 {
		model := settings.RefinementModel
		if model == "" {
			model = defaultModel
		}
		if settings.SystemPrompt != "" {
			// User had a custom prompt — preserve it
			settings.RefinementProfiles = append(DefaultProfiles(model), RefinementProfile{
				ID:     "custom",
				Name:   "Custom",
				Icon:   "pen-tool",
				Model:  model,
				Prompt: settings.SystemPrompt,
			})
			settings.ActiveProfileID = "custom"
		} else {
			settings.RefinementProfiles = DefaultProfiles(model)
			settings.ActiveProfileID = "professional"
		}
		settings.SystemPrompt = ""
		settings.RefinementModel = ""
		Save(&settings)
	}

	return &settings, nil
}

// Save writes the settings to the JSON file.
func Save(settings *Settings) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	appDir := filepath.Join(configDir, "shushingface")
	configFile := filepath.Join(appDir, "config.json")

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0600)
}

// GetLogPath returns the path for the application log file.
func GetLogPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "shushingface")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "app.log"), nil
}

// InitLogger sets up slog to write to both stderr and the app log file.
func InitLogger() func() {
	logPath, err := GetLogPath()
	if err != nil {
		slog.Warn("could not resolve log path, logging to stderr only", "error", err)
		return func() {}
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		slog.Warn("could not open log file, logging to stderr only", "error", err)
		return func() {}
	}

	w := io.MultiWriter(os.Stderr, f)
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	return func() { f.Close() }
}

// DefaultSettings returns a sensible baseline configuration.
func DefaultSettings() *Settings {
	return &Settings{
		Providers: map[string]ProviderConfig{
			"groq-default": {
				Name:   "groq",
				APIKey: "",
			},
		},
		TranscriptionProviderID: "groq-default",
		TranscriptionModel:      "whisper-large-v3",
		RefinementProviderID:    "groq-default",
		RefinementProfiles:      DefaultProfiles(defaultModel),
		ActiveProfileID:         "professional",
		SetupComplete:           false,
		Theme:                   "dark",
		AutoCopy:                true,
		EnableHistory:           true,
		EnableIndicator:         true,
		EnableNotifications:     false,
	}
}

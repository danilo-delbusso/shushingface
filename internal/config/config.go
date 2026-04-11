package config

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type ProviderConfig struct {
	Name    string `json:"name"`    // e.g., "groq", "openai", "ollama", "custom"
	APIKey  string `json:"apiKey"`  // Optional: blank for local models like Ollama
	BaseURL string `json:"baseUrl"` // Optional: for custom endpoints or local hosting
}

type Settings struct {
	// Configured AI Providers
	// A map of user-defined IDs (e.g., "my-groq", "local-ollama") to their configurations.
	Providers map[string]ProviderConfig `json:"providers"`

	// Active Routing: Which provider ID and model to use for each step
	TranscriptionProviderID string `json:"transcriptionProviderId"`
	TranscriptionModel      string `json:"transcriptionModel"`

	RefinementProviderID string `json:"refinementProviderId"`
	RefinementModel      string `json:"refinementModel"`

	// Refinement
	SystemPrompt string `json:"systemPrompt"`

	// Preferences
	GlobalHotkey        string `json:"globalHotkey"`
	AutoCopy            bool   `json:"autoCopy"`
	EnableHistory       bool   `json:"enableHistory"`
	EnableIndicator     bool   `json:"enableIndicator"`
	EnableNotifications bool   `json:"enableNotifications"`

	// Audio
	InputDeviceID string `json:"inputDeviceId,omitempty"`
}

// Load reads the settings from the OS user config directory.
func Load() (*Settings, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	appDir := filepath.Join(configDir, "sussurro")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, err
	}

	configFile := filepath.Join(appDir, "config.json")

	// If file doesn't exist, create a default one
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

	return &settings, nil
}

// Save writes the settings to the JSON file.
func Save(settings *Settings) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	appDir := filepath.Join(configDir, "sussurro")
	configFile := filepath.Join(appDir, "config.json")

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	// 0600 because it contains API keys
	return os.WriteFile(configFile, data, 0600)
}

// GetLogPath returns the path for the application log file.
func GetLogPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "sussurro")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "app.log"), nil
}

// InitLogger sets up slog to write to both stderr and the app log file.
// Returns a cleanup function to close the log file.
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

// DefaultSystemPrompt is the built-in refinement prompt.
const DefaultSystemPrompt = "You are a text transformer, NOT a conversational AI. " +
	"Your ONLY task is to rewrite the provided speech transcript into a clear, professional, yet conversational message. " +
	"CRITICAL RULES:\n" +
	"- DO NOT answer questions present in the transcript.\n" +
	"- DO NOT engage in conversation or acknowledge the user.\n" +
	"- DO NOT add any conversational filler, preambles (e.g., 'Here is the refined message:'), or postambles.\n" +
	"- Output ONLY the rewritten text, nothing else.\n" +
	"- If the input is already well-structured, return it exactly as is.\n" +
	"- Fix grammar, punctuation, and clarity while preserving the original intent.\n" +
	"- Use paragraph breaks or bullet points only if it significantly improves readability."

// DefaultSettings returns a sensible baseline configuration.
func DefaultSettings() *Settings {
	return &Settings{
		Providers: map[string]ProviderConfig{
			"groq-default": {
				Name:    "groq",
				APIKey:  "", // Must be filled by user
				BaseURL: "",
			},
		},
		TranscriptionProviderID: "groq-default",
		TranscriptionModel:      "whisper-large-v3",
		RefinementProviderID:    "groq-default",
		RefinementModel:         "llama-3.3-70b-versatile",
		SystemPrompt:            DefaultSystemPrompt,
		GlobalHotkey:            "Ctrl+Shift+R",
		AutoCopy:                true,
		EnableHistory:           true,
		EnableIndicator:         true,
		EnableNotifications:     false,
	}
}

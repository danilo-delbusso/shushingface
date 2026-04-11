package config

import (
	"encoding/json"
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

	// Preferences
	GlobalHotkey  string `json:"globalHotkey"`
	AutoCopy      bool   `json:"autoCopy"`
	EnableHistory bool   `json:"enableHistory"` // Opt-in local history tracking

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
		RefinementModel:         "llama3-70b-8192",
		GlobalHotkey:            "Ctrl+Shift+R",
		AutoCopy:                true,
		EnableHistory:           true,
	}
}

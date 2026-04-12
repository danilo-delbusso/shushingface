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

type FewShotExample struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type RefinementProfile struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Icon        string           `json:"icon"`   // lucide icon name
	Model       string           `json:"model"`
	Prompt      string           `json:"prompt"`
	Examples    []FewShotExample `json:"examples,omitempty"`
	Temperature float32          `json:"temperature,omitempty"`
	TopP        float32          `json:"topP,omitempty"`
}

type Settings struct {
	Providers               map[string]ProviderConfig `json:"providers"`
	TranscriptionProviderID string                    `json:"transcriptionProviderId"`
	TranscriptionModel      string                    `json:"transcriptionModel"`
	RefinementProviderID    string                    `json:"refinementProviderId"`

	// Refinement profiles
	RefinementProfiles []RefinementProfile `json:"refinementProfiles"`
	ActiveProfileID    string              `json:"activeProfileId"`
	GlobalRules        string              `json:"globalRules,omitempty"`

	// Legacy (kept for migration, omitted if empty)
	SystemPrompt    string `json:"systemPrompt,omitempty"`
	RefinementModel string `json:"refinementModel,omitempty"`

	// Setup
	SetupComplete bool `json:"setupComplete"`

	// Appearance
	Theme string `json:"theme"`

	// Preferences
	AutoPaste           bool `json:"autoPaste"`
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

const baseRules = "\n\nRules:\n" +
	"- Output only the rewritten text, nothing else.\n" +
	"- Preserve the speaker's original meaning, intent, and any questions exactly as stated.\n" +
	"- Clean up speech artifacts: filler words (um, uh, like, you know), false starts, and repetitions.\n" +
	"- Keep the speaker's natural phrasing and word choices where they already work — only fix what is actually broken.\n" +
	"- Never add words, ideas, or formality the speaker did not express.\n" +
	"- Return well-written input unchanged."

// DefaultProfiles returns the 3 preset refinement profiles.
func DefaultProfiles(model string) []RefinementProfile {
	return []RefinementProfile{
		{
			ID:    "casual",
			Name:  "Casual",
			Icon:  "coffee",
			Model: model,
			Prompt: "You are a speech-to-text editor. Rewrite the transcript so it reads like something the speaker would actually type — " +
				"relaxed, natural, the way you'd message a colleague you're comfortable with. " +
				"Keep contractions, casual phrasing, and personality. Keep all the meaning intact — don't drop points or compress." + baseRules,
			Examples: []FewShotExample{
				{
					Input:  "so um I was thinking we should we should probably move the meeting to thursday because like john can't make it on wednesday and I think it would be better if everyone was there you know",
					Output: "I think we should move the meeting to Thursday, John can't make it on Wednesday and it'd be better if everyone was there",
				},
				{
					Input:  "hey so the the deployment went fine but we noticed that the the login page is loading kind of slowly so we might want to look into that",
					Output: "hey so the deployment went fine but we noticed the login page is loading kind of slowly, so we might want to look into that",
				},
			},
			Temperature: 0.4,
			TopP:        0.9,
		},
		{
			ID:    "professional",
			Name:  "Professional",
			Icon:  "briefcase",
			Model: model,
			Prompt: "You are a speech-to-text editor. Rewrite the transcript into clear, professional text " +
				"suitable for emails and workplace communication. Use complete sentences and precise language, " +
				"but keep it human — avoid corporate jargon and stiff phrasing that nobody would actually write." + baseRules,
			Examples: []FewShotExample{
				{
					Input:  "so um I was thinking we should we should probably move the meeting to thursday because like john can't make it on wednesday and I think it would be better if everyone was there you know",
					Output: "I'd like to move the meeting to Thursday since John can't make it on Wednesday. It would be better to have everyone there.",
				},
				{
					Input:  "I just wanted to flag that uh the the API response times have been creeping up over the past week or so and I think we should probably look into it before it becomes a bigger issue",
					Output: "I wanted to flag that API response times have been creeping up over the past week. We should look into it before it becomes a bigger issue.",
				},
			},
			Temperature: 0.3,
			TopP:        0.9,
		},
		{
			ID:    "concise",
			Name:  "Concise",
			Icon:  "zap",
			Model: model,
			Prompt: "You are a speech-to-text editor. Compress the transcript to its essential meaning. " +
				"Strip filler, hedging, repetition, and unnecessary detail. One to two sentences. Every word earns its place." + baseRules,
			Examples: []FewShotExample{
				{
					Input:  "so um I was thinking we should we should probably move the meeting to thursday because like john can't make it on wednesday and I think it would be better if everyone was there you know",
					Output: "Move the meeting to Thursday — John can't make Wednesday.",
				},
				{
					Input:  "I just wanted to flag that uh the the API response times have been creeping up over the past week or so and I think we should probably look into it before it becomes a bigger issue",
					Output: "API response times are creeping up. We should investigate before it gets worse.",
				},
			},
			Temperature: 0.2,
			TopP:        0.9,
		},
	}
}

const defaultModel = "meta-llama/llama-4-scout-17b-16e-instruct"

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
		AutoPaste:               true,
		AutoCopy:                false,
		EnableHistory:           true,
		EnableIndicator:         true,
		EnableNotifications:     false,
	}
}

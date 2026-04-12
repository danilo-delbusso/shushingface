package config

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// ──────────────────────────────────────────────────
// Types
// ──────────────────────────────────────────────────

// Connection represents a named AI provider configuration.
type Connection struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ProviderID string `json:"providerId"`
	APIKey     string `json:"apiKey"`
	BaseURL    string `json:"baseUrl,omitempty"`
}

type FewShotExample struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type RefinementProfile struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Icon         string           `json:"icon"`                    // lucide icon name
	ConnectionID string           `json:"connectionId,omitempty"`  // override; empty = use global
	Model        string           `json:"model"`                   // override; empty = use global RefinementModel
	Prompt       string           `json:"prompt"`
	Examples     []FewShotExample `json:"examples,omitempty"`
	Temperature  float32          `json:"temperature,omitempty"`
	TopP         float32          `json:"topP,omitempty"`
}

type Settings struct {
	// Connections — multiple named AI service configurations
	Connections []Connection `json:"connections"`

	// Default assignments — connection ID + model for each function
	TranscriptionConnectionID string `json:"transcriptionConnectionId"`
	TranscriptionModel        string `json:"transcriptionModel"`
	TranscriptionLanguage     string `json:"transcriptionLanguage,omitempty"` // ISO 639-1, empty = auto-detect
	RefinementConnectionID    string `json:"refinementConnectionId"`
	RefinementModel           string `json:"refinementModel"`

	// Refinement profiles
	RefinementProfiles []RefinementProfile `json:"refinementProfiles"`
	ActiveProfileID    string              `json:"activeProfileId"`
	GlobalRules        string              `json:"globalRules,omitempty"`
	BuiltInRules       string              `json:"builtInRules,omitempty"`

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

	// Legacy fields — kept only for migration, cleared on load
	LegacyProviderID      string                          `json:"providerId,omitempty"`
	LegacyProviderAPIKey  string                          `json:"providerApiKey,omitempty"`
	LegacyProviderBaseURL string                          `json:"providerBaseUrl,omitempty"`
	LegacyProviders       map[string]legacyProviderConfig `json:"providers,omitempty"`
	LegacyTransProviderID string                          `json:"transcriptionProviderId,omitempty"`
	LegacyRefProviderID   string                          `json:"refinementProviderId,omitempty"`
	LegacySystemPrompt    string                          `json:"systemPrompt,omitempty"`
}

type legacyProviderConfig struct {
	Name    string `json:"name"`
	APIKey  string `json:"apiKey"`
	BaseURL string `json:"baseUrl"`
}

// ──────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────

// GetConnection returns the connection with the given ID, or nil.
func (s *Settings) GetConnection(id string) *Connection {
	for i := range s.Connections {
		if s.Connections[i].ID == id {
			return &s.Connections[i]
		}
	}
	return nil
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

// EffectiveRefinementConnectionID returns the connection to use for refinement,
// checking the active profile override first, then the global default.
func (s *Settings) EffectiveRefinementConnectionID() string {
	if p := s.ActiveProfile(); p != nil && p.ConnectionID != "" {
		return p.ConnectionID
	}
	return s.RefinementConnectionID
}

// EffectiveRefinementModel returns the model to use for refinement,
// checking the active profile override first, then the global default.
func (s *Settings) EffectiveRefinementModel() string {
	if p := s.ActiveProfile(); p != nil && p.Model != "" {
		return p.Model
	}
	return s.RefinementModel
}

// ──────────────────────────────────────────────────
// Built-in rules
// ──────────────────────────────────────────────────

const defaultBuiltInRules = "- Output only the rewritten text, nothing else.\n" +
	"- Keep all meaning intact — never drop points, details, or nuance the speaker expressed.\n" +
	"- Preserve the speaker's original intent and any questions exactly as stated.\n" +
	"- Clean up speech artifacts: filler words (um, uh, like, you know), false starts, and repetitions.\n" +
	"- Keep the speaker's natural phrasing and word choices where they already work — only fix what is actually broken.\n" +
	"- Never add words, ideas, or formality the speaker did not express.\n" +
	"- Return well-written input unchanged."

// DefaultBuiltInRules returns the factory-default built-in rules string.
func DefaultBuiltInRules() string { return defaultBuiltInRules }

// GetBuiltInRules returns the active built-in rules (user-customised or default).
func (s *Settings) GetBuiltInRules() string {
	if s.BuiltInRules != "" {
		return s.BuiltInRules
	}
	return defaultBuiltInRules
}

// ──────────────────────────────────────────────────
// Default profiles & settings
// ──────────────────────────────────────────────────

const DefaultRefinementModel = "meta-llama/llama-4-scout-17b-16e-instruct"
const DefaultTranscriptionModel = "whisper-large-v3"

// DefaultProfiles returns the 3 preset refinement profiles.
func DefaultProfiles() []RefinementProfile {
	return []RefinementProfile{
		{
			ID:   "casual",
			Name: "Casual",
			Icon: "coffee",
			Prompt: "You are a speech-to-text editor. Rewrite the transcript so it reads like something the speaker would actually type — " +
				"relaxed, natural, the way you'd message a colleague you're comfortable with. " +
				"Keep contractions, casual phrasing, and personality.",
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
			ID:   "professional",
			Name: "Professional",
			Icon: "briefcase",
			Prompt: "You are a speech-to-text editor. Rewrite the transcript into clear, professional text " +
				"suitable for emails and workplace communication. Use complete sentences and precise language, " +
				"but keep it human — avoid corporate jargon and stiff phrasing that nobody would actually write.",
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
			ID:   "concise",
			Name: "Concise",
			Icon: "zap",
			Prompt: "You are a speech-to-text editor. Compress the transcript to its essential meaning. " +
				"Strip filler, hedging, repetition, and unnecessary detail. One to two sentences. Every word earns its place.",
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

// DefaultSettings returns a sensible baseline configuration (no connections).
func DefaultSettings() *Settings {
	return &Settings{
		TranscriptionModel: DefaultTranscriptionModel,
		RefinementModel:    DefaultRefinementModel,
		RefinementProfiles: DefaultProfiles(),
		ActiveProfileID:    "professional",
		SetupComplete:      false,
		Theme:              "dark",
		AutoPaste:          true,
		AutoCopy:           false,
		EnableHistory:      true,
		EnableIndicator:    true,
		EnableNotifications: false,
	}
}

// ──────────────────────────────────────────────────
// Load / Save / Migration
// ──────────────────────────────────────────────────

func providerDisplayName(id string) string {
	switch id {
	case "groq":
		return "Groq"
	default:
		return id
	}
}

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

	migrated := false

	// Migrate: old multi-provider map → single provider fields (intermediate step)
	if settings.LegacyProviderID == "" && len(settings.LegacyProviders) > 0 {
		provID := settings.LegacyTransProviderID
		if provID == "" {
			provID = settings.LegacyRefProviderID
		}
		if old, ok := settings.LegacyProviders[provID]; ok {
			settings.LegacyProviderID = old.Name
			if settings.LegacyProviderID == "" {
				settings.LegacyProviderID = "groq"
			}
			settings.LegacyProviderAPIKey = old.APIKey
			settings.LegacyProviderBaseURL = old.BaseURL
		}
		settings.LegacyProviders = nil
		settings.LegacyTransProviderID = ""
		settings.LegacyRefProviderID = ""
		migrated = true
	}

	// Migrate: single provider → connections slice
	if len(settings.Connections) == 0 && settings.LegacyProviderID != "" {
		connID := "default"
		provID := settings.LegacyProviderID
		settings.Connections = []Connection{{
			ID:         connID,
			Name:       providerDisplayName(provID),
			ProviderID: provID,
			APIKey:     settings.LegacyProviderAPIKey,
			BaseURL:    settings.LegacyProviderBaseURL,
		}}
		settings.TranscriptionConnectionID = connID
		settings.RefinementConnectionID = connID
		settings.LegacyProviderID = ""
		settings.LegacyProviderAPIKey = ""
		settings.LegacyProviderBaseURL = ""
		migrated = true
	}

	// Migrate: old single prompt → profiles
	if len(settings.RefinementProfiles) == 0 {
		if settings.LegacySystemPrompt != "" {
			settings.RefinementProfiles = append(DefaultProfiles(), RefinementProfile{
				ID:     "custom",
				Name:   "Custom",
				Icon:   "pen-tool",
				Prompt: settings.LegacySystemPrompt,
			})
			settings.ActiveProfileID = "custom"
		} else {
			settings.RefinementProfiles = DefaultProfiles()
			settings.ActiveProfileID = "professional"
		}
		settings.LegacySystemPrompt = ""
		migrated = true
	}

	// Ensure models are set
	if settings.RefinementModel == "" {
		if p := settings.ActiveProfile(); p != nil && p.Model != "" {
			settings.RefinementModel = p.Model
		} else {
			settings.RefinementModel = DefaultRefinementModel
		}
		migrated = true
	}
	if settings.TranscriptionModel == "" {
		settings.TranscriptionModel = DefaultTranscriptionModel
		migrated = true
	}

	if migrated {
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

// ──────────────────────────────────────────────────
// Utility
// ──────────────────────────────────────────────────

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

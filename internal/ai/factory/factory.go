package factory

import (
	"context"
	"errors"
	"fmt"
	"os"

	"codeberg.org/dbus/shushingface/internal/ai"
	"codeberg.org/dbus/shushingface/internal/config"

	// Register providers — each init() calls ai.RegisterProvider.
	_ "codeberg.org/dbus/shushingface/internal/ai/groq"
)

// ProcessorPair holds the default transcriber and refiner built from config.
type ProcessorPair struct {
	Transcriber ai.Transcriber
	Refiner     ai.Refiner
}

// NewFromConfig builds a Transcriber and Refiner from the settings.
func NewFromConfig(cfg *config.Settings) (*ProcessorPair, error) {
	transcriber, err := BuildTranscriber(cfg)
	if err != nil {
		return nil, fmt.Errorf("transcription: %w", err)
	}

	refiner, err := BuildRefiner(cfg, "", "")
	if err != nil {
		return nil, fmt.Errorf("refinement: %w", err)
	}

	return &ProcessorPair{Transcriber: transcriber, Refiner: refiner}, nil
}

// BuildTranscriber creates a Transcriber from the default transcription connection.
func BuildTranscriber(cfg *config.Settings) (ai.Transcriber, error) {
	conn := cfg.GetConnection(cfg.TranscriptionConnectionID)
	if conn == nil {
		return &PlaceholderProcessor{Reason: "No transcription connection configured. Set one up in Connections."}, nil
	}
	return buildProcessor(conn, cfg.TranscriptionModel, "")
}

// BuildRefiner creates a Refiner from a connection+model.
// If connectionID or model are empty, the global defaults from cfg are used.
func BuildRefiner(cfg *config.Settings, connectionID, model string) (ai.Refiner, error) {
	connID := connectionID
	if connID == "" {
		connID = cfg.EffectiveRefinementConnectionID()
	}
	conn := cfg.GetConnection(connID)
	if conn == nil {
		return &PlaceholderProcessor{Reason: "No refinement connection configured. Set one up in Connections."}, nil
	}
	m := model
	if m == "" {
		m = cfg.EffectiveRefinementModel()
	}
	return buildProcessor(conn, "", m)
}

func buildProcessor(conn *config.Connection, transcriptionModel, refinementModel string) (ai.Processor, error) {
	provider, err := ai.GetProvider(conn.ProviderID)
	if err != nil {
		return nil, err
	}
	apiKey := resolveAPIKey(conn.APIKey, envKeyForProvider(conn.ProviderID))
	if apiKey == "" {
		return &PlaceholderProcessor{Reason: fmt.Sprintf("API key missing for connection %q.", conn.Name)}, nil
	}
	return provider.NewProcessor(apiKey, conn.BaseURL, transcriptionModel, refinementModel)
}

// PlaceholderProcessor is used when a connection is not yet configured.
type PlaceholderProcessor struct {
	Reason string
}

func (p *PlaceholderProcessor) Transcribe(_ context.Context, _ []byte) (string, error) {
	return "", errors.New(p.Reason)
}

func (p *PlaceholderProcessor) Refine(_ context.Context, _ string, _ ai.RefineOptions) (string, error) {
	return "", errors.New(p.Reason)
}

func resolveAPIKey(configuredKey, envVar string) string {
	if configuredKey != "" {
		return configuredKey
	}
	if envVar != "" {
		return os.Getenv(envVar)
	}
	return ""
}

func envKeyForProvider(providerID string) string {
	switch providerID {
	case "groq":
		return "GROQ_API_KEY"
	default:
		return ""
	}
}

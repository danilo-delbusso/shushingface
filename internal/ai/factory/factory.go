package factory

import (
	"context"
	"errors"
	"os"

	"codeberg.org/dbus/shushingface/internal/ai"
	"codeberg.org/dbus/shushingface/internal/config"

	// Register providers — each init() calls ai.RegisterProvider.
	_ "codeberg.org/dbus/shushingface/internal/ai/groq"
)

// NewFromConfig builds a Processor using the provider registry.
func NewFromConfig(cfg *config.Settings) (ai.Processor, error) {
	provider, err := ai.GetProvider(cfg.ProviderID)
	if err != nil {
		return nil, err
	}

	apiKey := resolveAPIKey(cfg.ProviderAPIKey, envKeyForProvider(cfg.ProviderID))
	if apiKey == "" {
		return &PlaceholderProcessor{Reason: "API key is missing. Please configure it in Settings."}, nil
	}

	refinementModel := cfg.EffectiveRefinementModel()

	return provider.NewProcessor(apiKey, cfg.ProviderBaseURL, cfg.TranscriptionModel, refinementModel)
}

// PlaceholderProcessor is used when a provider is not yet fully configured.
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
	return os.Getenv(envVar)
}

func envKeyForProvider(providerID string) string {
	switch providerID {
	case "groq":
		return "GROQ_API_KEY"
	default:
		return ""
	}
}

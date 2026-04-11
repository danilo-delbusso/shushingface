package factory

import (
	"context"
	"fmt"
	"os"

	"codeberg.org/dbus/sussurro/internal/ai"
	"codeberg.org/dbus/sussurro/internal/ai/groq"
	"codeberg.org/dbus/sussurro/internal/config"
)

// Router combines multiple Processors to handle Transcription and Refinement separately.
// It implements ai.Processor, making it a drop-in replacement for the core.Engine.
type Router struct {
	transcriber ai.Processor
	refiner     ai.Processor
}

func (r *Router) Transcribe(ctx context.Context, wavData []byte) (string, error) {
	return r.transcriber.Transcribe(ctx, wavData)
}

func (r *Router) Refine(ctx context.Context, transcript, systemPrompt string) (string, error) {
	return r.refiner.Refine(ctx, transcript, systemPrompt)
}

// NewFromConfig builds the AI Dependency Graph (the Router) based on current settings.
func NewFromConfig(cfg *config.Settings) (ai.Processor, error) {
	// Resolve Transcriber
	transProvider, ok := cfg.Providers[cfg.TranscriptionProviderID]
	if !ok {
		return nil, fmt.Errorf("transcription provider '%s' not found", cfg.TranscriptionProviderID)
	}

	var transcriber ai.Processor
	var err error
	apiKey := resolveAPIKey(transProvider.APIKey, "GROQ_API_KEY")
	if apiKey == "" && transProvider.Name != "ollama" {
		transcriber = &PlaceholderProcessor{Reason: "Transcription API key is missing. Please configure it in Settings."}
	} else {
		switch transProvider.Name {
		case "groq":
			transcriber, err = groq.NewProcessor(apiKey, cfg.TranscriptionModel, "")
		default:
			return nil, fmt.Errorf("unsupported transcription provider: %s", transProvider.Name)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error initializing transcriber: %v", err)
	}

	// Resolve Refiner
	refineProvider, ok := cfg.Providers[cfg.RefinementProviderID]
	if !ok {
		return nil, fmt.Errorf("refinement provider '%s' not found", cfg.RefinementProviderID)
	}

	var refiner ai.Processor
	apiKey = resolveAPIKey(refineProvider.APIKey, "GROQ_API_KEY")
	if apiKey == "" && refineProvider.Name != "ollama" {
		refiner = &PlaceholderProcessor{Reason: "Refinement API key is missing. Please configure it in Settings."}
	} else {
		switch refineProvider.Name {
		case "groq":
			refiner, err = groq.NewProcessor(apiKey, "", cfg.RefinementModel)
		default:
			return nil, fmt.Errorf("unsupported refinement provider: %s", refineProvider.Name)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("error initializing refiner: %v", err)
	}

	return &Router{
		transcriber: transcriber,
		refiner:     refiner,
	}, nil
}

// PlaceholderProcessor is used when a provider is not yet fully configured (e.g., missing API key).
// It allows the application to start so the user can reach the Settings screen.
type PlaceholderProcessor struct {
	Reason string
}

func (p *PlaceholderProcessor) Transcribe(ctx context.Context, wavData []byte) (string, error) {
	return "", fmt.Errorf(p.Reason)
}

func (p *PlaceholderProcessor) Refine(ctx context.Context, transcript, systemPrompt string) (string, error) {
	return "", fmt.Errorf(p.Reason)
}

// resolveAPIKey falls back to environment variables for backward compatibility
func resolveAPIKey(configuredKey, envVar string) string {
	if configuredKey != "" {
		return configuredKey
	}
	return os.Getenv(envVar)
}

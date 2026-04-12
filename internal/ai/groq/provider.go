package groq

import (
	"context"
	"strings"

	"codeberg.org/dbus/shushingface/internal/ai"
	"codeberg.org/dbus/shushingface/internal/ai/openaicompat"
)

const defaultBaseURL = "https://api.groq.com/openai/v1"

func init() { ai.RegisterProvider(&groqProvider{}) }

type groqProvider struct{}

func (g *groqProvider) ID() string          { return "groq" }
func (g *groqProvider) DisplayName() string { return "Groq" }

func (g *groqProvider) ListModels(ctx context.Context, apiKey, baseURL string) ([]ai.ModelInfo, error) {
	base := baseURL
	if base == "" {
		base = defaultBaseURL
	}
	return openaicompat.ListModels(ctx, apiKey, base, groqClassifier)
}

func (g *groqProvider) NewProcessor(apiKey, baseURL, transcriptionModel, refinementModel string) (ai.Processor, error) {
	base := baseURL
	if base == "" {
		base = defaultBaseURL
	}
	return openaicompat.NewProcessor(base, apiKey, transcriptionModel, refinementModel), nil
}

// groqClassifier filters out non-inference models specific to Groq's catalog.
func groqClassifier(id string) string {
	lower := strings.ToLower(id)
	switch {
	case strings.Contains(lower, "whisper"):
		return "transcription"
	case strings.Contains(lower, "guard") || strings.Contains(lower, "safeguard"):
		return ""
	case strings.Contains(lower, "orpheus"):
		return ""
	case strings.Contains(lower, "compound"):
		return ""
	default:
		return "chat"
	}
}

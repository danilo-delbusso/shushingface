package openai

import (
	"context"

	"codeberg.org/dbus/shushingface/internal/ai"
	"codeberg.org/dbus/shushingface/internal/ai/openaicompat"
)

const defaultBaseURL = "https://api.openai.com/v1"

func init() { ai.RegisterProvider(&openaiProvider{}) }

type openaiProvider struct{}

func (o *openaiProvider) ID() string          { return "openai" }
func (o *openaiProvider) DisplayName() string { return "OpenAI" }

func (o *openaiProvider) ListModels(ctx context.Context, apiKey, baseURL string) ([]ai.ModelInfo, error) {
	base := baseURL
	if base == "" {
		base = defaultBaseURL
	}
	return openaicompat.ListModels(ctx, apiKey, base, openaicompat.DefaultClassifier)
}

func (o *openaiProvider) NewProcessor(apiKey, baseURL, transcriptionModel, refinementModel string) (ai.Processor, error) {
	base := baseURL
	if base == "" {
		base = defaultBaseURL
	}
	return openaicompat.NewProcessor(base, apiKey, transcriptionModel, refinementModel), nil
}

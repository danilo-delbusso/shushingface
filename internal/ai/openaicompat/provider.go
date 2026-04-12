package openaicompat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"codeberg.org/dbus/shushingface/internal/ai"
)

func init() { ai.RegisterProvider(&provider{}) }

type provider struct{}

func (p *provider) ID() string          { return "openai-compatible" }
func (p *provider) DisplayName() string { return "OpenAI-Compatible" }

func (p *provider) ListModels(ctx context.Context, apiKey, baseURL string) ([]ai.ModelInfo, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required for OpenAI-compatible providers")
	}
	return ListModels(ctx, apiKey, baseURL, DefaultClassifier)
}

func (p *provider) NewProcessor(apiKey, baseURL, transcriptionModel, refinementModel string) (ai.Processor, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required for OpenAI-compatible providers")
	}
	return NewProcessor(baseURL, apiKey, transcriptionModel, refinementModel), nil
}

// ──────────────────────────────────────────────────
// Shared model listing — used by all OpenAI-compatible providers
// ──────────────────────────────────────────────────

// ModelClassifier decides the category for a model ID.
// Return "" to skip the model entirely.
type ModelClassifier func(id string) string

// DefaultClassifier is a generic heuristic: whisper = transcription, rest = chat.
func DefaultClassifier(id string) string {
	lower := strings.ToLower(id)
	if strings.Contains(lower, "whisper") {
		return "transcription"
	}
	return "chat"
}

// ListModels fetches models from any OpenAI-compatible /models endpoint.
func ListModels(ctx context.Context, apiKey, baseURL string, classify ModelClassifier) ([]ai.ModelInfo, error) {
	url := strings.TrimRight(baseURL, "/") + "/models"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models API returned %d", resp.StatusCode)
	}

	var body struct {
		Data []struct {
			ID      string `json:"id"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	out := make([]ai.ModelInfo, 0, len(body.Data))
	for _, m := range body.Data {
		cat := classify(m.ID)
		if cat == "" {
			continue
		}
		out = append(out, ai.ModelInfo{
			ID:       m.ID,
			Name:     m.ID,
			Category: cat,
		})
	}
	return out, nil
}

package openaicompat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"codeberg.org/dbus/shushingface/internal/ai"
)

// httpClient is shared across all model-listing calls.
var httpClient = &http.Client{Timeout: 30 * time.Second}

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

// ModelClassifier decides the category for a model ID. Return "" to skip it.
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

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("models API returned %d: %s", resp.StatusCode, apiErrorMessage(body))
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

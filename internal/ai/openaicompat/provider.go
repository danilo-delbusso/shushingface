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
		cat := classifyModel(m.ID)
		out = append(out, ai.ModelInfo{
			ID:       m.ID,
			Name:     m.ID,
			Category: cat,
		})
	}
	return out, nil
}

// classifyModel uses heuristics on the model ID.
// Since we can't know every provider's naming scheme, we're generous —
// whisper models are transcription, everything else is chat.
func classifyModel(id string) string {
	lower := strings.ToLower(id)
	if strings.Contains(lower, "whisper") {
		return "transcription"
	}
	return "chat"
}

func (p *provider) NewProcessor(apiKey, baseURL, transcriptionModel, refinementModel string) (ai.Processor, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required for OpenAI-compatible providers")
	}
	return NewProcessor(baseURL, apiKey, transcriptionModel, refinementModel), nil
}

package groq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"codeberg.org/dbus/shushingface/internal/ai"
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
	url := strings.TrimRight(base, "/") + "/models"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("groq models API returned %d", resp.StatusCode)
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
		if cat == "" {
			continue // skip non-inference models (guards, embeddings, etc.)
		}
		out = append(out, ai.ModelInfo{
			ID:       m.ID,
			Name:     m.ID,
			Category: cat,
		})
	}
	return out, nil
}

func classifyModel(id string) string {
	lower := strings.ToLower(id)
	switch {
	case strings.Contains(lower, "whisper"):
		return "transcription"
	case strings.Contains(lower, "guard") || strings.Contains(lower, "safeguard"):
		return "" // safety classifiers, not usable
	case strings.Contains(lower, "orpheus"):
		return "" // TTS models, not usable
	case strings.Contains(lower, "compound"):
		return "" // agentic orchestration, not usable for simple refinement
	default:
		return "chat"
	}
}

func (g *groqProvider) NewProcessor(apiKey, baseURL, transcriptionModel, refinementModel string) (ai.Processor, error) {
	return NewProcessor(apiKey, transcriptionModel, refinementModel)
}

package openaicompat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"codeberg.org/dbus/shushingface/internal/ai"
)

// apiErrorMessage extracts a human-readable message from an OpenAI-style
// error response body. Falls back to the raw body if parsing fails.
func apiErrorMessage(body []byte) string {
	var parsed struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &parsed) == nil && parsed.Error.Message != "" {
		return parsed.Error.Message
	}
	if len(body) > 200 {
		return string(body[:200]) + "…"
	}
	return string(body)
}

type processor struct {
	baseURL            string
	apiKey             string
	transcriptionModel string
	refinementModel    string
	client             *http.Client
}

// NewProcessor creates a processor for any OpenAI-compatible endpoint.
func NewProcessor(baseURL, apiKey, transcriptionModel, refinementModel string) *processor {
	return &processor{
		baseURL:            strings.TrimRight(baseURL, "/"),
		apiKey:             apiKey,
		transcriptionModel: transcriptionModel,
		refinementModel:    refinementModel,
		client:             &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *processor) Transcribe(ctx context.Context, wavData []byte, opts ai.TranscribeOptions) (string, error) {
	url := p.baseURL + "/audio/transcriptions"

	// Build multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("model", p.transcriptionModel); err != nil {
		return "", err
	}
	if opts.Language != "" {
		if err := writer.WriteField("language", opts.Language); err != nil {
			return "", err
		}
	}
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", err
	}
	if _, err := part.Write(wavData); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close response body", "error", err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading transcription response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("transcription API returned %d: %s", resp.StatusCode, apiErrorMessage(respBody))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	return result.Text, nil
}

func (p *processor) Refine(ctx context.Context, transcript string, opts ai.RefineOptions) (string, error) {
	url := p.baseURL + "/chat/completions"

	systemPrompt := opts.SystemPrompt
	if opts.Context != "" {
		systemPrompt += "\n\nThe user is currently typing in: " + opts.Context
	}

	// Assemble messages: system → few-shot pairs → real transcript
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	msgs := make([]message, 0, 2+len(opts.Examples)*2)
	msgs = append(msgs, message{Role: "system", Content: systemPrompt})
	for _, ex := range opts.Examples {
		msgs = append(msgs,
			message{Role: "user", Content: ex.Input},
			message{Role: "assistant", Content: ex.Output},
		)
	}
	msgs = append(msgs, message{Role: "user", Content: transcript})

	reqBody := struct {
		Model       string    `json:"model"`
		Messages    []message `json:"messages"`
		Temperature float32   `json:"temperature,omitempty"`
		TopP        float32   `json:"top_p,omitempty"`
	}{
		Model:       p.refinementModel,
		Messages:    msgs,
		Temperature: opts.Sampling.Temperature,
		TopP:        opts.Sampling.TopP,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close response body", "error", err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading chat response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("chat API returned %d: %s", resp.StatusCode, apiErrorMessage(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("API returned empty response")
	}
	return result.Choices[0].Message.Content, nil
}

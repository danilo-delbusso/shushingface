package groq

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"codeberg.org/dbus/shushingface/internal/ai"
	groqclient "github.com/conneroisu/groq-go"
)

// Client defines the interface for the Groq API operations we need,
// allowing for easy mocking during tests.
type Client interface {
	Transcribe(ctx context.Context, request groqclient.AudioRequest) (groqclient.AudioResponse, error)
	ChatCompletion(ctx context.Context, request groqclient.ChatCompletionRequest) (groqclient.ChatCompletionResponse, error)
}

// processor implements the ai.Processor interface using the Groq API.
type processor struct {
	client             Client
	transcriptionModel string
	refinementModel    string
}

// NewProcessor creates a new Groq processor instance that satisfies the ai.Processor interface.
func NewProcessor(apiKey, transcriptionModel, refinementModel string) (ai.Processor, error) {
	client, err := groqclient.NewClient(apiKey)
	if err != nil {
		return nil, err
	}
	return &processor{
		client:             client,
		transcriptionModel: transcriptionModel,
		refinementModel:    refinementModel,
	}, nil
}

func (p *processor) Transcribe(ctx context.Context, wavData []byte) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	transReq := groqclient.AudioRequest{
		FilePath: "audio.wav",
		Reader:   bytes.NewReader(wavData),
		Model:    groqclient.AudioModel(p.transcriptionModel),
	}

	transcription, err := p.client.Transcribe(ctx, transReq)
	if err != nil {
		return "", err
	}

	return transcription.Text, nil
}

func (p *processor) Refine(ctx context.Context, transcript string, opts ai.RefineOptions) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build system prompt, optionally injecting active-app context.
	systemPrompt := opts.SystemPrompt
	if opts.Context != "" {
		systemPrompt += "\n\nThe user is currently typing in: " + opts.Context
	}

	// Assemble messages: system → few-shot pairs → real transcript.
	msgs := make([]groqclient.ChatCompletionMessage, 0, 2+len(opts.Examples)*2)
	msgs = append(msgs, groqclient.ChatCompletionMessage{
		Role:    "system",
		Content: systemPrompt,
	})
	for _, ex := range opts.Examples {
		msgs = append(msgs,
			groqclient.ChatCompletionMessage{Role: "user", Content: ex.Input},
			groqclient.ChatCompletionMessage{Role: "assistant", Content: ex.Output},
		)
	}
	msgs = append(msgs, groqclient.ChatCompletionMessage{
		Role:    "user",
		Content: transcript,
	})

	chatReq := groqclient.ChatCompletionRequest{
		Model:       groqclient.ChatModel(p.refinementModel),
		Messages:    msgs,
		Temperature: opts.Sampling.Temperature,
		TopP:        opts.Sampling.TopP,
	}

	result, err := p.client.ChatCompletion(ctx, chatReq)
	if err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("groq returned empty response")
	}
	return result.Choices[0].Message.Content, nil
}

package groq

import (
	"context"
	"os"

	"codeberg.org/dbus/sussurro/internal/ai"
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
	// Create a temporary file for the WAV
	tmpFile, err := os.CreateTemp("", "sussurro-*.wav")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(wavData); err != nil {
		return "", err
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return "", err
	}

	transReq := groqclient.AudioRequest{
		FilePath: tmpFile.Name(),
		Reader:   tmpFile,
		Model:    groqclient.AudioModel(p.transcriptionModel),
	}

	transcription, err := p.client.Transcribe(ctx, transReq)
	if err != nil {
		return "", err
	}

	return transcription.Text, nil
}

func (p *processor) Refine(ctx context.Context, transcript string, systemPrompt string) (string, error) {
	chatReq := groqclient.ChatCompletionRequest{
		Model: groqclient.ChatModel(p.refinementModel),
		Messages: []groqclient.ChatCompletionMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: transcript,
			},
		},
	}

	result, err := p.client.ChatCompletion(ctx, chatReq)
	if err != nil {
		return "", err
	}

	return result.Choices[0].Message.Content, nil
}

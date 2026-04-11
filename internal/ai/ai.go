package ai

import "context"

// Processor defines the interface for an AI backend that can
// transcribe audio and refine the resulting text.
type Processor interface {
	// Transcribe takes raw WAV audio data and returns the transcribed text.
	Transcribe(ctx context.Context, wavData []byte) (transcript string, err error)
	// Refine takes a raw transcript and a system prompt, and returns the reworked text.
	Refine(ctx context.Context, transcript string, systemPrompt string) (refined string, err error)
}

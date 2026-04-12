package ai

import "context"

// FewShotPair is a single before/after example that demonstrates the
// expected transformation style to the model.
type FewShotPair struct {
	Input  string
	Output string
}

// SamplingParams controls generation behavior. Providers map these to
// their own API fields — not all providers support every parameter.
// Zero values mean "use provider default".
type SamplingParams struct {
	Temperature float32
	TopP        float32
}

// RefineOptions bundles everything the refiner needs beyond the raw transcript.
type RefineOptions struct {
	SystemPrompt string
	Examples     []FewShotPair  // rendered as user/assistant turns before the real transcript
	Context      string         // e.g. "Slack" or "Gmail" — the app the user is typing into
	Sampling     SamplingParams // generation parameters
}

// Processor defines the interface for an AI backend that can
// transcribe audio and refine the resulting text.
type Processor interface {
	// Transcribe takes raw WAV audio data and returns the transcribed text.
	Transcribe(ctx context.Context, wavData []byte) (transcript string, err error)
	// Refine takes a raw transcript and refinement options, and returns the reworked text.
	Refine(ctx context.Context, transcript string, opts RefineOptions) (refined string, err error)
}

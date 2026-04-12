package ai

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// ──────────────────────────────────────────────────
// Model & sampling types
// ──────────────────────────────────────────────────

// ModelInfo describes a single model available from a provider.
type ModelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`     // human-readable (may equal ID)
	Category string `json:"category"` // "transcription" or "chat"
}

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

// TranscribeOptions holds parameters for transcription.
type TranscribeOptions struct {
	// Language hint in ISO 639-1 format (e.g. "en", "fr").
	// Empty means auto-detect. When multiple are set, the first is used
	// as the primary hint (Whisper only accepts one).
	Language string
}

// ──────────────────────────────────────────────────
// Transcriber / Refiner / Processor interfaces
// ──────────────────────────────────────────────────

// Transcriber can transcribe audio to text.
type Transcriber interface {
	Transcribe(ctx context.Context, wavData []byte, opts TranscribeOptions) (transcript string, err error)
}

// Refiner can refine a transcript into polished text.
type Refiner interface {
	Refine(ctx context.Context, transcript string, opts RefineOptions) (refined string, err error)
}

// Processor combines both capabilities. Provider implementations satisfy this.
type Processor interface {
	Transcriber
	Refiner
}

// ──────────────────────────────────────────────────
// Provider — factory + model listing per AI service
// ──────────────────────────────────────────────────

// Provider is the top-level abstraction for an AI service (Groq, OpenAI, etc.).
type Provider interface {
	// ID returns the unique identifier (e.g. "groq").
	ID() string
	// DisplayName returns a human-readable name (e.g. "Groq").
	DisplayName() string
	// ListModels queries the provider API and returns available models.
	ListModels(ctx context.Context, apiKey, baseURL string) ([]ModelInfo, error)
	// NewProcessor creates a Processor for the given model selections.
	NewProcessor(apiKey, baseURL, transcriptionModel, refinementModel string) (Processor, error)
}

// ──────────────────────────────────────────────────
// Provider registry
// ──────────────────────────────────────────────────

var (
	registryMu sync.RWMutex
	registry   = map[string]Provider{}
)

// RegisterProvider adds a provider to the global registry.
func RegisterProvider(p Provider) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[p.ID()] = p
}

// GetProvider returns a registered provider by ID.
func GetProvider(id string) (Provider, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	p, ok := registry[id]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", id)
	}
	return p, nil
}

// ProviderInfo is a serialisable summary of a registered provider.
type ProviderInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// ListProviders returns all registered providers, sorted by ID.
func ListProviders() []ProviderInfo {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]ProviderInfo, 0, len(registry))
	for _, p := range registry {
		out = append(out, ProviderInfo{ID: p.ID(), DisplayName: p.DisplayName()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

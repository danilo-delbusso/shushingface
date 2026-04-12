package ai

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type ModelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`     // human-readable (may equal ID)
	Category string `json:"category"` // "transcription" or "chat"
}

type FewShotPair struct {
	Input  string
	Output string
}

// Zero values mean "use provider default".
type SamplingParams struct {
	Temperature float32
	TopP        float32
}

type RefineOptions struct {
	SystemPrompt string
	Examples     []FewShotPair  // rendered as user/assistant turns before the real transcript
	Context      string         // e.g. "Slack" or "Gmail" — the app the user is typing into
	Sampling     SamplingParams // generation parameters
}

type TranscribeOptions struct {
	Language string // ISO 639-1 hint; empty = auto-detect
}

type Transcriber interface {
	Transcribe(ctx context.Context, wavData []byte, opts TranscribeOptions) (transcript string, err error)
}

type Refiner interface {
	Refine(ctx context.Context, transcript string, opts RefineOptions) (refined string, err error)
}

type Processor interface {
	Transcriber
	Refiner
}

type Provider interface {
	ID() string
	DisplayName() string
	ListModels(ctx context.Context, apiKey, baseURL string) ([]ModelInfo, error)
	NewProcessor(apiKey, baseURL, transcriptionModel, refinementModel string) (Processor, error)
}

var (
	registryMu sync.RWMutex
	registry   = map[string]Provider{}
)

func RegisterProvider(p Provider) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[p.ID()] = p
}

func GetProvider(id string) (Provider, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	p, ok := registry[id]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", id)
	}
	return p, nil
}

type ProviderInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

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

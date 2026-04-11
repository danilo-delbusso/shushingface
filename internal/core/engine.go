package core

import (
	"context"
	"sync"

	"codeberg.org/dbus/sussurro/internal/ai"
	"codeberg.org/dbus/sussurro/internal/audio"
	"codeberg.org/dbus/sussurro/internal/audio/wav"
)

// Engine is the central orchestrator of the Sussurro application.
type Engine struct {
	mu           sync.RWMutex
	recorder     audio.Recorder
	processor    ai.Processor
	systemPrompt string
}

// NewEngine creates a new Sussurro Engine.
func NewEngine(recorder audio.Recorder, processor ai.Processor, systemPrompt string) *Engine {
	return &Engine{
		recorder:     recorder,
		processor:    processor,
		systemPrompt: systemPrompt,
	}
}

// SetProcessor updates the AI backend implementation at runtime.
func (e *Engine) SetProcessor(processor ai.Processor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.processor = processor
}

// GetProcessor returns the current AI processor (thread-safe snapshot).
func (e *Engine) GetProcessor() ai.Processor {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.processor
}

// SetSystemPrompt updates the refinement prompt at runtime.
func (e *Engine) SetSystemPrompt(prompt string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.systemPrompt = prompt
}

// StartRecording signals the underlying audio device to begin capturing audio.
func (e *Engine) StartRecording() error {
	return e.recorder.Start()
}

// StopAndProcess stops the current recording, encodes the audio to WAV,
// transcribes it, and then refines the transcript.
func (e *Engine) StopAndProcess(ctx context.Context) (transcript string, refined string, err error) {
	samples, err := e.recorder.Stop()
	if err != nil {
		return "", "", err
	}

	wavData, err := wav.Encode(samples, 16000)
	if err != nil {
		return "", "", err
	}

	e.mu.RLock()
	proc := e.processor
	prompt := e.systemPrompt
	e.mu.RUnlock()

	transcript, err = proc.Transcribe(ctx, wavData)
	if err != nil {
		return "", "", err
	}

	if transcript == "" {
		return "", "", nil
	}

	refined, err = proc.Refine(ctx, transcript, prompt)
	if err != nil {
		return transcript, "", err
	}

	return transcript, refined, nil
}

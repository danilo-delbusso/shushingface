package core

import (
	"context"
	"sync"

	"codeberg.org/dbus/shushingface/internal/ai"
	"codeberg.org/dbus/shushingface/internal/audio"
	"codeberg.org/dbus/shushingface/internal/audio/wav"
)

type Engine struct {
	mu          sync.RWMutex
	recorder    audio.Recorder
	transcriber ai.Transcriber
	refiner     ai.Refiner
}

func NewEngine(recorder audio.Recorder, transcriber ai.Transcriber, refiner ai.Refiner) *Engine {
	return &Engine{
		recorder:    recorder,
		transcriber: transcriber,
		refiner:     refiner,
	}
}

func (e *Engine) SetTranscriber(t ai.Transcriber) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.transcriber = t
}

func (e *Engine) SetRefiner(r ai.Refiner) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.refiner = r
}

func (e *Engine) GetRefiner() ai.Refiner {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.refiner
}

func (e *Engine) StartRecording() error {
	return e.recorder.Start()
}

// StopAndProcess stops recording, transcribes, and refines.
// If refinerOverride is non-nil it is used instead of the default refiner
// (for per-profile connection overrides).
func (e *Engine) StopAndProcess(ctx context.Context, opts ai.RefineOptions, refinerOverride ai.Refiner) (transcript string, refined string, err error) {
	samples, err := e.recorder.Stop()
	if err != nil {
		return "", "", err
	}

	wavData, err := wav.Encode(samples, 16000)
	if err != nil {
		return "", "", err
	}

	e.mu.RLock()
	t := e.transcriber
	r := e.refiner
	e.mu.RUnlock()

	if refinerOverride != nil {
		r = refinerOverride
	}

	transcript, err = t.Transcribe(ctx, wavData)
	if err != nil {
		return "", "", err
	}

	if transcript == "" {
		return "", "", nil
	}

	refined, err = r.Refine(ctx, transcript, opts)
	if err != nil {
		return transcript, "", err
	}

	return transcript, refined, nil
}

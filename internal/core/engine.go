package core

import (
	"context"
	"sync"

	"codeberg.org/dbus/shushingface/internal/ai"
	"codeberg.org/dbus/shushingface/internal/audio"
	"codeberg.org/dbus/shushingface/internal/audio/wav"
)

type Engine struct {
	mu        sync.RWMutex
	recorder  audio.Recorder
	processor ai.Processor
}

func NewEngine(recorder audio.Recorder, processor ai.Processor) *Engine {
	return &Engine{
		recorder:  recorder,
		processor: processor,
	}
}

func (e *Engine) SetProcessor(processor ai.Processor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.processor = processor
}

func (e *Engine) GetProcessor() ai.Processor {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.processor
}

func (e *Engine) StartRecording() error {
	return e.recorder.Start()
}

// StopAndProcess stops recording, transcribes, and refines with the given options.
func (e *Engine) StopAndProcess(ctx context.Context, opts ai.RefineOptions) (transcript string, refined string, err error) {
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
	e.mu.RUnlock()

	transcript, err = proc.Transcribe(ctx, wavData)
	if err != nil {
		return "", "", err
	}

	if transcript == "" {
		return "", "", nil
	}

	refined, err = proc.Refine(ctx, transcript, opts)
	if err != nil {
		return transcript, "", err
	}

	return transcript, refined, nil
}

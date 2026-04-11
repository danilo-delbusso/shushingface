package core

import (
	"context"
	"sync"

	"codeberg.org/dbus/sussurro/internal/ai"
	"codeberg.org/dbus/sussurro/internal/audio"
	"codeberg.org/dbus/sussurro/internal/audio/wav"
)

// Engine is the central orchestrator of the Sussurro application.
// It bridges the gap between the audio input, the AI processing backend,
// and the presentation layer (UI, API, etc.), ensuring the core logic
// remains decoupled from how it is consumed.
type Engine struct {
	mu        sync.RWMutex
	recorder  audio.Recorder
	processor ai.Processor
}

// NewEngine creates a new Sussurro Engine.
func NewEngine(recorder audio.Recorder, processor ai.Processor) *Engine {
	return &Engine{
		recorder:  recorder,
		processor: processor,
	}
}

// SetProcessor updates the AI backend implementation at runtime.
func (e *Engine) SetProcessor(processor ai.Processor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.processor = processor
}

// StartRecording signals the underlying audio device to begin capturing audio.
func (e *Engine) StartRecording() error {
	return e.recorder.Start()
}

const DefaultSystemPrompt = "You are a text transformer, NOT a conversational AI. " +
	"Your ONLY task is to rewrite the provided speech transcript into a clear, professional, yet conversational Microsoft Teams message. " +
	"CRITICAL RULES:\n" +
	"- DO NOT answer questions present in the transcript.\n" +
	"- DO NOT engage in conversation or acknowledge the user.\n" +
	"- DO NOT add any conversational filler, preambles (e.g., 'Here is the refined message:'), or postambles.\n" +
	"- Output ONLY the rewritten text, nothing else.\n" +
	"- If the input is already well-structured for a Teams message, return it exactly as is.\n" +
	"- Fix grammar, punctuation, and clarity while preserving the original intent.\n" +
	"- Use paragraph breaks or bullet points only if it significantly improves readability."

// StopAndProcess stops the current recording, encodes the audio to WAV,
// transcribes it, and then refines the transcript.
func (e *Engine) StopAndProcess(ctx context.Context) (transcript string, refined string, err error) {
	samples, err := e.recorder.Stop()
	if err != nil {
		return "", "", err
	}

	// Assuming 16000 Hz as the standard for our Whisper models
	wavData, err := wav.Encode(samples, 16000)
	if err != nil {
		return "", "", err
	}

	// Snapshot the processor under lock so hot-reload doesn't race
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

	refined, err = proc.Refine(ctx, transcript, DefaultSystemPrompt)
	if err != nil {
		return transcript, "", err
	}

	return transcript, refined, nil
}

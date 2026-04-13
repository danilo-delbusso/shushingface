package core

import (
	"context"
	"log/slog"
	"math"
	"strings"
	"sync"
	"time"

	"codeberg.org/dbus/shushingface/internal/ai"
	"codeberg.org/dbus/shushingface/internal/audio"
	"codeberg.org/dbus/shushingface/internal/audio/wav"
)

// sampleRate is the fixed capture rate for malgo recording.
const sampleRate = 16000

// Minimum clip duration and minimum RMS amplitude that must be present
// before we bother calling the transcription API. Whisper is trained on
// YouTube captions and hallucinates things like "Thank you." / "Thanks
// for watching!" when fed silence — cheaper + more correct to short-
// circuit here than to filter those strings after the fact.
const (
	minClipSamples = sampleRate / 4 // 250ms
	// 16-bit PCM RMS threshold. ~60 is quiet-room noise floor on most
	// mics; below this we assume the user didn't actually speak.
	minRMSAmplitude = 60.0
)

// Engine is the interface for the recording → transcription → refinement pipeline.
type Engine interface {
	StartRecording() error
	StopAndProcess(ctx context.Context, tOpts ai.TranscribeOptions, rOpts ai.RefineOptions, refinerOverride ai.Refiner) (transcript string, refined string, err error)
	SetTranscriber(t ai.Transcriber)
	SetRefiner(r ai.Refiner)
	GetRefiner() ai.Refiner
}

type engine struct {
	mu          sync.RWMutex
	recorder    audio.Recorder
	transcriber ai.Transcriber
	refiner     ai.Refiner
}

func NewEngine(recorder audio.Recorder, transcriber ai.Transcriber, refiner ai.Refiner) Engine {
	return &engine{
		recorder:    recorder,
		transcriber: transcriber,
		refiner:     refiner,
	}
}

func (e *engine) SetTranscriber(t ai.Transcriber) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.transcriber = t
}

func (e *engine) SetRefiner(r ai.Refiner) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.refiner = r
}

func (e *engine) GetRefiner() ai.Refiner {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.refiner
}

func (e *engine) StartRecording() error {
	return e.recorder.Start()
}

// StopAndProcess stops recording, transcribes, and refines.
// If refinerOverride is non-nil it is used instead of the default refiner.
func (e *engine) StopAndProcess(ctx context.Context, tOpts ai.TranscribeOptions, rOpts ai.RefineOptions, refinerOverride ai.Refiner) (transcript string, refined string, err error) {
	slog.Debug("engine.StopAndProcess: entry")

	t0 := time.Now()
	samples, err := e.recorder.Stop()
	slog.Debug("engine.StopAndProcess: recorder.Stop returned",
		"elapsed", time.Since(t0), "samples", len(samples), "error", err)
	if err != nil {
		return "", "", err
	}

	// Skip the API round trip entirely for no-speech clips:
	//   - empty: user hit start/stop back-to-back with no recording
	//   - too short: likely an accidental tap
	//   - too quiet: the mic captured background only
	// Whisper hallucinates "Thank you." / "Thanks for watching!" on
	// silence, so we must catch this before the transcription call.
	if len(samples) < minClipSamples {
		slog.Info("engine.StopAndProcess: clip too short, skipping transcription",
			"samples", len(samples), "threshold", minClipSamples)
		return "", "", nil
	}
	if rms := rmsInt16(samples); rms < minRMSAmplitude {
		slog.Info("engine.StopAndProcess: clip too quiet, skipping transcription",
			"rms", rms, "threshold", minRMSAmplitude)
		return "", "", nil
	}

	t0 = time.Now()
	wavData, err := wav.Encode(samples, sampleRate)
	slog.Debug("engine.StopAndProcess: wav.Encode returned",
		"elapsed", time.Since(t0), "bytes", len(wavData), "error", err)
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

	slog.Debug("engine.StopAndProcess: calling transcriber.Transcribe",
		"language", tOpts.Language, "bytes", len(wavData))
	t0 = time.Now()
	transcript, err = t.Transcribe(ctx, wavData, tOpts)
	slog.Debug("engine.StopAndProcess: transcriber.Transcribe returned",
		"elapsed", time.Since(t0), "chars", len(transcript), "error", err)
	if err != nil {
		return "", "", err
	}

	if transcript == "" {
		slog.Debug("engine.StopAndProcess: empty transcript, skipping refine")
		return "", "", nil
	}

	// Drop known Whisper hallucinations so they never surface to the user
	// or hit the refiner. These only match short transcripts — a legitimate
	// utterance containing "you" as a word is not affected.
	if isWhisperHallucination(transcript) {
		slog.Info("engine.StopAndProcess: dropping Whisper hallucination",
			"transcript", transcript)
		return "", "", nil
	}

	slog.Debug("engine.StopAndProcess: calling refiner.Refine", "chars", len(transcript))
	t0 = time.Now()
	refined, err = r.Refine(ctx, transcript, rOpts)
	slog.Debug("engine.StopAndProcess: refiner.Refine returned",
		"elapsed", time.Since(t0), "chars", len(refined), "error", err)
	if err != nil {
		return transcript, "", err
	}

	return transcript, refined, nil
}

// rmsInt16 computes the root-mean-square amplitude of a 16-bit PCM buffer.
// Used as a quick "did the user actually speak?" check before hitting the
// transcription API.
func rmsInt16(samples []int16) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sumSq float64
	for _, s := range samples {
		v := float64(s)
		sumSq += v * v
	}
	return math.Sqrt(sumSq / float64(len(samples)))
}

// whisperHallucinations are short phrases Whisper emits when the audio is
// silence or near-silence — artefacts of its YouTube-captions training
// data. Comparisons are case-insensitive and ignore trailing punctuation /
// whitespace, so "Thank you." and " thank you " both match.
var whisperHallucinations = map[string]struct{}{
	"":                      {},
	"you":                   {},
	"thank you":             {},
	"thanks":                {},
	"thanks for watching":   {},
	"thank you for watching": {},
	"please subscribe":      {},
	"subscribe":             {},
	"bye":                   {},
	"goodbye":               {},
	"okay":                  {},
	"ok":                    {},
	"hmm":                   {},
	"um":                    {},
	"uh":                    {},
}

// isWhisperHallucination returns true when transcript looks like one of
// the fixed phrases Whisper emits for silent input. We only treat short
// transcripts this way — a real utterance of "you" inside a longer
// sentence is fine.
func isWhisperHallucination(transcript string) bool {
	t := strings.ToLower(strings.TrimSpace(transcript))
	t = strings.TrimRight(t, ".!?,;: ")
	if len(t) > 30 {
		return false
	}
	_, ok := whisperHallucinations[t]
	return ok
}

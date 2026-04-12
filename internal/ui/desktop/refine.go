package desktop

import (
	"codeberg.org/dbus/shushingface/internal/ai"
	"codeberg.org/dbus/shushingface/internal/config"
)

// cfg must be a snapshot (not the shared pointer).
func (a *App) buildRefineOptions(cfg *config.Settings, activeApp string) ai.RefineOptions {
	opts := ai.RefineOptions{}

	if p := cfg.ActiveProfile(); p != nil {
		opts.SystemPrompt = p.Prompt
		opts.SystemPrompt += "\n\nRules:\n" + cfg.GetBuiltInRules()
		if cfg.GlobalRules != "" {
			opts.SystemPrompt += "\n\nUser rules (always apply):\n" + cfg.GlobalRules
		}

		opts.Sampling = ai.SamplingParams{
			Temperature: p.Temperature,
			TopP:        p.TopP,
		}
		for _, ex := range p.Examples {
			opts.Examples = append(opts.Examples, ai.FewShotPair{
				Input:  ex.Input,
				Output: ex.Output,
			})
		}
	}

	if activeApp != "" {
		opts.Context = activeApp
	}

	if a.history != nil {
		if records, err := a.history.GetHistory(2, 0); err == nil {
			for _, r := range records {
				if r.RawTranscript != "" && r.RefinedMessage != "" {
					opts.Examples = append(opts.Examples, ai.FewShotPair{
						Input:  r.RawTranscript,
						Output: r.RefinedMessage,
					})
				}
			}
		}
	}

	return opts
}

func (a *App) TestPrompt(sampleText, systemPrompt string) ProcessResult {
	a.mu.RLock()
	cfg := a.snapshotConfig()
	a.mu.RUnlock()

	refiner := a.engine.GetRefiner()
	prompt := systemPrompt
	prompt += "\n\nRules:\n" + cfg.GetBuiltInRules()
	if cfg.GlobalRules != "" {
		prompt += "\n\nUser rules (always apply):\n" + cfg.GlobalRules
	}
	opts := ai.RefineOptions{SystemPrompt: prompt}
	refined, err := refiner.Refine(a.ctx, sampleText, opts)
	if err != nil {
		return ProcessResult{Error: err.Error()}
	}
	return ProcessResult{Transcript: sampleText, Refined: refined}
}

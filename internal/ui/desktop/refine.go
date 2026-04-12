package desktop

import (
	"codeberg.org/dbus/shushingface/internal/ai"
)

// buildRefineOptions assembles RefineOptions from the active profile,
// active app context, and recent history.
func (a *App) buildRefineOptions(activeApp string) ai.RefineOptions {
	opts := ai.RefineOptions{}

	if p := a.cfg.ActiveProfile(); p != nil {
		opts.SystemPrompt = p.Prompt
		if a.cfg.GlobalRules != "" {
			opts.SystemPrompt += "\n\nGlobal rules (always apply):\n" + a.cfg.GlobalRules
		}
		opts.Sampling = ai.SamplingParams{
			Temperature: p.Temperature,
			TopP:        p.TopP,
		}
		// Static few-shot examples from the profile.
		for _, ex := range p.Examples {
			opts.Examples = append(opts.Examples, ai.FewShotPair{
				Input:  ex.Input,
				Output: ex.Output,
			})
		}
	}

	// Inject active app context when available.
	if activeApp != "" {
		opts.Context = activeApp
	}

	// Append recent history as dynamic few-shot examples so the model
	// calibrates to the user's personal style over time.
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
	proc := a.engine.GetProcessor()
	prompt := systemPrompt
	if a.cfg.GlobalRules != "" {
		prompt += "\n\nGlobal rules (always apply):\n" + a.cfg.GlobalRules
	}
	opts := ai.RefineOptions{SystemPrompt: prompt}
	refined, err := proc.Refine(a.ctx, sampleText, opts)
	if err != nil {
		return ProcessResult{Error: err.Error()}
	}
	return ProcessResult{Transcript: sampleText, Refined: refined}
}

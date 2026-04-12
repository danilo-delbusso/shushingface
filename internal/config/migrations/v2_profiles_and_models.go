package migrations

// V2ProfilesAndModels ensures refinement profiles and default models exist.
func V2ProfilesAndModels(data map[string]any) error {
	// Create default profiles if none exist
	profiles, _ := data["refinementProfiles"].([]any)
	if len(profiles) == 0 {
		customPrompt := strOr(data, "systemPrompt", "")
		profileMaps := defaultProfileMaps()
		if customPrompt != "" {
			profileMaps = append(profileMaps, map[string]any{
				"id":     "custom",
				"name":   "Custom",
				"icon":   "pen-tool",
				"prompt": customPrompt,
			})
			data["activeProfileId"] = "custom"
		} else if strOr(data, "activeProfileId", "") == "" {
			data["activeProfileId"] = "professional"
		}
		data["refinementProfiles"] = profileMaps
		delete(data, "systemPrompt")
	}

	// Ensure default models
	if strOr(data, "refinementModel", "") == "" {
		data["refinementModel"] = "meta-llama/llama-4-scout-17b-16e-instruct"
	}
	if strOr(data, "transcriptionModel", "") == "" {
		data["transcriptionModel"] = "whisper-large-v3"
	}

	return nil
}

func defaultProfileMaps() []any {
	return []any{
		map[string]any{
			"id":          "casual",
			"name":        "Casual",
			"icon":        "coffee",
			"prompt":      "You are a speech-to-text editor. Rewrite the transcript so it reads like something the speaker would actually type — relaxed, natural, the way you'd message a colleague you're comfortable with. Keep contractions, casual phrasing, and personality.",
			"temperature": 0.4,
			"topP":        0.9,
		},
		map[string]any{
			"id":          "professional",
			"name":        "Professional",
			"icon":        "briefcase",
			"prompt":      "You are a speech-to-text editor. Rewrite the transcript into clear, professional text suitable for emails and workplace communication. Use complete sentences and precise language, but keep it human — avoid corporate jargon and stiff phrasing that nobody would actually write.",
			"temperature": 0.3,
			"topP":        0.9,
		},
		map[string]any{
			"id":          "concise",
			"name":        "Concise",
			"icon":        "zap",
			"prompt":      "You are a speech-to-text editor. Compress the transcript to its essential meaning. Strip filler, hedging, repetition, and unnecessary detail. One to two sentences. Every word earns its place.",
			"temperature": 0.2,
			"topP":        0.9,
		},
	}
}

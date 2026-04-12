package config

import (
	"fmt"
	"log/slog"
)

// migration transforms raw config JSON from one version to the next.
// Operating on map[string]any avoids coupling to the current struct shape.
type migration struct {
	version     int
	description string
	up          func(data map[string]any) error
}

// currentConfigVersion is the version after all migrations have run.
// Increment this when adding a new migration.
const currentConfigVersion = 2

var migrations = []migration{
	{
		version:     1,
		description: "legacy providers to connections",
		up:          migrateV1LegacyToConnections,
	},
	{
		version:     2,
		description: "ensure profiles and default models",
		up:          migrateV2ProfilesAndModels,
	},
}

// migrateConfig runs all pending migrations on raw JSON data.
// Returns true if any migrations were applied.
func migrateConfig(data map[string]any) (bool, error) {
	v, _ := data["configVersion"].(float64)
	current := int(v)

	if current > currentConfigVersion {
		return false, fmt.Errorf(
			"config version %d is newer than this app supports (%d); please update shushingface",
			current, currentConfigVersion,
		)
	}

	if current == currentConfigVersion {
		return false, nil
	}

	for _, m := range migrations {
		if m.version > current {
			slog.Info("running config migration", "version", m.version, "description", m.description)
			if err := m.up(data); err != nil {
				return false, fmt.Errorf("config migration v%d (%s): %w", m.version, m.description, err)
			}
			data["configVersion"] = float64(m.version)
		}
	}

	return true, nil
}

// ──────────────────────────────────────────────────
// Migration V1: legacy provider formats → connections slice
// ──────────────────────────────────────────────────

func migrateV1LegacyToConnections(data map[string]any) error {
	// Already has connections? Nothing to do.
	if conns, ok := data["connections"].([]any); ok && len(conns) > 0 {
		return nil
	}

	// Step 1: old multi-provider map → flat provider fields
	if providers, ok := data["providers"].(map[string]any); ok && len(providers) > 0 {
		provID := strOr(data, "transcriptionProviderId", "")
		if provID == "" {
			provID = strOr(data, "refinementProviderId", "")
		}
		if old, ok := providers[provID].(map[string]any); ok {
			name := strOr(old, "name", "groq")
			if name == "" {
				name = "groq"
			}
			data["providerId"] = name
			data["providerApiKey"] = strOr(old, "apiKey", "")
			data["providerBaseUrl"] = strOr(old, "baseUrl", "")
		}
		delete(data, "providers")
		delete(data, "transcriptionProviderId")
		delete(data, "refinementProviderId")
	}

	// Step 2: flat provider fields → connections slice
	provID := strOr(data, "providerId", "")
	if provID != "" {
		conn := map[string]any{
			"id":         "default",
			"name":       providerDisplayName(provID),
			"providerId": provID,
			"apiKey":     strOr(data, "providerApiKey", ""),
			"baseUrl":    strOr(data, "providerBaseUrl", ""),
		}
		data["connections"] = []any{conn}
		data["transcriptionConnectionId"] = "default"
		data["refinementConnectionId"] = "default"
		delete(data, "providerId")
		delete(data, "providerApiKey")
		delete(data, "providerBaseUrl")
	}

	return nil
}

// ──────────────────────────────────────────────────
// Migration V2: ensure profiles + default models
// ──────────────────────────────────────────────────

func migrateV2ProfilesAndModels(data map[string]any) error {
	// Create default profiles if none exist
	profiles, _ := data["refinementProfiles"].([]any)
	if len(profiles) == 0 {
		customPrompt := strOr(data, "systemPrompt", "")
		defaults := DefaultProfiles()
		var profileMaps []any
		for _, p := range defaults {
			profileMaps = append(profileMaps, profileToMap(p))
		}
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
		data["refinementModel"] = DefaultRefinementModel
	}
	if strOr(data, "transcriptionModel", "") == "" {
		data["transcriptionModel"] = DefaultTranscriptionModel
	}

	return nil
}

// ──────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────

func strOr(m map[string]any, key, fallback string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return fallback
}

func profileToMap(p RefinementProfile) map[string]any {
	m := map[string]any{
		"id":     p.ID,
		"name":   p.Name,
		"icon":   p.Icon,
		"prompt": p.Prompt,
	}
	if p.Temperature != 0 {
		m["temperature"] = p.Temperature
	}
	if p.TopP != 0 {
		m["topP"] = p.TopP
	}
	if len(p.Examples) > 0 {
		var exs []any
		for _, ex := range p.Examples {
			exs = append(exs, map[string]any{"input": ex.Input, "output": ex.Output})
		}
		m["examples"] = exs
	}
	return m
}

package config

import (
	"fmt"
	"log/slog"
)

// currentConfigVersion is the version of the current config schema.
// Bump this and append to configMigrations when the schema changes.
const currentConfigVersion = 1

type configMigration struct {
	version     int
	description string
	up          func(data map[string]any) error
}

// Config migrations. v1 is the collapsed baseline — every field the app
// currently cares about is populated here. Add further entries (v2, v3…)
// when the schema evolves; each one operates on raw JSON so it stays
// decoupled from the current Settings struct.
var configMigrations = []configMigration{
	{version: 1, description: "initial schema", up: migrateInitial},
}

// migrateConfig runs all pending migrations on raw JSON data.
func migrateConfig(data map[string]any) (bool, error) {
	v, _ := data["configVersion"].(float64)
	current := int(v)

	if current >= currentConfigVersion {
		// Configs produced by a future version get accepted at face value
		// but we still clamp the version number down so the running binary
		// doesn't keep logging the mismatch.
		if current > currentConfigVersion {
			slog.Warn("config version is newer than expected, accepting as-is",
				"config", current, "app", currentConfigVersion)
			data["configVersion"] = float64(currentConfigVersion)
			return true, nil
		}
		return false, nil
	}

	for _, m := range configMigrations {
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

// migrateInitial installs a full baseline schema. It fills any field that
// DefaultSettings cares about and scrubs legacy fields from earlier
// pre-baseline configs. Since this is the only migration today, every
// upgrade path ends here.
func migrateInitial(data map[string]any) error {
	if _, ok := data["connections"]; !ok {
		data["connections"] = []any{}
	}

	if v, _ := data["refinementModel"].(string); v == "" {
		data["refinementModel"] = DefaultRefinementModel
	}
	if v, _ := data["transcriptionModel"].(string); v == "" {
		data["transcriptionModel"] = DefaultTranscriptionModel
	}

	profiles, _ := data["refinementProfiles"].([]any)
	if len(profiles) == 0 {
		defaults := DefaultProfiles()
		var profileMaps []any
		for _, p := range defaults {
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
			profileMaps = append(profileMaps, m)
		}
		data["refinementProfiles"] = profileMaps
	}
	if v, _ := data["activeProfileId"].(string); v == "" {
		data["activeProfileId"] = "professional"
	}

	if _, ok := data["shortcut"]; !ok {
		data["shortcut"] = ""
	}
	if _, ok := data["recordingMode"]; !ok {
		data["recordingMode"] = "toggle"
	}
	if _, ok := data["overlayEnabled"]; !ok {
		data["overlayEnabled"] = true
	}
	if _, ok := data["overlayOpacity"]; !ok {
		data["overlayOpacity"] = 0.4
	}
	if _, ok := data["debugLogging"]; !ok {
		data["debugLogging"] = false
	}

	for _, key := range []string{
		"providers", "providerId", "providerApiKey", "providerBaseUrl",
		"transcriptionProviderId", "refinementProviderId", "systemPrompt",
	} {
		delete(data, key)
	}

	return nil
}

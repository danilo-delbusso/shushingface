package config

import (
	"fmt"
	"log/slog"
)

// currentConfigVersion is the version of the current config schema.
// Bump this when adding a new migration.
const currentConfigVersion = 3

type configMigration struct {
	version     int
	description string
	up          func(data map[string]any) error
}

// Config migrations. Add new entries when the schema changes.
// Each migration operates on raw JSON (map[string]any) so it's
// decoupled from the current Settings struct.
var configMigrations = []configMigration{
	{version: 1, description: "initial schema", up: migrateInitial},
	{version: 2, description: "add shortcut field", up: migrateAddShortcut},
	{version: 3, description: "add recording mode and overlay", up: migrateAddOverlay},
}

func migrateAddShortcut(data map[string]any) error {
	if _, ok := data["shortcut"]; !ok {
		data["shortcut"] = ""
	}
	return nil
}

func migrateAddOverlay(data map[string]any) error {
	if _, ok := data["recordingMode"]; !ok {
		data["recordingMode"] = "toggle"
	}
	if _, ok := data["overlayEnabled"]; !ok {
		data["overlayEnabled"] = true
	}
	if _, ok := data["overlayOpacity"]; !ok {
		data["overlayOpacity"] = 0.4
	}
	return nil
}

// migrateConfig runs all pending migrations on raw JSON data.
func migrateConfig(data map[string]any) (bool, error) {
	v, _ := data["configVersion"].(float64)
	current := int(v)

	if current >= currentConfigVersion {
		// Already at or beyond current version — nothing to do.
		// During development, config versions may go backwards when
		// migrations are collapsed. Accept it gracefully.
		if current > currentConfigVersion {
			slog.Warn("config version is newer than expected, accepting as-is",
				"config", current, "app", currentConfigVersion)
			data["configVersion"] = float64(currentConfigVersion)
			return true, nil // save to downgrade the version number
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

// migrateInitial ensures a valid initial config structure.
// Handles fresh configs and any legacy formats.
func migrateInitial(data map[string]any) error {
	// Ensure connections exist
	if _, ok := data["connections"]; !ok {
		data["connections"] = []any{}
	}

	// Ensure default models
	if v, _ := data["refinementModel"].(string); v == "" {
		data["refinementModel"] = DefaultRefinementModel
	}
	if v, _ := data["transcriptionModel"].(string); v == "" {
		data["transcriptionModel"] = DefaultTranscriptionModel
	}

	// Ensure profiles exist
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

	// Clean up any legacy fields
	for _, key := range []string{
		"providers", "providerId", "providerApiKey", "providerBaseUrl",
		"transcriptionProviderId", "refinementProviderId", "systemPrompt",
	} {
		delete(data, key)
	}

	return nil
}

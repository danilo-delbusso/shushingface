package config

import (
	"fmt"
	"log/slog"

	"codeberg.org/dbus/shushingface/internal/config/migrations"
)

// migration transforms raw config JSON from one version to the next.
type migration struct {
	version     int
	description string
	up          func(data map[string]any) error
}

// currentConfigVersion is the version after all migrations have run.
// Increment this when adding a new migration.
const currentConfigVersion = 2

// Registry of config migrations. Each entry corresponds to a file in
// internal/config/migrations/. Add new migrations by:
// 1. Creating a new file in migrations/ with the migration function
// 2. Adding an entry here with the next version number
// 3. Incrementing currentConfigVersion
var configMigrations = []migration{
	{version: 1, description: "legacy providers to connections", up: migrations.V1LegacyToConnections},
	{version: 2, description: "ensure profiles and default models", up: migrations.V2ProfilesAndModels},
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

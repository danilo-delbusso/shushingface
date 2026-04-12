package history

import (
	"database/sql"
	"fmt"
	"log/slog"
)

type dbMigration struct {
	version     int
	description string
	up          func(tx *sql.Tx) error
}

var dbMigrations = []dbMigration{
	{
		version:     1,
		description: "create transcriptions table",
		up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS transcriptions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
				raw_transcript TEXT,
				refined_message TEXT,
				active_app TEXT
			)`)
			return err
		},
	},
	{
		version:     2,
		description: "add error column",
		up: func(tx *sql.Tx) error {
			_, err := tx.Exec(`ALTER TABLE transcriptions ADD COLUMN error TEXT DEFAULT ''`)
			return err
		},
	},
}

func runMigrations(db *sql.DB) error {
	// Create tracking table
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// Find current version
	var current int
	db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&current)

	// Run pending migrations
	for _, m := range dbMigrations {
		if m.version <= current {
			continue
		}
		slog.Info("running db migration", "version", m.version, "description", m.description)
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if err := m.up(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("db migration %d (%s): %w", m.version, m.description, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, m.version); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// bootstrapExistingDB detects databases created before the migration system
// and seeds schema_migrations so existing migrations don't re-run.
func bootstrapExistingDB(db *sql.DB) {
	// If schema_migrations already exists, already bootstrapped
	var name string
	if db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'`).Scan(&name) == nil {
		return
	}

	// If transcriptions table doesn't exist, this is a fresh DB — migrations will handle it
	if db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='transcriptions'`).Scan(&name) != nil {
		return
	}

	slog.Info("bootstrapping existing database into migration system")

	// Create tracking table and seed migration 1 (table exists)
	db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)
	db.Exec(`INSERT INTO schema_migrations (version) VALUES (1)`)

	// Check if error column already exists (migration 2)
	rows, err := db.Query(`PRAGMA table_info(transcriptions)`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var colName, colType string
		var notnull int
		var dflt sql.NullString
		var pk int
		rows.Scan(&cid, &colName, &colType, &notnull, &dflt, &pk)
		if colName == "error" {
			db.Exec(`INSERT INTO schema_migrations (version) VALUES (2)`)
			break
		}
	}
}

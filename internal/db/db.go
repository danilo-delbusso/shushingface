package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"codeberg.org/dbus/shushingface/internal/paths"

	// Register migrations via init() functions.
	_ "codeberg.org/dbus/shushingface/internal/db/migrations"
)

func Open() (*sql.DB, error) {
	stateDir, err := paths.State()
	if err != nil {
		return nil, err
	}
	dbPath := filepath.Join(stateDir, "shushingface.db")
	// Migrate from the legacy location (config dir) on first run after
	// the runtime-paths refactor. WAL/SHM siblings are recreated by
	// SQLite as needed, so we only move the main file.
	paths.MigrateFromConfig("shushingface.db", dbPath)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable WAL mode for better concurrent read/write
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		closeWarn(db, "after WAL pragma failure")
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		closeWarn(db, "after goose dialect failure")
		return nil, fmt.Errorf("setting goose dialect: %w", err)
	}
	if err := goose.UpContext(context.Background(), db, "."); err != nil {
		closeWarn(db, "after migration failure")
		return nil, fmt.Errorf("database migration: %w", err)
	}

	return db, nil
}

func closeWarn(db *sql.DB, what string) {
	if err := db.Close(); err != nil {
		slog.Warn("db close failed", "what", what, "error", err)
	}
}

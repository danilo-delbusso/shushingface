// Package db manages the shared SQLite database connection and migrations.
// All repositories (history, future caches, etc.) receive *sql.DB from here.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	// Register migrations via init() functions.
	_ "codeberg.org/dbus/shushingface/internal/db/migrations"
)

// Open opens the SQLite database and runs all pending migrations.
func Open() (*sql.DB, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	appDir := filepath.Join(configDir, "shushingface")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(appDir, "shushingface.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable WAL mode for better concurrent read/write
	db.Exec("PRAGMA journal_mode=WAL")

	goose.SetDialect("sqlite3")
	if err := goose.UpContext(context.Background(), db, "."); err != nil {
		db.Close()
		return nil, fmt.Errorf("database migration: %w", err)
	}

	return db, nil
}

package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	// Register migrations via init() functions.
	_ "codeberg.org/dbus/shushingface/internal/db/migrations"
)

func Open() (*sql.DB, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	appDir := filepath.Join(configDir, "shushingface")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(appDir, "shushingface.db")

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

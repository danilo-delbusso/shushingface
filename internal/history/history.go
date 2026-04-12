package history

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	// Register migrations via init() functions.
	_ "codeberg.org/dbus/shushingface/internal/history/migrations"
)

// Record represents a single transcription event stored in the local SQLite database.
type Record struct {
	ID             int64     `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	RawTranscript  string    `json:"rawTranscript"`
	RefinedMessage string    `json:"refinedMessage"`
	ActiveApp      string    `json:"activeApp"`
	Error          string    `json:"error,omitempty"`
}

// Manager handles the connection and queries to the local SQLite history database.
type Manager struct {
	db *sql.DB
}

// NewManager establishes a connection to the SQLite database and runs migrations.
func NewManager() (*Manager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	appDir := filepath.Join(configDir, "shushingface")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(appDir, "history.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Run pending migrations
	goose.SetDialect("sqlite3")
	if err := goose.UpContext(context.Background(), db, "."); err != nil {
		db.Close()
		return nil, fmt.Errorf("history db migration: %w", err)
	}

	return &Manager{db: db}, nil
}

// Insert adds a new transcription event to the local history database.
func (m *Manager) Insert(rawTranscript, refinedMessage, activeApp, errMsg string) (int64, error) {
	stmt := `INSERT INTO transcriptions (raw_transcript, refined_message, active_app, error) VALUES (?, ?, ?, ?)`
	res, err := m.db.Exec(stmt, rawTranscript, refinedMessage, activeApp, errMsg)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetHistory retrieves past transcription events, ordered from newest to oldest.
func (m *Manager) GetHistory(limit, offset int) ([]Record, error) {
	stmt := `SELECT id, timestamp, raw_transcript, refined_message, active_app, COALESCE(error, '') FROM transcriptions ORDER BY timestamp DESC LIMIT ? OFFSET ?`
	rows, err := m.db.Query(stmt, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []Record{}
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.ID, &r.Timestamp, &r.RawTranscript, &r.RefinedMessage, &r.ActiveApp, &r.Error); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// ClearHistory wipes all data from the transcriptions table.
func (m *Manager) ClearHistory() error {
	_, err := m.db.Exec(`DELETE FROM transcriptions`)
	return err
}

// Close gracefully closes the database connection.
func (m *Manager) Close() error {
	return m.db.Close()
}

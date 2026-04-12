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

	// Migrate from hand-rolled schema_migrations to goose (one-time bootstrap)
	bootstrapGoose(db)

	// Run pending migrations
	goose.SetDialect("sqlite3")
	if err := goose.UpContext(context.Background(), db, "."); err != nil {
		db.Close()
		return nil, fmt.Errorf("history db migration: %w", err)
	}

	return &Manager{db: db}, nil
}

// bootstrapGoose migrates from the old hand-rolled schema_migrations table
// to goose's goose_db_version table. Only runs once for existing users.
func bootstrapGoose(db *sql.DB) {
	// If goose table already exists, nothing to do
	var name string
	if db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='goose_db_version'`).Scan(&name) == nil {
		return
	}

	// Check if old schema_migrations table exists
	if db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'`).Scan(&name) != nil {
		// Also check if the transcriptions table exists without any migration tracking
		if db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='transcriptions'`).Scan(&name) != nil {
			return // Fresh database, goose will handle everything
		}
	}

	// Old database exists — seed goose with already-applied migrations
	db.Exec(`CREATE TABLE IF NOT EXISTS goose_db_version (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version_id INTEGER NOT NULL,
		is_applied INTEGER NOT NULL,
		tstamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	db.Exec(`INSERT INTO goose_db_version (version_id, is_applied) VALUES (0, 1)`)

	// Check what's already applied
	var tblExists bool
	db.QueryRow(`SELECT 1 FROM sqlite_master WHERE type='table' AND name='transcriptions'`).Scan(&tblExists)
	if tblExists {
		db.Exec(`INSERT INTO goose_db_version (version_id, is_applied) VALUES (1, 1)`)
	}

	// Check for error column
	rows, err := db.Query(`PRAGMA table_info(transcriptions)`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cid int
			var colName, colType string
			var notnull int
			var dflt sql.NullString
			var pk int
			rows.Scan(&cid, &colName, &colType, &notnull, &dflt, &pk)
			if colName == "error" {
				db.Exec(`INSERT INTO goose_db_version (version_id, is_applied) VALUES (2, 1)`)
				break
			}
		}
	}

	// Clean up old table
	db.Exec(`DROP TABLE IF EXISTS schema_migrations`)
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

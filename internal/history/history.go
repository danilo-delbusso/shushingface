// Package history provides a repository for transcription history records.
package history

import (
	"database/sql"
	"time"
)

// Record represents a single transcription event.
type Record struct {
	ID             int64     `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	RawTranscript  string    `json:"rawTranscript"`
	RefinedMessage string    `json:"refinedMessage"`
	ActiveApp      string    `json:"activeApp"`
	Error          string    `json:"error,omitempty"`
}

// Repository provides access to transcription history.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a history repository using the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Insert adds a new transcription event.
func (r *Repository) Insert(rawTranscript, refinedMessage, activeApp, errMsg string) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO transcriptions (raw_transcript, refined_message, active_app, error) VALUES (?, ?, ?, ?)`,
		rawTranscript, refinedMessage, activeApp, errMsg,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetHistory retrieves past transcription events, newest first.
func (r *Repository) GetHistory(limit, offset int) ([]Record, error) {
	rows, err := r.db.Query(
		`SELECT id, timestamp, raw_transcript, refined_message, active_app, COALESCE(error, '') FROM transcriptions ORDER BY timestamp DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var rec Record
		if err := rows.Scan(&rec.ID, &rec.Timestamp, &rec.RawTranscript, &rec.RefinedMessage, &rec.ActiveApp, &rec.Error); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

// Clear wipes all transcription history.
func (r *Repository) Clear() error {
	_, err := r.db.Exec(`DELETE FROM transcriptions`)
	return err
}

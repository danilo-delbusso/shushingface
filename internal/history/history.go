package history

import (
	"database/sql"
	"log/slog"
	"time"
)

type Record struct {
	ID             int64     `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	RawTranscript  string    `json:"rawTranscript"`
	RefinedMessage string    `json:"refinedMessage"`
	ActiveApp      string    `json:"activeApp"`
	Error          string    `json:"error,omitempty"`
}

// Store is the interface for reading and writing transcription history.
type Store interface {
	Insert(rawTranscript, refinedMessage, activeApp, errMsg string) (int64, error)
	GetHistory(limit, offset int) ([]Record, error)
	Clear() error
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

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

func (r *Repository) GetHistory(limit, offset int) ([]Record, error) {
	rows, err := r.db.Query(
		`SELECT id, timestamp, raw_transcript, refined_message, active_app, COALESCE(error, '') FROM transcriptions ORDER BY timestamp DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Warn("failed to close history rows", "error", err)
		}
	}()

	var records []Record
	for rows.Next() {
		var rec Record
		if err := rows.Scan(&rec.ID, &rec.Timestamp, &rec.RawTranscript, &rec.RefinedMessage, &rec.ActiveApp, &rec.Error); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func (r *Repository) Clear() error {
	_, err := r.db.Exec(`DELETE FROM transcriptions`)
	return err
}

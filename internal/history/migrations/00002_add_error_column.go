package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(up00002, down00002)
}

func up00002(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE transcriptions ADD COLUMN error TEXT DEFAULT ''`)
	return err
}

func down00002(ctx context.Context, tx *sql.Tx) error {
	// SQLite doesn't support DROP COLUMN before 3.35.0; safe to no-op
	return nil
}

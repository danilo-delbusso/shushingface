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
	_, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_transcriptions_timestamp ON transcriptions(timestamp DESC)`)
	return err
}

func down00002(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP INDEX IF EXISTS idx_transcriptions_timestamp`)
	return err
}

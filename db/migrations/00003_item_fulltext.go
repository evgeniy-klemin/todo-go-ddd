package migrations

import (
	"context"
	"database/sql"

	"github.com/evgeniy-klemin/todo/db/driver"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationNoTxContext(upItemFulltext, downItemFulltext)
}

// upItemFulltext creates a FULLTEXT index on MySQL. No-op on SQLite (not supported).
func upItemFulltext(ctx context.Context, db *sql.DB) error {
	if getDriver() != driver.MySQL {
		return nil
	}
	_, err := db.ExecContext(ctx, "CREATE FULLTEXT INDEX idx_item_fulltext ON item (name)")
	if err != nil && !isMySQLError(err, 1061) {
		return err
	}
	return nil
}

func downItemFulltext(ctx context.Context, db *sql.DB) error {
	if getDriver() != driver.MySQL {
		return nil
	}
	_, err := db.ExecContext(ctx, "DROP INDEX idx_item_fulltext ON item")
	if err != nil && !isMySQLError(err, 1091) {
		return err
	}
	return nil
}

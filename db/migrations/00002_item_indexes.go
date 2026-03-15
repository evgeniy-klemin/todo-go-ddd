package migrations

import (
	"context"
	"database/sql"
	"errors"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/evgeniy-klemin/todo/db/driver"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationNoTxContext(upItemIndexes, downItemIndexes)
}

func upItemIndexes(ctx context.Context, db *sql.DB) error {
	switch getDriver() {
	case driver.MySQL:
		_, err := db.ExecContext(ctx, "CREATE INDEX idx_item_position ON item (position)")
		if err != nil && !isMySQLError(err, 1061) {
			return err
		}
	default: // SQLite
		_, err := db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_item_position ON item (position)")
		if err != nil {
			return err
		}
	}
	return nil
}

func downItemIndexes(ctx context.Context, db *sql.DB) error {
	switch getDriver() {
	case driver.MySQL:
		_, err := db.ExecContext(ctx, "DROP INDEX idx_item_position ON item")
		if err != nil && !isMySQLError(err, 1091) {
			return err
		}
	default:
		_, err := db.ExecContext(ctx, "DROP INDEX IF EXISTS idx_item_position")
		if err != nil {
			return err
		}
	}
	return nil
}

// isMySQLError returns true when err is a MySQL error with the given number.
// MySQL DDL errors like 1061 (duplicate key name) are surfaced this way.
func isMySQLError(err error, number uint16) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == number
}

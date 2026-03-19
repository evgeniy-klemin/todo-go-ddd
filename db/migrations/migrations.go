package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/evgeniy-klemin/todo/db/driver"
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var embedMigrations embed.FS

var currentDriver string

func setDriver(d string) {
	currentDriver = d
}

func getDriver() string {
	return currentDriver
}

// Run applies all pending migrations to db using the given driver name.
func Run(db *sql.DB, d string) error {
	setDriver(d)
	goose.SetBaseFS(embedMigrations)
	dialect, err := gooseDialect(d)
	if err != nil {
		return err
	}
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func gooseDialect(d string) (string, error) {
	switch d {
	case driver.MySQL:
		return "mysql", nil
	case driver.SQLite:
		return "sqlite3", nil
	default:
		return "", fmt.Errorf("unsupported database driver: %q", d)
	}
}

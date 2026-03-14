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

func SetDriver(d string) {
	currentDriver = d
}

func Driver() string {
	return currentDriver
}

// Run applies all pending migrations to db using the given driver name.
func Run(db *sql.DB, d string) error {
	SetDriver(d)
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect(gooseDialect(d)); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func gooseDialect(d string) string {
	switch d {
	case driver.MySQL:
		return "mysql"
	default:
		return "sqlite3"
	}
}

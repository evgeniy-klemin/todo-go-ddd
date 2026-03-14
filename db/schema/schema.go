// Package schema provides centralized database schema definitions and migration helpers
// for the todo application. This avoids duplicating table/FTS5/trigger DDL across
// server startup and test setup code.
package schema

import (
	_ "embed"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

// DriverSQLite is the driver name for SQLite.
const DriverSQLite = "sqlite3"

// DriverMySQL is the driver name for MySQL.
const DriverMySQL = "mysql"

// ItemTableCreate is the DDL for the core item table, embedded from item.sql.
//
//go:embed item.sql
var ItemTableCreate string

// ItemIndexCreateSQLite is the DDL for the position index (SQLite supports IF NOT EXISTS).
//
//go:embed item_index_sqlite.sql
var ItemIndexCreateSQLite string

// ItemIndexCreateMySQL is the DDL for the position index (MySQL lacks IF NOT EXISTS for indexes).
//
//go:embed item_index_mysql.sql
var ItemIndexCreateMySQL string

// ItemFulltextCreateMySQL is the DDL for the MySQL FULLTEXT index on the name column.
//
//go:embed item_fulltext_mysql.sql
var ItemFulltextCreateMySQL string

// FTSTable creates the FTS5 virtual table for full-text search.
//
//go:embed item_fts.sql
var FTSTable string

// FTSTriggers contains the DDL for all three FTS sync triggers (INSERT, DELETE, UPDATE).
//
//go:embed item_fts_triggers.sql
var FTSTriggers string

// FTSRebuild re-indexes all existing rows in the item table into the FTS index.
//
//go:embed item_fts_rebuild.sql
var FTSRebuild string

// DropFTSTriggers removes FTS triggers (useful when FTS5 is not available but stale
// triggers remain from a previous run).
//
//go:embed item_fts_drop_triggers.sql
var DropFTSTriggers string

// Apply creates the base item table and position index. It does not set up FTS5.
// For MySQL, the index creation ignores "duplicate key" errors since MySQL does not
// support CREATE INDEX IF NOT EXISTS.
func Apply(db *sql.DB, driver string) error {
	if _, err := db.Exec(ItemTableCreate); err != nil {
		return err
	}
	indexDDL := ItemIndexCreateSQLite
	if driver == DriverMySQL {
		indexDDL = ItemIndexCreateMySQL
	}
	_, err := db.Exec(indexDDL)
	if err != nil && driver == DriverMySQL {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1061 {
			// Ignore "Duplicate key name" error (index already exists)
			return nil
		}
	}
	return err
}

// ApplyFTS creates the FTS5 virtual table, triggers, and rebuilds the index to
// cover any existing rows. Returns true if FTS5 is available and was set up
// successfully. If FTS5 is not compiled in, it cleans up any stale triggers and
// returns false.
func ApplyFTS(db *sql.DB) bool {
	if _, err := db.Exec(FTSTable); err != nil {
		// FTS5 not available — clean up stale triggers from a previous run
		db.Exec(DropFTSTriggers)
		return false
	}

	// FTSTriggers contains all three trigger DDL statements separated by semicolons.
	// go-sqlite3 uses sqlite3_exec internally, which handles multiple statements.
	if _, err := db.Exec(FTSTriggers); err != nil {
		return false
	}

	// Rebuild FTS index so any rows already in item are indexed
	db.Exec(FTSRebuild)
	return true
}

// ApplyAll creates the item table and attempts to set up FTS5. It returns whether
// FTS5 (or MySQL FULLTEXT) was successfully enabled and any error from creating the base table.
// When driver is "mysql", SQLite FTS5 setup is skipped but MySQL FULLTEXT index is created.
func ApplyAll(db *sql.DB, driver string) (ftsEnabled bool, err error) {
	if err := Apply(db, driver); err != nil {
		return false, err
	}
	if driver == DriverMySQL {
		_, indexErr := db.Exec(ItemFulltextCreateMySQL)
		if indexErr != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(indexErr, &mysqlErr) && mysqlErr.Number == 1061 {
				// Ignore "Duplicate key name" error (index already exists)
				return true, nil
			}
			return false, indexErr
		}
		return true, nil
	}
	return ApplyFTS(db), nil
}

// Package schema provides centralized database schema definitions and migration helpers
// for the todo application. This avoids duplicating table/FTS5/trigger DDL across
// server startup and test setup code.
package schema

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

// DriverSQLite is the driver name for SQLite.
const DriverSQLite = "sqlite3"

// DriverMySQL is the driver name for MySQL.
const DriverMySQL = "mysql"

// ItemTableCreate is the DDL for the core item table.
const ItemTableCreate = `
CREATE TABLE IF NOT EXISTS item (
	id VARCHAR(36) NOT NULL PRIMARY KEY,
	name VARCHAR(1000) NOT NULL,
	position INTEGER NOT NULL DEFAULT 1,
	done BOOL NOT NULL DEFAULT FALSE,
	created_at DATETIME NOT NULL
)`

// ItemIndexCreateSQLite is the DDL for the position index (SQLite supports IF NOT EXISTS).
const ItemIndexCreateSQLite = `CREATE INDEX IF NOT EXISTS idx_item_position ON item (position)`

// ItemIndexCreateMySQL is the DDL for the position index (MySQL lacks IF NOT EXISTS for indexes).
const ItemIndexCreateMySQL = `CREATE INDEX idx_item_position ON item (position)`

// ItemFulltextCreateMySQL is the DDL for the MySQL FULLTEXT index on the name column.
const ItemFulltextCreateMySQL = `CREATE FULLTEXT INDEX idx_item_fulltext ON item (name)`

// FTSTable creates the FTS5 virtual table for full-text search.
const FTSTable = `CREATE VIRTUAL TABLE IF NOT EXISTS item_fts USING fts5(name, content='item', content_rowid='rowid')`

// FTSTriggerInsert keeps the FTS index in sync on INSERT.
const FTSTriggerInsert = `CREATE TRIGGER IF NOT EXISTS item_ai AFTER INSERT ON item BEGIN
	INSERT INTO item_fts(rowid, name) VALUES (new.rowid, new.name);
END`

// FTSTriggerDelete keeps the FTS index in sync on DELETE.
const FTSTriggerDelete = `CREATE TRIGGER IF NOT EXISTS item_ad AFTER DELETE ON item BEGIN
	INSERT INTO item_fts(item_fts, rowid, name) VALUES('delete', old.rowid, old.name);
END`

// FTSTriggerUpdate keeps the FTS index in sync on UPDATE.
const FTSTriggerUpdate = `CREATE TRIGGER IF NOT EXISTS item_au AFTER UPDATE ON item BEGIN
	INSERT INTO item_fts(item_fts, rowid, name) VALUES('delete', old.rowid, old.name);
	INSERT INTO item_fts(rowid, name) VALUES (new.rowid, new.name);
END`

// FTSRebuild re-indexes all existing rows in the item table into the FTS index.
const FTSRebuild = `INSERT INTO item_fts(item_fts) VALUES('rebuild')`

// DropFTSTriggers removes FTS triggers (useful when FTS5 is not available but stale
// triggers remain from a previous run).
const DropFTSTriggers = `
DROP TRIGGER IF EXISTS item_ai;
DROP TRIGGER IF EXISTS item_ad;
DROP TRIGGER IF EXISTS item_au;
`

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

	for _, stmt := range []string{FTSTriggerInsert, FTSTriggerDelete, FTSTriggerUpdate} {
		if _, err := db.Exec(stmt); err != nil {
			return false
		}
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

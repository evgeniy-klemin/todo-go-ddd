package migrations_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/evgeniy-klemin/todo/db/driver"
	"github.com/evgeniy-klemin/todo/db/migrations"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestRun_CreatesItemTable(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, migrations.Run(db, driver.SQLite))

	row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='item'")
	var name string
	require.NoError(t, row.Scan(&name))
	assert.Equal(t, "item", name)
}

func TestRun_CreatesPositionIndex(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, migrations.Run(db, driver.SQLite))

	row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_item_position'")
	var name string
	require.NoError(t, row.Scan(&name))
	assert.Equal(t, "idx_item_position", name)
}

func TestRun_Idempotent(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, migrations.Run(db, driver.SQLite))
	require.NoError(t, migrations.Run(db, driver.SQLite))
}

func TestRun_GooseVersionTracked(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, migrations.Run(db, driver.SQLite))

	row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='goose_db_version'")
	var name string
	require.NoError(t, row.Scan(&name))
	assert.Equal(t, "goose_db_version", name)
}

// TestRun_SQLite_IndexMigration verifies that migration 00002 creates the
// position index on SQLite via the IF NOT EXISTS path (no MySQL DDL needed).
func TestRun_SQLite_IndexMigration(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, migrations.Run(db, driver.SQLite))

	row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_item_position'")
	var name string
	require.NoError(t, row.Scan(&name))
	assert.Equal(t, "idx_item_position", name)
}

func TestRun_UnsupportedDriver(t *testing.T) {
	db := newTestDB(t)
	err := migrations.Run(db, "postgres")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported database driver")
}

// TestRun_SQLite_FulltextMigrationIsNoOp verifies that migration 00003 is a
// no-op on SQLite — no FULLTEXT index is created (SQLite does not support it).
func TestRun_SQLite_FulltextMigrationIsNoOp(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, migrations.Run(db, driver.SQLite))

	// On SQLite the fulltext migration returns early, so no fulltext index
	// should appear in sqlite_master.
	row := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='index' AND name='idx_item_fulltext'")
	var count int
	require.NoError(t, row.Scan(&count))
	assert.Equal(t, 0, count)
}

package fts_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/evgeniy-klemin/todo/db/driver"
	"github.com/evgeniy-klemin/todo/db/fts"
	"github.com/evgeniy-klemin/todo/db/migrations"
)

// checkFTS5 skips t if FTS5 is not compiled into the sqlite3 library.
func checkFTS5(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec("CREATE VIRTUAL TABLE _fts5_probe USING fts5(x)"); err != nil {
		t.Skipf("FTS5 not available in this build, skipping: %v", err)
	}
	db.Exec("DROP TABLE _fts5_probe")
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	require.NoError(t, migrations.Run(db, driver.SQLite))
	return db
}

func TestApply_ReturnsTrueWhenFTS5Available(t *testing.T) {
	db := newTestDB(t)
	checkFTS5(t, db)

	ok := fts.Apply(db)
	assert.True(t, ok)

	row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='item_fts'")
	var name string
	require.NoError(t, row.Scan(&name))
	assert.Equal(t, "item_fts", name)
}

func TestApply_RebuildIndexesExistingRows(t *testing.T) {
	db := newTestDB(t)
	checkFTS5(t, db)

	_, err := db.Exec(
		`INSERT INTO item (id, name, position, done, created_at) VALUES ('1', 'buy milk', 1, false, datetime('now'))`,
	)
	require.NoError(t, err)

	ok := fts.Apply(db)
	require.True(t, ok)

	row := db.QueryRow("SELECT COUNT(*) FROM item_fts WHERE item_fts MATCH 'milk'")
	var count int
	require.NoError(t, row.Scan(&count))
	assert.Equal(t, 1, count)
}

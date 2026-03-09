package schema

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestApply_CreatesItemTable(t *testing.T) {
	db, err := sql.Open(DriverSQLite, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := Apply(db, DriverSQLite); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Verify the table exists
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='item'").Scan(&name)
	if err != nil {
		t.Fatalf("item table not found: %v", err)
	}
	if name != "item" {
		t.Errorf("expected table name 'item', got %q", name)
	}
}

func TestApplyFTS_ReturnsTrueWhenAvailable(t *testing.T) {
	db, err := sql.Open(DriverSQLite, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// FTS5 requires the base table to exist first (for content= sync)
	if err := Apply(db, DriverSQLite); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Check if FTS5 is available in this build
	if _, err := db.Exec("CREATE VIRTUAL TABLE _fts_probe USING fts5(x)"); err != nil {
		t.Skipf("FTS5 not available, skipping: %v", err)
	}
	db.Exec("DROP TABLE _fts_probe")

	got := ApplyFTS(db)
	if !got {
		t.Error("expected ApplyFTS to return true when FTS5 is available")
	}

	// Verify the FTS table exists
	var tblName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='item_fts'").Scan(&tblName)
	if err != nil {
		t.Fatalf("item_fts table not found: %v", err)
	}
}

func TestApplyAll_CreatesTableAndFTS(t *testing.T) {
	db, err := sql.Open(DriverSQLite, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Check if FTS5 is available
	if _, err := db.Exec("CREATE VIRTUAL TABLE _fts_probe USING fts5(x)"); err != nil {
		t.Skipf("FTS5 not available, skipping: %v", err)
	}
	db.Exec("DROP TABLE _fts_probe")
	// Re-open since we polluted it
	db.Close()
	db, err = sql.Open(DriverSQLite, ":memory:")
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer db.Close()

	fts, err := ApplyAll(db, DriverSQLite)
	if err != nil {
		t.Fatalf("ApplyAll: %v", err)
	}
	if !fts {
		t.Error("expected FTS enabled")
	}
}

func TestApplyFTS_RebuildIndexesExistingRows(t *testing.T) {
	db, err := sql.Open(DriverSQLite, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Check if FTS5 is available
	if _, err := db.Exec("CREATE VIRTUAL TABLE _fts_probe USING fts5(x)"); err != nil {
		t.Skipf("FTS5 not available, skipping: %v", err)
	}
	db.Exec("DROP TABLE _fts_probe")
	db.Close()
	db, err = sql.Open(DriverSQLite, ":memory:")
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer db.Close()

	// Create base table and insert data BEFORE FTS setup
	if err := Apply(db, DriverSQLite); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	_, err = db.Exec(`INSERT INTO item (id, name, position, done, created_at) VALUES ('aaa', 'Buy milk', 1, 0, '2025-01-01')`)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Now apply FTS — it should rebuild and index the existing row
	if !ApplyFTS(db) {
		t.Fatal("ApplyFTS returned false")
	}

	// Search for the pre-existing row via FTS
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM item_fts WHERE item_fts MATCH '"buy"*'`).Scan(&count)
	if err != nil {
		t.Fatalf("FTS query: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 FTS result for pre-existing row, got %d", count)
	}
}

func TestApply_Idempotent(t *testing.T) {
	db, err := sql.Open(DriverSQLite, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Apply twice should not error
	if err := Apply(db, DriverSQLite); err != nil {
		t.Fatalf("first Apply: %v", err)
	}
	if err := Apply(db, DriverSQLite); err != nil {
		t.Fatalf("second Apply: %v", err)
	}
}

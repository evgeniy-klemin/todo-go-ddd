package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	"github.com/evgeniy-klemin/todo/internal/item/domain"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE item (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			name VARCHAR(1000) NOT NULL,
			position INTEGER NOT NULL DEFAULT 1,
			done BOOL NOT NULL DEFAULT FALSE,
			created_at DATETIME NOT NULL
		);
		CREATE INDEX idx_item_position ON item (position);

		CREATE VIRTUAL TABLE IF NOT EXISTS item_fts USING fts5(name, content='item', content_rowid='rowid');

		CREATE TRIGGER item_ai AFTER INSERT ON item BEGIN
			INSERT INTO item_fts(rowid, name) VALUES (new.rowid, new.name);
		END;
		CREATE TRIGGER item_ad AFTER DELETE ON item BEGIN
			INSERT INTO item_fts(item_fts, rowid, name) VALUES('delete', old.rowid, old.name);
		END;
		CREATE TRIGGER item_au AFTER UPDATE ON item BEGIN
			INSERT INTO item_fts(item_fts, rowid, name) VALUES('delete', old.rowid, old.name);
			INSERT INTO item_fts(rowid, name) VALUES (new.rowid, new.name);
		END;
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	return db
}

func TestAddWithNextPosition_ConcurrentCreation_UniquePositions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	const numGoroutines = 20
	var wg sync.WaitGroup
	results := make([]*domain.Item, numGoroutines)
	errs := make([]error, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			item, err := domain.NewItem(fmt.Sprintf("Task %d", idx), 1) // placeholder position
			if err != nil {
				errs[idx] = fmt.Errorf("NewItem: %w", err)
				return
			}
			result, err := repo.AddWithNextPosition(ctx, item)
			if err != nil {
				errs[idx] = fmt.Errorf("AddWithNextPosition: %w", err)
				return
			}
			results[idx] = result
		}(i)
	}
	wg.Wait()

	// Check for errors
	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d failed: %v", i, err)
		}
	}

	// Check all positions are unique
	positions := make(map[int]bool)
	for i, item := range results {
		pos := item.Position().Int()
		if positions[pos] {
			t.Errorf("duplicate position %d found for goroutine %d", pos, i)
		}
		positions[pos] = true
	}

	// Check positions are 1..numGoroutines
	for i := 1; i <= numGoroutines; i++ {
		if !positions[i] {
			t.Errorf("expected position %d to be assigned, but it was not", i)
		}
	}
}

func TestAddWithNextPosition_SequentialCreation_IncrementsPosition(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		item, err := domain.NewItem(fmt.Sprintf("Task %d", i), 1)
		if err != nil {
			t.Fatalf("NewItem: %v", err)
		}
		result, err := repo.AddWithNextPosition(ctx, item)
		if err != nil {
			t.Fatalf("AddWithNextPosition: %v", err)
		}
		if result.Position().Int() != i {
			t.Errorf("expected position %d, got %d", i, result.Position().Int())
		}
	}
}

func TestAll_SearchReturnsMatchingItems(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	// Insert test items
	for _, name := range []string{"Buy milk", "Buy eggs", "Walk the dog", "Read a book"} {
		item, err := domain.NewItem(name, 1)
		if err != nil {
			t.Fatalf("NewItem: %v", err)
		}
		if _, err := repo.AddWithNextPosition(ctx, item); err != nil {
			t.Fatalf("AddWithNextPosition: %v", err)
		}
	}

	search := "buy"
	items, err := repo.All(ctx, nil, &search, nil, 1, 20, nil)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items matching 'buy', got %d", len(items))
	}
}

func TestAll_SearchNoMatchReturnsEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	item, err := domain.NewItem("Buy milk", 1)
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}
	if _, err := repo.AddWithNextPosition(ctx, item); err != nil {
		t.Fatalf("AddWithNextPosition: %v", err)
	}

	search := "xyz"
	items, err := repo.All(ctx, nil, &search, nil, 1, 20, nil)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items matching 'xyz', got %d", len(items))
	}
}

func TestAll_SearchCaseInsensitive(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	item, err := domain.NewItem("Buy Milk", 1)
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}
	if _, err := repo.AddWithNextPosition(ctx, item); err != nil {
		t.Fatalf("AddWithNextPosition: %v", err)
	}

	search := "buy milk"
	items, err := repo.All(ctx, nil, &search, nil, 1, 20, nil)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item matching 'buy milk' (case-insensitive), got %d", len(items))
	}
}

func TestAll_SearchPartialMatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	item, err := domain.NewItem("Buy milk and eggs", 1)
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}
	if _, err := repo.AddWithNextPosition(ctx, item); err != nil {
		t.Fatalf("AddWithNextPosition: %v", err)
	}

	search := "milk"
	items, err := repo.All(ctx, nil, &search, nil, 1, 20, nil)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item matching partial 'milk', got %d", len(items))
	}
}

func TestAll_NilSearchReturnsAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		item, err := domain.NewItem(fmt.Sprintf("Task %d", i), 1)
		if err != nil {
			t.Fatalf("NewItem: %v", err)
		}
		if _, err := repo.AddWithNextPosition(ctx, item); err != nil {
			t.Fatalf("AddWithNextPosition: %v", err)
		}
	}

	items, err := repo.All(ctx, nil, nil, nil, 1, 20, nil)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items with nil search, got %d", len(items))
	}
}

func TestCount_SearchFiltersCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	for _, name := range []string{"Buy milk", "Buy eggs", "Walk the dog"} {
		item, err := domain.NewItem(name, 1)
		if err != nil {
			t.Fatalf("NewItem: %v", err)
		}
		if _, err := repo.AddWithNextPosition(ctx, item); err != nil {
			t.Fatalf("AddWithNextPosition: %v", err)
		}
	}

	search := "buy"
	count, err := repo.Count(ctx, nil, &search)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2 for 'buy', got %d", count)
	}

	// nil search returns all
	count, err = repo.Count(ctx, nil, nil)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3 for nil search, got %d", count)
	}
}

func TestAll_SearchPrefixMatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	// Insert test items
	for _, name := range []string{"Buying groceries", "Buy milk", "Walk the dog"} {
		item, err := domain.NewItem(name, 1)
		if err != nil {
			t.Fatalf("NewItem: %v", err)
		}
		if _, err := repo.AddWithNextPosition(ctx, item); err != nil {
			t.Fatalf("AddWithNextPosition: %v", err)
		}
	}

	search := "buy"
	items, err := repo.All(ctx, nil, &search, nil, 1, 20, nil)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items matching prefix 'buy' (Buy milk, Buying groceries), got %d", len(items))
	}
}

func TestAll_SearchMultipleWords(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	for _, name := range []string{"Buy milk and eggs", "Buy bread", "Get milk"} {
		item, err := domain.NewItem(name, 1)
		if err != nil {
			t.Fatalf("NewItem: %v", err)
		}
		if _, err := repo.AddWithNextPosition(ctx, item); err != nil {
			t.Fatalf("AddWithNextPosition: %v", err)
		}
	}

	// Both "buy" and "milk" must be present (FTS5 AND semantics)
	search := "buy milk"
	items, err := repo.All(ctx, nil, &search, nil, 1, 20, nil)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item matching 'buy milk', got %d", len(items))
	}
}

func TestBuildFTSQuery(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"buy", `"buy"*`},
		{"buy milk", `"buy"* "milk"*`},
		{"Buy Milk", `"Buy"* "Milk"*`},
		{"", ""},
	}
	for _, tt := range tests {
		got := buildFTSQuery(tt.input)
		if got != tt.expected {
			t.Errorf("buildFTSQuery(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestAddWithNextPosition_WithExistingItems_ContinuesFromMax(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	// Insert an item with position 10 using regular Add
	item1, err := domain.NewItem("Existing Task", 10)
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}
	_, err = repo.Add(ctx, item1)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Now use AddWithNextPosition — should get position 11
	item2, err := domain.NewItem("New Task", 1)
	if err != nil {
		t.Fatalf("NewItem: %v", err)
	}
	result, err := repo.AddWithNextPosition(ctx, item2)
	if err != nil {
		t.Fatalf("AddWithNextPosition: %v", err)
	}
	if result.Position().Int() != 11 {
		t.Errorf("expected position 11, got %d", result.Position().Int())
	}
}

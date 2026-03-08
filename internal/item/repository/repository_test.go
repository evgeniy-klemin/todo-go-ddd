package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	"github.com/evgeniy-klemin/todo/db/schema"
	"github.com/evgeniy-klemin/todo/internal/item/domain"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if _, err := schema.ApplyAll(db); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	return db
}

func setupTestDBWithoutFTS(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := schema.Apply(db); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	return db
}

// insertItems is a test helper that inserts named items into the repo.
func insertItems(t *testing.T, repo *Repository, names ...string) {
	t.Helper()
	ctx := context.Background()
	for _, name := range names {
		item, err := domain.NewItem(name, 1)
		if err != nil {
			t.Fatalf("NewItem(%q): %v", name, err)
		}
		if _, err := repo.AddWithNextPosition(ctx, item); err != nil {
			t.Fatalf("AddWithNextPosition(%q): %v", name, err)
		}
	}
}

func TestSearch(t *testing.T) {
	t.Run("FTS5", func(t *testing.T) {
		t.Run("ReturnsMatchingItems", func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()
			repo := New(db)
			insertItems(t, repo, "Buy milk", "Buy eggs", "Walk the dog", "Read a book")

			search := "buy"
			items, err := repo.All(context.Background(), nil, &search, nil, 1, 20, nil)
			if err != nil {
				t.Fatalf("All: %v", err)
			}
			if len(items) != 2 {
				t.Errorf("expected 2 items matching 'buy', got %d", len(items))
			}
		})

		t.Run("NoMatch", func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()
			repo := New(db)
			insertItems(t, repo, "Buy milk")

			search := "xyz"
			items, err := repo.All(context.Background(), nil, &search, nil, 1, 20, nil)
			if err != nil {
				t.Fatalf("All: %v", err)
			}
			if len(items) != 0 {
				t.Errorf("expected 0 items matching 'xyz', got %d", len(items))
			}
		})

		t.Run("CaseInsensitive", func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()
			repo := New(db)
			insertItems(t, repo, "Buy Milk")

			search := "buy milk"
			items, err := repo.All(context.Background(), nil, &search, nil, 1, 20, nil)
			if err != nil {
				t.Fatalf("All: %v", err)
			}
			if len(items) != 1 {
				t.Errorf("expected 1 item matching 'buy milk' (case-insensitive), got %d", len(items))
			}
		})

		t.Run("PartialMatch", func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()
			repo := New(db)
			insertItems(t, repo, "Buy milk and eggs")

			search := "milk"
			items, err := repo.All(context.Background(), nil, &search, nil, 1, 20, nil)
			if err != nil {
				t.Fatalf("All: %v", err)
			}
			if len(items) != 1 {
				t.Errorf("expected 1 item matching partial 'milk', got %d", len(items))
			}
		})

		t.Run("PrefixMatch", func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()
			repo := New(db)
			insertItems(t, repo, "Buying groceries", "Buy milk", "Walk the dog")

			search := "buy"
			items, err := repo.All(context.Background(), nil, &search, nil, 1, 20, nil)
			if err != nil {
				t.Fatalf("All: %v", err)
			}
			if len(items) != 2 {
				t.Errorf("expected 2 items matching prefix 'buy', got %d", len(items))
			}
		})

		t.Run("MultipleWords", func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()
			repo := New(db)
			insertItems(t, repo, "Buy milk and eggs", "Buy bread", "Get milk")

			search := "buy milk"
			items, err := repo.All(context.Background(), nil, &search, nil, 1, 20, nil)
			if err != nil {
				t.Fatalf("All: %v", err)
			}
			if len(items) != 1 {
				t.Errorf("expected 1 item matching 'buy milk', got %d", len(items))
			}
		})

		t.Run("NilSearchReturnsAll", func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()
			repo := New(db)
			insertItems(t, repo, "Task 1", "Task 2", "Task 3")

			items, err := repo.All(context.Background(), nil, nil, nil, 1, 20, nil)
			if err != nil {
				t.Fatalf("All: %v", err)
			}
			if len(items) != 3 {
				t.Errorf("expected 3 items with nil search, got %d", len(items))
			}
		})
	})

	t.Run("Fallback", func(t *testing.T) {
		t.Run("LIKESearch", func(t *testing.T) {
			db := setupTestDBWithoutFTS(t)
			defer db.Close()
			repo := New(db)
			insertItems(t, repo, "Buy milk", "Buy eggs", "Walk the dog", "Read a book")

			search := "buy"
			items, err := repo.All(context.Background(), nil, &search, nil, 1, 20, nil)
			if err != nil {
				t.Fatalf("All with LIKE fallback: %v", err)
			}
			if len(items) != 2 {
				t.Errorf("expected 2 items matching 'buy' via LIKE fallback, got %d", len(items))
			}
		})

		t.Run("LIKECount", func(t *testing.T) {
			db := setupTestDBWithoutFTS(t)
			defer db.Close()
			repo := New(db)
			insertItems(t, repo, "Buy milk", "Buy eggs", "Walk the dog")

			search := "buy"
			count, err := repo.Count(context.Background(), nil, &search)
			if err != nil {
				t.Fatalf("Count with LIKE fallback: %v", err)
			}
			if count != 2 {
				t.Errorf("expected count 2 for 'buy' via LIKE fallback, got %d", count)
			}
		})
	})

	t.Run("BuildFTSQuery", func(t *testing.T) {
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
	})
}

func TestCount(t *testing.T) {
	t.Run("SearchFiltersCount", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := New(db)
		insertItems(t, repo, "Buy milk", "Buy eggs", "Walk the dog")

		search := "buy"
		count, err := repo.Count(context.Background(), nil, &search)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if count != 2 {
			t.Errorf("expected count 2 for 'buy', got %d", count)
		}

		// nil search returns all
		count, err = repo.Count(context.Background(), nil, nil)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if count != 3 {
			t.Errorf("expected count 3 for nil search, got %d", count)
		}
	})
}

func TestPosition(t *testing.T) {
	t.Run("ConcurrentCreation", func(t *testing.T) {
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
				item, err := domain.NewItem(fmt.Sprintf("Task %d", idx), 1)
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

		for i, err := range errs {
			if err != nil {
				t.Fatalf("goroutine %d failed: %v", i, err)
			}
		}

		positions := make(map[int]bool)
		for i, item := range results {
			pos := item.Position().Int()
			if positions[pos] {
				t.Errorf("duplicate position %d found for goroutine %d", pos, i)
			}
			positions[pos] = true
		}

		for i := 1; i <= numGoroutines; i++ {
			if !positions[i] {
				t.Errorf("expected position %d to be assigned, but it was not", i)
			}
		}
	})

	t.Run("SequentialIncrement", func(t *testing.T) {
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
	})

	t.Run("ContinuesFromMax", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		repo := New(db)
		ctx := context.Background()

		// Insert an item with position 10 using regular Add
		item1, err := domain.NewItem("Existing Task", 10)
		if err != nil {
			t.Fatalf("NewItem: %v", err)
		}
		if _, err = repo.Add(ctx, item1); err != nil {
			t.Fatalf("Add: %v", err)
		}

		// AddWithNextPosition should get position 11
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
	})
}

func TestHasFTS(t *testing.T) {
	t.Run("DetectsAbsence", func(t *testing.T) {
		db := setupTestDBWithoutFTS(t)
		defer db.Close()
		repo := New(db)
		if repo.hasFTS() {
			t.Error("expected hasFTS() = false for DB without FTS5 table")
		}
	})

	t.Run("DetectsPresence", func(t *testing.T) {
		// Check if FTS5 is available
		probe, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("open db: %v", err)
		}
		_, err = probe.Exec("CREATE VIRTUAL TABLE _fts_probe USING fts5(x)")
		probe.Close()
		if err != nil {
			t.Skipf("FTS5 not available, skipping: %v", err)
		}

		db := setupTestDB(t)
		defer db.Close()
		repo := New(db)
		if !repo.hasFTS() {
			t.Error("expected hasFTS() = true for DB with FTS5 table")
		}
	})
}

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

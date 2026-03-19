package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/evgeniy-klemin/todo/internal/item/domain"
	"github.com/evgeniy-klemin/todo/internal/item/repository"
	"github.com/evgeniy-klemin/todo/internal/item/repository/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite3: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE item (
		id VARCHAR(36) NOT NULL PRIMARY KEY,
		name VARCHAR(1000) NOT NULL,
		position INTEGER NOT NULL DEFAULT 1,
		done BOOL NOT NULL DEFAULT FALSE,
		created_at DATETIME NOT NULL
	)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	boil.SetDB(db)
	return db
}

func insertItem(t *testing.T, db *sql.DB, id string, name string, done bool) {
	t.Helper()
	item := models.Item{
		ID:        id,
		Name:      name,
		Position:  1,
		Done:      done,
		CreatedAt: time.Now().Truncate(time.Second),
	}
	if err := item.Insert(context.Background(), db, boil.Infer()); err != nil {
		t.Fatalf("insert item: %v", err)
	}
}

func newID(t *testing.T) string {
	t.Helper()
	id, err := domain.GenerateModelID()
	if err != nil {
		t.Fatalf("generate id: %v", err)
	}
	return id.String()
}

func TestCountWithFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.New(db)
	ctx := context.Background()

	// Insert 3 done and 2 not-done items
	insertItem(t, db, newID(t), "task done 1", true)
	insertItem(t, db, newID(t), "task done 2", true)
	insertItem(t, db, newID(t), "task done 3", true)
	insertItem(t, db, newID(t), "task not done 1", false)
	insertItem(t, db, newID(t), "task not done 2", false)

	t.Run("count all items no filter", func(t *testing.T) {
		count, err := repo.Count(ctx, nil)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if count != 5 {
			t.Errorf("expected 5, got %d", count)
		}
	})

	t.Run("count with done=true", func(t *testing.T) {
		done := true
		count, err := repo.Count(ctx, &done)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if count != 3 {
			t.Errorf("expected 3, got %d", count)
		}
	})

	t.Run("count with done=false", func(t *testing.T) {
		done := false
		count, err := repo.Count(ctx, &done)
		if err != nil {
			t.Fatalf("Count: %v", err)
		}
		if count != 2 {
			t.Errorf("expected 2, got %d", count)
		}
	})
}

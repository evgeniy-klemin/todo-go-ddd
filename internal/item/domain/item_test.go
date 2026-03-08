package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewItem_Valid(t *testing.T) {
	item, err := NewItem("Buy milk", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "Buy milk" {
		t.Errorf("expected name 'Buy milk', got '%s'", item.Name)
	}
	if item.Position != 1 {
		t.Errorf("expected position 1, got %d", item.Position)
	}
	if item.Done {
		t.Error("new item should not be done")
	}
	if item.ID.IsZero() {
		t.Error("ID should be generated")
	}
	if item.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewItem_EmptyName_ReturnsError(t *testing.T) {
	_, err := NewItem("", 1)
	if !errors.Is(err, ErrNameLength) {
		t.Errorf("expected ErrNameLength, got: %v", err)
	}
}

func TestNewItem_NameTooLong_ReturnsError(t *testing.T) {
	_, err := NewItem(strings.Repeat("x", NameMaxLength+1), 1)
	if !errors.Is(err, ErrNameLength) {
		t.Errorf("expected ErrNameLength, got: %v", err)
	}
}

func TestNewItem_ZeroPosition_ReturnsError(t *testing.T) {
	_, err := NewItem("Task", 0)
	if !errors.Is(err, ErrPositionValue) {
		t.Errorf("expected ErrPositionValue, got: %v", err)
	}
}

func TestItem_Complete(t *testing.T) {
	item, _ := NewItem("Task", 1)
	if item.Done {
		t.Fatal("item should not be done initially")
	}
	item.Complete()
	if !item.Done {
		t.Error("item should be done after Complete()")
	}
}

func TestItem_Uncomplete(t *testing.T) {
	item, _ := NewItem("Task", 1)
	item.Complete()
	if !item.Done {
		t.Fatal("item should be done after Complete()")
	}
	item.Uncomplete()
	if item.Done {
		t.Error("item should not be done after Uncomplete()")
	}
}

func TestItem_Rename_Valid(t *testing.T) {
	item, _ := NewItem("Old name", 1)
	if err := item.Rename("New name"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "New name" {
		t.Errorf("expected 'New name', got '%s'", item.Name)
	}
}

func TestItem_Rename_EmptyName_ReturnsError(t *testing.T) {
	item, _ := NewItem("Task", 1)
	err := item.Rename("")
	if !errors.Is(err, ErrNameLength) {
		t.Errorf("expected ErrNameLength, got: %v", err)
	}
	// name должно остаться неизменным при ошибке
	if item.Name != "Task" {
		t.Errorf("name should not change on error, got '%s'", item.Name)
	}
}

func TestItem_MoveTo_Valid(t *testing.T) {
	item, _ := NewItem("Task", 1)
	if err := item.MoveTo(5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Position != 5 {
		t.Errorf("expected position 5, got %d", item.Position)
	}
}

func TestItem_MoveTo_ZeroPosition_ReturnsError(t *testing.T) {
	item, _ := NewItem("Task", 1)
	err := item.MoveTo(0)
	if !errors.Is(err, ErrPositionValue) {
		t.Errorf("expected ErrPositionValue, got: %v", err)
	}
	// позиция должна остаться неизменной при ошибке
	if item.Position != 1 {
		t.Errorf("position should not change on error, got %d", item.Position)
	}
}

func TestReconstituteItem(t *testing.T) {
	id, _ := NewModelID("00000000-0000-0000-0000-000000000001")
	item := ReconstituteItem(id, "Task", 3, true, testTime())
	if item.Name != "Task" || item.Position != 3 || !item.Done {
		t.Errorf("ReconstituteItem did not set fields correctly: %+v", item)
	}
}

func testTime() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}

package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// --- Value Object Tests ---

func TestNewName_Valid(t *testing.T) {
	n, err := NewName("Buy milk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.String() != "Buy milk" {
		t.Errorf("expected 'Buy milk', got '%s'", n.String())
	}
}

func TestNewName_Empty_ReturnsError(t *testing.T) {
	_, err := NewName("")
	if !errors.Is(err, ErrNameLength) {
		t.Errorf("expected ErrNameLength, got: %v", err)
	}
}

func TestNewName_TooLong_ReturnsError(t *testing.T) {
	_, err := NewName(strings.Repeat("x", NameMaxLength+1))
	if !errors.Is(err, ErrNameLength) {
		t.Errorf("expected ErrNameLength, got: %v", err)
	}
}

func TestNewName_ExactMinLength(t *testing.T) {
	n, err := NewName("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.String() != "a" {
		t.Errorf("expected 'a', got '%s'", n.String())
	}
}

func TestNewName_ExactMaxLength(t *testing.T) {
	val := strings.Repeat("x", NameMaxLength)
	n, err := NewName(val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.String() != val {
		t.Errorf("expected string of length %d, got length %d", NameMaxLength, len(n.String()))
	}
}

func TestNewPosition_Valid(t *testing.T) {
	p, err := NewPosition(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Int() != 5 {
		t.Errorf("expected 5, got %d", p.Int())
	}
}

func TestNewPosition_Zero_ReturnsError(t *testing.T) {
	_, err := NewPosition(0)
	if !errors.Is(err, ErrPositionValue) {
		t.Errorf("expected ErrPositionValue, got: %v", err)
	}
}

func TestNewPosition_Negative_ReturnsError(t *testing.T) {
	_, err := NewPosition(-1)
	if !errors.Is(err, ErrPositionValue) {
		t.Errorf("expected ErrPositionValue, got: %v", err)
	}
}

func TestNewPosition_MinValue(t *testing.T) {
	p, err := NewPosition(PositionMin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Int() != PositionMin {
		t.Errorf("expected %d, got %d", PositionMin, p.Int())
	}
}

// --- Item Tests ---

func TestNewItem_Valid(t *testing.T) {
	item, err := NewItem("Buy milk", 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name().String() != "Buy milk" {
		t.Errorf("expected name 'Buy milk', got '%s'", item.Name().String())
	}
	if item.Position().Int() != 1 {
		t.Errorf("expected position 1, got %d", item.Position().Int())
	}
	if item.Done() {
		t.Error("new item should not be done")
	}
	id := item.ID()
	if id.IsZero() {
		t.Error("ID should be generated")
	}
	if item.CreatedAt().IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewItem_EmptyName_ReturnsError(t *testing.T) {
	_, err := NewItem("", 1, "")
	if !errors.Is(err, ErrNameLength) {
		t.Errorf("expected ErrNameLength, got: %v", err)
	}
}

func TestNewItem_NameTooLong_ReturnsError(t *testing.T) {
	_, err := NewItem(strings.Repeat("x", NameMaxLength+1), 1, "")
	if !errors.Is(err, ErrNameLength) {
		t.Errorf("expected ErrNameLength, got: %v", err)
	}
}

func TestNewItem_ZeroPosition_ReturnsError(t *testing.T) {
	_, err := NewItem("Task", 0, "")
	if !errors.Is(err, ErrPositionValue) {
		t.Errorf("expected ErrPositionValue, got: %v", err)
	}
}

func TestItem_Complete(t *testing.T) {
	item, _ := NewItem("Task", 1, "")
	if item.Done() {
		t.Fatal("item should not be done initially")
	}
	item.Complete()
	if !item.Done() {
		t.Error("item should be done after Complete()")
	}
}

func TestItem_Uncomplete(t *testing.T) {
	item, _ := NewItem("Task", 1, "")
	item.Complete()
	if !item.Done() {
		t.Fatal("item should be done after Complete()")
	}
	item.Uncomplete()
	if item.Done() {
		t.Error("item should not be done after Uncomplete()")
	}
}

func TestItem_Rename_Valid(t *testing.T) {
	item, _ := NewItem("Old name", 1, "")
	if err := item.Rename("New name"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name().String() != "New name" {
		t.Errorf("expected 'New name', got '%s'", item.Name().String())
	}
}

func TestItem_Rename_EmptyName_ReturnsError(t *testing.T) {
	item, _ := NewItem("Task", 1, "")
	err := item.Rename("")
	if !errors.Is(err, ErrNameLength) {
		t.Errorf("expected ErrNameLength, got: %v", err)
	}
	if item.Name().String() != "Task" {
		t.Errorf("name should not change on error, got '%s'", item.Name().String())
	}
}

func TestItem_MoveTo_Valid(t *testing.T) {
	item, _ := NewItem("Task", 1, "")
	if err := item.MoveTo(5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Position().Int() != 5 {
		t.Errorf("expected position 5, got %d", item.Position().Int())
	}
}

func TestItem_MoveTo_ZeroPosition_ReturnsError(t *testing.T) {
	item, _ := NewItem("Task", 1, "")
	err := item.MoveTo(0)
	if !errors.Is(err, ErrPositionValue) {
		t.Errorf("expected ErrPositionValue, got: %v", err)
	}
	if item.Position().Int() != 1 {
		t.Errorf("position should not change on error, got %d", item.Position().Int())
	}
}

func TestReconstituteItem(t *testing.T) {
	id, _ := NewModelID("00000000-0000-0000-0000-000000000001")
	item := ReconstituteItem(id, "Task", "", 3, true, testTime())
	if item.Name().String() != "Task" || item.Position().Int() != 3 || !item.Done() {
		t.Errorf("ReconstituteItem did not set fields correctly")
	}
}

// --- Description Value Object Tests ---

func TestNewDescription_Valid(t *testing.T) {
	d, err := NewDescription("A detailed task description")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.String() != "A detailed task description" {
		t.Errorf("expected 'A detailed task description', got '%s'", d.String())
	}
}

func TestNewDescription_Empty_Allowed(t *testing.T) {
	d, err := NewDescription("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.String() != "" {
		t.Errorf("expected empty string, got '%s'", d.String())
	}
}

func TestNewDescription_ExactMaxLength(t *testing.T) {
	val := strings.Repeat("x", DescriptionMaxLength)
	d, err := NewDescription(val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(d.String()) != DescriptionMaxLength {
		t.Errorf("expected length %d, got %d", DescriptionMaxLength, len(d.String()))
	}
}

func TestNewDescription_TooLong_ReturnsError(t *testing.T) {
	_, err := NewDescription(strings.Repeat("x", DescriptionMaxLength+1))
	if !errors.Is(err, ErrDescriptionLength) {
		t.Errorf("expected ErrDescriptionLength, got: %v", err)
	}
}

// --- Item with Description Tests ---

func TestNewItem_WithDescription_Valid(t *testing.T) {
	item, err := NewItem("Buy milk", 1, "Get whole milk from store")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Description().String() != "Get whole milk from store" {
		t.Errorf("expected description 'Get whole milk from store', got '%s'", item.Description().String())
	}
}

func TestNewItem_WithEmptyDescription_Valid(t *testing.T) {
	item, err := NewItem("Buy milk", 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Description().String() != "" {
		t.Errorf("expected empty description, got '%s'", item.Description().String())
	}
}

func TestNewItem_WithDescriptionTooLong_ReturnsError(t *testing.T) {
	_, err := NewItem("Buy milk", 1, strings.Repeat("x", DescriptionMaxLength+1))
	if !errors.Is(err, ErrDescriptionLength) {
		t.Errorf("expected ErrDescriptionLength, got: %v", err)
	}
}

func TestItem_ChangeDescription_Valid(t *testing.T) {
	item, _ := NewItem("Task", 1, "old desc")
	desc, _ := NewDescription("new desc")
	item.ChangeDescription(desc)
	if item.Description().String() != "new desc" {
		t.Errorf("expected 'new desc', got '%s'", item.Description().String())
	}
}

func TestReconstituteItem_WithDescription(t *testing.T) {
	id, _ := NewModelID("00000000-0000-0000-0000-000000000001")
	item := ReconstituteItem(id, "Task", "A description", 3, true, testTime())
	if item.Description().String() != "A description" {
		t.Errorf("expected description 'A description', got '%s'", item.Description().String())
	}
}

func testTime() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}

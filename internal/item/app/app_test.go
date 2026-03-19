package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/evgeniy-klemin/todo/internal/item/domain"
)

func testTime() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}

// --- mocks ---

type mockDomainRepository struct {
	getByIDFn             func(ctx context.Context, id domain.ModelID) (*domain.Item, error)
	addFn                 func(ctx context.Context, item *domain.Item) (*domain.Item, error)
	addWithNextPositionFn func(ctx context.Context, item *domain.Item) (*domain.Item, error)
	updateFn              func(ctx context.Context, id domain.ModelID, updater func(*domain.Item) error) (*domain.Item, error)
	listFn                func(ctx context.Context, filter domain.ListFilter, sort []domain.SortField, page, perPage int) ([]*domain.Item, error)
	countFn               func(ctx context.Context, filter domain.ListFilter) (int, error)
	listWithCursorFn      func(ctx context.Context, filter domain.ListFilter, sort []domain.SortField, limit int, cursor *domain.Cursor) ([]*domain.Item, error)
}

func (m *mockDomainRepository) GetByID(ctx context.Context, id domain.ModelID) (*domain.Item, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockDomainRepository) Add(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	if m.addFn != nil {
		return m.addFn(ctx, item)
	}
	return item, nil
}

func (m *mockDomainRepository) AddWithNextPosition(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	if m.addWithNextPositionFn != nil {
		return m.addWithNextPositionFn(ctx, item)
	}
	return item, nil
}

func (m *mockDomainRepository) Update(ctx context.Context, id domain.ModelID, updater func(*domain.Item) error) (*domain.Item, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, updater)
	}
	return nil, nil
}

func (m *mockDomainRepository) List(ctx context.Context, filter domain.ListFilter, sort []domain.SortField, page, perPage int) ([]*domain.Item, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter, sort, page, perPage)
	}
	return nil, nil
}

func (m *mockDomainRepository) Count(ctx context.Context, filter domain.ListFilter) (int, error) {
	if m.countFn != nil {
		return m.countFn(ctx, filter)
	}
	return 0, nil
}

func (m *mockDomainRepository) ListWithCursor(ctx context.Context, filter domain.ListFilter, sort []domain.SortField, limit int, cursor *domain.Cursor) ([]*domain.Item, error) {
	if m.listWithCursorFn != nil {
		return m.listWithCursorFn(ctx, filter, sort, limit, cursor)
	}
	return nil, nil
}

// --- helpers ---

func intPtr(v int) *int { return &v }

func newService(domainRepo domain.Repository) *ItemService {
	return NewItemService(domainRepo)
}

// --- tests ---

func TestCreate_WithExplicitPosition_CallsAdd(t *testing.T) {
	addCalled := false
	addWithNextPositionCalled := false
	domainRepo := &mockDomainRepository{
		addFn: func(_ context.Context, item *domain.Item) (*domain.Item, error) {
			addCalled = true
			return item, nil
		},
		addWithNextPositionFn: func(_ context.Context, item *domain.Item) (*domain.Item, error) {
			addWithNextPositionCalled = true
			return item, nil
		},
	}

	svc := newService(domainRepo)

	item, err := svc.Create(context.Background(), "Task 1", intPtr(5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Position == nil || *item.Position != 5 {
		t.Errorf("expected position 5, got %v", item.Position)
	}
	if !addCalled {
		t.Error("Add should be called when position is explicitly provided")
	}
	if addWithNextPositionCalled {
		t.Error("AddWithNextPosition should not be called when position is explicitly provided")
	}
}

func TestCreate_WithoutPosition_CallsAddWithNextPosition(t *testing.T) {
	addCalled := false
	addWithNextPositionCalled := false
	domainRepo := &mockDomainRepository{
		addFn: func(_ context.Context, item *domain.Item) (*domain.Item, error) {
			addCalled = true
			return item, nil
		},
		addWithNextPositionFn: func(_ context.Context, item *domain.Item) (*domain.Item, error) {
			addWithNextPositionCalled = true
			// Simulate assigning position 8
			if err := item.MoveTo(8); err != nil {
				return nil, err
			}
			return item, nil
		},
	}

	svc := newService(domainRepo)

	item, err := svc.Create(context.Background(), "Task 2", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Position == nil || *item.Position != 8 {
		t.Errorf("expected position 8, got %v", item.Position)
	}
	if addCalled {
		t.Error("Add should not be called when position is nil")
	}
	if !addWithNextPositionCalled {
		t.Error("AddWithNextPosition should be called when position is nil")
	}
}

func TestCreate_EmptyName_ReturnsValidationError(t *testing.T) {
	svc := newService(&mockDomainRepository{})

	_, err := svc.Create(context.Background(), "", intPtr(1))
	if err == nil {
		t.Fatal("expected validation error for empty name, got nil")
	}
	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation, got: %v", err)
	}
	if !errors.Is(err, domain.ErrNameLength) {
		t.Errorf("expected domain.ErrNameLength in chain, got: %v", err)
	}
}

func TestCreate_NameTooLong_ReturnsValidationError(t *testing.T) {
	longName := strings.Repeat("a", domain.NameMaxLength+1)
	svc := newService(&mockDomainRepository{})

	_, err := svc.Create(context.Background(), longName, intPtr(1))
	if err == nil {
		t.Fatal("expected validation error for long name, got nil")
	}
	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation, got: %v", err)
	}
	if !errors.Is(err, domain.ErrNameLength) {
		t.Errorf("expected domain.ErrNameLength in chain, got: %v", err)
	}
}

func TestCreate_AddWithNextPositionError_Propagated(t *testing.T) {
	repoErr := errors.New("db connection lost")
	domainRepo := &mockDomainRepository{
		addWithNextPositionFn: func(_ context.Context, _ *domain.Item) (*domain.Item, error) {
			return nil, repoErr
		},
	}

	svc := newService(domainRepo)

	_, err := svc.Create(context.Background(), "Task 4", nil)
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error to be propagated, got: %v", err)
	}
}

func TestCreate_DomainRepositoryAddError_Propagated(t *testing.T) {
	addErr := errors.New("insert failed")
	domainRepo := &mockDomainRepository{
		addFn: func(_ context.Context, _ *domain.Item) (*domain.Item, error) {
			return nil, addErr
		},
	}

	svc := newService(domainRepo)

	_, err := svc.Create(context.Background(), "Task 5", intPtr(2))
	if !errors.Is(err, addErr) {
		t.Errorf("expected add error to be propagated, got: %v", err)
	}
}

func boolPtr(v bool) *bool { return &v }

func TestUpdate_SetDoneTrue_CallsComplete(t *testing.T) {
	id, _ := domain.NewModelID("00000000-0000-0000-0000-000000000001")
	domainRepo := &mockDomainRepository{
		updateFn: func(_ context.Context, _ domain.ModelID, updater func(*domain.Item) error) (*domain.Item, error) {
			item := domain.ReconstituteItem(id, "Task", 1, false, testTime())
			if err := updater(item); err != nil {
				return nil, err
			}
			return item, nil
		},
	}

	svc := newService(domainRepo)
	result, err := svc.Update(context.Background(), &Item{
		ID:   "00000000-0000-0000-0000-000000000001",
		Done: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Done == nil || !*result.Done {
		t.Error("expected Done to be true after update")
	}
}

func TestUpdate_SetDoneFalse_CallsReopen(t *testing.T) {
	id, _ := domain.NewModelID("00000000-0000-0000-0000-000000000001")
	domainRepo := &mockDomainRepository{
		updateFn: func(_ context.Context, _ domain.ModelID, updater func(*domain.Item) error) (*domain.Item, error) {
			item := domain.ReconstituteItem(id, "Task", 1, true, testTime())
			if err := updater(item); err != nil {
				return nil, err
			}
			return item, nil
		},
	}

	svc := newService(domainRepo)
	result, err := svc.Update(context.Background(), &Item{
		ID:   "00000000-0000-0000-0000-000000000001",
		Done: boolPtr(false),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Done == nil || *result.Done {
		t.Error("expected Done to be false after update")
	}
}

func TestList_PassesSearchToRepository(t *testing.T) {
	var capturedSearch *string
	domainRepo := &mockDomainRepository{
		listFn: func(_ context.Context, filter domain.ListFilter, _ []domain.SortField, _, _ int) ([]*domain.Item, error) {
			capturedSearch = filter.Search
			return []*domain.Item{}, nil
		},
		countFn: func(_ context.Context, filter domain.ListFilter) (int, error) {
			return 0, nil
		},
	}

	svc := newService(domainRepo)
	searchTerm := "buy"
	_, err := svc.List(context.Background(), ListQuery{
		Search:  &searchTerm,
		Page:    1,
		PerPage: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSearch == nil {
		t.Fatal("expected search to be passed to repository, got nil")
	}
	if *capturedSearch != "buy" {
		t.Errorf("expected search 'buy', got '%s'", *capturedSearch)
	}
}

func TestList_NilSearch_PassesNilToRepository(t *testing.T) {
	searchChecked := false
	domainRepo := &mockDomainRepository{
		listFn: func(_ context.Context, filter domain.ListFilter, _ []domain.SortField, _, _ int) ([]*domain.Item, error) {
			searchChecked = true
			if filter.Search != nil {
				t.Errorf("expected nil search, got '%s'", *filter.Search)
			}
			return []*domain.Item{}, nil
		},
		countFn: func(_ context.Context, filter domain.ListFilter) (int, error) {
			return 0, nil
		},
	}

	svc := newService(domainRepo)
	_, err := svc.List(context.Background(), ListQuery{
		Page:    1,
		PerPage: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !searchChecked {
		t.Fatal("List was not called")
	}
}

func TestUpdate_InvalidID_ReturnsError(t *testing.T) {
	svc := newService(&mockDomainRepository{})
	_, err := svc.Update(context.Background(), &Item{
		ID:   "short",
		Done: boolPtr(true),
	})
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
	if !errors.Is(err, ErrValidation) {
		t.Errorf("expected ErrValidation, got: %v", err)
	}
}

func TestGetItemByID_NotFound_ReturnsErrNotFound(t *testing.T) {
	domainRepo := &mockDomainRepository{
		getByIDFn: func(_ context.Context, _ domain.ModelID) (*domain.Item, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := newService(domainRepo)
	_, err := svc.GetItemByID(context.Background(), "00000000-0000-0000-0000-000000000001")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected domain.ErrNotFound in chain, got: %v", err)
	}
}

func TestUpdate_NotFound_ReturnsErrNotFound(t *testing.T) {
	domainRepo := &mockDomainRepository{
		updateFn: func(_ context.Context, _ domain.ModelID, _ func(*domain.Item) error) (*domain.Item, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := newService(domainRepo)
	_, err := svc.Update(context.Background(), &Item{
		ID:   "00000000-0000-0000-0000-000000000001",
		Done: boolPtr(true),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

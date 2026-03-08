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
	addFn    func(ctx context.Context, item *domain.Item) (*domain.Item, error)
	updateFn func(ctx context.Context, id domain.ModelID, updater func(*domain.Item) error) (*domain.Item, error)
}

func (m *mockDomainRepository) GetByID(ctx context.Context, id domain.ModelID) (*domain.Item, error) {
	return nil, nil
}

func (m *mockDomainRepository) Add(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	if m.addFn != nil {
		return m.addFn(ctx, item)
	}
	return item, nil
}

func (m *mockDomainRepository) Update(ctx context.Context, id domain.ModelID, updater func(*domain.Item) error) (*domain.Item, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, updater)
	}
	return nil, nil
}

type mockAppRepository struct {
	allFn          func(ctx context.Context, done *bool, fields []ItemField, page, perPage int, sortFields SortFields) ([]Item, error)
	maxPositionFn  func(ctx context.Context) (int, error)
}

func (m *mockAppRepository) All(ctx context.Context, done *bool, fields []ItemField, page, perPage int, sortFields SortFields) ([]Item, error) {
	if m.allFn != nil {
		return m.allFn(ctx, done, fields, page, perPage, sortFields)
	}
	return nil, nil
}

func (m *mockAppRepository) Count(ctx context.Context, done *bool) (int, error) {
	return 0, nil
}

func (m *mockAppRepository) MaxPosition(ctx context.Context) (int, error) {
	if m.maxPositionFn != nil {
		return m.maxPositionFn(ctx)
	}
	return 0, nil
}

// --- helpers ---

func intPtr(v int) *int { return &v }

func newService(domainRepo domain.Repository, appRepo Repository) *ItemService {
	return NewItemService(domainRepo, appRepo)
}

// --- tests ---

func TestCreate_WithExplicitPosition(t *testing.T) {
	appRepoCallCount := 0
	appRepo := &mockAppRepository{
		allFn: func(_ context.Context, _ *bool, _ []ItemField, _, _ int, _ SortFields) ([]Item, error) {
			appRepoCallCount++
			return nil, nil
		},
	}

	svc := newService(&mockDomainRepository{}, appRepo)

	item, err := svc.Create(context.Background(), "Task 1", intPtr(5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Position != 5 {
		t.Errorf("expected position 5, got %d", item.Position)
	}
	if appRepoCallCount != 0 {
		t.Errorf("appRepository.All should not be called when position is provided")
	}
}

func TestCreate_WithoutPosition_UsesMaxPlusOne(t *testing.T) {
	appRepo := &mockAppRepository{
		maxPositionFn: func(_ context.Context) (int, error) {
			return 7, nil
		},
	}

	svc := newService(&mockDomainRepository{}, appRepo)

	item, err := svc.Create(context.Background(), "Task 2", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Position != 8 {
		t.Errorf("expected position 8, got %d", item.Position)
	}
}

func TestCreate_WithoutPosition_NoExistingItems_PositionIsOne(t *testing.T) {
	appRepo := &mockAppRepository{
		maxPositionFn: func(_ context.Context) (int, error) {
			return 0, nil
		},
	}

	svc := newService(&mockDomainRepository{}, appRepo)

	item, err := svc.Create(context.Background(), "Task 3", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Position != 1 {
		t.Errorf("expected position 1, got %d", item.Position)
	}
}

func TestCreate_EmptyName_ReturnsValidationError(t *testing.T) {
	svc := newService(&mockDomainRepository{}, &mockAppRepository{})

	_, err := svc.Create(context.Background(), "", intPtr(1))
	if err == nil {
		t.Fatal("expected validation error for empty name, got nil")
	}
	if !errors.Is(err, domain.ErrNameLength) {
		t.Errorf("expected ErrNameLength, got: %v", err)
	}
}

func TestCreate_NameTooLong_ReturnsValidationError(t *testing.T) {
	longName := strings.Repeat("a", domain.NameMaxLength+1)
	svc := newService(&mockDomainRepository{}, &mockAppRepository{})

	_, err := svc.Create(context.Background(), longName, intPtr(1))
	if err == nil {
		t.Fatal("expected validation error for long name, got nil")
	}
	if !errors.Is(err, domain.ErrNameLength) {
		t.Errorf("expected ErrNameLength, got: %v", err)
	}
}

func TestCreate_AppRepositoryMaxPositionError_Propagated(t *testing.T) {
	repoErr := errors.New("db connection lost")
	appRepo := &mockAppRepository{
		maxPositionFn: func(_ context.Context) (int, error) {
			return 0, repoErr
		},
	}

	svc := newService(&mockDomainRepository{}, appRepo)

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

	svc := newService(domainRepo, &mockAppRepository{})

	_, err := svc.Create(context.Background(), "Task 5", intPtr(2))
	if !errors.Is(err, addErr) {
		t.Errorf("expected add error to be propagated, got: %v", err)
	}
}

func TestCreate_WithExplicitPosition_MaxPositionNotCalled(t *testing.T) {
	maxPositionCalled := false
	appRepo := &mockAppRepository{
		maxPositionFn: func(_ context.Context) (int, error) {
			maxPositionCalled = true
			return 0, nil
		},
	}

	svc := newService(&mockDomainRepository{}, appRepo)
	_, _ = svc.Create(context.Background(), "Task", intPtr(3))

	if maxPositionCalled {
		t.Error("MaxPosition should not be called when position is explicitly provided")
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

	svc := newService(domainRepo, &mockAppRepository{})
	result, err := svc.Update(context.Background(), &Item{
		ID:   "00000000-0000-0000-0000-000000000001",
		Done: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Done {
		t.Error("expected Done to be true after update")
	}
}

func TestUpdate_SetDoneFalse_CallsUncomplete(t *testing.T) {
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

	svc := newService(domainRepo, &mockAppRepository{})
	result, err := svc.Update(context.Background(), &Item{
		ID:   "00000000-0000-0000-0000-000000000001",
		Done: boolPtr(false),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Done {
		t.Error("expected Done to be false after update")
	}
}

func TestUpdate_InvalidID_ReturnsError(t *testing.T) {
	svc := newService(&mockDomainRepository{}, &mockAppRepository{})
	_, err := svc.Update(context.Background(), &Item{
		ID:   "short",
		Done: boolPtr(true),
	})
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}

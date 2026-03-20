package app

import (
	"context"
	"errors"
	"log/slog"

	"github.com/evgeniy-klemin/todo/internal/item/domain"
)

type ItemService struct {
	repo domain.Repository
}

func NewItemService(repo domain.Repository) *ItemService {
	return &ItemService{
		repo: repo,
	}
}

func (s *ItemService) GetItemByID(ctx context.Context, id string) (*Item, error) {
	modelID, err := domain.NewModelID(id)
	if err != nil {
		return nil, Validation("get item by id", err)
	}
	item, err := s.repo.GetByID(ctx, modelID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, NotFound("get item by id", err)
		}
		return nil, err
	}
	return domainToAppItem(item, nil), nil
}

func (s *ItemService) Create(ctx context.Context, name string, position *int) (*Item, error) {
	if position != nil {
		item, err := domain.NewItem(name, *position)
		if err != nil {
			slog.WarnContext(ctx, "NewItem validation failed", "name", name, "position", *position, "error", err)
			return nil, Validation("create item", err)
		}
		result, err := s.repo.Add(ctx, item)
		if err != nil {
			slog.ErrorContext(ctx, "repo.Add failed", "error", err)
			return nil, err
		}
		resultID := result.ID()
		slog.InfoContext(ctx, "item created", "id", resultID.String(), "position", result.Position().Int())
		return domainToAppItem(result, nil), nil
	}

	// Auto-position: create with placeholder, repo will assign real position atomically
	item, err := domain.NewItem(name, 1)
	if err != nil {
		slog.WarnContext(ctx, "NewItem validation failed", "name", name, "error", err)
		return nil, Validation("create item", err)
	}
	result, err := s.repo.AddWithNextPosition(ctx, item)
	if err != nil {
		slog.ErrorContext(ctx, "repo.AddWithNextPosition failed", "error", err)
		return nil, err
	}
	resultID := result.ID()
	slog.InfoContext(ctx, "item created", "id", resultID.String(), "position", result.Position().Int())
	return domainToAppItem(result, nil), nil
}

// All fetches items using cursor-based pagination.
// cursorData is an opaque []byte returned by a previous call (nil = from start).
// Returns items, an opaque nextCursor []byte for the last item (nil if no items), and any error.
// The caller (ports layer) is responsible for deciding whether there is a next page and
// whether to use this cursor, based on how many items were requested vs returned.
func (s *ItemService) All(ctx context.Context, done *bool, search *string, fields []ItemField, limit int, cursorData []byte, sortFields SortFields) ([]Item, []byte, error) {
	filter := domain.ListFilter{Done: done, Search: search}
	domainSort := appSortFieldsToDomain(sortFields)

	domainItems, err := s.repo.ListWithCursor(ctx, filter, domainSort, limit, cursorData)
	if err != nil {
		return nil, nil, err
	}
	items := make([]Item, 0, len(domainItems))
	for _, d := range domainItems {
		items = append(items, *domainToAppItem(d, fields))
	}
	// Return cursor for the second-to-last item when multiple items were fetched.
	// The caller (ports) fetches limit=perPage+1 to detect hasNext; item[len-1] is
	// the probe. The cursor must point to item[len-2] = last displayed item, so
	// the next page starts after it and includes the probe item.
	// When only one item is returned, use it as the cursor anchor.
	var nextCursor []byte
	if len(domainItems) > 1 {
		nextCursor, err = s.repo.BuildCursor(domainItems[len(domainItems)-2], domainSort)
		if err != nil {
			return nil, nil, err
		}
	} else if len(domainItems) == 1 {
		nextCursor, err = s.repo.BuildCursor(domainItems[0], domainSort)
		if err != nil {
			return nil, nil, err
		}
	}
	return items, nextCursor, nil
}

func (s *ItemService) Count(ctx context.Context, done *bool, search *string) (int, error) {
	filter := domain.ListFilter{Done: done, Search: search}
	return s.repo.Count(ctx, filter)
}

func (s *ItemService) Update(ctx context.Context, reqItem *Item) (*Item, error) {
	modelID, err := domain.NewModelID(reqItem.ID)
	if err != nil {
		return nil, Validation("update item", err)
	}
	result, err := s.repo.Update(ctx, modelID, func(item *domain.Item) error {
		if reqItem.Done != nil {
			if *reqItem.Done {
				item.Complete()
			} else {
				item.Reopen()
			}
		}
		if reqItem.Position != nil {
			if err := item.MoveTo(*reqItem.Position); err != nil {
				return Validation("update item position", err)
			}
		}
		if reqItem.Name != nil {
			if err := item.Rename(*reqItem.Name); err != nil {
				return Validation("update item name", err)
			}
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, NotFound("update item", err)
		}
		return nil, err
	}
	return domainToAppItem(result, nil), nil
}

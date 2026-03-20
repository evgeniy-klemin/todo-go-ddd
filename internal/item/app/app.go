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

// All fetches items using cursor-based pagination and returns a cursor for the next page.
//
// Parameters:
//   - done: optional boolean filter; nil means no filter, true/false filters by done state.
//   - search: optional full-text search string; nil or empty means no text filter.
//   - fields: subset of item fields to populate in the returned items (nil = all fields).
//   - limit: maximum number of items to fetch; callers typically pass perPage+1 so that
//     the presence of an extra item signals that a next page exists.
//   - cursorData: opaque cursor from the previous call; pass nil to start from the beginning.
//   - sortFields: ordered list of sort columns and directions; must be consistent across pages.
//
// Returns:
//   - items: up to limit domain items converted to app.Item.
//   - nextCursor: opaque []byte pointing to item[len-2] when len > 1, or item[0] when len == 1,
//     or nil when no items were returned. The ports layer decides whether to expose this cursor
//     based on whether len(items) > perPage.
//   - err: any repository or serialization error.
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

// Count returns the total number of items matching the given filter.
// done and search have the same meaning as in All. The result is used
// to populate the X-Total-Count response header and is independent of pagination.
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

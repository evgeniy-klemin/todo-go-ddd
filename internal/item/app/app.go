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

func (s *ItemService) List(ctx context.Context, query ListQuery) (ListResult, error) {
	filter := domain.ListFilter{
		Done:   query.Done,
		Search: query.Search,
	}
	sortFields := appSortFieldsToDomain(query.SortFields)

	count, err := s.repo.Count(ctx, filter)
	if err != nil {
		return ListResult{}, err
	}
	domainItems, err := s.repo.List(ctx, filter, sortFields, query.Page, query.PerPage)
	if err != nil {
		return ListResult{}, err
	}

	items := make([]Item, 0, len(domainItems))
	for _, d := range domainItems {
		items = append(items, *domainToAppItem(d, query.Fields))
	}
	return ListResult{Items: items, TotalCount: count}, nil
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

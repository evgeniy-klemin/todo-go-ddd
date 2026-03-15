package app

import (
	"context"
	"errors"
	"log/slog"

	"github.com/evgeniy-klemin/todo/internal/item/domain"
)

type ItemService struct {
	domainRepository domain.Repository
	queryRepository  QueryRepository
}

func NewItemService(domainRepository domain.Repository, queryRepository QueryRepository) *ItemService {
	return &ItemService{
		domainRepository: domainRepository,
		queryRepository:  queryRepository,
	}
}

func domainToAppItem(d *domain.Item) *Item {
	name := d.Name().String()
	position := d.Position().Int()
	done := d.Done()
	createdAt := d.CreatedAt()
	id := d.ID()
	return &Item{
		ID:        id.String(),
		Name:      &name,
		Position:  &position,
		Done:      &done,
		CreatedAt: &createdAt,
	}
}

func (s *ItemService) GetItemByID(ctx context.Context, id string) (*Item, error) {
	modelID, err := domain.NewModelID(id)
	if err != nil {
		return nil, Validation("get item by id", err)
	}
	item, err := s.domainRepository.GetByID(ctx, modelID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, NotFound("get item by id", err)
		}
		return nil, err
	}
	return domainToAppItem(item), nil
}

func (s *ItemService) Create(ctx context.Context, name string, position *int) (*Item, error) {
	if position != nil {
		item, err := domain.NewItem(name, *position)
		if err != nil {
			slog.WarnContext(ctx, "NewItem validation failed", "name", name, "position", *position, "error", err)
			return nil, Validation("create item", err)
		}
		result, err := s.domainRepository.Add(ctx, item)
		if err != nil {
			slog.ErrorContext(ctx, "domainRepository.Add failed", "error", err)
			return nil, err
		}
		resultID := result.ID()
		slog.InfoContext(ctx, "item created", "id", resultID.String(), "position", result.Position().Int())
		return domainToAppItem(result), nil
	}

	// Auto-position: create with placeholder, repo will assign real position atomically
	item, err := domain.NewItem(name, 1)
	if err != nil {
		slog.WarnContext(ctx, "NewItem validation failed", "name", name, "error", err)
		return nil, Validation("create item", err)
	}
	result, err := s.domainRepository.AddWithNextPosition(ctx, item)
	if err != nil {
		slog.ErrorContext(ctx, "domainRepository.AddWithNextPosition failed", "error", err)
		return nil, err
	}
	resultID := result.ID()
	slog.InfoContext(ctx, "item created", "id", resultID.String(), "position", result.Position().Int())
	return domainToAppItem(result), nil
}

func (s *ItemService) List(ctx context.Context, query ListQuery) (ListResult, error) {
	count, err := s.queryRepository.Count(ctx, query.Done, query.Search)
	if err != nil {
		return ListResult{}, err
	}
	items, err := s.queryRepository.All(ctx, query.Done, query.Search, query.Fields, query.Page, query.PerPage, query.SortFields)
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, TotalCount: count}, nil
}

func (s *ItemService) Update(ctx context.Context, reqItem *Item) (*Item, error) {
	modelID, err := domain.NewModelID(reqItem.ID)
	if err != nil {
		return nil, Validation("update item", err)
	}
	result, err := s.domainRepository.Update(ctx, modelID, func(item *domain.Item) error {
		if reqItem.Done != nil {
			if *reqItem.Done {
				item.Complete()
			} else {
				item.Uncomplete()
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
	return domainToAppItem(result), nil
}

package app

import (
	"context"
	"log/slog"

	"github.com/evgeniy-klemin/todo/internal/item/domain"
)

type ItemService struct {
	domainRepository domain.Repository
	appRepository    Repository
}

func NewItemService(domainRepository domain.Repository, appRepository Repository) *ItemService {
	return &ItemService{
		domainRepository: domainRepository,
		appRepository:    appRepository,
	}
}

func (s *ItemService) GetItemByID(ctx context.Context, id string) (*domain.Item, error) {
	modelID, err := domain.NewModelID(id)
	if err != nil {
		return nil, err
	}
	return s.domainRepository.GetByID(ctx, modelID)
}

func (s *ItemService) Create(ctx context.Context, name string, position *int) (*domain.Item, error) {
	pos, err := s.resolvePosition(ctx, position)
	if err != nil {
		slog.ErrorContext(ctx, "resolvePosition failed", "error", err)
		return nil, err
	}
	item, err := domain.NewItem(name, pos)
	if err != nil {
		slog.WarnContext(ctx, "NewItem validation failed", "name", name, "position", pos, "error", err)
		return nil, err
	}
	result, err := s.domainRepository.Add(ctx, item)
	if err != nil {
		slog.ErrorContext(ctx, "domainRepository.Add failed", "error", err)
		return nil, err
	}
	slog.InfoContext(ctx, "item created", "id", result.ID.String(), "position", result.Position)
	return result, nil
}

func (s *ItemService) resolvePosition(ctx context.Context, position *int) (int, error) {
	if position != nil {
		return *position, nil
	}
	max, err := s.appRepository.MaxPosition(ctx)
	if err != nil {
		return 0, err
	}
	return max + 1, nil
}

func (s *ItemService) List(ctx context.Context, query ListQuery) (ListResult, error) {
	count, err := s.appRepository.Count(ctx, query.Done)
	if err != nil {
		return ListResult{}, err
	}
	items, err := s.appRepository.All(ctx, query.Done, query.Fields, query.Page, query.PerPage, query.SortFields)
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, TotalCount: count}, nil
}

func (s *ItemService) Update(ctx context.Context, reqItem *Item) (*domain.Item, error) {
	modelID, err := domain.NewModelID(reqItem.ID)
	if err != nil {
		return nil, err
	}
	return s.domainRepository.Update(ctx, modelID, func(item *domain.Item) error {
		if reqItem.Done != nil {
			if *reqItem.Done {
				item.Complete()
			} else {
				item.Uncomplete()
			}
		}
		if reqItem.Position != nil {
			if err := item.MoveTo(*reqItem.Position); err != nil {
				return err
			}
		}
		if reqItem.Name != nil {
			if err := item.Rename(*reqItem.Name); err != nil {
				return err
			}
		}
		return nil
	})
}

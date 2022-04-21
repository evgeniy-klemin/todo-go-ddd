package app

import (
	"context"
	"time"

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
	positionVal := 0
	if position != nil {
		positionVal = *position
	} else {
		items, err := s.appRepository.All(ctx, nil, []ItemField{ItemFieldPosition}, 1, 1, SortFields{SortField{Field: ItemFieldPosition, SortDirection: SortDirectionDesc}})
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if *item.Position > positionVal {
				positionVal = *item.Position
			}
		}
		positionVal += 1
	}

	id, err := domain.GenerateModelID()
	if err != nil {
		return nil, err
	}

	item := &domain.Item{
		ID:        id,
		Name:      name,
		Position:  positionVal,
		Done:      false,
		CreatedAt: time.Now().Truncate(time.Second),
	}
	if err := item.Validate(); err != nil {
		return nil, err
	}
	return s.domainRepository.Add(ctx, item)
}

func (s *ItemService) All(ctx context.Context, done *bool, fields []ItemField, page, perPage int, sortFields SortFields) ([]Item, error) {
	return s.appRepository.All(ctx, done, fields, page, perPage, sortFields)
}

func (s *ItemService) Count(ctx context.Context, done *bool) (int, error) {
	return s.appRepository.Count(ctx, done)
}

func (s *ItemService) Update(ctx context.Context, reqItem *Item) (*domain.Item, error) {
	modelID, err := domain.NewModelID(reqItem.ID)
	if err != nil {
		return nil, err
	}
	return s.domainRepository.Update(ctx, modelID, func(item *domain.Item) error {
		if reqItem.Done != nil {
			item.Done = *reqItem.Done
		}
		if reqItem.Position != nil {
			item.Position = *reqItem.Position
		}
		if reqItem.Name != nil {
			item.Name = *reqItem.Name
		}
		return nil
	})
}

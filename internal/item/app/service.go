package app

import (
	"context"

	"github.com/evgeniy-klemin/todo/internal/item/domain"
)

type Service interface {
	Create(ctx context.Context, name string, position *int) (*domain.Item, error)
	GetItemByID(ctx context.Context, id string) (*domain.Item, error)
	List(ctx context.Context, query ListQuery) (ListResult, error)
	Update(ctx context.Context, reqItem *Item) (*domain.Item, error)
}

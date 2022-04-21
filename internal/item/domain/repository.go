package domain

import "context"

type Repository interface {
	GetByID(ctx context.Context, id ModelID) (*Item, error)
	Add(ctx context.Context, item *Item) (*Item, error)
	Update(ctx context.Context, id ModelID, updater func(item *Item) error) (*Item, error)
}

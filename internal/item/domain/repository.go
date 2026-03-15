package domain

import "context"

type SortDirection int

const (
	SortAsc SortDirection = iota + 1
	SortDesc
)

type SortField struct {
	Field     string
	Direction SortDirection
}

type ListFilter struct {
	Done   *bool
	Search *string
}

type Repository interface {
	GetByID(ctx context.Context, id ModelID) (*Item, error)
	Add(ctx context.Context, item *Item) (*Item, error)
	AddWithNextPosition(ctx context.Context, item *Item) (*Item, error)
	Update(ctx context.Context, id ModelID, updater func(item *Item) error) (*Item, error)
	List(ctx context.Context, filter ListFilter, sort []SortField, page, perPage int) ([]*Item, error)
	Count(ctx context.Context, filter ListFilter) (int, error)
}

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
	// ListWithCursor fetches up to limit items after the given opaque cursor (nil = from start).
	// cursorData is a JSON-encoded cursor serialized by the app layer.
	ListWithCursor(ctx context.Context, filter ListFilter, sort []SortField, limit int, cursorData []byte) ([]*Item, error)
}

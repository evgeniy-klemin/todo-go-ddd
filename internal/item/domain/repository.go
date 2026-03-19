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

// Cursor is an opaque pagination cursor passed from the app layer.
// It is defined here to avoid a circular dependency; the app layer
// wraps it with encoding/decoding helpers.
type Cursor struct {
	Values []CursorValue
	ID     string
}

// CursorValue holds a single sort-field value for cursor pagination.
type CursorValue struct {
	Field     string
	Value     interface{}
	Direction string // "asc" or "desc"
}

type Repository interface {
	GetByID(ctx context.Context, id ModelID) (*Item, error)
	Add(ctx context.Context, item *Item) (*Item, error)
	AddWithNextPosition(ctx context.Context, item *Item) (*Item, error)
	Update(ctx context.Context, id ModelID, updater func(item *Item) error) (*Item, error)
	List(ctx context.Context, filter ListFilter, sort []SortField, page, perPage int) ([]*Item, error)
	Count(ctx context.Context, filter ListFilter) (int, error)
	// ListWithCursor fetches up to limit items after the given cursor (nil = from start).
	ListWithCursor(ctx context.Context, filter ListFilter, sort []SortField, limit int, cursor *Cursor) ([]*Item, error)
}

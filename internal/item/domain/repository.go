package domain

import "context"

// SortDirection indicates the ordering direction for list queries.
type SortDirection int

const (
	SortAsc SortDirection = iota + 1
	SortDesc
)

// SortField specifies a field and its ordering direction for list queries.
type SortField struct {
	Field     string
	Direction SortDirection
}

// ListFilter holds optional filter criteria for narrowing list query results.
type ListFilter struct {
	Done   *bool
	Search *string
}

// Repository defines the persistence interface for Item aggregates.
type Repository interface {
	GetByID(ctx context.Context, id ModelID) (*Item, error)
	Add(ctx context.Context, item *Item) (*Item, error)
	AddWithNextPosition(ctx context.Context, item *Item) (*Item, error)
	Update(ctx context.Context, id ModelID, updater func(item *Item) error) (*Item, error)
	List(ctx context.Context, filter ListFilter, sort []SortField, page, perPage int) ([]*Item, error)
	Count(ctx context.Context, filter ListFilter) (int, error)
}

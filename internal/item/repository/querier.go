package repository

import (
	"context"
	"database/sql"
	"time"
)

// dbItem is the internal transfer object between adapters and the repository.
type dbItem struct {
	ID        string
	Name      string
	Position  int64
	Done      bool
	CreatedAt time.Time
}

// listFilter carries the filter parameters for List/Count queries.
// It mirrors domain.ListFilter but avoids importing domain types into the adapter layer.
type listFilter struct {
	Done   *bool
	Search *string
}

// querier abstracts over the sqlc-generated query types for both SQLite and MySQL,
// allowing the repository to stay database-agnostic.
type querier interface {
	GetItemByID(ctx context.Context, id string) (dbItem, error)
	InsertItem(ctx context.Context, id, name string, position int64, done bool, createdAt time.Time) error
	UpdateItem(ctx context.Context, name string, position int64, done bool, id string) error
	MaxPosition(ctx context.Context) (int64, error)
	// ListItems executes a SELECT query applying filter, ORDER BY, LIMIT, OFFSET.
	ListItems(ctx context.Context, filter listFilter, orderBy string, limit, offset int) ([]dbItem, error)
	// CountItems executes a COUNT query applying filter.
	CountItems(ctx context.Context, filter listFilter) (int, error)
	// WithTx returns a new querier backed by the given transaction.
	WithTx(tx *sql.Tx) querier
}

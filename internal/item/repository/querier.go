package repository

import (
	"context"
	"database/sql"
	"strings"
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

// sortField carries a single ORDER BY field and direction.
type sortField struct {
	Field string
	Desc  bool
}

// buildOrderBy converts a slice of sortField into an ORDER BY clause string.
// Defaults to "position asc" when the slice is empty.
func buildOrderBy(sort []sortField) string {
	if len(sort) == 0 {
		return "position asc"
	}
	parts := make([]string, len(sort))
	for i, s := range sort {
		if s.Desc {
			parts[i] = s.Field + " desc"
		} else {
			parts[i] = s.Field + " asc"
		}
	}
	return strings.Join(parts, ", ")
}

// cursorValue holds a single sort-field value for cursor-based pagination.
type cursorValue struct {
	Field     string
	Value     interface{}
	Direction string // "asc" or "desc"
}

// cursorParam is the adapter-layer cursor type (mirrors domain.Cursor without importing it).
type cursorParam struct {
	Values []cursorValue
	ID     string
}

// querier abstracts over the sqlc-generated query types for both SQLite and MySQL,
// allowing the repository to stay database-agnostic.
type querier interface {
	GetItemByID(ctx context.Context, id string) (dbItem, error)
	InsertItem(ctx context.Context, id, name string, position int64, done bool, createdAt time.Time) error
	UpdateItem(ctx context.Context, name string, position int64, done bool, id string) error
	MaxPosition(ctx context.Context) (int64, error)
	// ListItems executes a SELECT query applying filter, ORDER BY, LIMIT, OFFSET.
	ListItems(ctx context.Context, filter listFilter, sort []sortField, limit, offset int) ([]dbItem, error)
	// ListItemsWithCursor executes a SELECT with cursor-based WHERE clause instead of OFFSET.
	ListItemsWithCursor(ctx context.Context, filter listFilter, sort []sortField, limit int, cursor *cursorParam) ([]dbItem, error)
	// CountItems executes a COUNT query applying filter.
	CountItems(ctx context.Context, filter listFilter) (int, error)
	// WithTx returns a new querier backed by the given transaction.
	WithTx(tx *sql.Tx) querier
}

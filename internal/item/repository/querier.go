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

// querier abstracts over the sqlc-generated query types for both SQLite and MySQL,
// allowing the repository to stay database-agnostic.
type querier interface {
	GetItemByID(ctx context.Context, id string) (dbItem, error)
	InsertItem(ctx context.Context, id, name string, position int64, done bool, createdAt time.Time) error
	UpdateItem(ctx context.Context, name string, position int64, done bool, id string) error
	MaxPosition(ctx context.Context) (int64, error)
	// WithTx returns a new querier backed by the given transaction.
	WithTx(tx *sql.Tx) querier
}


package repository

import (
	"context"
	"database/sql"
	"fmt"
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

// toInt64 converts the interface{} returned by MaxPosition queries to int64.
// SQLite returns int64 for integer columns; MySQL may return []byte (encoded as ASCII digits).
func toInt64(v interface{}) (int64, error) {
	switch n := v.(type) {
	case int64:
		return n, nil
	case int32:
		return int64(n), nil
	case []byte:
		// MySQL sometimes encodes numeric results as a byte slice of ASCII digits.
		var result int64
		for _, b := range n {
			if b < '0' || b > '9' {
				return 0, fmt.Errorf("unexpected byte %q in max_position value", b)
			}
			result = result*10 + int64(b-'0')
		}
		return result, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unexpected type %T for max_position", v)
	}
}

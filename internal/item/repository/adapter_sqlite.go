package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/evgeniy-klemin/todo/internal/item/repository/sqlitedb"
)

type sqliteAdapter struct {
	q *sqlitedb.Queries
}

func newSQLiteAdapter(db *sql.DB) *sqliteAdapter {
	return &sqliteAdapter{q: sqlitedb.New(db)}
}

func (a *sqliteAdapter) GetItemByID(ctx context.Context, id string) (dbItem, error) {
	item, err := a.q.GetItemByID(ctx, id)
	if err != nil {
		return dbItem{}, err
	}
	return dbItem{
		ID:        item.ID,
		Name:      item.Name,
		Position:  item.Position,
		Done:      item.Done,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (a *sqliteAdapter) InsertItem(ctx context.Context, id, name string, position int64, done bool, createdAt time.Time) error {
	return a.q.InsertItem(ctx, sqlitedb.InsertItemParams{
		ID:        id,
		Name:      name,
		Position:  position,
		Done:      done,
		CreatedAt: createdAt,
	})
}

func (a *sqliteAdapter) UpdateItem(ctx context.Context, name string, position int64, done bool, id string) error {
	return a.q.UpdateItem(ctx, sqlitedb.UpdateItemParams{
		Name:     name,
		Position: position,
		Done:     done,
		ID:       id,
	})
}

func (a *sqliteAdapter) MaxPosition(ctx context.Context) (int64, error) {
	v, err := a.q.MaxPosition(ctx)
	if err != nil {
		return 0, err
	}
	return toInt64(v)
}

func (a *sqliteAdapter) WithTx(tx *sql.Tx) querier {
	return &sqliteAdapter{q: a.q.WithTx(tx)}
}

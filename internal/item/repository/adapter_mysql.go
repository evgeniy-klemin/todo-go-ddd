package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/evgeniy-klemin/todo/internal/item/repository/mysqldb"
)

type mysqlAdapter struct {
	q *mysqldb.Queries
}

func newMySQLAdapter(db *sql.DB) *mysqlAdapter {
	return &mysqlAdapter{q: mysqldb.New(db)}
}

func (a *mysqlAdapter) GetItemByID(ctx context.Context, id string) (dbItem, error) {
	item, err := a.q.GetItemByID(ctx, id)
	if err != nil {
		return dbItem{}, err
	}
	return dbItem{
		ID:        item.ID,
		Name:      item.Name,
		Position:  int64(item.Position), // MySQL uses int32; normalize to int64
		Done:      item.Done,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (a *mysqlAdapter) InsertItem(ctx context.Context, id, name string, position int64, done bool, createdAt time.Time) error {
	return a.q.InsertItem(ctx, mysqldb.InsertItemParams{
		ID:        id,
		Name:      name,
		Position:  int32(position), // MySQL schema uses INT (int32)
		Done:      done,
		CreatedAt: createdAt,
	})
}

func (a *mysqlAdapter) UpdateItem(ctx context.Context, name string, position int64, done bool, id string) error {
	return a.q.UpdateItem(ctx, mysqldb.UpdateItemParams{
		Name:     name,
		Position: int32(position), // MySQL schema uses INT (int32)
		Done:     done,
		ID:       id,
	})
}

func (a *mysqlAdapter) MaxPosition(ctx context.Context) (int64, error) {
	v, err := a.q.MaxPosition(ctx)
	if err != nil {
		return 0, err
	}
	return toInt64(v)
}

func (a *mysqlAdapter) WithTx(tx *sql.Tx) querier {
	return &mysqlAdapter{q: a.q.WithTx(tx)}
}

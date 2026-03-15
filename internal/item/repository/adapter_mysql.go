package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
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
	if position > math.MaxInt32 {
		return fmt.Errorf("position %d overflows int32", position)
	}
	return a.q.InsertItem(ctx, mysqldb.InsertItemParams{
		ID:        id,
		Name:      name,
		Position:  int32(position), // MySQL schema uses INT (int32)
		Done:      done,
		CreatedAt: createdAt,
	})
}

func (a *mysqlAdapter) UpdateItem(ctx context.Context, name string, position int64, done bool, id string) error {
	if position > math.MaxInt32 {
		return fmt.Errorf("position %d overflows int32", position)
	}
	return a.q.UpdateItem(ctx, mysqldb.UpdateItemParams{
		Name:     name,
		Position: int32(position), // MySQL schema uses INT (int32)
		Done:     done,
		ID:       id,
	})
}

func (a *mysqlAdapter) MaxPosition(ctx context.Context) (int64, error) {
	return a.q.MaxPosition(ctx)
}

func (a *mysqlAdapter) WithTx(tx *sql.Tx) querier {
	return &mysqlAdapter{q: a.q.WithTx(tx)}
}

func (a *mysqlAdapter) SearchCondition(search string, ftsEnabled bool) (string, interface{}) {
	if ftsEnabled {
		return "MATCH(name) AGAINST(? IN BOOLEAN MODE)", buildMySQLFTSQuery(search)
	}
	return "LOWER(name) LIKE LOWER(?)", "%" + search + "%"
}

// buildMySQLFTSQuery converts a user search string into a MySQL FULLTEXT boolean mode query
// with prefix matching and AND logic.
// Example: "buy milk" -> "+buy* +milk*"
func buildMySQLFTSQuery(search string) string {
	words := strings.Fields(search)
	sanitized := make([]string, 0, len(words))
	for _, word := range words {
		word = strings.ReplaceAll(word, "+", "")
		word = strings.ReplaceAll(word, "-", "")
		word = strings.ReplaceAll(word, "*", "")
		if word == "" {
			continue
		}
		sanitized = append(sanitized, "+"+word+"*")
	}
	return strings.Join(sanitized, " ")
}

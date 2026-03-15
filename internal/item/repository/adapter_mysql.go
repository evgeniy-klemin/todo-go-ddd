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
	q          *mysqldb.Queries
	db         mysqldb.DBTX
	ftsEnabled bool
}

func newMySQLAdapter(db *sql.DB, ftsEnabled bool) *mysqlAdapter {
	return &mysqlAdapter{q: mysqldb.New(db), db: db, ftsEnabled: ftsEnabled}
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
	return &mysqlAdapter{q: a.q.WithTx(tx), db: tx, ftsEnabled: a.ftsEnabled}
}

func (a *mysqlAdapter) ListItems(ctx context.Context, filter listFilter, orderBy string, limit, offset int) ([]dbItem, error) {
	var conditions []string
	var args []interface{}

	if filter.Done != nil {
		conditions = append(conditions, "done=?")
		args = append(args, *filter.Done)
	}
	if filter.Search != nil && *filter.Search != "" {
		cond, arg := a.searchCondition(*filter.Search)
		conditions = append(conditions, cond)
		args = append(args, arg)
	}

	q := "SELECT id, name, position, done, created_at FROM item"
	if len(conditions) > 0 {
		q += " WHERE " + strings.Join(conditions, " AND ")
	}
	q += " ORDER BY " + orderBy
	q += " LIMIT ? OFFSET ?"
	queryArgs := make([]interface{}, len(args), len(args)+2)
	copy(queryArgs, args)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := a.db.QueryContext(ctx, q, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []dbItem
	for rows.Next() {
		var item dbItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Position, &item.Done, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("list items: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	return result, nil
}

func (a *mysqlAdapter) CountItems(ctx context.Context, filter listFilter) (int, error) {
	var conditions []string
	var args []interface{}

	if filter.Done != nil {
		conditions = append(conditions, "done=?")
		args = append(args, *filter.Done)
	}
	if filter.Search != nil && *filter.Search != "" {
		cond, arg := a.searchCondition(*filter.Search)
		conditions = append(conditions, cond)
		args = append(args, arg)
	}

	q := "SELECT COUNT(*) FROM item"
	if len(conditions) > 0 {
		q += " WHERE " + strings.Join(conditions, " AND ")
	}
	var count int
	if err := a.db.QueryRowContext(ctx, q, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count items: %w", err)
	}
	return count, nil
}

func (a *mysqlAdapter) searchCondition(search string) (string, interface{}) {
	if a.ftsEnabled {
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

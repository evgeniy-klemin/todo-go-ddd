package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/evgeniy-klemin/todo/internal/item/repository/sqlitedb"
)

type sqliteAdapter struct {
	q          *sqlitedb.Queries
	db         sqlitedb.DBTX
	ftsEnabled bool
}

func newSQLiteAdapter(db *sql.DB, ftsEnabled bool) *sqliteAdapter {
	return &sqliteAdapter{q: sqlitedb.New(db), db: db, ftsEnabled: ftsEnabled}
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
	return a.q.MaxPosition(ctx)
}

func (a *sqliteAdapter) WithTx(tx *sql.Tx) querier {
	return &sqliteAdapter{q: a.q.WithTx(tx), db: tx, ftsEnabled: a.ftsEnabled}
}

func (a *sqliteAdapter) ListItems(ctx context.Context, filter listFilter, orderBy string, limit, offset int) ([]dbItem, error) {
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

func (a *sqliteAdapter) CountItems(ctx context.Context, filter listFilter) (int, error) {
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

func (a *sqliteAdapter) searchCondition(search string) (string, interface{}) {
	if a.ftsEnabled {
		return "item.rowid IN (SELECT rowid FROM item_fts WHERE item_fts MATCH ?)", buildFTSQuery(search)
	}
	return "LOWER(name) LIKE LOWER(?)", "%" + search + "%"
}

// buildFTSQuery converts a user search string into an FTS5 query with prefix matching.
// Example: "buy milk" -> "\"buy\"* \"milk\"*" (each word gets prefix matching)
func buildFTSQuery(search string) string {
	words := strings.Fields(search)
	for i, word := range words {
		word = strings.ReplaceAll(word, "\"", "\"\"")
		words[i] = "\"" + word + "\"" + "*"
	}
	return strings.Join(words, " ")
}

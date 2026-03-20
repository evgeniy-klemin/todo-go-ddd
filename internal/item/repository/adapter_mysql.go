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

func (a *mysqlAdapter) ListItems(ctx context.Context, filter listFilter, sort []sortField, limit, offset int) ([]dbItem, error) {
	var conditions []string
	var args []any

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
	q += " ORDER BY " + buildOrderBy(sort)
	q += " LIMIT ? OFFSET ?"
	queryArgs := make([]any, len(args), len(args)+2)
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

func (a *mysqlAdapter) ListItemsWithCursor(ctx context.Context, filter listFilter, sort []sortField, limit int, cursor *cursorParam) ([]dbItem, error) {
	var conditions []string
	var args []any

	if filter.Done != nil {
		conditions = append(conditions, "done=?")
		args = append(args, *filter.Done)
	}
	if filter.Search != nil && *filter.Search != "" {
		cond, arg := a.searchCondition(*filter.Search)
		conditions = append(conditions, cond)
		args = append(args, arg)
	}

	if cursor != nil {
		whereClause, cursorArgs := a.buildCursorWhere(cursor)
		if whereClause != "" {
			conditions = append(conditions, "("+whereClause+")")
			args = append(args, cursorArgs...)
		}
	}

	q := "SELECT id, name, position, done, created_at FROM item"
	if len(conditions) > 0 {
		q += " WHERE " + strings.Join(conditions, " AND ")
	}
	q += " ORDER BY " + buildOrderBy(sort) + ", id asc"
	q += " LIMIT ?"
	args = append(args, limit)

	rows, err := a.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list items with cursor: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []dbItem
	for rows.Next() {
		var item dbItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Position, &item.Done, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("list items with cursor: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list items with cursor: %w", err)
	}
	return result, nil
}

func (a *mysqlAdapter) CountItems(ctx context.Context, filter listFilter) (int, error) {
	var conditions []string
	var args []any

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

// buildCursorWhere builds the expanded OR-form cursor WHERE clause.
// For each sort field, it builds the prefix-equality + current-field comparison terms.
// Finally, it adds an id tie-breaker term.
func (a *mysqlAdapter) buildCursorWhere(cursor *cursorParam) (string, []interface{}) {
	if cursor == nil || len(cursor.Values) == 0 {
		return "", nil
	}

	cmpOp := func(dir string) string {
		if dir == "desc" {
			return "<"
		}
		return ">"
	}

	var terms []string
	var args []interface{}

	n := len(cursor.Values)

	for i := 0; i < n; i++ {
		var parts []string
		var termArgs []interface{}
		for j := 0; j < i; j++ {
			cv := cursor.Values[j]
			parts = append(parts, fmt.Sprintf("%s = ?", cv.Field))
			termArgs = append(termArgs, cv.Value)
		}
		cv := cursor.Values[i]
		op := cmpOp(cv.Direction)
		parts = append(parts, fmt.Sprintf("%s %s ?", cv.Field, op))
		termArgs = append(termArgs, cv.Value)

		terms = append(terms, "("+strings.Join(parts, " AND ")+")")
		args = append(args, termArgs...)
	}

	var idParts []string
	var idArgs []interface{}
	for _, cv := range cursor.Values {
		idParts = append(idParts, fmt.Sprintf("%s = ?", cv.Field))
		idArgs = append(idArgs, cv.Value)
	}
	idParts = append(idParts, "id > ?")
	idArgs = append(idArgs, cursor.ID)
	terms = append(terms, "("+strings.Join(idParts, " AND ")+")")
	args = append(args, idArgs...)

	return strings.Join(terms, " OR "), args
}

func (a *mysqlAdapter) searchCondition(search string) (string, any) {
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

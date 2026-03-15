package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/evgeniy-klemin/todo/db/driver"
	"github.com/evgeniy-klemin/todo/internal/item/domain"
)

type Repository struct {
	db         *sql.DB
	q          querier
	mu         sync.Mutex
	driver     string
	ftsEnabled bool
}

func NewSQLite(db *sql.DB, ftsEnabled bool) *Repository {
	return &Repository{
		db:         db,
		q:          newSQLiteAdapter(db),
		driver:     driver.SQLite,
		ftsEnabled: ftsEnabled,
	}
}

func NewMySQL(db *sql.DB, ftsEnabled bool) *Repository {
	return &Repository{
		db:         db,
		q:          newMySQLAdapter(db),
		driver:     driver.MySQL,
		ftsEnabled: ftsEnabled,
	}
}

func (r *Repository) maxPosition(ctx context.Context) (int, error) {
	result, err := r.q.MaxPosition(ctx)
	if err != nil {
		return 0, fmt.Errorf("max position query: %w", err)
	}
	return int(result), nil
}

func (r *Repository) AddWithNextPosition(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	max, err := r.maxPosition(ctx)
	if err != nil {
		return nil, fmt.Errorf("max position: %w", err)
	}

	if err := item.MoveTo(max + 1); err != nil {
		return nil, fmt.Errorf("move to position: %w", err)
	}

	return r.Add(ctx, item)
}

func (r *Repository) GetByID(ctx context.Context, id domain.ModelID) (*domain.Item, error) {
	item, err := r.q.GetItemByID(ctx, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainItem(item)
}

func (r *Repository) getByID(ctx context.Context, tx *sql.Tx, id domain.ModelID) (*domain.Item, error) {
	txQ := r.q.WithTx(tx)
	item, err := txQ.GetItemByID(ctx, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainItem(item)
}

// toDomainItem converts a dbItem to a domain.Item via ReconstituteItem.
func toDomainItem(item dbItem) (*domain.Item, error) {
	modelID, err := domain.NewModelID(item.ID)
	if err != nil {
		return nil, err
	}
	return domain.ReconstituteItem(modelID, item.Name, int(item.Position), item.Done, item.CreatedAt), nil
}

func (r *Repository) Add(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	id := item.ID()
	err := r.q.InsertItem(
		ctx,
		id.String(),
		item.Name().String(),
		int64(item.Position().Int()),
		item.Done(),
		item.CreatedAt().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("insert item: %w", err)
	}
	return item, nil
}

func (r *Repository) List(
	ctx context.Context,
	filter domain.ListFilter,
	sort []domain.SortField,
	page, perPage int,
) ([]*domain.Item, error) {
	var orderBy []string
	for _, s := range sort {
		col := s.Field
		if s.Direction == domain.SortDesc {
			col += " desc"
		} else {
			col += " asc"
		}
		orderBy = append(orderBy, col)
	}
	if len(orderBy) == 0 {
		orderBy = append(orderBy, "position asc")
	}

	var conditions []string
	var args []interface{}

	if filter.Done != nil {
		conditions = append(conditions, "done=?")
		args = append(args, *filter.Done)
	}

	if filter.Search != nil && *filter.Search != "" {
		if r.ftsEnabled {
			if r.driver == driver.MySQL {
				mysqlQuery := buildMySQLFTSQuery(*filter.Search)
				conditions = append(conditions, "MATCH(name) AGAINST(? IN BOOLEAN MODE)")
				args = append(args, mysqlQuery)
			} else {
				ftsQuery := buildFTSQuery(*filter.Search)
				conditions = append(conditions, "item.rowid IN (SELECT rowid FROM item_fts WHERE item_fts MATCH ?)")
				args = append(args, ftsQuery)
			}
		} else {
			conditions = append(conditions, "LOWER(name) LIKE LOWER(?)")
			args = append(args, "%"+*filter.Search+"%")
		}
	}

	q := "SELECT id, name, position, done, created_at FROM item"
	if len(conditions) > 0 {
		q += " WHERE " + strings.Join(conditions, " AND ")
	}
	q += " ORDER BY " + strings.Join(orderBy, ", ")
	q += " LIMIT ? OFFSET ?"
	args = append(args, perPage, perPage*(page-1))

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	res := make([]*domain.Item, 0)
	for rows.Next() {
		var dbRow dbItem
		if err := rows.Scan(&dbRow.ID, &dbRow.Name, &dbRow.Position, &dbRow.Done, &dbRow.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		domainItem, err := toDomainItem(dbRow)
		if err != nil {
			return nil, fmt.Errorf("convert item: %w", err)
		}
		res = append(res, domainItem)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}
	return res, nil
}

func (r *Repository) Count(ctx context.Context, filter domain.ListFilter) (int, error) {
	var conditions []string
	var args []interface{}

	if filter.Done != nil {
		conditions = append(conditions, "done=?")
		args = append(args, *filter.Done)
	}

	if filter.Search != nil && *filter.Search != "" {
		if r.ftsEnabled {
			if r.driver == driver.MySQL {
				mysqlQuery := buildMySQLFTSQuery(*filter.Search)
				conditions = append(conditions, "MATCH(name) AGAINST(? IN BOOLEAN MODE)")
				args = append(args, mysqlQuery)
			} else {
				ftsQuery := buildFTSQuery(*filter.Search)
				conditions = append(conditions, "item.rowid IN (SELECT rowid FROM item_fts WHERE item_fts MATCH ?)")
				args = append(args, ftsQuery)
			}
		} else {
			conditions = append(conditions, "LOWER(name) LIKE LOWER(?)")
			args = append(args, "%"+*filter.Search+"%")
		}
	}

	q := "SELECT COUNT(*) FROM item"
	if len(conditions) > 0 {
		q += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count items: %w", err)
	}
	return count, nil
}

func (r *Repository) Update(
	ctx context.Context,
	id domain.ModelID,
	updater func(item *domain.Item) error,
) (*domain.Item, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	domainItem, err := r.getByID(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if err := updater(domainItem); err != nil {
		return nil, err
	}

	itemID := domainItem.ID()
	txQ := r.q.WithTx(tx)
	err = txQ.UpdateItem(
		ctx,
		domainItem.Name().String(),
		int64(domainItem.Position().Int()),
		domainItem.Done(),
		itemID.String(),
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return domainItem, nil
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

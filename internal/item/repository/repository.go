package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/evgeniy-klemin/todo/db/schema"
	"github.com/evgeniy-klemin/todo/internal/item/app"
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
		driver:     schema.DriverSQLite,
		ftsEnabled: ftsEnabled,
	}
}

func NewMySQL(db *sql.DB, ftsEnabled bool) *Repository {
	return &Repository{
		db:         db,
		q:          newMySQLAdapter(db),
		driver:     schema.DriverMySQL,
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
		return nil, err
	}
	return toDomainItem(item)
}

func (r *Repository) getByID(ctx context.Context, tx *sql.Tx, id domain.ModelID) (*domain.Item, error) {
	txQ := r.q.WithTx(tx)
	item, err := txQ.GetItemByID(ctx, id.String())
	if err != nil {
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

func (r *Repository) All(
	ctx context.Context,
	done *bool,
	search *string,
	fields []app.ItemField,
	page, perPage int,
	sortFields app.SortFields,
) ([]app.Item, error) {
	for _, field := range fields {
		switch field {
		case app.ItemFieldName, app.ItemFieldPosition, app.ItemFieldDone, app.ItemFieldCreatedAt:
		default:
			return nil, fmt.Errorf("field %d not found", field)
		}
	}

	var orderBy []string
	for _, sortField := range sortFields {
		var col string
		switch sortField.Field {
		case app.ItemFieldName:
			col = "name"
		case app.ItemFieldPosition:
			col = "position"
		case app.ItemFieldDone:
			col = "done"
		case app.ItemFieldCreatedAt:
			col = "created_at"
		default:
			return nil, fmt.Errorf("field %d not found", sortField.Field)
		}
		switch sortField.SortDirection {
		case app.SortDirectionAsc:
			col += " asc"
		case app.SortDirectionDesc:
			col += " desc"
		}
		orderBy = append(orderBy, col)
	}
	if len(orderBy) == 0 {
		orderBy = append(orderBy, "position asc")
	}

	var conditions []string
	var args []interface{}

	if done != nil {
		conditions = append(conditions, "done=?")
		args = append(args, *done)
	}

	if search != nil && *search != "" {
		if r.ftsEnabled {
			if r.driver == schema.DriverMySQL {
				mysqlQuery := buildMySQLFTSQuery(*search)
				conditions = append(conditions, "MATCH(name) AGAINST(? IN BOOLEAN MODE)")
				args = append(args, mysqlQuery)
			} else {
				ftsQuery := buildFTSQuery(*search)
				conditions = append(conditions, "item.rowid IN (SELECT rowid FROM item_fts WHERE item_fts MATCH ?)")
				args = append(args, ftsQuery)
			}
		} else {
			conditions = append(conditions, "LOWER(name) LIKE LOWER(?)")
			args = append(args, "%"+*search+"%")
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
	defer rows.Close()

	if len(fields) == 0 {
		fields = app.DefaultItemFields
	}

	res := make([]app.Item, 0)
	for rows.Next() {
		var (
			id        string
			name      string
			position  int64
			doneVal   bool
			createdAt time.Time
		)
		if err := rows.Scan(&id, &name, &position, &doneVal, &createdAt); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		item := app.Item{ID: id}
		for _, field := range fields {
			switch field {
			case app.ItemFieldName:
				n := name
				item.Name = &n
			case app.ItemFieldPosition:
				p := int(position)
				item.Position = &p
			case app.ItemFieldDone:
				d := doneVal
				item.Done = &d
			case app.ItemFieldCreatedAt:
				t := createdAt
				item.CreatedAt = &t
			default:
				return nil, fmt.Errorf("field %d not found", field)
			}
		}
		res = append(res, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}
	return res, nil
}

func (r *Repository) Count(ctx context.Context, done *bool, search *string) (int, error) {
	var conditions []string
	var args []interface{}

	if done != nil {
		conditions = append(conditions, "done=?")
		args = append(args, *done)
	}

	if search != nil && *search != "" {
		if r.ftsEnabled {
			if r.driver == schema.DriverMySQL {
				mysqlQuery := buildMySQLFTSQuery(*search)
				conditions = append(conditions, "MATCH(name) AGAINST(? IN BOOLEAN MODE)")
				args = append(args, mysqlQuery)
			} else {
				ftsQuery := buildFTSQuery(*search)
				conditions = append(conditions, "item.rowid IN (SELECT rowid FROM item_fts WHERE item_fts MATCH ?)")
				args = append(args, ftsQuery)
			}
		} else {
			conditions = append(conditions, "LOWER(name) LIKE LOWER(?)")
			args = append(args, "%"+*search+"%")
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

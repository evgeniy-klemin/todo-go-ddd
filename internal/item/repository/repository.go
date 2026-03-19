package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/evgeniy-klemin/todo/internal/item/domain"
)

type Repository struct {
	db *sql.DB
	q  querier
	mu sync.Mutex
}

func NewSQLite(db *sql.DB, ftsEnabled bool) *Repository {
	return &Repository{
		db: db,
		q:  newSQLiteAdapter(db, ftsEnabled),
	}
}

func NewMySQL(db *sql.DB, ftsEnabled bool) *Repository {
	return &Repository{
		db: db,
		q:  newMySQLAdapter(db, ftsEnabled),
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
	dbSort := make([]sortField, len(sort))
	for i, s := range sort {
		dbSort[i] = sortField{Field: s.Field, Desc: s.Direction == domain.SortDesc}
	}

	dbRows, err := r.q.ListItems(ctx, listFilter{Done: filter.Done, Search: filter.Search}, dbSort, perPage, perPage*(page-1))
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}

	res := make([]*domain.Item, 0, len(dbRows))
	for _, dbRow := range dbRows {
		domainItem, err := toDomainItem(dbRow)
		if err != nil {
			return nil, fmt.Errorf("convert item: %w", err)
		}
		res = append(res, domainItem)
	}
	return res, nil
}

// ListWithCursor fetches up to limit items starting after cursor (nil = from start).
// It delegates cursor WHERE clause building to the adapter via ListItemsWithCursor.
func (r *Repository) ListWithCursor(
	ctx context.Context,
	filter domain.ListFilter,
	sort []domain.SortField,
	limit int,
	cursor *domain.Cursor,
) ([]*domain.Item, error) {
	dbSort := make([]sortField, len(sort))
	for i, s := range sort {
		dbSort[i] = sortField{Field: s.Field, Desc: s.Direction == domain.SortDesc}
	}

	var dbCursor *cursorParam
	if cursor != nil {
		dbCursor = &cursorParam{
			Values: make([]cursorValue, len(cursor.Values)),
			ID:     cursor.ID,
		}
		for i, cv := range cursor.Values {
			dbCursor.Values[i] = cursorValue{
				Field:     cv.Field,
				Value:     cv.Value,
				Direction: cv.Direction,
			}
		}
	}

	dbRows, err := r.q.ListItemsWithCursor(ctx, listFilter{Done: filter.Done, Search: filter.Search}, dbSort, limit, dbCursor)
	if err != nil {
		return nil, fmt.Errorf("query items with cursor: %w", err)
	}

	res := make([]*domain.Item, 0, len(dbRows))
	for _, dbRow := range dbRows {
		domainItem, err := toDomainItem(dbRow)
		if err != nil {
			return nil, fmt.Errorf("convert item: %w", err)
		}
		res = append(res, domainItem)
	}
	return res, nil
}

func (r *Repository) Count(ctx context.Context, filter domain.ListFilter) (int, error) {
	count, err := r.q.CountItems(ctx, listFilter{Done: filter.Done, Search: filter.Search})
	if err != nil {
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

// buildCursorWhere builds the expanded OR-form cursor WHERE clause.
// For each sort field, it builds the prefix-equality + current-field comparison terms.
// Finally, it adds an id tie-breaker term.
func buildCursorWhere(cursor *cursorParam) (string, []interface{}) {
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

	// Build n terms: for i in [0..n-1], prefix equality on [0..i-1] + comparison on [i]
	for i := 0; i < n; i++ {
		var parts []string
		var termArgs []interface{}
		// Prefix equalities
		for j := 0; j < i; j++ {
			cv := cursor.Values[j]
			parts = append(parts, fmt.Sprintf("%s = ?", cv.Field))
			termArgs = append(termArgs, cv.Value)
		}
		// Current field comparison
		cv := cursor.Values[i]
		op := cmpOp(cv.Direction)
		parts = append(parts, fmt.Sprintf("%s %s ?", cv.Field, op))
		termArgs = append(termArgs, cv.Value)

		terms = append(terms, "("+strings.Join(parts, " AND ")+")")
		args = append(args, termArgs...)
	}

	// Add id tie-breaker: all sort fields equal + id > cursor.ID
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

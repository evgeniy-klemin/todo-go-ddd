package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

// ListWithCursor fetches up to limit items after the position encoded in cursorData.
// cursorData is an opaque []byte previously returned by BuildCursor; pass nil to start
// from the beginning of the result set. Internally it is JSON-decoded into a cursorParam
// and forwarded to the adapter's ListItemsWithCursor which translates it into a keyset
// WHERE clause (see buildCursorWhere). filter and sort must match those used when the
// cursor was built, otherwise results are undefined.
func (r *Repository) ListWithCursor(
	ctx context.Context,
	filter domain.ListFilter,
	sort []domain.SortField,
	limit int,
	cursorData []byte,
) ([]*domain.Item, error) {
	dbSort := make([]sortField, len(sort))
	for i, s := range sort {
		dbSort[i] = sortField{Field: s.Field, Desc: s.Direction == domain.SortDesc}
	}

	var dbCursor *cursorParam
	if len(cursorData) > 0 {
		dbCursor = &cursorParam{}
		if err := json.Unmarshal(cursorData, dbCursor); err != nil {
			return nil, fmt.Errorf("unmarshal cursor: %w", err)
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

// BuildCursor serializes item and sort into an opaque cursor []byte for use with ListWithCursor.
// It captures the current field values of item for each sort field so that the next call to
// ListWithCursor can reconstruct the keyset WHERE clause. The serialization format (JSON cursorParam)
// is an implementation detail of the repository package; callers must treat it as opaque bytes.
// sort must be the same slice passed to the original ListWithCursor call.
func (r *Repository) BuildCursor(item *domain.Item, sort []domain.SortField) ([]byte, error) {
	id := item.ID()
	cp := &cursorParam{
		ID: id.String(),
	}
	for _, sf := range sort {
		cv := cursorValue{
			Direction: "asc",
		}
		if sf.Direction == domain.SortDesc {
			cv.Direction = "desc"
		}
		switch sf.Field {
		case "name":
			cv.Field = "name"
			cv.Value = item.Name().String()
		case "position":
			cv.Field = "position"
			cv.Value = item.Position().Int()
		case "done":
			cv.Field = "done"
			cv.Value = item.Done()
		case "created_at":
			cv.Field = "created_at"
			cv.Value = item.CreatedAt().Unix()
		default:
			cv.Field = sf.Field
		}
		cp.Values = append(cp.Values, cv)
	}
	return json.Marshal(cp)
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

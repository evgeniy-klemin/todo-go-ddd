package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/evgeniy-klemin/todo/internal/item/app"
	"github.com/evgeniy-klemin/todo/internal/item/domain"
	"github.com/evgeniy-klemin/todo/internal/item/repository/models"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetByID(ctx context.Context, id domain.ModelID) (*domain.Item, error) {
	return r.getByID(ctx, id, false)
}

func (r *Repository) getByID(ctx context.Context, id domain.ModelID, forUpdate bool) (*domain.Item, error) {
	var qms []qm.QueryMod
	qms = append(qms, qm.Where(models.ItemColumns.ID+"=?", id.String()))
	if forUpdate {
		qms = append(qms, qm.For("UPDATE"))
	}
	item, err := models.Items(qms...).One(ctx, r.db)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	modelID, err := domain.NewModelID(item.ID)
	if err != nil {
		return nil, err
	}
	return &domain.Item{
		ID:        modelID,
		Name:      item.Name,
		Position:  int(item.Position),
		Done:      item.Done,
		CreatedAt: item.CreatedAt,
	}, nil
}

func (r *Repository) Add(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	dbItem := models.Item{
		ID:        item.ID.String(),
		Name:      item.Name,
		Position:  item.Position,
		Done:      item.Done,
		CreatedAt: item.CreatedAt.UTC(),
	}

	if err := dbItem.Insert(ctx, r.db, boil.Infer()); err != nil {
		return nil, errors.Wrap(err, "ошибка добавления в базу")
	}

	return item, nil
}

func (r *Repository) All(
	ctx context.Context,
	done *bool,
	fields []app.ItemField,
	limit int,
	cursor *app.Cursor,
	sortFields app.SortFields,
) ([]app.Item, error) {
	var columns []string
	for _, field := range fields {
		switch field {
		case app.ItemFieldName:
			columns = append(columns, models.ItemColumns.Name)
		case app.ItemFieldPosition:
			columns = append(columns, models.ItemColumns.Position)
		case app.ItemFieldDone:
			columns = append(columns, models.ItemColumns.Done)
		case app.ItemFieldCreatedAt:
			columns = append(columns, models.ItemColumns.CreatedAt)
		default:
			return nil, errors.Errorf("Field %d not found", field)
		}
	}
	if len(columns) > 0 {
		columns = append(columns, models.ItemColumns.ID)
	}

	var orderBy []string
	for _, sortField := range sortFields {
		var order string
		switch sortField.Field {
		case app.ItemFieldName:
			order = models.ItemColumns.Name
		case app.ItemFieldPosition:
			order = models.ItemColumns.Position
		case app.ItemFieldDone:
			order = models.ItemColumns.Done
		case app.ItemFieldCreatedAt:
			order = models.ItemColumns.CreatedAt
		default:
			return nil, errors.Errorf("Field %d not found", sortFields)
		}
		switch sortField.SortDirection {
		case app.SortDirectionAsc:
			order = order + " asc"
		case app.SortDirectionDesc:
			order = order + " desc"
		}
		orderBy = append(orderBy, order)
	}
	if len(orderBy) == 0 {
		orderBy = append(orderBy, models.ItemColumns.Position+" asc")
	}
	// Always append id ASC as tie-breaker
	orderBy = append(orderBy, models.ItemColumns.ID+" asc")

	var query []qm.QueryMod

	if done != nil {
		query = append(query, qm.Where(models.ItemColumns.Done+"=?", *done))
	}

	if cursor != nil {
		whereClause, args := buildCursorWhere(cursor)
		if whereClause != "" {
			query = append(query, qm.Where(whereClause, args...))
		}
	}

	query = append(query, qm.Select(columns...))
	query = append(query, qm.Limit(limit))
	query = append(query, qm.OrderBy(strings.Join(orderBy, ", ")))

	dbItems, err := models.Items(query...).All(ctx, r.db)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	res := make([]app.Item, 0)
	for _, dbItem := range dbItems {
		item := app.Item{
			ID: dbItem.ID,
		}
		if len(fields) == 0 {
			fields = app.DefaultItemFields
		}
		for _, field := range fields {
			switch field {
			case app.ItemFieldName:
				item.Name = &dbItem.Name
			case app.ItemFieldPosition:
				position := int(dbItem.Position)
				item.Position = &position
			case app.ItemFieldDone:
				item.Done = &dbItem.Done
			case app.ItemFieldCreatedAt:
				item.CreatedAt = &dbItem.CreatedAt
			default:
				return nil, errors.Errorf("Field %d not found", field)
			}
		}
		res = append(res, item)
	}
	return res, nil
}

// buildCursorWhere builds the expanded OR-form cursor WHERE clause.
// For each sort field, it builds the prefix-equality + current-field comparison terms.
// Finally, it adds an id tie-breaker term.
func buildCursorWhere(cursor *app.Cursor) (string, []interface{}) {
	if cursor == nil || len(cursor.Values) == 0 {
		return "", nil
	}

	colName := func(field string) string {
		switch field {
		case "name":
			return models.ItemColumns.Name
		case "position":
			return models.ItemColumns.Position
		case "done":
			return models.ItemColumns.Done
		case "created_at":
			return models.ItemColumns.CreatedAt
		default:
			return field
		}
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
			parts = append(parts, fmt.Sprintf("%s = ?", colName(cv.Field)))
			termArgs = append(termArgs, cv.Value)
		}
		// Current field comparison
		cv := cursor.Values[i]
		op := cmpOp(cv.Direction)
		parts = append(parts, fmt.Sprintf("%s %s ?", colName(cv.Field), op))
		termArgs = append(termArgs, cv.Value)

		terms = append(terms, "("+strings.Join(parts, " AND ")+")")
		args = append(args, termArgs...)
	}

	// Add id tie-breaker: all sort fields equal + id > cursor.ID
	var idParts []string
	var idArgs []interface{}
	for _, cv := range cursor.Values {
		idParts = append(idParts, fmt.Sprintf("%s = ?", colName(cv.Field)))
		idArgs = append(idArgs, cv.Value)
	}
	idParts = append(idParts, fmt.Sprintf("%s > ?", models.ItemColumns.ID))
	idArgs = append(idArgs, cursor.ID)
	terms = append(terms, "("+strings.Join(idParts, " AND ")+")")
	args = append(args, idArgs...)

	return strings.Join(terms, " OR "), args
}

func (r *Repository) Count(ctx context.Context, done *bool) (int, error) {
	var query []qm.QueryMod
	if done != nil {
		query = append(query, qm.Where(models.ItemColumns.Done+"=?", *done))
	}
	count, err := models.Items(query...).Count(ctx, r.db)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return int(count), nil
}

func (r *Repository) Update(ctx context.Context, id domain.ModelID, updater func(item *domain.Item) error) (*domain.Item, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	domainItem, err := r.getByID(ctx, id, true)
	if err != nil {
		return nil, err
	}

	if err := updater(domainItem); err != nil {
		return nil, err
	}

	dbItem := &models.Item{
		ID:        domainItem.ID.String(),
		Name:      domainItem.Name,
		Position:  domainItem.Position,
		Done:      domainItem.Done,
		CreatedAt: domainItem.CreatedAt,
	}
	_, err = dbItem.Update(ctx, r.db, boil.Blacklist(models.ItemColumns.CreatedAt))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return domainItem, nil
}

package repository

import (
	"context"
	"database/sql"
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
	page, perPage int,
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

	var query []qm.QueryMod

	if done != nil {
		query = append(query, qm.Where(models.ItemColumns.Done+"=?", *done))
	}

	query = append(query, qm.Select(columns...))
	query = append(query, qm.Limit(perPage))
	query = append(query, qm.Offset(perPage*(page-1)))
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

func (r *Repository) Count(ctx context.Context, done *bool) (int, error) {
	var query []qm.QueryMod

	if done != nil {
		query = append(query, qm.Where(models.ItemColumns.Done+"=?", *done))
	}

	count, err := models.Items(query...).Count(ctx, r.db)
	return int(count), err
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

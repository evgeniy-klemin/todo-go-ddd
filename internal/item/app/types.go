package app

import (
	"time"

	"github.com/evgeniy-klemin/todo/internal/item/domain"
)

type ItemField int

const (
	ItemFieldName ItemField = iota + 1
	ItemFieldPosition
	ItemFieldDone
	ItemFieldCreatedAt
)

var DefaultItemFields = []ItemField{
	ItemFieldName,
	ItemFieldPosition,
	ItemFieldDone,
	ItemFieldCreatedAt,
}

type Item struct {
	ID        string
	Name      *string
	Position  *int
	Done      *bool
	CreatedAt *time.Time
}

type SortDirection int

const (
	SortDirectionAsc SortDirection = iota + 1
	SortDirectionDesc
)

type SortField struct {
	Field         ItemField
	SortDirection SortDirection
}
type SortFields []SortField

type ListQuery struct {
	Done       *bool
	Search     *string
	Fields     []ItemField
	Page       int
	PerPage    int
	SortFields SortFields
}

type ListResult struct {
	Items      []Item
	TotalCount int
}

func appSortFieldsToDomain(sf SortFields) []domain.SortField {
	res := make([]domain.SortField, len(sf))
	for i, f := range sf {
		var field string
		switch f.Field {
		case ItemFieldName:
			field = "name"
		case ItemFieldPosition:
			field = "position"
		case ItemFieldDone:
			field = "done"
		case ItemFieldCreatedAt:
			field = "created_at"
		}
		dir := domain.SortAsc
		if f.SortDirection == SortDirectionDesc {
			dir = domain.SortDesc
		}
		res[i] = domain.SortField{Field: field, Direction: dir}
	}
	return res
}

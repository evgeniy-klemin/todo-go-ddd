package app

import (
	"time"
)

type ItemField int

const (
	ItemFieldName ItemField = iota + 1
	ItemFieldDescription
	ItemFieldPosition
	ItemFieldDone
	ItemFieldCreatedAt
)

var DefaultItemFields = []ItemField{
	ItemFieldName,
	ItemFieldDescription,
	ItemFieldPosition,
	ItemFieldDone,
	ItemFieldCreatedAt,
}

type Item struct {
	ID          string
	Name        *string
	Description *string
	Position    *int
	Done        *bool
	CreatedAt   *time.Time
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

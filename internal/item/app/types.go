package app

import (
	"time"
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

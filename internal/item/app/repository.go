package app

import "context"

type Repository interface {
	All(ctx context.Context, done *bool, fields []ItemField, limit int, cursor *Cursor, sortFields SortFields) ([]Item, error)
}

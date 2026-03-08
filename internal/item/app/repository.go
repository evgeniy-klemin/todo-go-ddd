package app

import "context"

type QueryRepository interface {
	All(ctx context.Context, done *bool, fields []ItemField, page, perPage int, sortFields SortFields) ([]Item, error)
	Count(ctx context.Context, done *bool) (int, error)
}

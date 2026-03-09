package app

import "context"

type QueryRepository interface {
	All(ctx context.Context, done *bool, search *string, fields []ItemField, page, perPage int, sortFields SortFields) ([]Item, error)
	Count(ctx context.Context, done *bool, search *string) (int, error)
}

package domain

import "context"

type SortDirection int

const (
	SortAsc SortDirection = iota + 1
	SortDesc
)

type SortField struct {
	Field     string
	Direction SortDirection
}

type ListFilter struct {
	Done   *bool
	Search *string
}

type Repository interface {
	GetByID(ctx context.Context, id ModelID) (*Item, error)
	Add(ctx context.Context, item *Item) (*Item, error)
	AddWithNextPosition(ctx context.Context, item *Item) (*Item, error)
	Update(ctx context.Context, id ModelID, updater func(item *Item) error) (*Item, error)
	Count(ctx context.Context, filter ListFilter) (int, error)
	// ListWithCursor fetches up to limit items after the position encoded in cursorData.
	// Pass cursorData = nil to start from the beginning of the result set.
	// The format of cursorData is owned by the repository implementation; callers must
	// treat it as opaque and only pass values previously returned by BuildCursor.
	// filter and sort must be identical to those used when BuildCursor was called.
	ListWithCursor(ctx context.Context, filter ListFilter, sort []SortField, limit int, cursorData []byte) ([]*Item, error)
	// BuildCursor serializes item and the active sort fields into an opaque cursor []byte.
	// The returned value should be passed to the next call to ListWithCursor so that
	// pagination continues from the position of item.
	// sort must be the same slice that was passed to the originating ListWithCursor call.
	BuildCursor(item *Item, sort []SortField) ([]byte, error)
}

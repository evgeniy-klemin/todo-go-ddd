package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// dbItem is the internal transfer object between adapters and the repository.
type dbItem struct {
	ID        string
	Name      string
	Position  int64
	Done      bool
	CreatedAt time.Time
}

// listFilter carries the filter parameters for List/Count queries.
// It mirrors domain.ListFilter but avoids importing domain types into the adapter layer.
type listFilter struct {
	Done   *bool
	Search *string
}

// sortField carries a single ORDER BY field and direction.
type sortField struct {
	Field string
	Desc  bool
}

// buildOrderBy converts a slice of sortField into an ORDER BY clause string.
// Defaults to "position asc" when the slice is empty.
func buildOrderBy(sort []sortField) string {
	if len(sort) == 0 {
		return "position asc"
	}
	parts := make([]string, len(sort))
	for i, s := range sort {
		if s.Desc {
			parts[i] = s.Field + " desc"
		} else {
			parts[i] = s.Field + " asc"
		}
	}
	return strings.Join(parts, ", ")
}

// cursorValue holds the sort-field name, its current value, and the sort direction
// ("asc" or "desc") for one column in a cursor-based pagination query.
// It is used to reconstruct the WHERE clause that resumes a page from a known position.
type cursorValue struct {
	Field     string
	Value     any
	Direction string // "asc" or "desc"
}

// cursorParam is the adapter-layer cursor type. It carries the last-seen item's ID
// plus one cursorValue per sort field so that buildCursorWhere can reconstruct the
// WHERE clause for the next page. It is JSON-marshalled by BuildCursor and
// unmarshalled by ListWithCursor; the format is internal to the repository package.
type cursorParam struct {
	Values []cursorValue
	ID     string
}

// buildCursorWhere builds the expanded OR-form cursor WHERE clause used for
// keyset (cursor-based) pagination. It does not use any database-specific syntax
// beyond standard SQL comparison operators, so it is shared by all adapters.
//
// For a cursor with n sort fields it produces n+1 OR terms:
//   - For i in [0..n-1]: equality on fields [0..i-1], then comparison (> or <) on field i.
//   - A final tie-breaker term: equality on all sort fields, then "id > cursor.ID".
//
// Example (two sort fields: position asc, name asc, id = "abc"):
//
//	(position > ?) OR
//	(position = ? AND name > ?) OR
//	(position = ? AND name = ? AND id > ?)
//
// Parameters:
//   - cursor: the deserialized cursor from the previous page's last item; nil or empty Values returns ("", nil).
//
// Returns the WHERE clause string (without the leading "WHERE") and the
// corresponding positional argument slice.
func buildCursorWhere(cursor *cursorParam) (string, []any) {
	if cursor == nil || len(cursor.Values) == 0 {
		return "", nil
	}

	cmpOp := func(dir string) string {
		if dir == "desc" {
			return "<"
		}
		return ">"
	}

	var terms []string
	var args []any

	n := len(cursor.Values)

	// Build n terms: for i in [0..n-1], prefix equality on [0..i-1] + comparison on [i]
	for i := 0; i < n; i++ {
		var parts []string
		var termArgs []any
		// Prefix equalities
		for j := 0; j < i; j++ {
			cv := cursor.Values[j]
			parts = append(parts, fmt.Sprintf("%s = ?", cv.Field))
			termArgs = append(termArgs, cv.Value)
		}
		// Current field comparison
		cv := cursor.Values[i]
		op := cmpOp(cv.Direction)
		parts = append(parts, fmt.Sprintf("%s %s ?", cv.Field, op))
		termArgs = append(termArgs, cv.Value)

		terms = append(terms, "("+strings.Join(parts, " AND ")+")")
		args = append(args, termArgs...)
	}

	// Add id tie-breaker: all sort fields equal + id > cursor.ID
	var idParts []string
	var idArgs []any
	for _, cv := range cursor.Values {
		idParts = append(idParts, fmt.Sprintf("%s = ?", cv.Field))
		idArgs = append(idArgs, cv.Value)
	}
	idParts = append(idParts, "id > ?")
	idArgs = append(idArgs, cursor.ID)
	terms = append(terms, "("+strings.Join(idParts, " AND ")+")")
	args = append(args, idArgs...)

	return strings.Join(terms, " OR "), args
}

// querier abstracts over the sqlc-generated query types for both SQLite and MySQL,
// allowing the repository to stay database-agnostic.
type querier interface {
	GetItemByID(ctx context.Context, id string) (dbItem, error)
	InsertItem(ctx context.Context, id, name string, position int64, done bool, createdAt time.Time) error
	UpdateItem(ctx context.Context, name string, position int64, done bool, id string) error
	MaxPosition(ctx context.Context) (int64, error)
	// ListItems executes a SELECT query applying filter, ORDER BY, LIMIT, and OFFSET.
	// sort determines column order; limit and offset implement page-based slicing.
	ListItems(ctx context.Context, filter listFilter, sort []sortField, limit, offset int) ([]dbItem, error)
	// ListItemsWithCursor executes a SELECT with a keyset WHERE clause instead of OFFSET.
	// cursor is nil when fetching the first page; subsequent pages pass the cursorParam
	// returned by BuildCursor for the last visible item.
	// limit is the maximum number of rows to return (caller typically requests perPage+1
	// to detect whether a next page exists).
	ListItemsWithCursor(ctx context.Context, filter listFilter, sort []sortField, limit int, cursor *cursorParam) ([]dbItem, error)
	// CountItems executes a COUNT(*) query applying the same filter as ListItems/ListItemsWithCursor.
	// The result is used to populate the X-Total-Count response header.
	CountItems(ctx context.Context, filter listFilter) (int, error)
	// WithTx returns a new querier backed by the given transaction.
	// All operations on the returned querier run within tx.
	WithTx(tx *sql.Tx) querier
}

package app

import (
	"github.com/evgeniy-klemin/todo/internal/utils/pagination"
)

// Cursor and CursorValue are type aliases from the pagination package.
type Cursor = pagination.Cursor
type CursorValue = pagination.CursorValue

// EncodeCursor encodes a cursor to a base64 string.
var EncodeCursor = pagination.EncodeCursor

// DecodeCursor decodes a cursor from a base64 string.
var DecodeCursor = pagination.DecodeCursor

func BuildCursorFromItem(item Item, sortFields SortFields) *Cursor {
	var values []CursorValue
	for _, sf := range sortFields {
		cv := CursorValue{
			Direction: "asc",
		}
		if sf.SortDirection == SortDirectionDesc {
			cv.Direction = "desc"
		}
		switch sf.Field {
		case ItemFieldName:
			cv.Field = "name"
			if item.Name != nil {
				cv.Value = *item.Name
			}
		case ItemFieldPosition:
			cv.Field = "position"
			if item.Position != nil {
				cv.Value = *item.Position
			}
		case ItemFieldDone:
			cv.Field = "done"
			if item.Done != nil {
				cv.Value = *item.Done
			}
		case ItemFieldCreatedAt:
			cv.Field = "created_at"
			if item.CreatedAt != nil {
				cv.Value = item.CreatedAt.Unix()
			}
		}
		values = append(values, cv)
	}
	return &Cursor{
		Values: values,
		ID:     item.ID,
	}
}

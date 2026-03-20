package app

import (
	"encoding/base64"
	"encoding/json"
)

// Cursor holds pagination state for cursor-based listing.
type Cursor struct {
	Values []CursorValue
	ID     string
}

// CursorValue holds a single sort-field value for cursor pagination.
type CursorValue struct {
	Field     string
	Value     any
	Direction string // "asc" or "desc"
}

// EncodeCursor encodes a cursor to a base64 string.
func EncodeCursor(c *Cursor) (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

// DecodeCursor decodes a cursor from a base64 string.
func DecodeCursor(s string) (*Cursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var c Cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

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

package app

import (
	"encoding/base64"
	"encoding/json"
)

type CursorValue struct {
	Field     string      `json:"f"`
	Value     interface{} `json:"v"`
	Direction string      `json:"d"`
}

type Cursor struct {
	Values []CursorValue `json:"v"`
	ID     string        `json:"id"`
}

func EncodeCursor(c *Cursor) (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

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

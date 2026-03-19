package pagination

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

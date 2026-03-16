package domain

import (
	"fmt"

	"github.com/gofrs/uuid"
)

// ModelID is a value object wrapping a UUID string that uniquely identifies a domain entity.
type ModelID struct {
	value string
}

func NewModelID(value string) (ModelID, error) {
	if len(value) != 36 {
		return ModelID{}, fmt.Errorf("value must be 36 symbols: %s", value)
	}
	return ModelID{value}, nil
}

func GenerateModelID() (ModelID, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return ModelID{}, fmt.Errorf("invalid generate uuid: %w", err)
	}
	return ModelID{fmt.Sprintf("%s", id)}, nil
}

func (m *ModelID) IsZero() bool {
	return *m == ModelID{}
}

func (m *ModelID) String() string {
	return m.value
}

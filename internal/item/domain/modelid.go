package domain

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

type ModelID struct {
	value string
}

func NewModelID(value string) (ModelID, error) {
	if len(value) != 36 {
		return ModelID{}, errors.Errorf("Value must be 36 symbols: %s", value)
	}
	return ModelID{value}, nil
}

func GenerateModelID() (ModelID, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return ModelID{}, errors.Wrap(err, "invalid generate uuid")
	}
	return ModelID{fmt.Sprintf("%s", id)}, nil
}

func (m *ModelID) IsZero() bool {
	return *m == ModelID{}
}

func (m *ModelID) String() string {
	return m.value
}

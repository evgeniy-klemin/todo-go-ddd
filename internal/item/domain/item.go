package domain

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrNameLength    = fmt.Errorf("name has wrong size")
	ErrPositionValue = fmt.Errorf("position has wrong value")
)

const (
	NameMinLength = 1
	NameMaxLength = 1000
	PositionMin   = 1
)

type Item struct {
	ID        ModelID
	Name      string
	Position  int
	Done      bool
	CreatedAt time.Time
}

func (i *Item) Validate() error {
	// Name
	if len(i.Name) < NameMinLength {
		return errors.Wrapf(ErrNameLength, "'%s' value less then min %d - got %d", i.Name, NameMinLength, len(i.Name))
	}
	if len(i.Name) > NameMaxLength {
		return errors.Wrapf(ErrNameLength, "'%s' value more then max %d - got %d", i.Name, NameMaxLength, len(i.Name))
	}
	// Position
	if i.Position < PositionMin {
		return errors.Wrapf(ErrPositionValue, "%d value less then min %d", i.Position, PositionMin)
	}
	return nil
}

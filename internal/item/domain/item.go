package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrNameLength    = errors.New("name has wrong size")
	ErrPositionValue = errors.New("position has wrong value")
	ErrNotFound      = errors.New("item not found")
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

func NewItem(name string, position int) (*Item, error) {
	id, err := GenerateModelID()
	if err != nil {
		return nil, err
	}
	item := &Item{
		ID:        id,
		Name:      name,
		Position:  position,
		Done:      false,
		CreatedAt: time.Now().Truncate(time.Second),
	}
	if err := item.Validate(); err != nil {
		return nil, err
	}
	return item, nil
}

func ReconstituteItem(id ModelID, name string, position int, done bool, createdAt time.Time) *Item {
	return &Item{
		ID:        id,
		Name:      name,
		Position:  position,
		Done:      done,
		CreatedAt: createdAt,
	}
}

func (i *Item) Complete() {
	i.Done = true
}

func (i *Item) Uncomplete() {
	i.Done = false
}

func (i *Item) Rename(name string) error {
	if len(name) < NameMinLength {
		return fmt.Errorf("%w: '%s' value less then min %d - got %d", ErrNameLength, name, NameMinLength, len(name))
	}
	if len(name) > NameMaxLength {
		return fmt.Errorf("%w: '%s' value more then max %d - got %d", ErrNameLength, name, NameMaxLength, len(name))
	}
	i.Name = name
	return nil
}

func (i *Item) MoveTo(position int) error {
	if position < PositionMin {
		return fmt.Errorf("%w: %d value less then min %d", ErrPositionValue, position, PositionMin)
	}
	i.Position = position
	return nil
}

func (i *Item) Validate() error {
	// Name
	if len(i.Name) < NameMinLength {
		return fmt.Errorf("%w: '%s' value less then min %d - got %d", ErrNameLength, i.Name, NameMinLength, len(i.Name))
	}
	if len(i.Name) > NameMaxLength {
		return fmt.Errorf("%w: '%s' value more then max %d - got %d", ErrNameLength, i.Name, NameMaxLength, len(i.Name))
	}
	// Position
	if i.Position < PositionMin {
		return fmt.Errorf("%w: %d value less then min %d", ErrPositionValue, i.Position, PositionMin)
	}
	return nil
}

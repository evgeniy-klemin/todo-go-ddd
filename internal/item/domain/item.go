package domain

import (
	"fmt"
	"time"
)

// Domain errors returned by item operations.
var (
	ErrNameLength    = fmt.Errorf("name has wrong size")
	ErrPositionValue = fmt.Errorf("position has wrong value")
	ErrNotFound      = fmt.Errorf("item not found")
)

const (
	NameMinLength = 1
	NameMaxLength = 1000
	PositionMin   = 1
)

// Name is a value object representing a validated item name.
type Name struct {
	value string
}

func NewName(value string) (Name, error) {
	if len(value) < NameMinLength {
		return Name{}, fmt.Errorf("%w: '%s' value less than min %d - got %d", ErrNameLength, value, NameMinLength, len(value))
	}
	if len(value) > NameMaxLength {
		return Name{}, fmt.Errorf("%w: '%s' value more than max %d - got %d", ErrNameLength, value, NameMaxLength, len(value))
	}
	return Name{value: value}, nil
}

func (n Name) String() string { return n.value }

// Position is a value object representing a validated item position.
type Position struct {
	value int
}

func NewPosition(value int) (Position, error) {
	if value < PositionMin {
		return Position{}, fmt.Errorf("%w: %d value less than min %d", ErrPositionValue, value, PositionMin)
	}
	return Position{value: value}, nil
}

func (p Position) Int() int { return p.value }

// Item is the aggregate root with private fields.
type Item struct {
	id        ModelID
	name      Name
	position  Position
	done      bool
	createdAt time.Time
}

// Getters
func (i *Item) ID() ModelID        { return i.id }
func (i *Item) Name() Name         { return i.name }
func (i *Item) Position() Position  { return i.position }
func (i *Item) Done() bool          { return i.done }
func (i *Item) CreatedAt() time.Time { return i.createdAt }

// NewItem creates a new Item with validation via value objects.
func NewItem(name string, position int) (*Item, error) {
	id, err := GenerateModelID()
	if err != nil {
		return nil, err
	}
	n, err := NewName(name)
	if err != nil {
		return nil, err
	}
	p, err := NewPosition(position)
	if err != nil {
		return nil, err
	}
	return &Item{
		id:        id,
		name:      n,
		position:  p,
		done:      false,
		createdAt: time.Now().Truncate(time.Second),
	}, nil
}

// ReconstituteItem recreates an Item from persistence without validation.
func ReconstituteItem(id ModelID, name string, position int, done bool, createdAt time.Time) *Item {
	return &Item{
		id:        id,
		name:      Name{value: name},
		position:  Position{value: position},
		done:      done,
		createdAt: createdAt,
	}
}

// Behavior methods

func (i *Item) Complete() {
	i.done = true
}

func (i *Item) Uncomplete() {
	i.done = false
}

func (i *Item) Rename(name string) error {
	n, err := NewName(name)
	if err != nil {
		return err
	}
	i.name = n
	return nil
}

func (i *Item) MoveTo(position int) error {
	p, err := NewPosition(position)
	if err != nil {
		return err
	}
	i.position = p
	return nil
}

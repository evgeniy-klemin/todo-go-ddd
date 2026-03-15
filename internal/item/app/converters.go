package app

import "github.com/evgeniy-klemin/todo/internal/item/domain"

func domainToAppItem(d *domain.Item) *Item {
	name := d.Name().String()
	position := d.Position().Int()
	done := d.Done()
	createdAt := d.CreatedAt()
	id := d.ID()
	return &Item{
		ID:        id.String(),
		Name:      &name,
		Position:  &position,
		Done:      &done,
		CreatedAt: &createdAt,
	}
}

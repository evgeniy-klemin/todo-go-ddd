package app

import "github.com/evgeniy-klemin/todo/internal/item/domain"

func domainToAppItem(d *domain.Item, fields []ItemField) *Item {
	if len(fields) == 0 {
		fields = DefaultItemFields
	}
	id := d.ID()
	item := &Item{ID: id.String()}
	for _, field := range fields {
		switch field {
		case ItemFieldName:
			name := d.Name().String()
			item.Name = &name
		case ItemFieldPosition:
			position := d.Position().Int()
			item.Position = &position
		case ItemFieldDone:
			done := d.Done()
			item.Done = &done
		case ItemFieldCreatedAt:
			createdAt := d.CreatedAt()
			item.CreatedAt = &createdAt
		}
	}
	return item
}

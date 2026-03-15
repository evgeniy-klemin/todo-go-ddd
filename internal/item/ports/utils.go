package ports

import (
	"strings"

	"github.com/evgeniy-klemin/todo/internal/item/app"
	"github.com/pkg/errors"
)

func getAppFieldsFromGetParam(fields *string) ([]app.ItemField, error) {
	var res []app.ItemField
	if fields != nil {
		strFields := strings.Split(*fields, ",")
		if len(strFields) > 0 {
			for _, strField := range strFields {
				var field app.ItemField
				switch strField {
				case "name":
					field = app.ItemFieldName
				case "position":
					field = app.ItemFieldPosition
				case "done":
					field = app.ItemFieldDone
				case "created_at":
					field = app.ItemFieldCreatedAt
				default:
					return nil, errors.Errorf("Field %s not found", strField)
				}
				res = append(res, field)
			}
		}
	}
	if len(res) == 0 {
		res = app.DefaultItemFields
	}
	return res, nil
}

func getSortFieldsFromGetParam(sort *string) (app.SortFields, error) {
	var sortFields app.SortFields
	if sort != nil {
		strFields := strings.Split(*sort, ",")
		if len(strFields) > 0 {
			for _, strField := range strFields {
				sortDirection := app.SortDirectionAsc
				if strings.HasPrefix(strField, "-") {
					sortDirection = app.SortDirectionDesc
					strField = strings.TrimLeft(strField, "-")
				}
				var field app.ItemField
				switch strField {
				case "name":
					field = app.ItemFieldName
				case "position":
					field = app.ItemFieldPosition
				case "done":
					field = app.ItemFieldDone
				case "created_at":
					field = app.ItemFieldCreatedAt
				default:
					return nil, errors.Errorf("Field %s not found", strField)
				}
				sortFields = append(sortFields, app.SortField{
					Field:         field,
					SortDirection: sortDirection,
				})
			}
		}
	}
	return sortFields, nil
}

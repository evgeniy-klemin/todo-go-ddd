package ports

import (
	"fmt"
	"strings"
	"time"

	"github.com/evgeniy-klemin/todo/internal/item/app"
)

func appItemsToRespItems(appItems []app.Item) []ItemResponse {
	var res []ItemResponse
	for _, appItem := range appItems {
		var respCreatedAt *time.Time
		if appItem.CreatedAt != nil {
			createdAt := (*appItem.CreatedAt).UTC()
			respCreatedAt = &createdAt
		}
		res = append(res, ItemResponse{
			Id:        appItem.ID,
			Name:      appItem.Name,
			Position:  appItem.Position,
			Done:      appItem.Done,
			CreatedAt: respCreatedAt,
		})
	}
	return res
}

func appItemToResp(appItem *app.Item) *ItemResponse {
	var respCreatedAt *time.Time
	if appItem.CreatedAt != nil {
		createdAt := (*appItem.CreatedAt).UTC()
		respCreatedAt = &createdAt
	}
	return &ItemResponse{
		Id:        appItem.ID,
		Name:      appItem.Name,
		Position:  appItem.Position,
		Done:      appItem.Done,
		CreatedAt: respCreatedAt,
	}
}

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
					return nil, fmt.Errorf("field %s not found", strField)
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
					return nil, fmt.Errorf("field %s not found", strField)
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

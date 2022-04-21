package ports

import (
	"strings"
	"time"

	"github.com/evgeniy-klemin/todo/internal/item/app"
	"github.com/evgeniy-klemin/todo/internal/item/domain"
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

func domainItemToResp(domainItem *domain.Item) *ItemResponse {
	return &ItemResponse{
		Id:        domainItem.ID.String(),
		Name:      &domainItem.Name,
		Position:  &domainItem.Position,
		Done:      &domainItem.Done,
		CreatedAt: &domainItem.CreatedAt,
	}
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

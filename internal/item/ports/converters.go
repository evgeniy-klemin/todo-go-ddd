package ports

import (
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

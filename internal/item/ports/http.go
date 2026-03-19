//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=oapicfg/types.yaml ../../../docs/todo.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=oapicfg/api.yaml ../../../docs/todo.yaml
package ports

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/evgeniy-klemin/todo/internal/item/app"
)

type HttpServer struct {
	itemService *app.ItemService
}

func NewHttpServer(itemService *app.ItemService) *HttpServer {
	return &HttpServer{
		itemService: itemService,
	}
}

// (GET /items)
func (h *HttpServer) GetItems(ctx echo.Context, params GetItemsParams) error {
	fields, err := getAppFieldsFromGetParam(params.Fields)
	if err != nil {
		return err
	}
	perPage := 20
	if params.PerPage != nil {
		perPage = int(*params.PerPage)
	}
	sortFields, err := getSortFieldsFromGetParam(params.Sort)
	if err != nil {
		return err
	}
	if len(sortFields) == 0 {
		sortFields = app.SortFields{{Field: app.ItemFieldPosition, SortDirection: app.SortDirectionAsc}}
	}

	var cursor *app.Cursor
	if params.Cursor != nil && *params.Cursor != "" {
		cursor, err = app.DecodeCursor(string(*params.Cursor))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid cursor")
		}
	}

	items, err := h.itemService.All(ctx.Request().Context(), params.Done, fields, perPage+1, cursor, sortFields)
	if err != nil {
		ctx.Error(err)
		return err
	}

	hasNext := len(items) > perPage
	if hasNext {
		items = items[:perPage]
	}

	var nextCursorEncoded string
	if hasNext && len(items) > 0 {
		lastItem := items[len(items)-1]
		nextCursor := app.BuildCursorFromItem(lastItem, sortFields)
		nextCursorEncoded, err = app.EncodeCursor(nextCursor)
		if err != nil {
			ctx.Error(err)
			return err
		}
	}

	totalCount, err := h.itemService.Count(ctx.Request().Context(), params.Done)
	if err != nil {
		ctx.Error(err)
		return err
	}

	respItems := appItemsToRespItems(items)
	ctx.Response().Header().Set("X-Per-Page", strconv.Itoa(perPage))
	ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(totalCount))
	if hasNext {
		ctx.Response().Header().Set("X-Next-Cursor", nextCursorEncoded)
	}
	ctx.Response().Header().Set("Link", cursorLinks(ctx.Request(), perPage, hasNext, nextCursorEncoded))

	return ctx.JSON(http.StatusOK, respItems)
}

// Create New User
// (POST /items)
func (h *HttpServer) PostItems(ctx echo.Context) error {
	var itemPost ItemPost
	if err := ctx.Bind(&itemPost); err != nil {
		ctx.Error(err)
		return errors.Wrap(err, "некорректные параметры")
	}
	var position *int
	if itemPost.Position != nil {
		positionVal := int(*itemPost.Position)
		position = &positionVal
	}
	item, err := h.itemService.Create(ctx.Request().Context(), itemPost.Name, position)
	if err != nil {
		ctx.Error(err)
		return err
	}

	respItem := domainItemToResp(item)

	return ctx.JSON(http.StatusCreated, respItem)
}

// Get Item Info by Item ID
// (GET /items/{item_id})
func (h *HttpServer) GetItemsItemId(ctx echo.Context, itemId ItemId) error {
	item, err := h.itemService.GetItemByID(ctx.Request().Context(), string(itemId))
	if err != nil {
		ctx.Error(err)
		return err
	}

	respItem := domainItemToResp(item)

	return ctx.JSON(http.StatusOK, respItem)
}

// Update Item
// (PATCH /items/{item_id})
func (h *HttpServer) PatchItemsItemid(ctx echo.Context, itemId ItemId) error {
	var itemPatch ItemPatch
	if err := ctx.Bind(&itemPatch); err != nil {
		ctx.Error(err)
		return errors.Wrap(err, "некорректные параметры")
	}

	appItem := &app.Item{
		ID:       string(itemId),
		Name:     itemPatch.Name,
		Position: itemPatch.Position,
		Done:     itemPatch.Done,
	}

	item, err := h.itemService.Update(ctx.Request().Context(), appItem)
	if err != nil {
		ctx.Error(err)
		return err
	}

	respItem := domainItemToResp(item)

	return ctx.JSON(http.StatusOK, respItem)
}

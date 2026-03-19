//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=oapicfg/types.yaml ../../../docs/todo.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=oapicfg/api.yaml ../../../docs/todo.yaml
package ports

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/evgeniy-klemin/todo/internal/item/app"
)

// errorStatusMap maps app error kinds to HTTP status codes.
var errorStatusMap = map[error]int{
	app.ErrNotFound:   http.StatusNotFound,
	app.ErrValidation: http.StatusUnprocessableEntity,
}

// ItemService defines the port that the HTTP handler requires from the application layer.
type ItemService interface {
	Create(ctx context.Context, name string, position *int) (*app.Item, error)
	GetItemByID(ctx context.Context, id string) (*app.Item, error)
	List(ctx context.Context, query app.ListQuery) (app.ListResult, error)
	Update(ctx context.Context, reqItem *app.Item) (*app.Item, error)
	All(ctx context.Context, done *bool, fields []app.ItemField, limit int, cursor *app.Cursor, sortFields app.SortFields) ([]app.Item, error)
	Count(ctx context.Context, done *bool) (int, error)
}

func httpError(ctx echo.Context, err error) error {
	var appErr *app.AppError
	if errors.As(err, &appErr) {
		if status, ok := errorStatusMap[appErr.Kind]; ok {
			return ctx.JSON(status, map[string]string{"error": appErr.UserMessage})
		}
	}
	return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

type HttpServer struct {
	itemService ItemService
}

func NewHttpServer(itemService ItemService) *HttpServer {
	return &HttpServer{
		itemService: itemService,
	}
}

// (GET /items)
func (h *HttpServer) GetItems(ctx echo.Context, params GetItemsParams) error {
	fields, err := getAppFieldsFromGetParam(params.Fields)
	if err != nil {
		return httpError(ctx, err)
	}
	sortFields, err := getSortFieldsFromGetParam(params.Sort)
	if err != nil {
		return httpError(ctx, err)
	}
	if len(sortFields) == 0 {
		sortFields = app.SortFields{{Field: app.ItemFieldPosition, SortDirection: app.SortDirectionAsc}}
	}

	perPage := 20
	if params.PerPage != nil {
		perPage = int(*params.PerPage)
	}

	// Cursor-based pagination: used when _cursor param is present (no search support).
	// Offset-based pagination: used otherwise (supports search via q param).
	if params.Cursor != nil && *params.Cursor != "" {
		return h.getItemsWithCursor(ctx, params, fields, sortFields, perPage)
	}
	return h.getItemsWithOffset(ctx, params, fields, sortFields, perPage)
}

func (h *HttpServer) getItemsWithCursor(ctx echo.Context, params GetItemsParams, fields []app.ItemField, sortFields app.SortFields, perPage int) error {
	cursor, err := app.DecodeCursor(string(*params.Cursor))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid cursor")
	}

	items, err := h.itemService.All(ctx.Request().Context(), params.Done, fields, perPage+1, cursor, sortFields)
	if err != nil {
		return httpError(ctx, err)
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
			return httpError(ctx, err)
		}
	}

	totalCount, err := h.itemService.Count(ctx.Request().Context(), params.Done)
	if err != nil {
		return httpError(ctx, err)
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

func (h *HttpServer) getItemsWithOffset(ctx echo.Context, params GetItemsParams, fields []app.ItemField, sortFields app.SortFields, perPage int) error {
	// Parse _page from raw query string since it's not in the OpenAPI spec (removed in master)
	page := 1
	if p := ctx.QueryParam("_page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	query := app.ListQuery{
		Done:       params.Done,
		Search:     params.Q,
		Fields:     fields,
		Page:       page,
		PerPage:    perPage,
		SortFields: sortFields,
	}

	result, err := h.itemService.List(ctx.Request().Context(), query)
	if err != nil {
		return httpError(ctx, err)
	}

	ctx.Response().Header().Set("X-Page", strconv.Itoa(query.Page))
	ctx.Response().Header().Set("X-Per-Page", strconv.Itoa(query.PerPage))
	ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(result.TotalCount))

	return ctx.JSON(http.StatusOK, appItemsToRespItems(result.Items))
}

// Create New User
// (POST /items)
func (h *HttpServer) PostItems(ctx echo.Context) error {
	var itemPost ItemPost
	if err := ctx.Bind(&itemPost); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "некорректные параметры"})
	}
	var position *int
	if itemPost.Position != nil {
		positionVal := int(*itemPost.Position)
		position = &positionVal
	}
	item, err := h.itemService.Create(ctx.Request().Context(), itemPost.Name, position)
	if err != nil {
		return httpError(ctx, err)
	}

	respItem := appItemToResp(item)

	return ctx.JSON(http.StatusCreated, respItem)
}

// Get Item Info by Item ID
// (GET /items/{item_id})
func (h *HttpServer) GetItemsItemId(ctx echo.Context, itemId ItemId) error {
	item, err := h.itemService.GetItemByID(ctx.Request().Context(), string(itemId))
	if err != nil {
		return httpError(ctx, err)
	}

	respItem := appItemToResp(item)

	return ctx.JSON(http.StatusOK, respItem)
}

// Update Item
// (PATCH /items/{item_id})
func (h *HttpServer) PatchItemsItemid(ctx echo.Context, itemId ItemId) error {
	var itemPatch ItemPatch
	if err := ctx.Bind(&itemPatch); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "некорректные параметры"})
	}

	appItem := &app.Item{
		ID:       string(itemId),
		Name:     itemPatch.Name,
		Position: itemPatch.Position,
		Done:     itemPatch.Done,
	}

	item, err := h.itemService.Update(ctx.Request().Context(), appItem)
	if err != nil {
		return httpError(ctx, err)
	}

	respItem := appItemToResp(item)

	return ctx.JSON(http.StatusOK, respItem)
}

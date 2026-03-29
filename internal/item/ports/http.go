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
	Update(ctx context.Context, reqItem *app.Item) (*app.Item, error)
	// All fetches items using cursor-based pagination.
	// done and search are optional filters (nil = no filter).
	// fields restricts which item fields are populated (nil = all).
	// limit is the maximum number of items to return; the handler passes perPage+1 to detect hasNext.
	// cursorData is an opaque []byte from the previous response (nil = first page).
	// sortFields defines column order and must be consistent across pages.
	// Returns: items slice, opaque nextCursor for the next call (nil if no items), and any error.
	All(ctx context.Context, done *bool, search *string, fields []app.ItemField, limit int, cursorData []byte, sortFields app.SortFields) ([]app.Item, []byte, error)
	// Count returns the total number of items matching done and search filters,
	// independent of pagination. Used to set the X-Total-Count response header.
	Count(ctx context.Context, done *bool, search *string) (int, error)
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

	var cursorData []byte
	if params.Cursor != nil && *params.Cursor != "" {
		cursorData, err = decodeCursor(string(*params.Cursor))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid cursor")
		}
	}

	items, nextCursor, err := h.itemService.All(ctx.Request().Context(), params.Done, params.Q, fields, perPage+1, cursorData, sortFields)
	if err != nil {
		return httpError(ctx, err)
	}

	hasNext := len(items) > perPage
	if hasNext {
		items = items[:perPage]
	}

	var nextCursorEncoded string
	if hasNext && len(nextCursor) > 0 {
		nextCursorEncoded = encodeCursor(nextCursor)
	}

	totalCount, err := h.itemService.Count(ctx.Request().Context(), params.Done, params.Q)
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

// Create New Item
// (POST /items)
func (h *HttpServer) PostItems(ctx echo.Context) error {
	var itemPost ItemPost
	if err := ctx.Bind(&itemPost); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request parameters"})
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
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request parameters"})
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

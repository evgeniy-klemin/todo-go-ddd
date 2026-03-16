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
	query, err := buildListQuery(params)
	if err != nil {
		return httpError(ctx, err)
	}

	result, err := h.itemService.List(ctx.Request().Context(), query)
	if err != nil {
		return httpError(ctx, err)
	}

	paginator := NewPaginator(query.Page, query.PerPage, result.TotalCount)
	ctx.Response().Header().Set("X-Page", strconv.Itoa(query.Page))
	ctx.Response().Header().Set("X-Per-Page", strconv.Itoa(query.PerPage))
	ctx.Response().Header().Set("X-Total-Count", strconv.Itoa(result.TotalCount))
	ctx.Response().Header().Set("Link", links(ctx.Request(), paginator))

	return ctx.JSON(http.StatusOK, appItemsToRespItems(result.Items))
}

func buildListQuery(params GetItemsParams) (app.ListQuery, error) {
	fields, err := getAppFieldsFromGetParam(params.Fields)
	if err != nil {
		return app.ListQuery{}, err
	}
	sortFields, err := getSortFieldsFromGetParam(params.Sort)
	if err != nil {
		return app.ListQuery{}, err
	}
	page := 1
	if params.Page != nil {
		page = int(*params.Page)
	}
	perPage := 20
	if params.PerPage != nil {
		perPage = int(*params.PerPage)
	}
	return app.ListQuery{
		Done:       params.Done,
		Search:     params.Q,
		Fields:     fields,
		Page:       page,
		PerPage:    perPage,
		SortFields: sortFields,
	}, nil
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

package ports

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/evgeniy-klemin/todo/internal/item/app"
	"github.com/evgeniy-klemin/todo/internal/item/domain"
)

func testTime() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}

// --- mock app.Service ---

type mockService struct {
	createFn  func(ctx context.Context, name string, position *int) (*domain.Item, error)
	getByIDFn func(ctx context.Context, id string) (*domain.Item, error)
	updateFn  func(ctx context.Context, reqItem *app.Item) (*domain.Item, error)
	listFn    func(ctx context.Context, query app.ListQuery) (app.ListResult, error)
}

func (m *mockService) Create(ctx context.Context, name string, position *int) (*domain.Item, error) {
	return m.createFn(ctx, name, position)
}
func (m *mockService) GetItemByID(ctx context.Context, id string) (*domain.Item, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockService) Update(ctx context.Context, reqItem *app.Item) (*domain.Item, error) {
	return m.updateFn(ctx, reqItem)
}
func (m *mockService) List(ctx context.Context, query app.ListQuery) (app.ListResult, error) {
	if m.listFn != nil {
		return m.listFn(ctx, query)
	}
	return app.ListResult{}, nil
}

func newTestServer(svc app.Service) *HttpServer {
	return NewHttpServer(svc)
}

func newEchoContext(method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// --- tests ---

func TestPostItems_Returns201OnSuccess(t *testing.T) {
	id, _ := domain.NewModelID("00000000-0000-0000-0000-000000000001")
	svc := &mockService{
		createFn: func(_ context.Context, name string, position *int) (*domain.Item, error) {
			item, _ := domain.NewItem(name, 1)
			item.ID = id
			return item, nil
		},
	}

	server := newTestServer(svc)
	ctx, rec := newEchoContext(http.MethodPost, "/items", `{"name":"Buy milk"}`)

	if err := server.PostItems(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestPostItems_InvalidName_Returns422(t *testing.T) {
	svc := &mockService{
		createFn: func(_ context.Context, name string, position *int) (*domain.Item, error) {
			return nil, domain.ErrNameLength
		},
	}

	server := newTestServer(svc)
	ctx, rec := newEchoContext(http.MethodPost, "/items", `{"name":""}`)

	_ = server.PostItems(ctx)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

func TestGetItems_Returns200(t *testing.T) {
	svc := &mockService{
		listFn: func(_ context.Context, _ app.ListQuery) (app.ListResult, error) {
			return app.ListResult{}, nil
		},
	}

	server := newTestServer(svc)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := server.GetItems(ctx, GetItemsParams{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestPatchItemsItemid_DoneTrue_Returns200(t *testing.T) {
	id, _ := domain.NewModelID("00000000-0000-0000-0000-000000000001")
	svc := &mockService{
		updateFn: func(_ context.Context, reqItem *app.Item) (*domain.Item, error) {
			return domain.ReconstituteItem(id, "Task", 1, true, testTime()), nil
		},
	}

	server := newTestServer(svc)
	ctx, rec := newEchoContext(http.MethodPatch, "/items/00000000-0000-0000-0000-000000000001", `{"done":true}`)

	if err := server.PatchItemsItemid(ctx, "00000000-0000-0000-0000-000000000001"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	done, ok := body["done"].(bool)
	if !ok || !done {
		t.Errorf("expected done=true in response, got %v", body["done"])
	}
}

func TestGetItemsItemId_NotFound_Returns404(t *testing.T) {
	svc := &mockService{
		getByIDFn: func(_ context.Context, _ string) (*domain.Item, error) {
			return nil, domain.ErrNotFound
		},
	}

	server := newTestServer(svc)
	ctx, rec := newEchoContext(http.MethodGet, "/items/123", "")

	_ = server.GetItemsItemId(ctx, "00000000-0000-0000-0000-000000000001")

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if _, ok := body["error"]; !ok {
		t.Error("expected 'error' field in response body")
	}
}

package ports

import (
	"context"
	"encoding/json"
	"fmt"
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

// --- mock ItemService ---

type mockService struct {
	createFn  func(ctx context.Context, name string, position *int) (*app.Item, error)
	getByIDFn func(ctx context.Context, id string) (*app.Item, error)
	updateFn  func(ctx context.Context, reqItem *app.Item) (*app.Item, error)
	allFn     func(ctx context.Context, done *bool, search *string, fields []app.ItemField, limit int, cursorData []byte, sortFields app.SortFields) ([]app.Item, []byte, error)
	countFn   func(ctx context.Context, done *bool, search *string) (int, error)
}

func (m *mockService) Create(ctx context.Context, name string, position *int) (*app.Item, error) {
	return m.createFn(ctx, name, position)
}
func (m *mockService) GetItemByID(ctx context.Context, id string) (*app.Item, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockService) Update(ctx context.Context, reqItem *app.Item) (*app.Item, error) {
	return m.updateFn(ctx, reqItem)
}

func (m *mockService) All(ctx context.Context, done *bool, search *string, fields []app.ItemField, limit int, cursorData []byte, sortFields app.SortFields) ([]app.Item, []byte, error) {
	if m.allFn != nil {
		return m.allFn(ctx, done, search, fields, limit, cursorData, sortFields)
	}
	return nil, nil, nil
}

func (m *mockService) Count(ctx context.Context, done *bool, search *string) (int, error) {
	if m.countFn != nil {
		return m.countFn(ctx, done, search)
	}
	return 0, nil
}

func intPtr(v int) *int       { return &v }
func boolPtr(v bool) *bool    { return &v }
func strPtr(v string) *string { return &v }
func timePtr(v time.Time) *time.Time { return &v }

func newTestServer(svc ItemService) *HttpServer {
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
	svc := &mockService{
		createFn: func(_ context.Context, name string, position *int) (*app.Item, error) {
			t := testTime()
			return &app.Item{ID: "00000000-0000-0000-0000-000000000001", Name: strPtr(name), Position: intPtr(1), Done: boolPtr(false), CreatedAt: timePtr(t)}, nil
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
		createFn: func(_ context.Context, name string, position *int) (*app.Item, error) {
			return nil, app.Validation("create item", domain.ErrNameLength)
		},
	}

	server := newTestServer(svc)
	ctx, rec := newEchoContext(http.MethodPost, "/items", `{"name":""}`)

	_ = server.PostItems(ctx)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

func TestGetItems_WithSearchParam_PassesToService(t *testing.T) {
	var capturedSearch *string
	svc := &mockService{
		allFn: func(_ context.Context, _ *bool, search *string, _ []app.ItemField, _ int, _ []byte, _ app.SortFields) ([]app.Item, []byte, error) {
			capturedSearch = search
			return nil, nil, nil
		},
	}

	server := newTestServer(svc)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/items?q=buy", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	search := "buy"
	if err := server.GetItems(ctx, GetItemsParams{Q: &search}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedSearch == nil {
		t.Fatal("expected search to be set, got nil")
	}
	if *capturedSearch != "buy" {
		t.Errorf("expected search 'buy', got '%s'", *capturedSearch)
	}
}

func TestGetItems_Returns200(t *testing.T) {
	svc := &mockService{}

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
	svc := &mockService{
		updateFn: func(_ context.Context, reqItem *app.Item) (*app.Item, error) {
			tt := testTime()
			return &app.Item{ID: "00000000-0000-0000-0000-000000000001", Name: strPtr("Task"), Position: intPtr(1), Done: boolPtr(true), CreatedAt: timePtr(tt)}, nil
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

	var body map[string]any
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
		getByIDFn: func(_ context.Context, _ string) (*app.Item, error) {
			return nil, app.NotFound("get item by id", fmt.Errorf("item not found"))
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

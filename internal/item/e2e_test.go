package item_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"

	"github.com/evgeniy-klemin/todo/db/schema"
	item "github.com/evgeniy-klemin/todo/internal/item"
	"github.com/evgeniy-klemin/todo/internal/item/ports"
)

func setupE2EDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if _, err := schema.ApplyAll(db); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	return db
}

func setupE2EServer(t *testing.T) (*echo.Echo, *sql.DB) {
	t.Helper()
	db := setupE2EDB(t)
	container := item.NewContainer(db)

	e := echo.New()
	container.RegisterHandlers(e)
	return e, db
}

func createItem(t *testing.T, e *echo.Echo, name string) {
	t.Helper()
	body := `{"name":"` + name + `"}`
	req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("createItem(%q): expected 201, got %d: %s", name, rec.Code, rec.Body.String())
	}
}

func searchItems(t *testing.T, e *echo.Echo, query string) []ports.ItemResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/items?q="+url.QueryEscape(query), nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("searchItems(%q): expected 200, got %d: %s", query, rec.Code, rec.Body.String())
	}
	var items []ports.ItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatalf("searchItems(%q): unmarshal: %v", query, err)
	}
	return items
}

// getItems sends a GET /items request. The queryString parameter must be pre-encoded
// by the caller (e.g. "q=task&_per_page=2"). Callers are responsible for proper URL encoding.
func getItems(t *testing.T, e *echo.Echo, queryString string) []ports.ItemResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/items?"+queryString, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("getItems(%q): expected 200, got %d: %s", queryString, rec.Code, rec.Body.String())
	}
	var items []ports.ItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatalf("getItems(%q): unmarshal: %v", queryString, err)
	}
	return items
}

func patchItem(t *testing.T, e *echo.Echo, id string, body string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPatch, "/items/"+id, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patchItem(%q): expected 200, got %d: %s", id, rec.Code, rec.Body.String())
	}
}

func TestE2E_SearchByExactWord(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	createItem(t, e, "Buy milk")
	createItem(t, e, "Buy eggs")
	createItem(t, e, "Walk dog")

	items := searchItems(t, e, "buy")
	if len(items) != 2 {
		t.Errorf("expected 2 results for 'buy', got %d", len(items))
	}
}

func TestE2E_SearchByPrefix(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	createItem(t, e, "Buying groceries")
	createItem(t, e, "Walk the dog")

	items := searchItems(t, e, "buy")
	if len(items) != 1 {
		t.Errorf("expected 1 result for prefix 'buy' matching 'Buying groceries', got %d", len(items))
	}
}

func TestE2E_SearchCaseInsensitive(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	createItem(t, e, "Buy Milk")

	items := searchItems(t, e, "buy")
	if len(items) != 1 {
		t.Errorf("expected 1 result for case-insensitive 'buy', got %d", len(items))
	}
}

func TestE2E_SearchNoResults(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	createItem(t, e, "Buy milk")

	items := searchItems(t, e, "xyz")
	if len(items) != 0 {
		t.Errorf("expected 0 results for 'xyz', got %d", len(items))
	}
}

func TestE2E_SearchWithPagination(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	// Create 5 items with "Task" in the name
	for i := 1; i <= 5; i++ {
		createItem(t, e, "Task item number "+strings.Repeat("x", i))
	}
	createItem(t, e, "Walk dog")

	// Search for "task" with pagination: page 1, 2 per page
	items := getItems(t, e, "q=task&_per_page=2&_page=1")
	if len(items) != 2 {
		t.Errorf("expected 2 results on page 1, got %d", len(items))
	}

	// Page 2
	items = getItems(t, e, "q=task&_per_page=2&_page=2")
	if len(items) != 2 {
		t.Errorf("expected 2 results on page 2, got %d", len(items))
	}

	// Page 3 — should have 1 remaining
	items = getItems(t, e, "q=task&_per_page=2&_page=3")
	if len(items) != 1 {
		t.Errorf("expected 1 result on page 3, got %d", len(items))
	}
}

func TestE2E_SearchCombinedWithDoneFilter(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	createItem(t, e, "Buy milk")
	createItem(t, e, "Buy eggs")
	createItem(t, e, "Walk dog")

	// Get all items to find the ID of "Buy milk"
	allItems := getItems(t, e, "")
	var buyMilkID string
	for _, it := range allItems {
		if it.Name != nil && *it.Name == "Buy milk" {
			buyMilkID = it.Id
			break
		}
	}
	if buyMilkID == "" {
		t.Fatal("could not find 'Buy milk' item")
	}

	// Mark "Buy milk" as done
	patchItem(t, e, buyMilkID, `{"done":true}`)

	// Search for "buy" with done=true — should only get "Buy milk"
	items := getItems(t, e, "q=buy&done=true")
	if len(items) != 1 {
		t.Errorf("expected 1 done result for 'buy', got %d", len(items))
	}
	if items[0].Name != nil && *items[0].Name != "Buy milk" {
		t.Errorf("expected 'Buy milk', got %v", items[0].Name)
	}
}

func TestE2E_CreateAndSearchImmediately(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	createItem(t, e, "Unique search term alpha")

	// Should be immediately searchable via FTS trigger
	items := searchItems(t, e, "alpha")
	if len(items) != 1 {
		t.Errorf("expected 1 result for 'alpha' immediately after creation, got %d", len(items))
	}
}

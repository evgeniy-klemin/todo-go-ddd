package item_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"

	"github.com/evgeniy-klemin/todo/db/driver"
	"github.com/evgeniy-klemin/todo/db/fts"
	"github.com/evgeniy-klemin/todo/db/migrations"
	item "github.com/evgeniy-klemin/todo/internal/item"
	"github.com/evgeniy-klemin/todo/internal/item/ports"
)

func setupE2EDB(t *testing.T) (*sql.DB, bool) {
	t.Helper()
	db, err := sql.Open(driver.SQLite, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := migrations.Run(db, driver.SQLite); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	ftsEnabled := fts.Apply(db)
	return db, ftsEnabled
}

func setupE2EServer(t *testing.T) (*echo.Echo, *sql.DB) {
	t.Helper()
	db, ftsEnabled := setupE2EDB(t)
	container := item.NewContainer(db, driver.SQLite, ftsEnabled)

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

	// Search for "task" with cursor pagination: 2 per page, first page
	items, rec := getItemsRaw(t, e, "q=task&_per_page=2")
	if len(items) != 2 {
		t.Errorf("expected 2 results on page 1, got %d", len(items))
	}
	nextCursor := rec.Header().Get("X-Next-Cursor")
	if nextCursor == "" {
		t.Fatal("expected X-Next-Cursor header for page 1")
	}

	// Page 2 using cursor
	items, rec = getItemsRaw(t, e, "q=task&_per_page=2&_cursor="+url.QueryEscape(nextCursor))
	if len(items) != 2 {
		t.Errorf("expected 2 results on page 2, got %d", len(items))
	}
	nextCursor = rec.Header().Get("X-Next-Cursor")
	if nextCursor == "" {
		t.Fatal("expected X-Next-Cursor header for page 2")
	}

	// Page 3 — should have 1 remaining
	items, _ = getItemsRaw(t, e, "q=task&_per_page=2&_cursor="+url.QueryEscape(nextCursor))
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

// getItemsRaw sends a GET /items request and returns both items and the recorder
// so callers can inspect response headers like X-Total-Count.
func getItemsRaw(t *testing.T, e *echo.Echo, queryString string) ([]ports.ItemResponse, *httptest.ResponseRecorder) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/items?"+queryString, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("getItemsRaw(%q): expected 200, got %d: %s", queryString, rec.Code, rec.Body.String())
	}
	var items []ports.ItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatalf("getItemsRaw(%q): unmarshal: %v", queryString, err)
	}
	return items, rec
}

func TestE2E_CursorPaginationBasic(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	for i := 1; i <= 5; i++ {
		createItem(t, e, fmt.Sprintf("Task %d", i))
	}

	// Page 1: 2 items, X-Next-Cursor present
	items, rec := getItemsRaw(t, e, "_per_page=2")
	if len(items) != 2 {
		t.Errorf("page 1: expected 2 items, got %d", len(items))
	}
	nextCursor := rec.Header().Get("X-Next-Cursor")
	if nextCursor == "" {
		t.Fatal("page 1: expected X-Next-Cursor header")
	}

	// Page 2: 2 more items, X-Next-Cursor present
	items, rec = getItemsRaw(t, e, "_per_page=2&_cursor="+url.QueryEscape(nextCursor))
	if len(items) != 2 {
		t.Errorf("page 2: expected 2 items, got %d", len(items))
	}
	nextCursor = rec.Header().Get("X-Next-Cursor")
	if nextCursor == "" {
		t.Fatal("page 2: expected X-Next-Cursor header")
	}

	// Page 3: 1 item, no X-Next-Cursor
	items, rec = getItemsRaw(t, e, "_per_page=2&_cursor="+url.QueryEscape(nextCursor))
	if len(items) != 1 {
		t.Errorf("page 3: expected 1 item, got %d", len(items))
	}
	if got := rec.Header().Get("X-Next-Cursor"); got != "" {
		t.Errorf("page 3: expected no X-Next-Cursor, got %q", got)
	}
}

func TestE2E_NoCursorOnLastPage(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	createItem(t, e, "Item 1")
	createItem(t, e, "Item 2")
	createItem(t, e, "Item 3")

	items, rec := getItemsRaw(t, e, "_per_page=5")
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
	if got := rec.Header().Get("X-Next-Cursor"); got != "" {
		t.Errorf("expected no X-Next-Cursor on last page, got %q", got)
	}
}

func TestE2E_FirstPageNoCursor(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	createItem(t, e, "Alpha")
	createItem(t, e, "Beta")
	createItem(t, e, "Gamma")

	items := getItems(t, e, "")
	if len(items) != 3 {
		t.Errorf("expected 3 items from first page (no cursor), got %d", len(items))
	}
}

func TestE2E_InvalidCursor(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/items?_cursor=invalid_base64_garbage!!!", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid cursor, got %d", rec.Code)
	}
}

func TestE2E_SortWithCursor(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	// Create items — positions are assigned sequentially (1,2,3,4,5)
	for i := 1; i <= 5; i++ {
		createItem(t, e, fmt.Sprintf("Item %d", i))
	}

	// Sort by position descending, 2 per page
	items, rec := getItemsRaw(t, e, "_sort=-position&_per_page=2")
	if len(items) != 2 {
		t.Errorf("page 1: expected 2 items, got %d", len(items))
	}
	// First item should have highest position (5)
	if items[0].Position == nil || *items[0].Position != 5 {
		t.Errorf("page 1 first item: expected position 5, got %v", items[0].Position)
	}
	nextCursor := rec.Header().Get("X-Next-Cursor")
	if nextCursor == "" {
		t.Fatal("page 1: expected X-Next-Cursor header")
	}

	// Follow cursor — remaining items still in desc order
	items, _ = getItemsRaw(t, e, "_sort=-position&_per_page=2&_cursor="+url.QueryEscape(nextCursor))
	if len(items) != 2 {
		t.Errorf("page 2: expected 2 items, got %d", len(items))
	}
	// Should be positions 3, 2
	if items[0].Position == nil || *items[0].Position != 3 {
		t.Errorf("page 2 first item: expected position 3, got %v", items[0].Position)
	}
}

func TestE2E_EmptyDB(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	items, rec := getItemsRaw(t, e, "")
	if len(items) != 0 {
		t.Errorf("expected empty list, got %d items", len(items))
	}
	if got := rec.Header().Get("X-Next-Cursor"); got != "" {
		t.Errorf("expected no X-Next-Cursor for empty DB, got %q", got)
	}
}

func TestE2E_PerPageHeader(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	_, rec := getItemsRaw(t, e, "_per_page=5")
	if got := rec.Header().Get("X-Per-Page"); got != "5" {
		t.Errorf("expected X-Per-Page: 5, got %q", got)
	}
}

func TestE2E_LinkHeader(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	for i := 1; i <= 3; i++ {
		createItem(t, e, fmt.Sprintf("Link item %d", i))
	}

	// With hasNext: Link header should contain rel=next and rel=first
	_, rec := getItemsRaw(t, e, "_per_page=2")
	linkHeader := rec.Header().Get("Link")
	if linkHeader == "" {
		t.Fatal("expected Link header, got empty")
	}
	if !strings.Contains(linkHeader, "rel=next") {
		t.Errorf("expected Link header to contain rel=next, got %q", linkHeader)
	}
	if !strings.Contains(linkHeader, "rel=first") {
		t.Errorf("expected Link header to contain rel=first, got %q", linkHeader)
	}

	// Without hasNext (all items fit): Link header should contain only rel=first
	_, rec = getItemsRaw(t, e, "_per_page=10")
	linkHeader = rec.Header().Get("Link")
	if linkHeader == "" {
		t.Fatal("expected Link header even on last page, got empty")
	}
	if strings.Contains(linkHeader, "rel=next") {
		t.Errorf("expected no rel=next on last page, got %q", linkHeader)
	}
	if !strings.Contains(linkHeader, "rel=first") {
		t.Errorf("expected rel=first on last page, got %q", linkHeader)
	}
}

func TestE2E_TotalCountWithTextSearch(t *testing.T) {
	e, db := setupE2EServer(t)
	defer db.Close()

	createItem(t, e, "buy groceries")
	createItem(t, e, "buy milk")
	createItem(t, e, "clean house")
	createItem(t, e, "task report")
	createItem(t, e, "task review")

	// Test q=buy — should return 2 items with X-Total-Count=2
	items, rec := getItemsRaw(t, e, "q=buy")
	if len(items) != 2 {
		t.Errorf("q=buy: expected 2 items, got %d", len(items))
	}
	if got := rec.Header().Get("X-Total-Count"); got != "2" {
		t.Errorf("q=buy: expected X-Total-Count=2, got %q", got)
	}

	// Test q=task — should return 2 items with X-Total-Count=2
	items, rec = getItemsRaw(t, e, "q=task")
	if len(items) != 2 {
		t.Errorf("q=task: expected 2 items, got %d", len(items))
	}
	if got := rec.Header().Get("X-Total-Count"); got != "2" {
		t.Errorf("q=task: expected X-Total-Count=2, got %q", got)
	}

	// Test q=nonexistent — should return empty list with X-Total-Count=0
	items, rec = getItemsRaw(t, e, "q=nonexistent")
	if len(items) != 0 {
		t.Errorf("q=nonexistent: expected 0 items, got %d", len(items))
	}
	if got := rec.Header().Get("X-Total-Count"); got != "0" {
		t.Errorf("q=nonexistent: expected X-Total-Count=0, got %q", got)
	}

	// Test q=buy&_per_page=1 — should return 1 item but X-Total-Count=2 (filtered total)
	items, rec = getItemsRaw(t, e, "q=buy&_per_page=1")
	if len(items) != 1 {
		t.Errorf("q=buy&_per_page=1: expected 1 item, got %d", len(items))
	}
	if got := rec.Header().Get("X-Total-Count"); got != "2" {
		t.Errorf("q=buy&_per_page=1: expected X-Total-Count=2, got %q", got)
	}
}

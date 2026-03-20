package item

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/evgeniy-klemin/todo/internal/item/domain"
	"github.com/evgeniy-klemin/todo/internal/item/repository"
	"github.com/evgeniy-klemin/todo/test/setup"
)

type SearchSuite struct {
	setup.MySQLSuite
}

func TestSearch(t *testing.T) {
	suite.Run(t, new(SearchSuite))
}

func (s *SearchSuite) insertItems(names ...string) {
	ctx := context.Background()
	repo := repository.NewMySQL(s.DB, s.FTSEnabled)
	for _, name := range names {
		item, err := domain.NewItem(name, 1)
		s.Require().NoError(err)
		_, err = repo.AddWithNextPosition(ctx, item)
		s.Require().NoError(err)
	}
}

func (s *SearchSuite) TestExactWordMatch() {
	s.insertItems("Buy milk", "Buy eggs", "Walk the dog")
	repo := repository.NewMySQL(s.DB, s.FTSEnabled)

	search := "buy"
	items, err := repo.ListWithCursor(context.Background(), domain.ListFilter{Search: &search}, nil, 20, nil)
	s.Require().NoError(err)
	s.Equal(2, len(items), "expected 2 items matching 'buy'")
}

func (s *SearchSuite) TestPrefixMatch() {
	s.insertItems("Buying groceries", "Buy milk", "Walk the dog")
	repo := repository.NewMySQL(s.DB, s.FTSEnabled)

	search := "buy"
	items, err := repo.ListWithCursor(context.Background(), domain.ListFilter{Search: &search}, nil, 20, nil)
	s.Require().NoError(err)
	s.Equal(2, len(items), "expected 2 items matching prefix 'buy'")
}

func (s *SearchSuite) TestCaseInsensitive() {
	s.insertItems("Buy Milk")
	repo := repository.NewMySQL(s.DB, s.FTSEnabled)

	search := "buy"
	items, err := repo.ListWithCursor(context.Background(), domain.ListFilter{Search: &search}, nil, 20, nil)
	s.Require().NoError(err)
	s.Equal(1, len(items))
}

func (s *SearchSuite) TestMultipleWords() {
	s.insertItems("Buy milk and eggs", "Buy bread", "Get milk")
	repo := repository.NewMySQL(s.DB, s.FTSEnabled)

	search := "buy milk"
	items, err := repo.ListWithCursor(context.Background(), domain.ListFilter{Search: &search}, nil, 20, nil)
	s.Require().NoError(err)
	s.Equal(1, len(items), "expected 1 item matching both 'buy' AND 'milk'")
}

func (s *SearchSuite) TestNoResults() {
	s.insertItems("Buy milk")
	repo := repository.NewMySQL(s.DB, s.FTSEnabled)

	search := "xyz"
	items, err := repo.ListWithCursor(context.Background(), domain.ListFilter{Search: &search}, nil, 20, nil)
	s.Require().NoError(err)
	s.Equal(0, len(items))
}

func (s *SearchSuite) TestNilSearchReturnsAll() {
	s.insertItems("Task 1", "Task 2", "Task 3")
	repo := repository.NewMySQL(s.DB, s.FTSEnabled)

	items, err := repo.ListWithCursor(context.Background(), domain.ListFilter{}, nil, 20, nil)
	s.Require().NoError(err)
	s.Equal(3, len(items))
}

func (s *SearchSuite) TestSearchWithCursorPagination() {
	for i := 1; i <= 5; i++ {
		s.insertItems(fmt.Sprintf("Task item %d", i))
	}
	s.insertItems("Walk dog")
	repo := repository.NewMySQL(s.DB, s.FTSEnabled)

	search := "task"
	sortFields := []domain.SortField{{Field: "position", Direction: domain.SortAsc}}

	// First page: limit 2
	items, err := repo.ListWithCursor(context.Background(), domain.ListFilter{Search: &search}, sortFields, 2, nil)
	s.Require().NoError(err)
	s.Equal(2, len(items))

	// Build cursor from last item and get next page
	cursorData, err := repo.BuildCursor(items[len(items)-1], sortFields)
	s.Require().NoError(err)

	items2, err := repo.ListWithCursor(context.Background(), domain.ListFilter{Search: &search}, sortFields, 2, cursorData)
	s.Require().NoError(err)
	s.Equal(2, len(items2))

	// Build cursor from last item and get final page
	cursorData2, err := repo.BuildCursor(items2[len(items2)-1], sortFields)
	s.Require().NoError(err)

	items3, err := repo.ListWithCursor(context.Background(), domain.ListFilter{Search: &search}, sortFields, 2, cursorData2)
	s.Require().NoError(err)
	s.Equal(1, len(items3))
}

func (s *SearchSuite) TestCountWithSearch() {
	s.insertItems("Buy milk", "Buy eggs", "Walk the dog")
	repo := repository.NewMySQL(s.DB, s.FTSEnabled)

	search := "buy"
	count, err := repo.Count(context.Background(), domain.ListFilter{Search: &search})
	s.Require().NoError(err)
	s.Equal(2, count)
}

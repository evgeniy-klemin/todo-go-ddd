package item

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/evgeniy-klemin/todo/db/schema"
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
	repo := repository.New(s.DB, schema.DriverMySQL, s.FTSEnabled)
	for _, name := range names {
		item, err := domain.NewItem(name, 1)
		s.Require().NoError(err)
		_, err = repo.AddWithNextPosition(ctx, item)
		s.Require().NoError(err)
	}
}

func (s *SearchSuite) TestExactWordMatch() {
	s.insertItems("Buy milk", "Buy eggs", "Walk the dog")
	repo := repository.New(s.DB, schema.DriverMySQL, s.FTSEnabled)

	search := "buy"
	items, err := repo.All(context.Background(), nil, &search, nil, 1, 20, nil)
	s.Require().NoError(err)
	s.Equal(2, len(items), "expected 2 items matching 'buy'")
}

func (s *SearchSuite) TestPrefixMatch() {
	s.insertItems("Buying groceries", "Buy milk", "Walk the dog")
	repo := repository.New(s.DB, schema.DriverMySQL, s.FTSEnabled)

	search := "buy"
	items, err := repo.All(context.Background(), nil, &search, nil, 1, 20, nil)
	s.Require().NoError(err)
	s.Equal(2, len(items), "expected 2 items matching prefix 'buy'")
}

func (s *SearchSuite) TestCaseInsensitive() {
	s.insertItems("Buy Milk")
	repo := repository.New(s.DB, schema.DriverMySQL, s.FTSEnabled)

	search := "buy"
	items, err := repo.All(context.Background(), nil, &search, nil, 1, 20, nil)
	s.Require().NoError(err)
	s.Equal(1, len(items))
}

func (s *SearchSuite) TestMultipleWords() {
	s.insertItems("Buy milk and eggs", "Buy bread", "Get milk")
	repo := repository.New(s.DB, schema.DriverMySQL, s.FTSEnabled)

	search := "buy milk"
	items, err := repo.All(context.Background(), nil, &search, nil, 1, 20, nil)
	s.Require().NoError(err)
	s.Equal(1, len(items), "expected 1 item matching both 'buy' AND 'milk'")
}

func (s *SearchSuite) TestNoResults() {
	s.insertItems("Buy milk")
	repo := repository.New(s.DB, schema.DriverMySQL, s.FTSEnabled)

	search := "xyz"
	items, err := repo.All(context.Background(), nil, &search, nil, 1, 20, nil)
	s.Require().NoError(err)
	s.Equal(0, len(items))
}

func (s *SearchSuite) TestNilSearchReturnsAll() {
	s.insertItems("Task 1", "Task 2", "Task 3")
	repo := repository.New(s.DB, schema.DriverMySQL, s.FTSEnabled)

	items, err := repo.All(context.Background(), nil, nil, nil, 1, 20, nil)
	s.Require().NoError(err)
	s.Equal(3, len(items))
}

func (s *SearchSuite) TestSearchWithPagination() {
	for i := 1; i <= 5; i++ {
		s.insertItems(fmt.Sprintf("Task item %d", i))
	}
	s.insertItems("Walk dog")
	repo := repository.New(s.DB, schema.DriverMySQL, s.FTSEnabled)

	search := "task"
	// Page 1
	items, err := repo.All(context.Background(), nil, &search, nil, 1, 2, nil)
	s.Require().NoError(err)
	s.Equal(2, len(items))

	// Page 3 — 1 remaining
	items, err = repo.All(context.Background(), nil, &search, nil, 3, 2, nil)
	s.Require().NoError(err)
	s.Equal(1, len(items))
}

func (s *SearchSuite) TestCountWithSearch() {
	s.insertItems("Buy milk", "Buy eggs", "Walk the dog")
	repo := repository.New(s.DB, schema.DriverMySQL, s.FTSEnabled)

	search := "buy"
	count, err := repo.Count(context.Background(), nil, &search)
	s.Require().NoError(err)
	s.Equal(2, count)
}

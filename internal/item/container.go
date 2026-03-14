package item

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"github.com/evgeniy-klemin/todo/db/schema"
	"github.com/evgeniy-klemin/todo/internal/item/app"
	"github.com/evgeniy-klemin/todo/internal/item/ports"
	"github.com/evgeniy-klemin/todo/internal/item/repository"
)

type Container struct {
	httpServer *ports.HttpServer
}

func (c *Container) RegisterHandlers(e *echo.Echo) {
	ports.RegisterHandlers(e, c.httpServer)
}

func NewContainer(db *sql.DB, driver string, ftsEnabled bool) *Container {
	var repo *repository.Repository
	switch driver {
	case schema.DriverMySQL:
		repo = repository.NewMySQL(db, ftsEnabled)
	default:
		repo = repository.NewSQLite(db, ftsEnabled)
	}
	service := app.NewItemService(repo, repo)
	httpServer := ports.NewHttpServer(service)
	return &Container{httpServer: httpServer}
}

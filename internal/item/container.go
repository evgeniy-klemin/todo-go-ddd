package item

import (
	"database/sql"

	"github.com/labstack/echo/v4"

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

func NewContainer(db *sql.DB) *Container {
	repository := repository.New(db)
	service := app.NewItemService(repository, repository)
	httpServer := ports.NewHttpServer(service)
	return &Container{
		httpServer: httpServer,
	}
}

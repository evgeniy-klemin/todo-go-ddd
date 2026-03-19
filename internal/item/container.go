package item

import (
	"database/sql"
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/evgeniy-klemin/todo/db/driver"
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

func NewContainer(db *sql.DB, drv string, ftsEnabled bool) *Container {
	var repo *repository.Repository
	switch drv {
	case driver.MySQL:
		repo = repository.NewMySQL(db, ftsEnabled)
	case driver.SQLite:
		repo = repository.NewSQLite(db, ftsEnabled)
	default:
		panic(fmt.Sprintf("unsupported database driver: %s", drv))
	}
	service := app.NewItemService(repo)
	httpServer := ports.NewHttpServer(service)
	return &Container{httpServer: httpServer}
}

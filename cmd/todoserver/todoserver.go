package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"

	"github.com/evgeniy-klemin/todo"
	"github.com/evgeniy-klemin/todo/db/driver"
	"github.com/evgeniy-klemin/todo/db/fts"
	"github.com/evgeniy-klemin/todo/db/migrations"
	item "github.com/evgeniy-klemin/todo/internal/item"
	itemports "github.com/evgeniy-klemin/todo/internal/item/ports"
)

func registerDoc(e *echo.Echo) {
	assetHandler := http.FileServer(todo.GetFileSystem())
	e.GET("/docs/*", echo.WrapHandler(http.StripPrefix("/docs/", assetHandler)))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Path() == "/" {
				_ = c.Redirect(http.StatusPermanentRedirect, "/docs/")
			}
			return next(c)
		}
	})
}

func setOapiValidator(e *echo.Echo) {
	itemSwagger, err := itemports.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}
	itemSwagger.Servers = nil
	e.Use(middleware.OapiRequestValidatorWithOptions(itemSwagger, &middleware.Options{
		Skipper: func(c echo.Context) bool {
			switch c.Path() {
			case "/docs/*", "/":
				return true
			}
			return false
		},
	}))
}

func server(port int) {
	e := echo.New()
	// Log all requests
	e.Use(echomiddleware.Logger())
	// Swagger validation
	setOapiValidator(e)

	drv := os.Getenv("DB_DRIVER")
	if drv == "" {
		drv = driver.SQLite
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		switch drv {
		case driver.MySQL:
			dsn = "todo:todo@tcp(localhost)/todotest?parseTime=true"
		default:
			dsn = "file:todotest.db?cache=shared"
		}
	}

	db, err := sql.Open(drv, dsn)
	if err != nil {
		panic(err)
	}
	if err := migrations.Run(db, drv); err != nil {
		panic(err)
	}
	var ftsEnabled bool
	if drv == driver.MySQL {
		ftsEnabled = true // FULLTEXT index created by goose migration 00003
	} else {
		ftsEnabled = fts.Apply(db)
	}
	if !ftsEnabled {
		switch drv {
		case driver.MySQL:
			fmt.Fprintf(os.Stderr, "Warning: MySQL FULLTEXT index not available, falling back to LIKE search\n")
		default:
			fmt.Fprintf(os.Stderr, "Warning: FTS5 not available, falling back to LIKE search\n")
		}
	}

	// Containers
	itemContainer := item.NewContainer(db, drv, ftsEnabled)

	// Register http handlers
	itemContainer.RegisterHandlers(e)

	registerDoc(e)

	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", port)))
}

func main() {
	var port = flag.Int("port", 3000, "Port for test HTTP server")
	flag.Parse()

	server(*port)
}

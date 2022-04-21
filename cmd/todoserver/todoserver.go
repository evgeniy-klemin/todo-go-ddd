package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"

	"github.com/evgeniy-klemin/todo"
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

	// db, err := sql.Open("sqlite3", "file:todotest.db?cache=shared")
	db, err := sql.Open("mysql", "todo:todo@tcp(localhost)/todotest?parseTime=true")
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// Containers
	itemContainer := item.NewContainer(db)

	// Register http handlers
	itemContainer.RegisterHandlers(e)

	registerDoc(e)

	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", port)))
}

func fixtures() {
	// db, err := sql.Open("sqlite3", "file:todotest.db?cache=shared")
	db, err := sql.Open("mysql", "todo:todo@tcp(localhost)/todotest")
	if err != nil {
		panic(err)
	}
	f, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("mysql"),
		testfixtures.Directory("db/fixtures"),
	)
	if err != nil {
		panic(err)
	}
	if err := f.Load(); err != nil {
		panic(err)
	}
}

func main() {
	var port = flag.Int("port", 8080, "Port for test HTTP server")
	flag.Parse()

	fixtures()
	server(*port)
}

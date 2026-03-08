# TODO DDD

> Example the todo API with DDD

## Prerequisites

- `go install github.com/pressly/goose/v3/cmd/goose@latest` - migrate DB
- `go get github.com/deepmap/oapi-codegen/pkg/codegen@v1.8.2` - dependency for sqlboiler
- `go install github.com/volatiletech/sqlboiler/v4@latest` - ORM
- `go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-mysql@latest` - ORM (MySQL)

## Setup

- `docker-compose up` - run mysql server
- `make generate` - regenerate openapi and orm
- `make migrate` - migrate DB

## Run

- `go run cmd/todoserver/todoserver.go` - run server, check url http://localhost:3000/ (use `-port` flag to override)
- `go run cmd/todoclient/todoclient.go` - run client - concurrent 10000 rest queries to server

## TODO

[x] implement retrieve items by page number
[x] sort by field position
[x] fix sort - replace to ordered map
[x] implement filter by done field
[x] move pagination to Header
[ ] tests for CRUD
[ ] improve search with full text search

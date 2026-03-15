# Todo Go DDD

A todo API built with Go using Domain-Driven Design (DDD) architecture. Supports SQLite and MySQL as database backends.

## Prerequisites

- Go 1.25+
- Docker & Docker Compose (for MySQL)
- `go get github.com/deepmap/oapi-codegen/pkg/codegen@v1.8.2` — OpenAPI codegen
- `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` — sqlc query codegen

## Quick Start

### SQLite (default, no setup needed)

```bash
make run
```

### MySQL

```bash
make run-mysql
```

The server starts at http://localhost:3000. API docs at http://localhost:3000/docs/.

## Configuration

| Variable    | Default   | Description                            |
|-------------|-----------|----------------------------------------|
| `DB_DRIVER` | `sqlite3` | Database driver (`sqlite3` or `mysql`) |
| `DB_DSN`    | auto      | Database connection string             |

Default DSN per driver:
- **sqlite3**: `file:todotest.db?cache=shared`
- **mysql**: `todo:todo@tcp(localhost)/todotest?parseTime=true`

## Make Targets

| Target                 | Description                                       |
|------------------------|---------------------------------------------------|
| `make run`             | Run server with SQLite                            |
| `make run-mysql`       | Start MySQL via docker-compose and run server     |
| `make test`            | Run unit tests (SQLite, no external dependencies) |
| `make test-integration`| Run integration tests against MySQL               |
| `make generate`        | Regenerate OpenAPI and sqlc query code            |

## Code Generation

This project uses two codegen tools:

- **OpenAPI**: `go generate ./...` regenerates HTTP handlers from `docs/todo.yaml`
- **sqlc**: `sqlc generate` regenerates type-safe query code from SQL files in `db/queries/`

Run `make generate` to invoke both.

After modifying any file under `db/queries/`, run `sqlc generate` to keep the generated packages (`internal/item/repository/sqlitedb/` and `internal/item/repository/mysqldb/`) in sync.

## Run

- `go run cmd/todoserver/todoserver.go` — run server (use `-port` flag to override port)
- `go run cmd/todoclient/todoclient.go` — run client (concurrent 10000 REST queries, uses port 8080)

## Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend dev server starts at http://localhost:5173 and proxies API requests to the backend.

## Project Structure

```
cmd/todoserver/       Server entrypoint
internal/item/
  domain/             Domain models and business logic
  app/                Application services
  ports/              HTTP handlers (OpenAPI)
  repository/         Database access (sqlc)
db/
  schema/             Schema definitions (single source of truth)
  fixtures/           Test data
test/
  cases/item/         MySQL integration tests
  setup/              Test suite setup (MySQL)
docs/                 OpenAPI/Swagger specs
frontend/             React frontend
```

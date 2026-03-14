.PHONY: run run-mysql test test-integration generate

DOCKER_COMPOSE := $(shell command -v docker-compose 2>/dev/null || echo "docker compose")

run:
	go run cmd/todoserver/todoserver.go

run-mysql:
	$(DOCKER_COMPOSE) up -d
	DB_DRIVER=mysql DB_DSN="todo:todo@tcp(localhost)/todotest?parseTime=true" go run cmd/todoserver/todoserver.go

test:
	go test ./db/... ./internal/... -count=1

test-integration:
	$(DOCKER_COMPOSE) up -d
	go test ./test/... -count=1 -v

generate:
	go generate ./...

.PHONY: run run-mysql test test-integration test-e2e generate lint lint-gopls

DOCKER_COMPOSE := $(shell command -v docker-compose 2>/dev/null || echo "docker compose")

LOCAL_BIN := $(CURDIR)/.bin
GOLANGCI_LINT_VERSION := v2.11.3
GOLANGCI_LINT := $(LOCAL_BIN)/golangci-lint

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

frontend/node_modules:
	cd frontend && npm install

frontend/node_modules/.playwright-browsers-installed: frontend/node_modules
	cd frontend && npx playwright install --with-deps
	@touch $@

test-e2e: frontend/node_modules/.playwright-browsers-installed
	cd frontend && npx playwright test

generate:
	go generate ./...
	sqlc generate

$(GOLANGCI_LINT):
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

lint-gopls:
	gopls check $$(find . -name "*.go" -not -path "./frontend/*" | tr '\n' ' ')

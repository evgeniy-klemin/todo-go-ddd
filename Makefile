.PHONY: run run-mysql test test-integration generate

run:
	go run cmd/todoserver/todoserver.go

run-mysql:
	docker-compose up -d
	DB_DRIVER=mysql DB_DSN="todo:todo@tcp(localhost)/todotest?parseTime=true" go run cmd/todoserver/todoserver.go

test:
	go test ./db/... ./internal/... -count=1

test-integration:
	docker-compose up -d
	go test ./test/... -count=1 -v

generate:
	cd internal/item && sqlboiler mysql
	go generate ./...

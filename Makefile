migrate:
	goose -dir db/migrations mysql "todo:todo@tcp(localhost)/todotest?parseTime=true" up

generate:
	cd internal/item && sqlboiler mysql
	go generate ./...

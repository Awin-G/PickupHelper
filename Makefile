.PHONY: run build test test-unit test-integration migrate-up migrate-down migrate-reset tidy vet fmt lint

APP_ENV ?= dev
BIN := bin/server
MYSQL_DSN := "root:1973@tcp(127.0.0.1:3306)/pickup_helper?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4"
MYSQL_DB := pickup_helper

run:
	APP_ENV=$(APP_ENV) go run cmd/server/main.go

build:
	go build -o $(BIN) cmd/server/main.go

test:
	go test ./... -race -count=1

test-unit:
	go test ./internal/... -race -count=1 -short

test-integration:
	go test -tags=integration ./test/... -race -count=1 -v -timeout 300s

migrate-up:
	@mysql -uroot -p1973 -e "CREATE DATABASE IF NOT EXISTS $(MYSQL_DB) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null || mariadb -uroot -p1973 -e "CREATE DATABASE IF NOT EXISTS $(MYSQL_DB) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
	goose -dir migrations mysql $(MYSQL_DSN) up

migrate-down:
	goose -dir migrations mysql $(MYSQL_DSN) down

migrate-reset:
	goose -dir migrations mysql $(MYSQL_DSN) reset

tidy:
	go mod tidy

vet:
	go vet ./...

vet-integration:
	go vet -tags=integration ./...

fmt:
	gofmt -s -w .

lint: vet
	@echo "lint placeholder (no linter configured)"

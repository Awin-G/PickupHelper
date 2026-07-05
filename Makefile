.PHONY: run build test test-unit test-integration migrate-up migrate-down tidy vet fmt lint

APP_ENV ?= dev
BIN := bin/server
MIGRATE_DSN := "root:1973@tcp(127.0.0.1:3306)/pickup_helper?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4"

run:
	APP_ENV=$(APP_ENV) go run cmd/server/main.go

build:
	go build -o $(BIN) cmd/server/main.go

test:
	go test ./... -race -count=1

test-unit:
	go test ./internal/... -race -count=1 -short

test-integration:
	go test ./test/... -race -count=1 -tags=integration

migrate-up:
	goose -dir migrations mysql $(MIGRATE_DSN) up

migrate-down:
	goose -dir migrations mysql $(MIGRATE_DSN) down

tidy:
	go mod tidy

vet:
	go vet ./...

fmt:
	gofmt -s -w .

lint: vet
	@echo "lint placeholder (no linter configured)"

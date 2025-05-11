.PHONY: dev test build migrate

dev:                     # hot reload
	air

test:                    # unit tests + race detector
	go test ./... -race -short

build:
	go build -o bin/adeptd ./cmd/adeptd

migrate:
	go run ./internal/db/migrate.go up

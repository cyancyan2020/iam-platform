BINARY := iam-platform
GO := go
MAIN := cmd/main.go

.PHONY: run build migrate-up migrate-down test lint

run:
	$(GO) run $(MAIN)

build:
	$(GO) build -o bin/$(BINARY) $(MAIN)

migrate-up:
	migrate -path db/migrations -database "mysql://root:1234@tcp(127.0.0.1:3306)/iam_platform?charset=utf8mb4&parseTime=True&loc=Local" up

migrate-down:
	migrate -path db/migrations -database "mysql://root:1234@tcp(127.0.0.1:3306)/iam_platform?charset=utf8mb4&parseTime=True&loc=Local" down

test:
	$(GO) test ./...

lint:
	golangci-lint run ./...

.PHONY: help build run test clean migrate-up migrate-down test-unit test-e2e test-all test-coverage

help:
	@echo "Available commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build         - Build all services"
	@echo "  make run-all       - Run all services"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make test          - Run all tests (unit + e2e)"
	@echo "  make test-unit     - Run unit tests only (fast, no Docker needed)"
	@echo "  make test-e2e      - Run E2E tests only (requires Docker)"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make test-watch    - Run unit tests in watch mode"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up    - Run database migrations"
	@echo "  make migrate-down  - Rollback database migrations"

build:
	go build -o bin/commands ./cmd/commands
	go build -o bin/projections ./cmd/projections
	go build -o bin/api ./cmd/api
	go build -o bin/slackbot ./cmd/slackbot
	go build -o bin/scheduler ./cmd/scheduler

run-commands:
	go run cmd/commands/main.go

run-projections:
	go run cmd/projections/main.go

run-api:
	go run cmd/api/main.go

run-slackbot:
	go run cmd/slackbot/main.go

run-scheduler:
	go run cmd/scheduler/main.go

# Run all tests
test: test-unit test-e2e

# Run unit tests only (fast, no Docker needed)
test-unit:
	@echo "Running unit tests..."
	@go test -v ./internal/... ./cmd/...

# Run E2E tests only (requires Docker)
test-e2e:
	@echo "Running E2E tests (requires Docker)..."
	@go test -v -timeout=5m ./tests/e2e/...

# Run all tests
test-all:
	@echo "Running all tests..."
	@go test -v -timeout=5m ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Watch mode for unit tests (requires entr: brew install entr)
test-watch:
	@echo "Watching for changes... (Ctrl+C to stop)"
	@find . -name '*.go' | entr -c go test -v ./internal/... ./cmd/...

migrate-up:
	migrate -path migrations -database ${EVENT_STORE_URL} up
	migrate -path migrations -database ${PROJECTION_DB_URL} up

migrate-down:
	migrate -path migrations -database ${PROJECTION_DB_URL} down
	migrate -path migrations -database ${EVENT_STORE_URL} down

clean:
	rm -rf bin/
	go clean

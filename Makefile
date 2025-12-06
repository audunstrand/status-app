.PHONY: help build run test clean migrate-up migrate-down

help:
	@echo "Available commands:"
	@echo "  make build        - Build all services"
	@echo "  make run-all      - Run all services"
	@echo "  make test         - Run tests"
	@echo "  make migrate-up   - Run database migrations"
	@echo "  make migrate-down - Rollback database migrations"
	@echo "  make clean        - Clean build artifacts"

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

test:
	go test -v ./...

migrate-up:
	migrate -path migrations -database ${EVENT_STORE_URL} up
	migrate -path migrations -database ${PROJECTION_DB_URL} up

migrate-down:
	migrate -path migrations -database ${PROJECTION_DB_URL} down
	migrate -path migrations -database ${EVENT_STORE_URL} down

clean:
	rm -rf bin/
	go clean

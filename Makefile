.PHONY: help build test clean

help:
	@echo "Commands:"
	@echo "  make build    - Build all services to bin/"
	@echo "  make test     - Run all tests"
	@echo "  make clean    - Remove build artifacts"

build:
	@mkdir -p bin
	@go build -o bin/commands ./cmd/commands
	@go build -o bin/api ./cmd/api
	@go build -o bin/projections ./cmd/projections
	@go build -o bin/slackbot ./cmd/slackbot
	@go build -o bin/scheduler ./cmd/scheduler
	@echo "Built all services to bin/"

test:
	@go test ./...

clean:
	@rm -rf bin/
	@go clean

.PHONY: help build test test-unit test-e2e clean

help:
	@echo "Commands:"
	@echo "  make build      - Build all services to bin/"
	@echo "  make test       - Run all tests"
	@echo "  make test-unit  - Run unit tests"
	@echo "  make test-e2e   - Run E2E tests"
	@echo "  make clean      - Remove build artifacts"

build:
	@mkdir -p bin
	@go build -o bin/backend ./cmd/backend
	@go build -o bin/slackbot ./cmd/slackbot
	@go build -o bin/scheduler ./cmd/scheduler
	@echo "Built all services to bin/"

test:
	@go test ./...

test-unit:
	@go test ./...

test-e2e:
	@go test ./...

clean:
	@rm -rf bin/
	@go clean

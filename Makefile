.PHONY: help build test test-unit test-e2e clean deploy-dashboard

help:
	@echo "Commands:"
	@echo "  make build             - Build all services to bin/"
	@echo "  make test              - Run all tests"
	@echo "  make test-unit         - Run unit tests"
	@echo "  make test-e2e          - Run E2E tests"
	@echo "  make deploy-dashboard  - Deploy Grafana dashboard (requires GRAFANA_URL and GRAFANA_API_KEY)"
	@echo "  make clean             - Remove build artifacts"

build:
	@mkdir -p bin
	@go build -o bin/backend ./cmd/backend
	@go build -o bin/slackbot ./cmd/slackbot
	@go build -o bin/scheduler ./cmd/scheduler
	@go build -o bin/migrate ./cmd/migrate
	@echo "Built all services to bin/"

test:
	@go test ./...

test-unit:
	@go test ./...

test-e2e:
	@go test ./...

deploy-dashboard:
	@if [ -z "$$GRAFANA_URL" ]; then \
		echo "Error: GRAFANA_URL not set"; \
		echo "Usage: GRAFANA_URL=https://your-grafana GRAFANA_API_KEY=key make deploy-dashboard"; \
		exit 1; \
	fi
	@if [ -z "$$GRAFANA_API_KEY" ]; then \
		echo "Error: GRAFANA_API_KEY not set"; \
		echo "Usage: GRAFANA_URL=https://your-grafana GRAFANA_API_KEY=key make deploy-dashboard"; \
		exit 1; \
	fi
	@echo "Deploying Grafana dashboard..."
	@python3 scripts/deploy-grafana-dashboard.py

clean:
	@rm -rf bin/
	@go clean

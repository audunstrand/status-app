.PHONY: test clean up down logs

# Run E2E tests with Docker
test:
	@echo "Running Docker-based E2E tests..."
	go test -v -timeout 5m ./...

# Start services for manual testing
up:
	docker-compose -f docker-compose.test.yml up -d
	@echo "Services started. Waiting for health checks..."
	@sleep 10
	@echo "Commands service: http://localhost:8081"
	@echo "API service: http://localhost:8082"

# Stop services
down:
	docker-compose -f docker-compose.test.yml down -v

# View logs
logs:
	docker-compose -f docker-compose.test.yml logs -f

# Clean up everything
clean:
	docker-compose -f docker-compose.test.yml down -v --remove-orphans
	docker system prune -f

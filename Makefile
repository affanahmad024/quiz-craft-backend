.PHONY: help run build test docker-up docker-down migrate-up migrate-down clean

# Default target
help:
	@echo "Available targets:"
	@echo "  run          - Run the application"
	@echo "  build        - Build the application"
	@echo "  test         - Run tests"
	@echo "  docker-up    - Start PostgreSQL and Redis with docker-compose"
	@echo "  docker-down  - Stop docker containers"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback last migration"
	@echo "  clean        - Clean build artifacts"

# Run the application
run:
	go run cmd/server/main.go

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Start docker services (PostgreSQL + Redis)
docker-up:
	docker-compose -f deployments/docker/docker-compose.yml up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@timeout /t 5 /nobreak > nul
	@echo "PostgreSQL and Redis are ready!"

# Stop docker services
docker-down:
	docker-compose -f deployments/docker/docker-compose.yml down

# Stop and remove volumes
docker-clean:
	docker-compose -f deployments/docker/docker-compose.yml down -v

# View docker logs
docker-logs:
	docker-compose -f deployments/docker/docker-compose.yml logs -f

# Run migrations (after building)
migrate-up: build
	./bin/server migrate up

# Rollback migration (after building)
migrate-down: build
	./bin/server migrate down

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf uploads/
	rm -rf tmp/

# Development setup (install dependencies and start docker)
setup: deps docker-up
	@echo "Setup complete! Run 'make run' to start the server"

# Full restart (stop docker, clean, start docker, run)
restart: docker-down docker-up run

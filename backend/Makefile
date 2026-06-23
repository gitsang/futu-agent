.PHONY: build run clean test fmt lint docker-build docker-up docker-down

# Variables
APP_NAME=futu-agent-backend
BUILD_DIR=./dist/builds
GO=go
GOFMT=gofmt
GOIMPORTS=goimports

# Build
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

# Run locally
run: build
	@echo "Running $(APP_NAME)..."
	$(BUILD_DIR)/$(APP_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -race ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w .
	$(GOIMPORTS) -w .

# Lint code
lint:
	@echo "Linting code..."
	$(GO) vet ./...

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME) -f Dockerfile .

docker-up:
	@echo "Starting services..."
	docker-compose up -d

docker-down:
	@echo "Stopping services..."
	docker-compose down

docker-logs:
	docker-compose logs -f

# Database commands
db-migrate:
	@echo "Running database migrations..."
	$(GO) run ./cmd/migrate

# Development
dev:
	@echo "Starting development server..."
	$(GO) run ./cmd/server

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Build and run the application"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-up    - Start Docker services"
	@echo "  docker-down  - Stop Docker services"
	@echo "  dev          - Start development server"
	@echo "  help         - Show this help"

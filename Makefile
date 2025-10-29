.PHONY: help build run test clean sqlc migrate docker-up docker-down docker-logs docker-build init-dirs

# Detect OS and set shell accordingly
ifeq ($(OS),Windows_NT)
SHELL := cmd.exe
.SHELLFLAGS := /c
endif

# Initialize directories
init-dirs:
ifeq ($(OS),Windows_NT)
	@if not exist bin mkdir bin
else
	@mkdir -p bin
endif

# Default target
help:
	@echo "Available commands:"
	@echo "  make build          - Build all service binaries"
	@echo "  make build-gateway  - Build gateway service"
	@echo "  make build-svedprint - Build svedprint service"
	@echo "  make build-admin    - Build admin service"
	@echo "  make build-print    - Build print service"
	@echo "  make run-gateway    - Run gateway service"
	@echo "  make run-svedprint  - Run svedprint service"
	@echo "  make run-admin      - Run admin service"
	@echo "  make run-print      - Run print service"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make sqlc           - Generate sqlc code for all services"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Docker commands:"
	@echo "  make docker-up      - Start all services with docker-compose"
	@echo "  make docker-down    - Stop all services"
	@echo "  make docker-build   - Rebuild all services"
	@echo "  make docker-logs    - View logs from all services"
	@echo "  make docker-clean   - Stop and remove all containers, volumes"
	@echo ""
	@echo "Development:"
	@echo "  make dev-setup      - Initial setup (copy .env, install tools)"
	@echo "  make tidy           - Run go mod tidy"

# Build commands
build: init-dirs build-gateway build-svedprint build-admin build-print

build-gateway: init-dirs
	@echo "Building gateway service..."
	@go build -o bin/gateway ./cmd/gateway

build-svedprint: init-dirs
	@echo "Building svedprint service..."
	@go build -o bin/svedprint ./cmd/svedprint

build-admin: init-dirs
	@echo "Building admin service..."
	@go build -o bin/svedprint-admin ./cmd/svedprint-admin

build-print: init-dirs
	@echo "Building print service..."
	@go build -o bin/svedprint-print ./cmd/svedprint-print

# Run commands (requires environment variables)
run-gateway:
	@go run ./cmd/gateway/main.go

run-svedprint:
	@go run ./cmd/svedprint/main.go

run-admin:
	@go run ./cmd/svedprint-admin/main.go

run-print:
	@go run ./cmd/svedprint-print/main.go

# Test commands
test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# sqlc generation
sqlc:
	@echo "Generating sqlc code for all services..."
	@sqlc generate -f svedprint-sqlc.yaml
	@sqlc generate -f svedprint-admin-sqlc.yaml
	@sqlc generate -f gateway-sqlc.yaml
	@echo "sqlc generation complete"

# Clean
clean:
	@echo "Cleaning build artifacts..."
ifeq ($(OS),Windows_NT)
	@if exist bin rmdir /s /q bin
	@if exist coverage.out del /q coverage.out
	@if exist coverage.html del /q coverage.html
else
	@rm -rf bin/
	@rm -f coverage.out coverage.html
endif
	@echo "Clean complete"

# Docker commands
docker-up:
	@echo "Starting all services with docker-compose..."
	@docker-compose up -d
	@echo "Services started. Run 'make docker-logs' to view logs"

docker-down:
	@echo "Stopping all services..."
	@docker-compose down

docker-build:
	@echo "Rebuilding all services..."
	@docker-compose up -d --build

docker-logs:
	@docker-compose logs -f

docker-clean:
	@echo "Stopping and removing all containers and volumes..."
	@docker-compose down -v
	@echo "Docker clean complete"

docker-ps:
	@docker-compose ps

# Development setup
dev-setup:
	@echo "Setting up development environment..."
ifeq ($(OS),Windows_NT)
	@if not exist .env (copy .env.example .env && echo Created .env file from .env.example && echo Please edit .env and set required variables) else (echo .env file already exists)
	@echo Checking for required tools...
	@where sqlc >nul 2>&1 || echo Warning: sqlc not found. Install with: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@where migrate >nul 2>&1 || echo Warning: migrate not found. Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
else
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
		echo "Please edit .env and set required variables"; \
	else \
		echo ".env file already exists"; \
	fi
	@echo "Checking for required tools..."
	@which sqlc > /dev/null 2>&1 || echo "Warning: sqlc not found. Install with: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"
	@which migrate > /dev/null 2>&1 || echo "Warning: migrate not found. Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
endif
	@echo "Setup complete"

tidy:
	@echo "Running go mod tidy..."
	@go mod tidy
	@echo "Dependencies updated"

# Database migrations (example, adjust path as needed)
migrate-up:
	@echo "Running migrations..."
	@migrate -path db/svedprint/migrations -database "${DATABASE_URL}" up

migrate-down:
	@echo "Rolling back last migration..."
	@migrate -path db/svedprint/migrations -database "${DATABASE_URL}" down 1

# Linting (if you install golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

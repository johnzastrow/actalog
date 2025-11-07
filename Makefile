.PHONY: help build run test clean lint fmt docker-build docker-up docker-down migrate-up migrate-down

# Variables
APP_NAME=actalog
BINARY=bin/$(APP_NAME)
MAIN_PATH=cmd/$(APP_NAME)/main.go
DOCKER_COMPOSE=docker-compose

# Go build cache directories (Windows-friendly, keeps everything in project)
PROJECT_DIR=$(shell pwd)
CACHE_DIR=$(PROJECT_DIR)/.cache
GO_BUILD_CACHE=$(CACHE_DIR)/go-build
GO_MOD_CACHE=$(CACHE_DIR)/go-mod

# Export Go environment variables to use project directory
export GOCACHE=$(GO_BUILD_CACHE)
export GOMODCACHE=$(GO_MOD_CACHE)
export GOTMPDIR=$(CACHE_DIR)/tmp

# Default target
help: ## Show this help message
	@echo "ActaLog - Makefile Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application binary
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin $(GO_BUILD_CACHE) $(GO_MOD_CACHE) $(CACHE_DIR)/tmp
	@go build -o $(BINARY) $(MAIN_PATH)
	@echo "Build complete: $(BINARY)"

run: ## Run the application
	@echo "Running $(APP_NAME)..."
	@mkdir -p $(GO_BUILD_CACHE) $(GO_MOD_CACHE) $(CACHE_DIR)/tmp
	@go run $(MAIN_PATH)

dev: ## Run in development mode with auto-reload (requires air)
	@mkdir -p $(GO_BUILD_CACHE) $(GO_MOD_CACHE) $(CACHE_DIR)/tmp
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not found. Install with: go install github.com/air-verse/air@latest"; \
		echo "Falling back to 'go run'..."; \
		go run $(MAIN_PATH); \
	fi

test: ## Run all tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v -race ./test/unit/...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@go test -v -race ./test/integration/...

coverage: ## Show test coverage
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

lint: ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install from: https://golangci-lint.run/welcome/install/"; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w . 2>/dev/null || echo "goimports not found. Install with: go install golang.org/x/tools/cmd/goimports@latest"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf .cache/
	@rm -f coverage.out coverage.html
	@rm -f *.db *.sqlite *.sqlite3
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):latest .

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	@$(DOCKER_COMPOSE) up -d

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@$(DOCKER_COMPOSE) down

docker-logs: ## View Docker container logs
	@$(DOCKER_COMPOSE) logs -f

migrate-create: ## Create a new migration (usage: make migrate-create name=create_users_table)
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter is required. Usage: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	@mkdir -p migrations
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	touch migrations/$${timestamp}_$(name).up.sql; \
	touch migrations/$${timestamp}_$(name).down.sql; \
	echo "Created migration files:"; \
	echo "  migrations/$${timestamp}_$(name).up.sql"; \
	echo "  migrations/$${timestamp}_$(name).down.sql"

version: ## Show application version
	@go run $(MAIN_PATH) -version 2>/dev/null || echo "Build the app first with 'make build'"

.DEFAULT_GOAL := help

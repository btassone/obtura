# Obtura Makefile

# === Configuration ===
BINARY_NAME = obtura
BINARY_PATH = ./tmp/main
DOCKER_COMPOSE = docker-compose

# Go configuration
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOMOD = $(GOCMD) mod
GOFMT = $(GOCMD) fmt
GOVET = $(GOCMD) vet
LDFLAGS = -ldflags="-s -w"

.DEFAULT_GOAL := help

# === Core Commands ===

help: ## Show this help message
	@echo "Obtura Framework"
	@echo ""
	@echo "Usage: make [command]"
	@echo ""
	@echo "Commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Initial project setup
	@echo "Setting up Obtura..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/a-h/templ/cmd/templ@latest
	@mkdir -p tmp web/static/css web/static/js database/migrations database/seeders
	@$(GOMOD) download
	@echo "Setup complete! Run 'make dev' to start developing."

dev: ## Start development server with hot reload
	air

build: ## Build the application
	@templ generate
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) ./cmd/obtura

run: build ## Build and run the application
	$(BINARY_PATH) serve

test: ## Run tests
	$(GOTEST) -v -race -cover ./...

test-short: ## Run short tests only
	$(GOTEST) -v -short ./...

test-verbose: ## Run tests with verbose output
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-unit: ## Run unit tests only
	$(GOTEST) -v -race -short -tags=unit ./...

test-integration: ## Run integration tests
	$(GOTEST) -v -race -tags=integration ./test/integration/...

test-bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

test-specific: ## Run specific test (use: make test-specific name=TestFunctionName)
	@test -n "$(name)" || (echo "Error: name is required" && exit 1)
	$(GOTEST) -v -race -run $(name) ./...

clean: ## Clean build artifacts
	@rm -rf tmp/
	@rm -f $(BINARY_NAME) $(BINARY_NAME)-*
	@rm -rf coverage/
	@rm -f coverage.out coverage.html report.xml
	@find . -name "*_templ.go" -type f -delete
	@echo "Clean complete"

# === Code Quality ===

lint: ## Run linters (fmt, vet, golangci-lint)
	$(GOFMT) ./...
	$(GOVET) ./...
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

test-coverage: ## Generate test coverage report
	@./scripts/test-coverage.sh

test-ci: ## Run tests in CI mode
	@mkdir -p coverage
	$(GOTEST) -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -func=coverage/coverage.out

test-watch: ## Run tests in watch mode
	@if command -v gotestsum >/dev/null; then \
		gotestsum --watch -- -race ./...; \
	else \
		echo "gotestsum not installed. Install with: go install gotest.tools/gotestsum@latest"; \
	fi

# === Database ===

db-migrate: build ## Run pending migrations
	$(BINARY_PATH) migrate

db-rollback: build ## Rollback last migration
	$(BINARY_PATH) rollback

db-seed: build ## Run database seeders
	$(BINARY_PATH) seed

db-setup: db-migrate db-seed ## Setup database (migrate + seed)

# === Code Generation ===

new-migration: ## Create new migration (use: make new-migration name=create_users_table)
	@test -n "$(name)" || (echo "Error: name is required" && exit 1)
	$(BINARY_PATH) make:migration $(name)

new-model: ## Create new model (use: make new-model name=User)
	@test -n "$(name)" || (echo "Error: name is required" && exit 1)
	$(BINARY_PATH) make:model $(name)

new-controller: ## Create new controller (use: make new-controller name=UserController)
	@test -n "$(name)" || (echo "Error: name is required" && exit 1)
	$(BINARY_PATH) make:controller $(name)

new-plugin: ## Create new plugin (use: make new-plugin name=MyPlugin)
	@test -n "$(name)" || (echo "Error: name is required" && exit 1)
	$(BINARY_PATH) make:plugin $(name)

# === Production ===

build-prod: ## Build optimized binary for production
	@templ generate
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/obtura

build-all: ## Build for all platforms (Linux, Windows, macOS)
	@echo "Building for all platforms..."
	@templ generate
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/obtura
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/obtura
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/obtura
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/obtura
	@echo "Build complete! Binaries in dist/"

# === Docker ===

docker-up: ## Start Docker containers
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop Docker containers
	$(DOCKER_COMPOSE) down

docker-logs: ## View Docker container logs
	$(DOCKER_COMPOSE) logs -f

# === Utilities ===

deps: ## Update Go dependencies
	$(GOMOD) tidy
	$(GOMOD) download

tools: ## Check required development tools
	@echo "Development Tools Status:"
	@echo "========================"
	@command -v go >/dev/null && echo "✓ Go" || echo "✗ Go (required)"
	@command -v templ >/dev/null && echo "✓ Templ" || echo "✗ Templ (required - install with: go install github.com/a-h/templ/cmd/templ@latest)"
	@command -v air >/dev/null && echo "✓ Air" || echo "✗ Air (required - install with: go install github.com/cosmtrek/air@latest)"
	@command -v golangci-lint >/dev/null && echo "✓ golangci-lint" || echo "✗ golangci-lint (optional)"
	@command -v docker >/dev/null && echo "✓ Docker" || echo "✗ Docker (optional)"

.PHONY: help setup dev build run test test-short test-verbose test-unit \
        test-integration test-bench test-specific clean lint test-coverage \
        test-ci test-watch db-migrate db-rollback db-seed db-setup \
        new-migration new-model new-controller new-plugin \
        build-prod build-all docker-up docker-down docker-logs \
        deps tools
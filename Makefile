# Variables
APP_NAME := mini-ewallet
DOCKER_COMPOSE := docker-compose
GO := go
GOFLAGS := -v
MIGRATE := migrate
DB_URL := postgres://postgres:postgres@localhost:5432/mini_ewallet?sslmode=disable

# Colors
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

.PHONY: install-deps
install-deps: ## Install Go dependencies
	@echo "$(YELLOW)Installing dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod tidy

.PHONY: build
build: ## Build the application
	@echo "$(YELLOW)Building application...$(NC)"
	$(GO) build $(GOFLAGS) -o bin/$(APP_NAME) cmd/api/main.go
	@echo "$(GREEN)Build complete!$(NC)"

.PHONY: run
run: ## Run the application locally
	@echo "$(YELLOW)Running application...$(NC)"
	$(GO) run cmd/api/main.go

.PHONY: test
test: ## Run tests
	@echo "$(YELLOW)Running tests...$(NC)"
	$(GO) test -v -cover ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: lint
lint: ## Run linter
	@echo "$(YELLOW)Running linter...$(NC)"
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	@echo "$(YELLOW)Formatting code...$(NC)"
	$(GO) fmt ./...

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	docker build -t $(APP_NAME):latest .

.PHONY: docker-start
docker-start: ## Start all services with Docker Compose
	@echo "$(YELLOW)Starting services...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Services started!$(NC)"

.PHONY: docker-stop
docker-stop: ## Stop all services
	@echo "$(YELLOW)Stopping services...$(NC)"
	$(DOCKER_COMPOSE) stop

.PHONY: docker-down
docker-down: ## Stop and remove all services
	@echo "$(YELLOW)Removing services...$(NC)"
	$(DOCKER_COMPOSE) down -v

.PHONY: docker-logs
docker-logs: ## Show logs from all services
	$(DOCKER_COMPOSE) logs -f

.PHONY: migrate-up
migrate-up: ## Apply database migrations
	@echo "$(YELLOW)Applying migrations...$(NC)"
	$(MIGRATE) -path migrations -database "$(DB_URL)" up
	@echo "$(GREEN)Migrations applied!$(NC)"

.PHONY: migrate-down
migrate-down: ## Rollback database migrations
	@echo "$(YELLOW)Rolling back migrations...$(NC)"
	$(MIGRATE) -path migrations -database "$(DB_URL)" down 1

.PHONY: migrate-create
migrate-create: ## Create a new migration file (usage: make migrate-create name=migration_name)
	@echo "$(YELLOW)Creating migration: $(name)$(NC)"
	$(MIGRATE) create -ext sql -dir migrations -seq $(name)
	@echo "$(GREEN)Migration created!$(NC)"

.PHONY: migrate-force
migrate-force: ## Force migration version (usage: make migrate-force version=1)
	@echo "$(YELLOW)Forcing migration version to: $(version)$(NC)"
	$(MIGRATE) -path migrations -database "$(DB_URL)" force $(version)

.PHONY: docker-migrate
docker-migrate: ## Run migrations in Docker
	@echo "$(YELLOW)Running migrations in Docker...$(NC)"
	docker run --rm -v $(PWD)/migrations:/migrations --network mini-ewallet_default migrate/migrate \
		-path=/migrations -database "postgres://postgres:postgres@postgres:5432/mini_ewallet?sslmode=disable" up

.PHONY: seed
seed: ## Seed the database with sample data
	@echo "$(YELLOW)Seeding database...$(NC)"
	$(GO) run cmd/seed/main.go

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning...$(NC)"
	rm -rf bin/ coverage.out coverage.html

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(YELLOW)Installing development tools...$(NC)"
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)Tools installed!$(NC)"

.PHONY: dev
dev: docker-start migrate-up run ## Start development environment

.PHONY: prod-build
prod-build: ## Build for production
	@echo "$(YELLOW)Building for production...$(NC)"
	CGO_ENABLED=0 GOOS=linux $(GO) build -a -installsuffix cgo -o bin/$(APP_NAME) cmd/api/main.go
	@echo "$(GREEN)Production build complete!$(NC)"
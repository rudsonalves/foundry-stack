SHELL := /bin/bash
.DEFAULT_GOAL := help

# Default configuration. All variables can be overridden via the command line.
# Example: make build APP_NAME=foundry-stack-api
APP_NAME ?= foundry-stack-api
CMD_PATH ?= ./cmd
BIN_DIR ?= ./bin
BIN_PATH ?= $(BIN_DIR)/$(APP_NAME)
MIGRATIONS_DIR ?= ./migrations
ENV ?= dev
ENV_FILE := .env.$(ENV)
TEST_ENV_FILE ?= .env.test
COMPOSE := docker compose --env-file $(ENV_FILE)
POSTGRES_SERVICE := postgres

MODULE ?=
NAME ?=
MIGRATION_NAME ?=

.PHONY: help setup env-check db-up db-down db-status db-logs db-shell db-reset \
	db-setup mod-init tidy fmt vet lint test test-integration test-race build run \
	clean check migrate migrate-up migrate-down migrate-version migrate-create \
	migration-create swagger-up swagger-down swagger-logs dev stag prod \
	colima-up

help: ## Show available targets and common variables
	@awk 'BEGIN {FS = ":.*##"; printf "\nTargets:\n"} /^[a-zA-Z0-9_.-]+:.*##/ { printf "  %-18s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf "\nCommon variables:\n"
	@printf "  %-18s %s\n" "APP_NAME" "Name of the generated binary (default: $(APP_NAME))"
	@printf "  %-18s %s\n" "CMD_PATH" "Main package used by build/run (default: $(CMD_PATH))"
	@printf "  %-18s %s\n" "BIN_DIR" "Output directory for the binary (default: $(BIN_DIR))"
	@printf "  %-18s %s\n" "ENV" "Environment loaded from .env.ENV (default: $(ENV))"
	@printf "  %-18s %s\n" "MIGRATIONS_DIR" "Migrations directory (default: $(MIGRATIONS_DIR))"

# Bootstrap and dependencies
setup: tidy ## Prepare project dependencies

mod-init: ## Initialize go.mod (usage: make mod-init MODULE=github.com/you/repo)
	@if [ -z "$(MODULE)" ]; then \
		echo "Error: MODULE is required. Example: make mod-init MODULE=github.com/you/repo"; \
		exit 1; \
	fi
	@if [ -f go.mod ]; then \
		echo "go.mod already exists, skipping."; \
	else \
		go mod init $(MODULE); \
	fi

tidy: ## Run go mod tidy
	@if [ ! -f go.mod ]; then \
		echo "go.mod not found. Run: make mod-init MODULE=github.com/you/repo"; \
		exit 1; \
	fi
	@go mod tidy

# Quality
fmt: ## Format Go code
	@go fmt ./...

vet: ## Run go vet
	@go vet ./...

lint: ## Run golangci-lint, if installed
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "Error: golangci-lint not installed. Install: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

test: ## Run tests
	@go test ./...

test-race: ## Run tests with the race detector
	@go test -race ./...

test-integration: env-check ## Run integration tests against PostgreSQL
	@set -a; . "$(ENV_FILE)"; set +a; \
	go test ./internal/users/infrastructure -count=1 -v

check: fmt vet test ## Run formatting, vet, and tests

# Build and execution
build: ## Build the application binary
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_PATH) $(CMD_PATH)
	@echo "Built: $(BIN_PATH)"

run: ## Run the application locally
	@go run $(CMD_PATH)

clean: ## Remove build artifacts
	@rm -rf $(BIN_DIR)

dev: ## Run with .env.dev and the -debug flag
	go run $(CMD_PATH) -dev -debug

stag: ## Run with .env.stag
	go run $(CMD_PATH) -stag

prod: ## Run with .env.prod
	go run $(CMD_PATH) -prod

# Database and migrations
env-check:
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "Error: environment file not found: $(ENV_FILE)"; \
		exit 1; \
	fi

colima-up: ## Start Colima (macOS) for Docker
	@if ! command -v colima >/dev/null 2>&1; then \
		echo "Error: colima not installed. Install: https://github.com/abiosoft/colima"; \
		exit 1; \
	fi
	@colima start

db-up: env-check ## Start PostgreSQL (usage: make db-up ENV=dev)
	@$(COMPOSE) up -d --wait $(POSTGRES_SERVICE)

db-down: env-check ## Stop and remove the selected environment containers
	@$(COMPOSE) down

db-status: env-check ## Show PostgreSQL status for the selected environment
	@$(COMPOSE) ps $(POSTGRES_SERVICE)

db-logs: env-check ## Follow PostgreSQL logs for the selected environment
	@$(COMPOSE) logs -f $(POSTGRES_SERVICE)

db-shell: env-check ## Open psql in PostgreSQL for the selected environment
	@$(COMPOSE) exec $(POSTGRES_SERVICE) \
		sh -c 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"'

db-reset: env-check ## Recreate PostgreSQL and its volume (not allowed in production)
	@if [ "$(ENV)" = "prod" ]; then \
		echo "Error: db-reset is not allowed in production."; \
		exit 1; \
	fi
	@$(COMPOSE) down -v
	@$(COMPOSE) up -d --wait $(POSTGRES_SERVICE)

db-setup: db-up migrate-up ## Start PostgreSQL and apply all pending migrations

migrate: migrate-up ## Apply all pending migrations (usage: make migrate ENV=dev)

migrate-up: env-check ## Apply all pending migrations for ENV
	@if [ ! -d "$(MIGRATIONS_DIR)" ]; then \
		echo "Error: migrations directory not found: $(MIGRATIONS_DIR)"; \
		exit 1; \
	fi
	@if [ -z "$$(find "$(MIGRATIONS_DIR)" -type f -name '*.sql' -print -quit)" ]; then \
		echo "Error: no migration files found in $(MIGRATIONS_DIR)."; \
		exit 1; \
	fi
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "Error: migrate CLI not found. See: https://github.com/golang-migrate/migrate"; \
		exit 1; \
	fi
	@echo "Applying database schema changes via migrations only..."
	@set -a; . "$(ENV_FILE)"; set +a; \
		migrate -path $(MIGRATIONS_DIR) -database "$$DATABASE_URL" up

migrate-down: env-check ## Revert one migration step for ENV
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "Error: migrate CLI not found. See: https://github.com/golang-migrate/migrate"; \
		exit 1; \
	fi
	@set -a; . "$(ENV_FILE)"; set +a; \
		migrate -path $(MIGRATIONS_DIR) -database "$$DATABASE_URL" down 1

migrate-version: env-check ## Show the current migration version for ENV
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "Error: migrate CLI not found. See: https://github.com/golang-migrate/migrate"; \
		exit 1; \
	fi
	@set -a; . "$(ENV_FILE)"; set +a; \
		migrate -path $(MIGRATIONS_DIR) -database "$$DATABASE_URL" version

migrate-create: ## Create migration files (usage: make migrate-create NAME=create_users_table)
	@name="$(MIGRATION_NAME)"; \
	if [ -z "$$name" ]; then \
		name="$(NAME)"; \
	fi; \
	if [ -z "$$name" ]; then \
		echo "Error: MIGRATION_NAME is required. Example: make migrate-create MIGRATION_NAME=create_users_table"; \
		echo "Tip: NAME also works for backward compatibility."; \
		exit 1; \
	fi
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "Error: migrate CLI not found. See: https://github.com/golang-migrate/migrate"; \
		exit 1; \
	fi
	@name="$(MIGRATION_NAME)"; \
	if [ -z "$$name" ]; then \
		name="$(NAME)"; \
	fi; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq "$$name"

migration-create: migrate-create ## Alias for migrate-create

# API documentation
swagger-up: env-check ## Start Swagger UI at http://localhost:8081
	@$(COMPOSE) up -d swagger-ui
	@echo "Swagger UI: http://localhost:8081"

swagger-down: env-check ## Stop and remove the Swagger UI container
	@$(COMPOSE) rm -f -s swagger-ui

swagger-logs: env-check ## Follow Swagger UI logs
	@$(COMPOSE) logs -f swagger-ui

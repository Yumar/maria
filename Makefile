BACKEND := backend

.DEFAULT_GOAL := help

.PHONY: help build run fmt lint test test-integration generate check-generated db-up db-down

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the server binary into backend/bin/server
	cd $(BACKEND) && go build -o bin/server ./cmd/server

run: ## Run the server locally (reads env vars; see .env.example)
	cd $(BACKEND) && go run ./cmd/server

fmt: ## Format Go code
	cd $(BACKEND) && golangci-lint fmt

lint: ## Run golangci-lint
	cd $(BACKEND) && golangci-lint run ./...

test: ## Run unit tests (no Docker required)
	cd $(BACKEND) && go test -race ./...

test-integration: ## Run integration tests (requires Docker)
	@echo "test-integration: not configured yet (T-002)"

generate: ## Run code generators
	@echo "generate: not configured yet (T-002)"

check-generated: ## Fail if generated code is stale
	@echo "check-generated: not configured yet (T-002)"

db-up: ## Start local Postgres and wait until healthy
	docker compose up -d --wait postgres

db-down: ## Stop local Postgres (data volume is kept)
	docker compose down

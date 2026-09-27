BACKEND := backend

.DEFAULT_GOAL := help

.PHONY: help build run fmt lint test test-integration generate check-generated db-up db-down migrate-up migrate-down migrate-status migrate-new

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the server binary into backend/bin/server
	cd $(BACKEND) && go build -o bin/server ./cmd/server

run: ## Run the server locally (reads env vars; see .env.example)
	cd $(BACKEND) && go run ./cmd/server

fmt: ## Format Go code
	cd $(BACKEND) && golangci-lint fmt

lint: ## Run golangci-lint (default + integration build tag)
	cd $(BACKEND) && golangci-lint run ./...
	cd $(BACKEND) && golangci-lint run --build-tags=integration ./...

test: ## Run unit tests (no Docker required)
	cd $(BACKEND) && go test -race ./...

test-integration: ## Run integration tests (requires Docker)
	cd $(BACKEND) && go test -race -tags=integration ./...

generate: ## Run code generators (sqlc)
	cd $(BACKEND) && go tool sqlc generate

check-generated: generate ## Fail if generated code is stale
	git diff --exit-code -- $(BACKEND)/internal/db/sqlc

db-up: ## Start local Postgres and wait until healthy
	docker compose up -d --wait postgres

db-down: ## Stop local Postgres (data volume is kept)
	docker compose down

migrate-up: ## Apply all pending goose migrations (uses DATABASE_URL)
	cd $(BACKEND) && go tool goose -s -dir internal/db/migrations postgres "$$DATABASE_URL" up

migrate-down: ## Roll back the last goose migration (uses DATABASE_URL)
	cd $(BACKEND) && go tool goose -s -dir internal/db/migrations postgres "$$DATABASE_URL" down

migrate-status: ## Show goose migration status (uses DATABASE_URL)
	cd $(BACKEND) && go tool goose -s -dir internal/db/migrations postgres "$$DATABASE_URL" status

migrate-new: ## Create a new sequential migration (make migrate-new name=add_foo)
	cd $(BACKEND) && go tool goose -s -dir internal/db/migrations create $(name) sql

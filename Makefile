.PHONY: help dev-up dev-down migrate-up migrate-down migrate-create test test-coverage run clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev-up: ## Start development environment (PostgreSQL, Redis)
	docker-compose up -d

dev-down: ## Stop development environment
	docker-compose down

dev-logs: ## Show logs from development environment
	docker-compose logs -f

migrate-up: ## Run database migrations up
	@echo "Installing golang-migrate if not present..."
	@command -v migrate >/dev/null 2>&1 || brew install golang-migrate
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable" up

migrate-down: ## Run database migrations down
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/issue_tracker?sslmode=disable" down

migrate-create: ## Create new migration file (usage: make migrate-create name=<name>)
	@if [ -z "$(name)" ]; then \
		echo "Error: Please specify name (e.g., make migrate-create name=add_users)"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(name)

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

run: ## Run the application
	go run cmd/server/main.go

build: ## Build the application
	go build -o bin/issue-tracker cmd/server/main.go

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

install-tools: ## Install development tools
	go install github.com/golang/mock/mockgen@latest

lint: ## Run linter
	@command -v golangci-lint >/dev/null 2>&1 || brew install golangci-lint
	golangci-lint run

.DEFAULT_GOAL := help

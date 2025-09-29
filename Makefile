.PHONY: help build test lint fmt proto-gen clean run-server run-grpc-server

# Colors
CYAN := \033[36m
RESET := \033[0m

help: ## Display this help message
	@echo "$(CYAN)Available commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(CYAN)%-15s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build commands
build: ## Build the application
	@echo "$(CYAN)Building application...$(RESET)"
	@go build -o bin/server ./cmd/server
	@go build -o bin/grpc-server ./cmd/grpc-server

# Test commands
test: ## Run all tests
	@echo "$(CYAN)Running tests...$(RESET)"
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "$(CYAN)Running tests with coverage...$(RESET)"
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

test-contract: ## Run contract tests only
	@echo "$(CYAN)Running contract tests...$(RESET)"
	@go test -v ./tests/contract/...

test-integration: ## Run integration tests only
	@echo "$(CYAN)Running integration tests...$(RESET)"
	@go test -v ./tests/integration/...

# Code quality commands
lint: ## Run golangci-lint
	@echo "$(CYAN)Running linter...$(RESET)"
	@golangci-lint run

lint-fix: ## Run golangci-lint with fix
	@echo "$(CYAN)Running linter with fixes...$(RESET)"
	@golangci-lint run --fix

fmt: ## Format code with gofmt and goimports
	@echo "$(CYAN)Formatting code...$(RESET)"
	@gofmt -s -w .
	@goimports -w .

# Proto generation commands
proto-gen: ## Generate Go code from proto files
	@echo "$(CYAN)Generating proto files...$(RESET)"
	@mkdir -p internal/proto
	@protoc --proto_path=specs/001-nft/contracts \
		--go_out=internal/proto --go_opt=paths=source_relative \
		--go-grpc_out=internal/proto --go-grpc_opt=paths=source_relative \
		specs/001-nft/contracts/*.proto

proto-lint: ## Lint proto files
	@echo "$(CYAN)Linting proto files...$(RESET)"
	@buf lint specs/001-nft/contracts

# Database commands
migrate-up: ## Run database migrations up
	@echo "$(CYAN)Running migrations up...$(RESET)"
	@go run cmd/migrate/main.go up

migrate-down: ## Run database migrations down
	@echo "$(CYAN)Running migrations down...$(RESET)"
	@go run cmd/migrate/main.go down

migrate-create: ## Create new migration file (usage: make migrate-create NAME=create_users)
	@echo "$(CYAN)Creating migration: $(NAME)$(RESET)"
	@go run cmd/migrate/main.go create $(NAME)

# Development commands
run-server: ## Run the HTTP server
	@echo "$(CYAN)Starting HTTP server...$(RESET)"
	@go run ./cmd/server

run-grpc-server: ## Run the gRPC server
	@echo "$(CYAN)Starting gRPC server...$(RESET)"
	@go run ./cmd/grpc-server

dev: ## Run in development mode with hot reload
	@echo "$(CYAN)Starting development server...$(RESET)"
	@air

# Docker commands
docker-build: ## Build Docker images
	@echo "$(CYAN)Building Docker images...$(RESET)"
	@docker build -t nft-platform:latest .

docker-dev: ## Start development environment with Docker
	@echo "$(CYAN)Starting development environment...$(RESET)"
	@./scripts/docker-dev.sh up

docker-dev-app: ## Start development environment with application servers
	@echo "$(CYAN)Starting development environment with app servers...$(RESET)"
	@docker-compose --profile app up -d

docker-dev-hot: ## Start development environment with hot reload
	@echo "$(CYAN)Starting development environment with hot reload...$(RESET)"
	@docker-compose --profile hot-reload up -d

docker-run: ## Run with Docker Compose (infrastructure only)
	@echo "$(CYAN)Starting infrastructure services...$(RESET)"
	@docker-compose up -d postgres redis kafka kafka-ui hardhat-node

docker-run-all: ## Run all services with Docker Compose
	@echo "$(CYAN)Starting all services...$(RESET)"
	@docker-compose --profile app --profile monitoring up -d

docker-down: ## Stop Docker Compose
	@echo "$(CYAN)Stopping Docker Compose...$(RESET)"
	@docker-compose --profile "*" down

docker-down-clean: ## Stop Docker Compose and remove volumes
	@echo "$(CYAN)Stopping and cleaning Docker Compose...$(RESET)"
	@docker-compose --profile "*" down -v --remove-orphans

docker-logs: ## Show Docker Compose logs
	@echo "$(CYAN)Showing Docker logs...$(RESET)"
	@docker-compose logs -f

docker-ps: ## Show Docker Compose service status
	@echo "$(CYAN)Docker service status...$(RESET)"
	@docker-compose ps

docker-shell-db: ## Open PostgreSQL shell in Docker
	@echo "$(CYAN)Opening PostgreSQL shell...$(RESET)"
	@docker-compose exec postgres psql -U postgres -d nft_platform

docker-shell-redis: ## Open Redis CLI in Docker
	@echo "$(CYAN)Opening Redis CLI...$(RESET)"
	@docker-compose exec redis redis-cli

docker-kafka-topics: ## List Kafka topics in Docker
	@echo "$(CYAN)Listing Kafka topics...$(RESET)"
	@docker-compose exec kafka kafka-topics --bootstrap-server localhost:9092 --list

docker-reset: ## Reset Docker environment (clean rebuild)
	@echo "$(CYAN)Resetting Docker environment...$(RESET)"
	@./scripts/docker-dev.sh reset

# Utility commands
clean: ## Clean build artifacts
	@echo "$(CYAN)Cleaning...$(RESET)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean -testcache

deps: ## Download and tidy dependencies
	@echo "$(CYAN)Downloading dependencies...$(RESET)"
	@go mod download
	@go mod tidy

setup: deps proto-gen ## Initial setup for development
	@echo "$(CYAN)Setup complete!$(RESET)"

# Pre-commit hook
pre-commit: fmt lint test ## Run all pre-commit checks
	@echo "$(CYAN)All pre-commit checks passed!$(RESET)"
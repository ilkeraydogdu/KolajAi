# KolajAI Enterprise Marketplace - Makefile
# Production-ready build and deployment automation

# Variables
APP_NAME := kolajai-marketplace
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | awk '{print $$3}')

# Docker
DOCKER_REGISTRY := kolajAI
DOCKER_IMAGE := $(DOCKER_REGISTRY)/$(APP_NAME)
DOCKER_TAG := $(VERSION)

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -w -s"

# Colors
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo "$(GREEN)KolajAI Enterprise Marketplace - Available Commands$(NC)"
	@echo "=================================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

## Development Commands
.PHONY: dev
dev: ## Start development server with hot reload
	@echo "$(GREEN)Starting development server...$(NC)"
	@go run $(LDFLAGS) cmd/server/main.go

.PHONY: dev-frontend
dev-frontend: ## Start frontend development server
	@echo "$(GREEN)Starting frontend development server...$(NC)"
	@npm run dev

.PHONY: watch
watch: ## Watch for changes and rebuild
	@echo "$(GREEN)Watching for changes...$(NC)"
	@air -c .air.toml

## Build Commands
.PHONY: build
build: clean ## Build the application
	@echo "$(GREEN)Building application...$(NC)"
	@mkdir -p dist
	@CGO_ENABLED=1 go build $(LDFLAGS) -o dist/$(APP_NAME) cmd/server/main.go
	@echo "$(GREEN)Build completed: dist/$(APP_NAME)$(NC)"

.PHONY: build-frontend
build-frontend: ## Build frontend assets
	@echo "$(GREEN)Building frontend assets...$(NC)"
	@npm ci --silent
	@npm run build
	@echo "$(GREEN)Frontend build completed$(NC)"

.PHONY: build-all
build-all: build-frontend build ## Build everything

.PHONY: cross-build
cross-build: clean ## Cross-compile for multiple platforms
	@echo "$(GREEN)Cross-compiling for multiple platforms...$(NC)"
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/$(APP_NAME)-linux-amd64 cmd/server/main.go
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/$(APP_NAME)-darwin-amd64 cmd/server/main.go
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/$(APP_NAME)-windows-amd64.exe cmd/server/main.go
	@echo "$(GREEN)Cross-compilation completed$(NC)"

## Testing Commands
.PHONY: test
test: ## Run all tests
	@echo "$(GREEN)Running tests...$(NC)"
	@go test -v -race -timeout=30s ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: test-frontend
test-frontend: ## Run frontend tests
	@echo "$(GREEN)Running frontend tests...$(NC)"
	@npm test -- --watchAll=false --coverage

.PHONY: test-all
test-all: test test-frontend ## Run all tests (backend + frontend)

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@go test -bench=. -benchmem ./...

## Quality Assurance Commands
.PHONY: lint
lint: ## Run linters
	@echo "$(GREEN)Running linters...$(NC)"
	@golangci-lint run ./...
	@npm run lint

.PHONY: lint-fix
lint-fix: ## Fix linting issues
	@echo "$(GREEN)Fixing linting issues...$(NC)"
	@golangci-lint run --fix ./...
	@npm run lint:fix

.PHONY: format
format: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...
	@goimports -w .
	@npm run format

.PHONY: security-scan
security-scan: ## Run security scans
	@echo "$(GREEN)Running security scans...$(NC)"
	@gosec ./...
	@npm audit

.PHONY: deps-check
deps-check: ## Check for dependency updates
	@echo "$(GREEN)Checking for dependency updates...$(NC)"
	@go list -u -m all
	@npm outdated

## Docker Commands
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

.PHONY: docker-push
docker-push: docker-build ## Push Docker image to registry
	@echo "$(GREEN)Pushing Docker image to registry...$(NC)"
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@docker push $(DOCKER_IMAGE):latest
	@echo "$(GREEN)Docker image pushed$(NC)"

.PHONY: docker-run
docker-run: ## Run Docker container locally
	@echo "$(GREEN)Running Docker container...$(NC)"
	@docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):latest

.PHONY: docker-clean
docker-clean: ## Clean Docker images and containers
	@echo "$(GREEN)Cleaning Docker images and containers...$(NC)"
	@docker system prune -f
	@docker image prune -f

## Deployment Commands
.PHONY: deploy-dev
deploy-dev: ## Deploy to development environment
	@echo "$(GREEN)Deploying to development environment...$(NC)"
	@./scripts/deploy.sh development $(VERSION)

.PHONY: deploy-staging
deploy-staging: ## Deploy to staging environment
	@echo "$(GREEN)Deploying to staging environment...$(NC)"
	@./scripts/deploy.sh staging $(VERSION)

.PHONY: deploy-prod
deploy-prod: ## Deploy to production environment
	@echo "$(YELLOW)Are you sure you want to deploy to production? [y/N]$(NC)" && read ans && [ $${ans:-N} = y ]
	@echo "$(GREEN)Deploying to production environment...$(NC)"
	@./scripts/deploy.sh production $(VERSION)

.PHONY: rollback
rollback: ## Rollback deployment
	@echo "$(GREEN)Rolling back deployment...$(NC)"
	@./scripts/deploy.sh rollback

## Database Commands
.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "$(GREEN)Running database migrations...$(NC)"
	@go run cmd/migrate/main.go

.PHONY: db-seed
db-seed: ## Seed database with sample data
	@echo "$(GREEN)Seeding database...$(NC)"
	@go run cmd/seed/main.go

.PHONY: db-reset
db-reset: ## Reset database (WARNING: This will delete all data)
	@echo "$(RED)WARNING: This will delete all data. Are you sure? [y/N]$(NC)" && read ans && [ $${ans:-N} = y ]
	@echo "$(GREEN)Resetting database...$(NC)"
	@rm -f kolajAi.db
	@$(MAKE) db-migrate
	@$(MAKE) db-seed

.PHONY: db-backup
db-backup: ## Backup database
	@echo "$(GREEN)Backing up database...$(NC)"
	@mkdir -p backups
	@cp kolajAi.db backups/kolajAi-$(shell date +%Y%m%d-%H%M%S).db
	@echo "$(GREEN)Database backed up$(NC)"

## Monitoring Commands
.PHONY: logs
logs: ## Show application logs
	@docker-compose logs -f app

.PHONY: health
health: ## Check application health
	@curl -f http://localhost:8080/health | jq .

.PHONY: metrics
metrics: ## Show application metrics
	@curl -s http://localhost:8080/metrics

.PHONY: status
status: ## Show service status
	@docker-compose ps

## Maintenance Commands
.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@rm -rf dist/
	@rm -rf node_modules/.cache/
	@rm -f coverage.out coverage.html
	@go clean -cache -testcache -modcache

.PHONY: deps
deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@npm ci

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy
	@npm update

.PHONY: generate
generate: ## Generate code (mocks, docs, etc.)
	@echo "$(GREEN)Generating code...$(NC)"
	@go generate ./...

## Documentation Commands
.PHONY: docs
docs: ## Generate documentation
	@echo "$(GREEN)Generating documentation...$(NC)"
	@godoc -http=:6060 &
	@echo "$(GREEN)Documentation server started at http://localhost:6060$(NC)"

.PHONY: api-docs
api-docs: ## Generate API documentation
	@echo "$(GREEN)Generating API documentation...$(NC)"
	@swag init -g cmd/server/main.go -o docs/

## Utility Commands
.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/cosmtrek/air@latest

.PHONY: setup
setup: install-tools deps ## Setup development environment
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@cp .env.example .env
	@echo "$(YELLOW)Please update .env file with your configuration$(NC)"

.PHONY: version
version: ## Show version information
	@echo "App Name: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(GO_VERSION)"

## Docker Compose Commands
.PHONY: up
up: ## Start all services with Docker Compose
	@echo "$(GREEN)Starting all services...$(NC)"
	@docker-compose up -d

.PHONY: down
down: ## Stop all services
	@echo "$(GREEN)Stopping all services...$(NC)"
	@docker-compose down

.PHONY: restart
restart: ## Restart services
	@echo "$(GREEN)Restarting services...$(NC)"
	@docker-compose restart

.PHONY: ps
ps: ## Show running containers
	@docker-compose ps

## Production Specific Commands
.PHONY: prod-check
prod-check: ## Run production readiness checks
	@echo "$(GREEN)Running production readiness checks...$(NC)"
	@$(MAKE) test-all
	@$(MAKE) lint
	@$(MAKE) security-scan
	@echo "$(GREEN)Production readiness checks completed$(NC)"

.PHONY: backup
backup: ## Create full system backup
	@echo "$(GREEN)Creating system backup...$(NC)"
	@docker-compose run --rm backup

.PHONY: restore
restore: ## Restore from backup
	@echo "$(GREEN)Restoring from backup...$(NC)"
	@echo "$(YELLOW)Please specify backup file to restore from$(NC)"

# Default target
.DEFAULT_GOAL := help
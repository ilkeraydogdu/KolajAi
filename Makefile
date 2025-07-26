# KolajAI Marketplace Makefile

.PHONY: build test clean run seed tools lint vet fmt deps integration-test unit-test all-tests

# Build the main application
build:
	@echo "Building KolajAI server..."
	go build -o server ./cmd/server

# Build all tools
build-tools:
	@echo "Building tools..."
	go build -o seed ./cmd/seed
	go build -o db-tools ./cmd/db-tools

# Run unit tests
unit-test:
	@echo "Running unit tests..."
	go test ./internal/models -v
	go test ./internal/database -v
	go test ./internal/services -v

# Run integration tests
integration-test:
	@echo "Running integration tests..."
	go test . -v

# Run all tests
all-tests: unit-test integration-test
	@echo "All tests completed!"

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run the server
run: build
	@echo "Starting KolajAI server..."
	./server

# Run database seeding
seed: build-tools
	@echo "Seeding database..."
	./seed

# Run database tools
db-info: build-tools
	@echo "Getting database info..."
	./db-tools info

# Lint the code
lint:
	@echo "Running linter..."
	golangci-lint run

# Vet the code
vet:
	@echo "Running go vet..."
	go vet ./...

# Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -f server seed db-tools
	rm -f *.db
	rm -f *.log
	rm -f coverage.out coverage.html
	rm -f test_*.db

# Development setup
dev-setup: deps fmt vet build-tools
	@echo "Development environment setup complete!"

# Full build and test pipeline
ci: deps fmt vet all-tests build build-tools
	@echo "CI pipeline completed successfully!"

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the main server application"
	@echo "  build-tools    - Build all command-line tools"
	@echo "  unit-test      - Run unit tests"
	@echo "  integration-test - Run integration tests"
	@echo "  all-tests      - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  run            - Build and run the server"
	@echo "  seed           - Run database seeding"
	@echo "  db-info        - Get database information"
	@echo "  lint           - Run linter"
	@echo "  vet            - Run go vet"
	@echo "  fmt            - Format code"
	@echo "  deps           - Install dependencies"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev-setup      - Setup development environment"
	@echo "  ci             - Run full CI pipeline"
	@echo "  help           - Show this help message"
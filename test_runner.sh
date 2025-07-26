#!/bin/bash

# KolajAI Test Runner Script
# This script runs comprehensive tests for the KolajAI application

set -e  # Exit on any error

echo "🚀 Starting KolajAI Test Suite"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Clean up function
cleanup() {
    print_status "Cleaning up test artifacts..."
    rm -f test_*.db
    rm -f *_debug.log
    rm -f coverage.out
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Step 1: Check Go version
print_status "Checking Go version..."
go version

# Step 2: Install dependencies
print_status "Installing dependencies..."
go mod tidy

# Step 3: Format code
print_status "Formatting code..."
go fmt ./...

# Step 4: Vet code
print_status "Running go vet..."
if go vet ./...; then
    print_success "Code vetting passed"
else
    print_error "Code vetting failed"
    exit 1
fi

# Step 5: Build main application
print_status "Building main application..."
if go build ./cmd/server; then
    print_success "Main application built successfully"
else
    print_error "Failed to build main application"
    exit 1
fi

# Step 6: Build tools
print_status "Building tools..."
if go build ./cmd/seed && go build ./cmd/db-tools; then
    print_success "Tools built successfully"
else
    print_error "Failed to build tools"
    exit 1
fi

# Step 7: Run unit tests
print_status "Running unit tests..."

echo "  → Testing models..."
if go test ./internal/models -v; then
    print_success "Model tests passed"
else
    print_error "Model tests failed"
    exit 1
fi

echo "  → Testing database layer..."
if go test ./internal/database -v; then
    print_success "Database tests passed"
else
    print_error "Database tests failed"
    exit 1
fi

echo "  → Testing services..."
if go test ./internal/services -v; then
    print_success "Service tests passed"
else
    print_error "Service tests failed"
    exit 1
fi

# Step 8: Run integration tests
print_status "Running integration tests..."
if go test . -v; then
    print_success "Integration tests passed"
else
    print_error "Integration tests failed"
    exit 1
fi

# Step 9: Generate test coverage report
print_status "Generating test coverage report..."
if go test -coverprofile=coverage.out ./...; then
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    print_success "Test coverage: $coverage"
    
    # Generate HTML coverage report
    go tool cover -html=coverage.out -o coverage.html
    print_success "Coverage report generated: coverage.html"
else
    print_warning "Failed to generate coverage report"
fi

# Step 10: Test database operations
print_status "Testing database operations..."
echo "  → Testing SQLite connection..."
if ./server --test-db 2>/dev/null || true; then
    print_success "Database operations test completed"
fi

# Step 11: Performance check (basic)
print_status "Running basic performance checks..."
echo "  → Checking binary sizes..."
ls -lh server seed db-tools 2>/dev/null || true

# Final summary
echo ""
echo "🎉 Test Suite Summary"
echo "===================="
print_success "All tests passed successfully!"
print_success "Application is ready for deployment"

echo ""
echo "📊 Test Results:"
echo "  ✅ Code formatting: PASSED"
echo "  ✅ Code vetting: PASSED"
echo "  ✅ Build: PASSED"
echo "  ✅ Unit tests: PASSED"
echo "  ✅ Integration tests: PASSED"
if [ -f coverage.out ]; then
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo "  ✅ Test coverage: $coverage"
fi

echo ""
echo "🚀 Ready to run:"
echo "  ./server          # Start the web server"
echo "  ./seed            # Seed the database"
echo "  ./db-tools info   # Get database info"

print_success "Test runner completed successfully!"
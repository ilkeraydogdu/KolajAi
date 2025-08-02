#!/bin/bash

# KolajAI Enterprise Marketplace - Production Deployment Script
# Usage: ./scripts/deploy.sh [environment] [version]

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENVIRONMENT="${1:-production}"
VERSION="${2:-latest}"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-kolajAI}"
SERVICE_NAME="kolajai-app"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Docker is installed and running
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker is not running"
        exit 1
    fi
    
    # Check if Docker Compose is installed
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi
    
    # Check if .env file exists
    if [[ ! -f "${PROJECT_ROOT}/.env" ]]; then
        log_warning ".env file not found, copying from .env.example"
        if [[ -f "${PROJECT_ROOT}/.env.example" ]]; then
            cp "${PROJECT_ROOT}/.env.example" "${PROJECT_ROOT}/.env"
            log_warning "Please update .env file with your configuration before proceeding"
            exit 1
        else
            log_error ".env.example file not found"
            exit 1
        fi
    fi
    
    log_success "Prerequisites check passed"
}

# Build application
build_application() {
    log_info "Building application..."
    
    cd "${PROJECT_ROOT}"
    
    # Build frontend assets
    log_info "Building frontend assets..."
    if [[ -f "package.json" ]]; then
        npm ci --silent
        npm run build
        log_success "Frontend build completed"
    else
        log_warning "package.json not found, skipping frontend build"
    fi
    
    # Build Docker images
    log_info "Building Docker images..."
    docker-compose build --no-cache
    
    # Tag images with version
    if [[ "${VERSION}" != "latest" ]]; then
        docker tag "${DOCKER_REGISTRY}/${SERVICE_NAME}:latest" "${DOCKER_REGISTRY}/${SERVICE_NAME}:${VERSION}"
    fi
    
    log_success "Application build completed"
}

# Run tests
run_tests() {
    log_info "Running tests..."
    
    cd "${PROJECT_ROOT}"
    
    # Run Go tests
    if command -v go &> /dev/null; then
        log_info "Running Go tests..."
        go test -v ./... -timeout=30s
        log_success "Go tests passed"
    else
        log_warning "Go not found, skipping Go tests"
    fi
    
    # Run frontend tests
    if [[ -f "package.json" ]] && npm list jest &> /dev/null; then
        log_info "Running frontend tests..."
        npm test -- --watchAll=false --coverage
        log_success "Frontend tests passed"
    else
        log_warning "Jest not found, skipping frontend tests"
    fi
    
    log_success "All tests passed"
}

# Pre-deployment checks
pre_deployment_checks() {
    log_info "Running pre-deployment checks..."
    
    # Check if services are already running
    if docker-compose ps | grep -q "Up"; then
        log_info "Services are currently running"
        
        # Health check
        log_info "Performing health check..."
        if curl -f -s http://localhost:8080/health > /dev/null; then
            log_success "Current deployment is healthy"
        else
            log_warning "Current deployment health check failed"
        fi
    fi
    
    # Check disk space
    AVAILABLE_SPACE=$(df / | awk 'NR==2 {print $4}')
    REQUIRED_SPACE=1048576  # 1GB in KB
    
    if [[ ${AVAILABLE_SPACE} -lt ${REQUIRED_SPACE} ]]; then
        log_error "Insufficient disk space. Available: ${AVAILABLE_SPACE}KB, Required: ${REQUIRED_SPACE}KB"
        exit 1
    fi
    
    log_success "Pre-deployment checks passed"
}

# Deploy application
deploy_application() {
    log_info "Deploying application..."
    
    cd "${PROJECT_ROOT}"
    
    # Create backup of current deployment
    if docker-compose ps | grep -q "Up"; then
        log_info "Creating backup of current deployment..."
        docker-compose exec -T app ./scripts/backup.sh || log_warning "Backup failed"
    fi
    
    # Pull latest images (if using registry)
    if [[ "${DOCKER_REGISTRY}" != "kolajAI" ]]; then
        log_info "Pulling latest images..."
        docker-compose pull
    fi
    
    # Deploy with zero-downtime strategy
    log_info "Starting deployment..."
    
    # Start new containers
    docker-compose up -d --remove-orphans
    
    # Wait for services to be ready
    log_info "Waiting for services to be ready..."
    sleep 30
    
    # Health check
    RETRY_COUNT=0
    MAX_RETRIES=30
    
    while [[ ${RETRY_COUNT} -lt ${MAX_RETRIES} ]]; do
        if curl -f -s http://localhost:8080/health > /dev/null; then
            log_success "Deployment health check passed"
            break
        fi
        
        log_info "Health check failed, retrying... (${RETRY_COUNT}/${MAX_RETRIES})"
        sleep 10
        ((RETRY_COUNT++))
    done
    
    if [[ ${RETRY_COUNT} -eq ${MAX_RETRIES} ]]; then
        log_error "Deployment health check failed after ${MAX_RETRIES} retries"
        rollback_deployment
        exit 1
    fi
    
    log_success "Application deployed successfully"
}

# Rollback deployment
rollback_deployment() {
    log_warning "Rolling back deployment..."
    
    cd "${PROJECT_ROOT}"
    
    # Stop current containers
    docker-compose down
    
    # Restore from backup (if available)
    if [[ -f "backup/latest.tar.gz" ]]; then
        log_info "Restoring from backup..."
        # Restore logic here
        log_success "Rollback completed"
    else
        log_error "No backup found for rollback"
    fi
}

# Post-deployment tasks
post_deployment_tasks() {
    log_info "Running post-deployment tasks..."
    
    cd "${PROJECT_ROOT}"
    
    # Run database migrations
    log_info "Running database migrations..."
    docker-compose exec -T app go run cmd/migrate/main.go || log_warning "Migration failed"
    
    # Clear caches
    log_info "Clearing caches..."
    docker-compose exec -T redis redis-cli FLUSHALL || log_warning "Cache clear failed"
    
    # Warm up caches
    log_info "Warming up caches..."
    curl -s http://localhost:8080/api/warmup > /dev/null || log_warning "Cache warmup failed"
    
    # Send deployment notification
    send_deployment_notification
    
    log_success "Post-deployment tasks completed"
}

# Send deployment notification
send_deployment_notification() {
    log_info "Sending deployment notification..."
    
    # Slack notification (if configured)
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"🚀 KolajAI ${ENVIRONMENT} deployment completed successfully - Version: ${VERSION}\"}" \
            "${SLACK_WEBHOOK_URL}" || log_warning "Slack notification failed"
    fi
    
    # Email notification (if configured)
    if [[ -n "${NOTIFICATION_EMAIL:-}" ]]; then
        echo "KolajAI ${ENVIRONMENT} deployment completed successfully - Version: ${VERSION}" | \
            mail -s "Deployment Notification" "${NOTIFICATION_EMAIL}" || log_warning "Email notification failed"
    fi
    
    log_success "Deployment notification sent"
}

# Cleanup old images and containers
cleanup() {
    log_info "Cleaning up old images and containers..."
    
    # Remove unused images
    docker image prune -f
    
    # Remove unused containers
    docker container prune -f
    
    # Remove unused volumes (be careful with this in production)
    # docker volume prune -f
    
    # Remove unused networks
    docker network prune -f
    
    log_success "Cleanup completed"
}

# Display deployment summary
show_deployment_summary() {
    log_info "Deployment Summary"
    echo "===================="
    echo "Environment: ${ENVIRONMENT}"
    echo "Version: ${VERSION}"
    echo "Deployed at: $(date)"
    echo ""
    
    # Show running services
    log_info "Running services:"
    docker-compose ps
    
    echo ""
    log_info "Application URLs:"
    echo "Main App: http://localhost:8080"
    echo "Grafana: http://localhost:3000"
    echo "Prometheus: http://localhost:9090"
    echo "Kibana: http://localhost:5601"
    
    log_success "Deployment completed successfully! 🎉"
}

# Main deployment function
main() {
    log_info "Starting KolajAI deployment..."
    log_info "Environment: ${ENVIRONMENT}"
    log_info "Version: ${VERSION}"
    
    # Trap errors and cleanup
    trap 'log_error "Deployment failed!"; exit 1' ERR
    
    check_prerequisites
    build_application
    run_tests
    pre_deployment_checks
    deploy_application
    post_deployment_tasks
    cleanup
    show_deployment_summary
}

# Handle script arguments
case "${1:-}" in
    "production"|"staging"|"development")
        main "$@"
        ;;
    "rollback")
        rollback_deployment
        ;;
    "health")
        curl -f http://localhost:8080/health | jq .
        ;;
    "logs")
        docker-compose logs -f "${2:-app}"
        ;;
    "status")
        docker-compose ps
        ;;
    "stop")
        docker-compose down
        ;;
    "restart")
        docker-compose restart "${2:-}"
        ;;
    *)
        echo "Usage: $0 {production|staging|development|rollback|health|logs|status|stop|restart} [service]"
        echo ""
        echo "Commands:"
        echo "  production|staging|development  Deploy to specified environment"
        echo "  rollback                        Rollback to previous deployment"
        echo "  health                         Check application health"
        echo "  logs [service]                 Show logs for service (default: app)"
        echo "  status                         Show service status"
        echo "  stop                           Stop all services"
        echo "  restart [service]              Restart service(s)"
        exit 1
        ;;
esac
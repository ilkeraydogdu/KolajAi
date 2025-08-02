# Multi-stage build for optimized production image
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -extldflags '-static'" \
    -a -installsuffix cgo \
    -o kolajAi \
    ./cmd/server/main.go

# Production stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    sqlite \
    curl \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1001 -S kolajAi && \
    adduser -u 1001 -S kolajAi -G kolajAi

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/kolajAi .

# Copy configuration files
COPY --from=builder /app/config.yaml ./config.yaml
COPY --from=builder /app/config.production.yaml ./config.production.yaml

# Copy web assets
COPY --from=builder /app/web ./web

# Create necessary directories
RUN mkdir -p /var/log/kolajAI && \
    mkdir -p /app/data && \
    chown -R kolajAi:kolajAi /app /var/log/kolajAI

# Set permissions
RUN chmod +x /app/kolajAi

# Switch to non-root user
USER kolajAi

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV APP_ENV=production

# Run the application
CMD ["./kolajAi"]
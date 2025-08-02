# Multi-stage Docker build for KolajAI Enterprise Marketplace

# Stage 1: Frontend Build
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production --silent

# Copy source files
COPY web/static ./web/static
COPY webpack.config.js ./
COPY tailwind.config.js ./
COPY postcss.config.js ./

# Build frontend assets
RUN npm run build

# Stage 2: Go Build
FROM golang:1.21-alpine AS go-builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Copy built frontend assets from previous stage
COPY --from=frontend-builder /app/dist ./web/static/dist

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/server

# Stage 3: Production Image
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    sqlite \
    curl \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1001 -S kolajai && \
    adduser -u 1001 -S kolajai -G kolajai

# Set timezone
ENV TZ=Europe/Istanbul
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app

# Copy binary from builder stage
COPY --from=go-builder /app/main .

# Copy static files and templates
COPY --from=go-builder /app/web ./web
COPY --from=go-builder /app/configs ./configs

# Create necessary directories
RUN mkdir -p /app/data /app/logs /app/uploads /app/temp && \
    chown -R kolajai:kolajai /app

# Switch to non-root user
USER kolajai

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Expose port
EXPOSE 8080

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8080
ENV DB_PATH=/app/data/kolajAi.db
ENV UPLOAD_PATH=/app/uploads
ENV LOG_PATH=/app/logs

# Run the application
CMD ["./main"]

# Labels for metadata
LABEL maintainer="KolajAI Team <team@kolaj.ai>"
LABEL version="1.0.0"
LABEL description="KolajAI Enterprise Marketplace"
LABEL org.opencontainers.image.source="https://github.com/kolajAI/marketplace"
LABEL org.opencontainers.image.documentation="https://docs.kolaj.ai"
LABEL org.opencontainers.image.licenses="MIT"
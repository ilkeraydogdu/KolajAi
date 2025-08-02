#!/bin/bash

# KolajAI Build Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="kolajAi"
VERSION=${VERSION:-"2.0.0"}
BUILD_DIR="build"
BINARY_NAME="kolajAi"
LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

echo -e "${GREEN}🚀 Building KolajAI Enterprise Marketplace v${VERSION}${NC}"

# Clean previous builds
echo -e "${YELLOW}🧹 Cleaning previous builds...${NC}"
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

# Check Go version
echo -e "${YELLOW}🔍 Checking Go version...${NC}"
go version

# Download dependencies
echo -e "${YELLOW}📦 Downloading dependencies...${NC}"
go mod download
go mod verify

# Run tests
echo -e "${YELLOW}🧪 Running tests...${NC}"
go test -v -race -coverprofile=coverage.out ./...

# Generate test coverage report
echo -e "${YELLOW}📊 Generating coverage report...${NC}"
go tool cover -html=coverage.out -o ${BUILD_DIR}/coverage.html
go tool cover -func=coverage.out | tail -1

# Build for different platforms
platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a platform_split <<< "$platform"
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    
    output_name=${BINARY_NAME}
    if [ $GOOS = "windows" ]; then
        output_name+=".exe"
    fi
    
    output_path="${BUILD_DIR}/${BINARY_NAME}-${GOOS}-${GOARCH}/${output_name}"
    
    echo -e "${YELLOW}🔨 Building for ${GOOS}/${GOARCH}...${NC}"
    
    mkdir -p "${BUILD_DIR}/${BINARY_NAME}-${GOOS}-${GOARCH}"
    
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=1 go build \
        -ldflags="${LDFLAGS}" \
        -o $output_path \
        ./cmd/server/main.go
    
    # Copy configuration files
    cp config.yaml "${BUILD_DIR}/${BINARY_NAME}-${GOOS}-${GOARCH}/"
    cp config.production.yaml "${BUILD_DIR}/${BINARY_NAME}-${GOOS}-${GOARCH}/"
    
    # Copy web assets
    cp -r web "${BUILD_DIR}/${BINARY_NAME}-${GOOS}-${GOARCH}/"
    
    # Create archive
    cd ${BUILD_DIR}
    if [ $GOOS = "windows" ]; then
        zip -r "${BINARY_NAME}-${GOOS}-${GOARCH}-v${VERSION}.zip" "${BINARY_NAME}-${GOOS}-${GOARCH}/"
    else
        tar -czf "${BINARY_NAME}-${GOOS}-${GOARCH}-v${VERSION}.tar.gz" "${BINARY_NAME}-${GOOS}-${GOARCH}/"
    fi
    cd ..
    
    echo -e "${GREEN}✅ Built ${output_path}${NC}"
done

# Build Docker image
echo -e "${YELLOW}🐳 Building Docker image...${NC}"
docker build -t ${APP_NAME}:${VERSION} .
docker tag ${APP_NAME}:${VERSION} ${APP_NAME}:latest

# Generate checksums
echo -e "${YELLOW}🔐 Generating checksums...${NC}"
cd ${BUILD_DIR}
find . -name "*.tar.gz" -o -name "*.zip" | xargs sha256sum > checksums.txt
cd ..

# Build summary
echo -e "${GREEN}🎉 Build completed successfully!${NC}"
echo -e "${GREEN}📁 Build artifacts are in the ${BUILD_DIR} directory${NC}"
echo -e "${GREEN}🐳 Docker image: ${APP_NAME}:${VERSION}${NC}"

# File sizes
echo -e "${YELLOW}📏 Build sizes:${NC}"
ls -lh ${BUILD_DIR}/*.tar.gz ${BUILD_DIR}/*.zip 2>/dev/null || true

echo -e "${GREEN}✨ Build process finished!${NC}"
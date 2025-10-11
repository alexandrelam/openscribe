# OpenScribe Makefile

# Binary name
BINARY_NAME=openscribe

# Build directory
BUILD_DIR=bin

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "\
	-X 'github.com/alexandrelam/openscribe/internal/cli.Version=$(VERSION)' \
	-X 'github.com/alexandrelam/openscribe/internal/cli.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/alexandrelam/openscribe/internal/cli.BuildDate=$(BUILD_DATE)'"

# Build the project
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/openscribe

# Build for macOS ARM64
build-darwin-arm64:
	@echo "Building $(BINARY_NAME) $(VERSION) for macOS ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/openscribe

# Build for macOS x86_64
build-darwin-amd64:
	@echo "Building $(BINARY_NAME) $(VERSION) for macOS x86_64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/openscribe

# Build all architectures
build-all: build-darwin-arm64 build-darwin-amd64
	@echo "All builds complete!"

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

# Run the application
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Display help
help:
	@echo "OpenScribe Makefile Commands:"
	@echo "  make build               - Build the binary"
	@echo "  make build-darwin-arm64  - Build for macOS ARM64"
	@echo "  make build-darwin-amd64  - Build for macOS x86_64"
	@echo "  make build-all           - Build all architectures"
	@echo "  make install             - Install the binary to GOPATH/bin"
	@echo "  make run                 - Build and run the application"
	@echo "  make clean               - Remove build artifacts"
	@echo "  make test                - Run tests"
	@echo "  make deps                - Download and tidy dependencies"
	@echo "  make fmt                 - Format code"
	@echo "  make lint                - Run linter"
	@echo "  make help                - Display this help message"

.PHONY: build build-darwin-arm64 build-darwin-amd64 build-all install run clean test deps fmt lint help

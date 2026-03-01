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

# Build for macOS ARM64 (Apple Silicon)
build-darwin-arm64:
	@echo "Building $(BINARY_NAME) $(VERSION) for macOS ARM64 (Apple Silicon)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/openscribe

# Build all architectures (only ARM64 for Apple Silicon)
build-all: build-darwin-arm64
	@echo "Build complete!"

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

# Build with Moonshine support (requires vendored native libs)
build-moonshine:
	@echo "Building $(BINARY_NAME) $(VERSION) with Moonshine support..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -tags moonshine -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/openscribe

# Download and extract Moonshine native dependencies from XCFramework
moonshine-deps:
	@echo "Downloading Moonshine XCFramework..."
	@mkdir -p third_party/moonshine/include third_party/moonshine/lib
	@curl -sL -o /tmp/Moonshine.xcframework.zip \
		"https://github.com/moonshine-ai/moonshine-swift/releases/download/v0.0.49/Moonshine.xcframework.zip"
	@echo "Extracting native library and header..."
	@unzip -jo /tmp/Moonshine.xcframework.zip \
		"Moonshine.xcframework/macos-arm64_x86_64/libmoonshine.a" \
		-d third_party/moonshine/lib/
	@unzip -jo /tmp/Moonshine.xcframework.zip \
		"Moonshine.xcframework/macos-arm64_x86_64/Headers/moonshine-c-api.h" \
		-d third_party/moonshine/include/
	@rm -f /tmp/Moonshine.xcframework.zip
	@echo "Done! Native deps extracted to third_party/moonshine/"

# Display help
help:
	@echo "OpenScribe Makefile Commands:"
	@echo "  make build               - Build the binary"
	@echo "  make build-darwin-arm64  - Build for macOS ARM64 (Apple Silicon)"
	@echo "  make build-all           - Build for Apple Silicon"
	@echo "  make install             - Install the binary to GOPATH/bin"
	@echo "  make run                 - Build and run the application"
	@echo "  make clean               - Remove build artifacts"
	@echo "  make test                - Run tests"
	@echo "  make deps                - Download and tidy dependencies"
	@echo "  make fmt                 - Format code"
	@echo "  make lint                - Run linter"
	@echo "  make build-moonshine     - Build with Moonshine support (requires native libs)"
	@echo "  make moonshine-deps      - Show instructions for Moonshine native deps"
	@echo "  make help                - Display this help message"
	@echo ""
	@echo "Note: OpenScribe only supports macOS Apple Silicon (M1/M2/M3/M4)"

.PHONY: build build-darwin-arm64 build-all build-moonshine moonshine-deps install run clean test deps fmt lint help

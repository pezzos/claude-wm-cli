.PHONY: build test lint clean install dev help

# Build variables
BINARY_NAME=claude-wm-cli
BUILD_DIR=build
GO_FILES=$(shell find . -name "*.go" -type f -not -path "./vendor/*")

# Version information
VERSION ?= dev
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X claude-wm-cli/cmd.Version=$(VERSION) -X claude-wm-cli/cmd.GitCommit=$(GIT_COMMIT) -X claude-wm-cli/cmd.BuildTime=$(BUILD_TIME)"

# Build the binary
build:
	@echo "Building $(BINARY_NAME) $(VERSION) ($(GIT_COMMIT))..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/claude-wm

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@echo "Running linter..."
	@$(shell go env GOPATH)/bin/golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Install dependencies
install:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Development setup
dev: install
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest

# Cross-platform build targets
build-all: build-linux build-windows build-darwin
	@echo "All platform binaries built!"

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/claude-wm
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/claude-wm

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/claude-wm
	@GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe ./cmd/claude-wm

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/claude-wm
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/claude-wm

# Create release packages
package: build-all
	@echo "Creating release packages..."
	@mkdir -p $(BUILD_DIR)/packages
	@for binary in $(BUILD_DIR)/$(BINARY_NAME)-*; do \
		if [ -f "$$binary" ]; then \
			base=$$(basename "$$binary"); \
			platform=$$(echo "$$base" | sed 's/$(BINARY_NAME)-//'); \
			mkdir -p $(BUILD_DIR)/packages/$$platform; \
			cp "$$binary" $(BUILD_DIR)/packages/$$platform/; \
			cp docs/README.md $(BUILD_DIR)/packages/$$platform/; \
			cp LICENSE $(BUILD_DIR)/packages/$$platform/ 2>/dev/null || true; \
			cd $(BUILD_DIR)/packages && tar -czf $$platform.tar.gz $$platform/; \
			cd ../..; \
		fi \
	done
	@echo "Release packages created in $(BUILD_DIR)/packages/"

# Run the application
run:
	@go run ./cmd/claude-wm $(ARGS)

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w $(GO_FILES)
	@$(shell go env GOPATH)/bin/goimports -w $(GO_FILES)

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary for current platform"
	@echo "  build-all     - Build binaries for all platforms"
	@echo "  build-linux   - Build Linux binaries (amd64, arm64)"
	@echo "  build-windows - Build Windows binaries (amd64, arm64)"
	@echo "  build-darwin  - Build macOS binaries (amd64, arm64)"
	@echo "  package       - Create release packages for all platforms"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install dependencies"
	@echo "  dev           - Setup development environment"
	@echo "  run           - Run the application (use ARGS for arguments)"
	@echo "  fmt           - Format code"
	@echo "  help          - Show this help"

# Default target
default: help
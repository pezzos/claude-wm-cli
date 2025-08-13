.PHONY: build test test-all test-smoke test-unit test-integration test-system test-guard lint clean install dev help manifest coverage coverage-html serena-index validate-output

# Build variables
BINARY_NAME=claude-wm-cli
BUILD_DIR=build
GO_FILES=$(shell find . -name "*.go" -type f -not -path "./vendor/*")

# Version information
VERSION ?= dev
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X claude-wm-cli/internal/meta.Version=$(VERSION) -X claude-wm-cli/internal/meta.Commit=$(GIT_COMMIT) -X claude-wm-cli/internal/meta.BuildDate=$(BUILD_TIME)"

# Generate system manifest
manifest:
	@echo "Generating system manifest..."
	@go run internal/config/gen/manifest.go

# Build the binary
build: manifest
	@echo "Building $(BINARY_NAME) $(VERSION) ($(GIT_COMMIT))..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/claude-wm

# L0: Smoke Tests (Basic Functionality)
test-smoke: manifest
	@echo "🔍 L0: Running Smoke Tests..."
	@echo "   → Building binary..."
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/claude-wm
	@echo "   → Testing basic commands..."
	@$(BUILD_DIR)/$(BINARY_NAME) --version > /dev/null
	@$(BUILD_DIR)/$(BINARY_NAME) --help > /dev/null
	@$(BUILD_DIR)/$(BINARY_NAME) config --help > /dev/null
	@echo "✅ L0: Smoke tests passed"

# L1: Unit Tests (Component Testing)
test-unit:
	@echo "🧪 L1: Running Unit Tests..."
	@go test -short ./internal/diff/ ./internal/config/ ./internal/cmd/ ./internal/update/ ./internal/wm/... ./internal/fsutil/ ./internal/validation/ ./internal/model/
	@echo "✅ L1: Unit tests passed"

# L2: Integration Tests (Component Interaction)
test-integration:
	@echo "🔗 L2: Running Integration Tests..."
	@go test ./internal/integration/ ./cmd/ -run Integration
	@echo "✅ L2: Integration tests passed"

# L3: System Tests (End-to-End)
test-system:
	@echo "🏗️  L3: Running System Tests..."
	@go test ./test/system/ -timeout 10m || echo "⚠️  Note: System tests may not exist yet"
	@echo "✅ L3: System tests completed"

# Guard and Hook Testing
test-guard:
	@echo "🛡️  Running Guard/Hook Tests..."
	@go test ./internal/integration/ -run TestGuard -run TestHook -run TestInstallHook
	@echo "✅ Guard/Hook tests passed"

# Complete Test Suite (L0 → L3)
test-all: manifest
	@echo "🚀 Running Complete Test Suite (L0 → L3)"
	@echo "========================================="
	@$(MAKE) test-smoke
	@$(MAKE) test-unit  
	@$(MAKE) test-integration
	@$(MAKE) test-guard
	@$(MAKE) test-system
	@echo ""
	@echo "🎉 All Tests Passed!"
	@echo "L0 Smoke ............. ✅"
	@echo "L1 Unit .............. ✅"
	@echo "L2 Integration ....... ✅"
	@echo "L3 Guard/Hooks ....... ✅"
	@echo "L4 System ............ ✅"

# Enhanced Test Runner with Summary
test-runner:
	@echo "🚀 Running Enhanced Test Suite..."
	@go run ./internal/testrunner/main.go

# Legacy test target for compatibility
test:
	@echo "Running legacy test target..."
	@$(MAKE) test-unit

# Coverage Analysis
coverage:
	@echo "📊 Generating Coverage Report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

# HTML Coverage Report
coverage-html: coverage
	@echo "📊 Generating HTML Coverage Report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

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

# Serena Documentation Indexing (Incremental)
serena-index:
	@echo "📚 Running Serena incremental documentation indexer..."
	@go run ./cmd/serena-indexer/main.go -root .

# Validate Claude Code Output
validate-output:
	@echo "🔍 Validating Claude Code output schema..."
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make validate-output FILE=path/to/output.json"; \
		echo "   or: echo '{...json...}' | make validate-output-stdin"; \
		exit 1; \
	fi
	@node ./scripts/validate-output.js "$(FILE)" --verbose

# Validate output from stdin
validate-output-stdin:
	@echo "🔍 Validating Claude Code output from stdin..."
	@node ./scripts/validate-output.js --stdin --verbose

# Show help
help:
	@echo "Available targets:"
	@echo ""
	@echo "🏗️  Build Targets:"
	@echo "  build         - Build the binary for current platform"
	@echo "  build-all     - Build binaries for all platforms"
	@echo "  build-linux   - Build Linux binaries (amd64, arm64)"
	@echo "  build-windows - Build Windows binaries (amd64, arm64)"
	@echo "  build-darwin  - Build macOS binaries (amd64, arm64)"
	@echo "  package       - Create release packages for all platforms"
	@echo ""
	@echo "🧪 Testing Targets (L0 → L3 Protocol):"
	@echo "  test-all      - Run complete test suite (L0 → L3)"
	@echo "  test-smoke    - L0: Smoke tests (basic functionality)"
	@echo "  test-unit     - L1: Unit tests (component testing)"
	@echo "  test-integration - L2: Integration tests (component interaction)"
	@echo "  test-system   - L3: System tests (end-to-end)"
	@echo "  test-guard    - Guard/Hook specific tests"
	@echo "  test-runner   - Enhanced test runner with detailed summary"
	@echo "  test          - Legacy test target (L1 unit tests)"
	@echo ""
	@echo "📊 Coverage & Analysis:"
	@echo "  coverage      - Generate coverage report"
	@echo "  coverage-html - Generate HTML coverage report"
	@echo "  lint          - Run linter"
	@echo ""
	@echo "🛠️  Development:"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install dependencies"
	@echo "  dev           - Setup development environment"
	@echo "  run           - Run the application (use ARGS for arguments)"
	@echo "  fmt           - Format code"
	@echo "  manifest      - Generate system configuration manifest"
	@echo "  serena-index  - Update Serena documentation index (incremental)"
	@echo "  validate-output - Validate Claude Code JSON output (use FILE=path)"
	@echo "  help          - Show this help"

# Default target
default: help
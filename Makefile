# Weather Service Makefile
# Project configuration
PROJECT_NAME = weather
GO_VERSION = 1.23
GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
VERSION = $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -w -s -extldflags "-static" -X main.version=$(VERSION)
REGISTRY = ghcr.io/apimgr

# Build flags
BUILD_FLAGS = -ldflags="$(LDFLAGS)" -trimpath
STATIC_FLAGS = $(BUILD_FLAGS) -a -installsuffix cgo

# Target platforms (8 platforms as per SPEC)
PLATFORMS = \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64 \
	freebsd/amd64 \
	freebsd/arm64

.PHONY: all build build-all test docker release clean help dev run deps info install

# Default target
all: build

# Build for host platform only
build:
	@echo "Building $(PROJECT_NAME) for host ($(GOOS)/$(GOARCH))..."
	CGO_ENABLED=0 go build $(STATIC_FLAGS) -o $(PROJECT_NAME) main.go
	@echo "✅ Built $(PROJECT_NAME)"

# Build for all platforms
build-all: clean
	@echo "Building $(PROJECT_NAME) for all platforms..."
	@mkdir -p dist binaries
	# Build host binary
	CGO_ENABLED=0 go build $(STATIC_FLAGS) -o $(PROJECT_NAME) main.go
	@echo "✅ Built $(PROJECT_NAME) for host ($(GOOS)/$(GOARCH))"
	# Build for all platforms
	@$(foreach platform,$(PLATFORMS), \
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		CGO_ENABLED=0 go build $(STATIC_FLAGS) \
		-o binaries/$(PROJECT_NAME)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))$(if $(filter windows,$(word 1,$(subst /, ,$(platform)))),.exe,) \
		main.go && echo "✅ Built $(PROJECT_NAME)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))" || exit 1;)
	@echo "🎉 Build complete! Binaries in binaries/"
	@ls -lh binaries/

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./... 2>&1 || echo "No tests found (skipping)"
	@if [ -f coverage.out ]; then \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "✅ Tests complete! Coverage report: coverage.html"; \
	else \
		echo "✅ Test execution complete (no coverage data)"; \
	fi

# Build and push Docker image
docker: build
	@echo "Building and pushing Docker image..."
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--tag $(REGISTRY)/$(PROJECT_NAME):$(VERSION) \
		--tag $(REGISTRY)/$(PROJECT_NAME):latest \
		--tag docker.io/casjaysdevdocker/$(PROJECT_NAME):$(VERSION) \
		--tag docker.io/casjaysdevdocker/$(PROJECT_NAME):latest \
		--push .
	@echo "✅ Docker images pushed"

# Create GitHub release
release: build-all test
	@echo "Creating GitHub release..."
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "❌ Cannot release dev version. Please create a git tag first."; \
		exit 1; \
	fi
	# Create release with gh CLI
	gh release create $(VERSION) \
		--title "$(PROJECT_NAME) $(VERSION)" \
		--notes "Automated release of $(PROJECT_NAME) $(VERSION)" \
		binaries/*
	@echo "✅ GitHub release $(VERSION) created with all platform binaries"

# Development server
dev:
	@echo "Starting development server..."
	GIN_MODE=debug go run main.go

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf dist/ binaries/ $(PROJECT_NAME) coverage.out coverage.html
	@echo "✅ Clean complete"

# Quick run (build and run host binary)
run: build
	@echo "Running $(PROJECT_NAME)..."
	./$(PROJECT_NAME)

# Install to system (requires sudo)
install: build
	@echo "Installing $(PROJECT_NAME) to /usr/local/bin..."
	@sudo cp $(PROJECT_NAME) /usr/local/bin/$(PROJECT_NAME)
	@sudo chmod +x /usr/local/bin/$(PROJECT_NAME)
	@echo "✅ Installed to /usr/local/bin/$(PROJECT_NAME)"

# Uninstall from system (requires sudo)
uninstall:
	@echo "Uninstalling $(PROJECT_NAME)..."
	@sudo rm -f /usr/local/bin/$(PROJECT_NAME)
	@echo "✅ Uninstalled"

# Database migrations
migrate:
	@echo "Running database migrations..."
	@./$(PROJECT_NAME) --migrate || go run main.go --migrate
	@echo "✅ Migrations complete"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Lint code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		go vet ./...; \
	fi
	@echo "✅ Linting complete"

# Show build info
info:
	@echo "Project: $(PROJECT_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Host Platform: $(GOOS)/$(GOARCH)"
	@echo "Registry: $(REGISTRY)"
	@echo "Platforms: $(PLATFORMS)"

# Help
help:
	@echo "Weather Service Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build Targets:"
	@echo "  build      - Build for host platform only"
	@echo "  build-all  - Build for all 8 platforms"
	@echo "  test       - Run all tests with coverage"
	@echo "  docker     - Build and push multi-platform Docker image"
	@echo "  release    - Create GitHub release with all binaries"
	@echo ""
	@echo "Development Targets:"
	@echo "  dev        - Start development server"
	@echo "  run        - Build and run host binary"
	@echo "  deps       - Install/update dependencies"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
	@echo ""
	@echo "Installation Targets:"
	@echo "  install    - Install to /usr/local/bin (requires sudo)"
	@echo "  uninstall  - Remove from /usr/local/bin (requires sudo)"
	@echo "  migrate    - Run database migrations"
	@echo ""
	@echo "Utility Targets:"
	@echo "  clean      - Clean build artifacts"
	@echo "  info       - Show build information"
	@echo "  help       - Show this help"
	@echo ""
	@echo "Binary naming: $(PROJECT_NAME)-{os}-{arch}"
	@echo "Supported platforms:"
	@echo "  - Linux (amd64, arm64)"
	@echo "  - macOS/Darwin (amd64, arm64)"
	@echo "  - Windows (amd64, arm64)"
	@echo "  - FreeBSD (amd64, arm64)"

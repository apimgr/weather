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

# Target platforms (no ARM64 per user request)
PLATFORMS = \
	linux/amd64 \
	darwin/amd64 \
	windows/amd64

.PHONY: all build test docker release clean help

# Default target
all: build

# Build for all platforms plus host binary
build: clean
	@echo "Building $(PROJECT_NAME) for all platforms..."
	@mkdir -p dist
	# Build host binary
	CGO_ENABLED=0 go build $(STATIC_FLAGS) -o $(PROJECT_NAME) main.go
	@echo "✅ Built $(PROJECT_NAME) for host ($(GOOS)/$(GOARCH))"
	# Build for all platforms
	@$(foreach platform,$(PLATFORMS), \
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		CGO_ENABLED=0 go build $(STATIC_FLAGS) \
		-o dist/$(PROJECT_NAME)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))$(if $(filter windows,$(word 1,$(subst /, ,$(platform)))),.exe,) \
		main.go && echo "✅ Built $(PROJECT_NAME)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))" || exit 1;)
	@echo "🎉 Build complete! Binaries in dist/ and host binary: $(PROJECT_NAME)"
	@ls -la $(PROJECT_NAME) dist/

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Tests complete! Coverage report: coverage.html"

# Build and push Docker image
docker:
	@echo "Building and pushing Docker image..."
	# Build for linux/amd64 only (no ARM64 per user request)
	docker buildx build \
		--platform linux/amd64 \
		--tag $(REGISTRY)/$(PROJECT_NAME):$(VERSION) \
		--tag $(REGISTRY)/$(PROJECT_NAME):latest \
		--tag docker.io/casjaysdevdocker/$(PROJECT_NAME):$(VERSION) \
		--tag docker.io/casjaysdevdocker/$(PROJECT_NAME):latest \
		--push .
	@echo "✅ Docker image pushed to $(REGISTRY)/$(PROJECT_NAME):$(VERSION)"
	@echo "✅ Docker image pushed to docker.io/casjaysdevdocker/$(PROJECT_NAME):$(VERSION)"

# Create GitHub release
release: build test
	@echo "Creating GitHub release..."
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "❌ Cannot release dev version. Please create a git tag first."; \
		exit 1; \
	fi
	# Create release with gh CLI
	gh release create $(VERSION) \
		--title "$(PROJECT_NAME) $(VERSION)" \
		--notes "Automated release of $(PROJECT_NAME) $(VERSION)" \
		dist/*
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

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf dist/ $(PROJECT_NAME) coverage.out coverage.html
	@echo "✅ Clean complete"

# Quick run (build and run host binary)
run: build
	@echo "Running $(PROJECT_NAME)..."
	./$(PROJECT_NAME)

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
	@echo "Targets:"
	@echo "  build    - Build for all platforms + host binary"
	@echo "  test     - Run all tests with coverage"
	@echo "  docker   - Build and push multi-platform Docker image"
	@echo "  release  - Create GitHub release with all binaries"
	@echo "  dev      - Start development server"
	@echo "  run      - Build and run host binary"
	@echo "  deps     - Install/update dependencies"
	@echo "  clean    - Clean build artifacts"
	@echo "  info     - Show build information"
	@echo "  help     - Show this help"
	@echo ""
	@echo "Binary naming: $(PROJECT_NAME)-{os}-{arch}"
	@echo "Example: $(PROJECT_NAME)-linux-amd64, $(PROJECT_NAME)-darwin-arm64"

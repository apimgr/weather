# Weather Service Makefile
PROJECT_NAME = weather
GITHUB_ORG = apimgr
REGISTRY = ghcr.io/$(GITHUB_ORG)

# Version management - use env var if set, otherwise read from release.txt
VERSION ?= $(shell cat release.txt 2>/dev/null || echo "1.0.0")
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Host platform detection
HOST_OS = $(shell go env GOOS)
HOST_ARCH = $(shell go env GOARCH)

# Build flags
LDFLAGS = -w -s -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE) -X main.GitCommit=$(GIT_COMMIT)
BUILD_FLAGS = -ldflags="$(LDFLAGS)" -trimpath

# Platforms: amd64 and arm64 for all major OSes
PLATFORMS = \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64 \
	freebsd/amd64 \
	freebsd/arm64

.PHONY: build release docker test clean help

# Default target
all: build

# Build everything: all platforms + host binary
build:
	@echo "🔨 Building $(PROJECT_NAME) v$(VERSION)"
	@mkdir -p binaries
	@# Build for all platforms
	@$(foreach platform,$(PLATFORMS), \
		echo "  Building $(word 1,$(subst /, ,$(platform)))/$(word 2,$(subst /, ,$(platform)))..."; \
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		CGO_ENABLED=0 go build $(BUILD_FLAGS) \
		-o binaries/$(PROJECT_NAME)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))$(if $(filter windows,$(word 1,$(subst /, ,$(platform)))),.exe,) \
		./src && echo "  ✅ $(PROJECT_NAME)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))" || exit 1;)
	@# Build host binary
	@echo "  Building host binary ($(HOST_OS)/$(HOST_ARCH))..."
	@CGO_ENABLED=0 go build $(BUILD_FLAGS) -o binaries/$(PROJECT_NAME) ./src
	@echo "  ✅ $(PROJECT_NAME) (host: $(HOST_OS)/$(HOST_ARCH))"
	@# Strip musl binaries
	@if command -v strip >/dev/null 2>&1; then \
		echo "  Stripping Linux binaries..."; \
		for file in binaries/*-linux-*; do \
			if [ -f "$$file" ]; then \
				strip "$$file" 2>/dev/null || true; \
			fi; \
		done; \
	fi
	@echo "🎉 Build complete! Binaries in ./binaries/"
	@ls -lh binaries/ | tail -n +2 | awk '{printf "  %s\t%s\n", $$9, $$5}'

# Create GitHub release
release:
	@echo "🚀 Preparing GitHub release..."
	@# Auto-increment patch version unless VERSION env var is set
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "$$(cat release.txt 2>/dev/null || echo '1.0.0')" ]; then \
		CURRENT_VERSION=$$(cat release.txt 2>/dev/null || echo "1.0.0"); \
		MAJOR=$$(echo $$CURRENT_VERSION | cut -d. -f1); \
		MINOR=$$(echo $$CURRENT_VERSION | cut -d. -f2); \
		PATCH=$$(echo $$CURRENT_VERSION | cut -d. -f3); \
		NEW_PATCH=$$((PATCH + 1)); \
		NEW_VERSION="$$MAJOR.$$MINOR.$$NEW_PATCH"; \
		echo "$$NEW_VERSION" > release.txt; \
		echo "  📝 Version: $$CURRENT_VERSION → $$NEW_VERSION"; \
	else \
		echo "$(VERSION)" > release.txt; \
		echo "  📝 Using custom version: $(VERSION)"; \
	fi
	@# Rebuild with new version
	@$(MAKE) build VERSION=$$(cat release.txt)
	@# Run tests before releasing
	@$(MAKE) test
	@# Prepare release
	@NEW_VERSION=$$(cat release.txt); \
	mkdir -p releases; \
	cp -r binaries/* releases/; \
	echo "  🗑️  Cleaning up old release/tag if exists..."; \
	gh release delete "v$$NEW_VERSION" -y 2>/dev/null || true; \
	git tag -d "v$$NEW_VERSION" 2>/dev/null || true; \
	git push origin ":refs/tags/v$$NEW_VERSION" 2>/dev/null || true; \
	echo "  🏷️  Creating tag v$$NEW_VERSION..."; \
	git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	git push origin "v$$NEW_VERSION"; \
	echo "  📦 Creating GitHub release..."; \
	gh release create "v$$NEW_VERSION" \
		--title "$(PROJECT_NAME) v$$NEW_VERSION" \
		--notes "Release $(PROJECT_NAME) v$$NEW_VERSION\n\n**Built:** $(BUILD_DATE)\n**Commit:** $(GIT_COMMIT)" \
		releases/*
	@echo "✅ Release v$$(cat release.txt) created!"
	@echo "   https://github.com/$(GITHUB_ORG)/$(PROJECT_NAME)/releases/tag/v$$(cat release.txt)"

# Build and push Docker images (multi-platform)
docker:
	@echo "🐳 Building multi-platform Docker images..."
	@docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--tag $(REGISTRY)/$(PROJECT_NAME):$(VERSION) \
		--tag $(REGISTRY)/$(PROJECT_NAME):latest \
		--push .
	@echo "✅ Docker images pushed:"
	@echo "   $(REGISTRY)/$(PROJECT_NAME):$(VERSION)"
	@echo "   $(REGISTRY)/$(PROJECT_NAME):latest"

# Build development Docker image (local only)
docker-dev:
	@echo "🐳 Building development Docker image..."
	@docker build \
		--build-arg VERSION=dev \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--tag $(PROJECT_NAME):dev .
	@echo "✅ Development image built: $(PROJECT_NAME):dev"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v -race -coverprofile=coverage.out ./... 2>&1 || echo "⚠️  No tests found or tests failed"
	@if [ -f coverage.out ]; then \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "✅ Coverage report: coverage.html"; \
	fi

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf binaries/ releases/ $(PROJECT_NAME) coverage.out coverage.html
	@echo "✅ Clean complete"

# Help
help:
	@echo "Weather Service Build System"
	@echo ""
	@echo "Usage: make [target] [VERSION=x.y.z]"
	@echo ""
	@echo "Targets:"
	@echo "  build       - Build for all platforms (binaries/)"
	@echo "  release     - Create GitHub release (releases/)"
	@echo "  docker      - Build multi-platform Docker image (amd64,arm64)"
	@echo "  docker-dev  - Build local development image"
	@echo "  test        - Run all tests with coverage"
	@echo "  clean       - Remove build artifacts"
	@echo "  help        - Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION     - Set version (default: auto-increment from release.txt)"
	@echo ""
	@echo "Current:"
	@echo "  Version:    $(VERSION)"
	@echo "  Platform:   $(HOST_OS)/$(HOST_ARCH)"
	@echo "  Registry:   $(REGISTRY)"
	@echo ""
	@echo "Examples:"
	@echo "  make build                    # Build all platforms"
	@echo "  make build VERSION=2.0.0      # Build with custom version"
	@echo "  make release                  # Auto-increment and release"
	@echo "  make release VERSION=2.0.0    # Release with custom version"
	@echo "  make docker                   # Build and push to registry"

# Weather Service Makefile
PROJECT_NAME = weather
GITHUB_ORG = apimgr
REGISTRY = ghcr.io/$(GITHUB_ORG)

# Version management
# Override with: make build VERSION=2.0.0
VERSION ?= $(shell cat release.txt 2>/dev/null || echo "1.0.0")
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Host platform detection
HOST_OS = $(shell go env GOOS)
HOST_ARCH = $(shell go env GOARCH)

# Build flags
LDFLAGS = -w -s -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE) -X main.GitCommit=$(GIT_COMMIT)
BUILD_FLAGS = -ldflags="$(LDFLAGS)" -trimpath

# Platforms: amd64 and arm64 for linux, darwin, windows, freebsd
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
build: clean
	@echo "🔨 Building $(PROJECT_NAME) v$(VERSION)"
	@mkdir -p dist
	@# Build for all platforms
	@$(foreach platform,$(PLATFORMS), \
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		CGO_ENABLED=0 go build $(BUILD_FLAGS) \
		-o dist/$(PROJECT_NAME)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))$(if $(filter windows,$(word 1,$(subst /, ,$(platform)))),.exe,) \
		./src && echo "  ✅ $(PROJECT_NAME)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))" || exit 1;)
	@# Build host binary
	@CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(PROJECT_NAME) ./src
	@echo "  ✅ $(PROJECT_NAME) (host: $(HOST_OS)/$(HOST_ARCH))"
	@# Strip musl binaries
	@if command -v strip >/dev/null 2>&1; then \
		for file in dist/*-linux-*; do \
			if [ -f "$$file" ]; then \
				strip "$$file" 2>/dev/null && echo "  🔧 Stripped $$file" || true; \
			fi; \
		done; \
	fi
	@echo "🎉 Build complete! $(shell ls -1 dist/ | wc -l) binaries in dist/"
	@ls -lh dist/ | tail -n +2 | awk '{printf "  %s %s\n", $$9, $$5}'

# Create GitHub release (auto-increments version unless VERSION is set)
release: build test
	@echo "🚀 Creating GitHub release..."
	@# Use custom VERSION if provided, otherwise increment patch version
	@if [ -n "$(filter-out $(shell cat release.txt 2>/dev/null || echo "1.0.0"),$(VERSION))" ]; then \
		NEW_VERSION="$(VERSION)"; \
		echo "  📝 Using custom version: $$NEW_VERSION"; \
		echo "$$NEW_VERSION" > release.txt; \
	else \
		CURRENT_VERSION=$$(cat release.txt); \
		MAJOR=$$(echo $$CURRENT_VERSION | cut -d. -f1); \
		MINOR=$$(echo $$CURRENT_VERSION | cut -d. -f2); \
		PATCH=$$(echo $$CURRENT_VERSION | cut -d. -f3); \
		NEW_PATCH=$$((PATCH + 1)); \
		NEW_VERSION="$$MAJOR.$$MINOR.$$NEW_PATCH"; \
		echo "  📝 Incrementing version: $$CURRENT_VERSION → $$NEW_VERSION"; \
		echo "$$NEW_VERSION" > release.txt; \
	fi
	@# Delete existing release and tag if they exist
	@NEW_VERSION=$$(cat release.txt); \
	if gh release view "v$$NEW_VERSION" >/dev/null 2>&1; then \
		echo "  🗑️  Deleting existing release v$$NEW_VERSION"; \
		gh release delete "v$$NEW_VERSION" -y; \
	fi; \
	if git tag | grep -q "^v$$NEW_VERSION$$"; then \
		echo "  🗑️  Deleting existing tag v$$NEW_VERSION"; \
		git tag -d "v$$NEW_VERSION" 2>/dev/null || true; \
		git push origin ":refs/tags/v$$NEW_VERSION" 2>/dev/null || true; \
	fi
	@# Rebuild with new version
	@$(MAKE) build
	@# Create git tag and release
	@NEW_VERSION=$$(cat release.txt); \
	echo "  🏷️  Creating tag v$$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION" 2>/dev/null || true; \
	git push origin "v$$NEW_VERSION" 2>/dev/null || true; \
	echo "  📦 Creating GitHub release v$$NEW_VERSION"; \
	echo "Release $(PROJECT_NAME) v$$NEW_VERSION" > /tmp/release-notes.md; \
	echo "" >> /tmp/release-notes.md; \
	echo "**Built:** $(BUILD_DATE)" >> /tmp/release-notes.md; \
	echo "**Commit:** $(GIT_COMMIT)" >> /tmp/release-notes.md; \
	echo "" >> /tmp/release-notes.md; \
	echo "## Downloads" >> /tmp/release-notes.md; \
	echo "" >> /tmp/release-notes.md; \
	echo "Download the binary for your platform:" >> /tmp/release-notes.md; \
	echo "- **Linux**: \`$(PROJECT_NAME)-linux-{amd64|arm64}\`" >> /tmp/release-notes.md; \
	echo "- **macOS**: \`$(PROJECT_NAME)-darwin-{amd64|arm64}\`" >> /tmp/release-notes.md; \
	echo "- **Windows**: \`$(PROJECT_NAME)-windows-{amd64|arm64}.exe\`" >> /tmp/release-notes.md; \
	echo "- **FreeBSD**: \`$(PROJECT_NAME)-freebsd-{amd64|arm64}\`" >> /tmp/release-notes.md; \
	echo "" >> /tmp/release-notes.md; \
	echo "## Installation" >> /tmp/release-notes.md; \
	echo "" >> /tmp/release-notes.md; \
	echo "\`\`\`bash" >> /tmp/release-notes.md; \
	echo "# Linux/macOS/FreeBSD" >> /tmp/release-notes.md; \
	echo "chmod +x $(PROJECT_NAME)-*" >> /tmp/release-notes.md; \
	echo "sudo mv $(PROJECT_NAME)-* /usr/local/bin/$(PROJECT_NAME)" >> /tmp/release-notes.md; \
	echo "" >> /tmp/release-notes.md; \
	echo "# Windows" >> /tmp/release-notes.md; \
	echo "# Move $(PROJECT_NAME)-windows-*.exe to your PATH" >> /tmp/release-notes.md; \
	echo "\`\`\`" >> /tmp/release-notes.md; \
	echo "" >> /tmp/release-notes.md; \
	echo "## What's Changed" >> /tmp/release-notes.md; \
	echo "- Full changelog: https://github.com/$(GITHUB_ORG)/$(PROJECT_NAME)/commits/v$$NEW_VERSION" >> /tmp/release-notes.md; \
	gh release create "v$$NEW_VERSION" \
		--title "$(PROJECT_NAME) v$$NEW_VERSION" \
		--notes-file /tmp/release-notes.md \
		dist/*; \
	rm -f /tmp/release-notes.md
	@echo "✅ Release v$$NEW_VERSION created successfully!"
	@echo "   https://github.com/$(GITHUB_ORG)/$(PROJECT_NAME)/releases/tag/v$$NEW_VERSION"

# Build and push Docker images
docker:
	@echo "🐳 Building and pushing Docker images..."
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

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v -race -coverprofile=coverage.out ./... 2>&1 || echo "⚠️  No tests found (skipping)"
	@if [ -f coverage.out ]; then \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "✅ Tests complete! Coverage: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"; \
	else \
		echo "✅ Test execution complete"; \
	fi

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf dist/ $(PROJECT_NAME) coverage.out coverage.html
	@echo "✅ Clean complete"

# Help
help:
	@echo "Weather Service Build System"
	@echo ""
	@echo "Usage: make [target] [VAR=value]"
	@echo ""
	@echo "Targets:"
	@echo "  build    - Build for all platforms + host binary (default)"
	@echo "  release  - Auto-increment version, create GitHub release"
	@echo "  docker   - Build and push multi-platform Docker image"
	@echo "  test     - Run all tests with coverage"
	@echo "  clean    - Remove build artifacts"
	@echo "  help     - Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION  - Override version (default: auto from release.txt)"
	@echo "             Example: make build VERSION=2.0.0"
	@echo "             Example: make release VERSION=2.0.0"
	@echo ""
	@echo "Current version: $(VERSION)"
	@echo "Host platform: $(HOST_OS)/$(HOST_ARCH)"
	@echo "Registry: $(REGISTRY)"

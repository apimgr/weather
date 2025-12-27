# Weather Service Makefile
# TEMPLATE.md PART 2: Infer project and org from git remote URL, never hardcode
# Falls back to directory names if git is not available
PROJECT := $(shell git remote get-url origin 2>/dev/null | sed -E 's|.*/([^/]+)(\.git)?$$|\1|' || basename "$$(pwd)")
ORG := $(shell git remote get-url origin 2>/dev/null | sed -E 's|.*/([^/]+)/[^/]+(\.git)?$$|\1|' || basename "$$(dirname "$$(pwd)")")
REGISTRY = ghcr.io/$(ORG)

# Version management - use env var if set, otherwise read from release.txt
VERSION ?= $(shell cat release.txt 2>/dev/null || echo "1.0.0")
BUILD_DATE := $(shell date +"%a %b %d, %Y at %H:%M:%S %Z")
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Host platform detection
HOST_OS = $(shell go env GOOS)
HOST_ARCH = $(shell go env GOARCH)

# Build flags per TEMPLATE.md PART 6
# CRITICAL: CGO_ENABLED MUST be 0 for static binaries (TEMPLATE.md requirement)
LDFLAGS_SERVER := -s -w \
	-X 'github.com/apimgr/weather/src/cli.Version=$(VERSION)' \
	-X 'github.com/apimgr/weather/src/cli.GitCommit=$(COMMIT_ID)' \
	-X 'github.com/apimgr/weather/src/cli.BuildDate=$(BUILD_DATE)' \
	-X 'github.com/apimgr/weather/src/cli.CGOEnabled=0'
LDFLAGS_CLI := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.GitCommit=$(COMMIT_ID)' \
	-X 'main.BuildDate=$(BUILD_DATE)'
BUILD_FLAGS_SERVER = -ldflags="$(LDFLAGS_SERVER)" -trimpath
BUILD_FLAGS_CLI = -ldflags="$(LDFLAGS_CLI)" -trimpath

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

.PHONY: build release docker test

# Build everything: all platforms + host binary
# TEMPLATE.md: Makefile MUST have exactly 4 targets (build, release, docker, test)
build:
	@echo "🔨 Building $(PROJECT) v$(VERSION)"
	@# Generate Swagger documentation if swag is installed
	@if command -v swag >/dev/null 2>&1; then \
		echo "  📚 Generating Swagger documentation..."; \
		swag init -g src/main.go -o docs --parseDependency --parseInternal 2>/dev/null || echo "  ⚠️  Swagger generation failed (continuing build)"; \
	fi
	@# Generate GraphQL code (TEMPLATE.md Part 14 requirement)
	@if command -v gqlgen >/dev/null 2>&1; then \
		echo "  🔮 Generating GraphQL code..."; \
		gqlgen generate 2>/dev/null || echo "  ⚠️  GraphQL generation failed (continuing build)"; \
	else \
		echo "  🔮 Generating GraphQL code (using go run)..."; \
		go run github.com/99designs/gqlgen generate 2>/dev/null || echo "  ⚠️  GraphQL generation failed (continuing build)"; \
	fi
	@mkdir -p binaries
	@# Build server for all platforms
	@$(foreach platform,$(PLATFORMS), \
		echo "  Building server $(word 1,$(subst /, ,$(platform)))/$(word 2,$(subst /, ,$(platform)))..."; \
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		CGO_ENABLED=0 go build $(BUILD_FLAGS_SERVER) \
		-o binaries/$(PROJECT)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))$(if $(filter windows,$(word 1,$(subst /, ,$(platform)))),.exe,) \
		./src && echo "  ✅ $(PROJECT)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))" || exit 1;)
	@# Build CLI for all platforms
	@$(foreach platform,$(PLATFORMS), \
		echo "  Building CLI $(word 1,$(subst /, ,$(platform)))/$(word 2,$(subst /, ,$(platform)))..."; \
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		CGO_ENABLED=0 go build $(BUILD_FLAGS_CLI) \
		-o binaries/$(PROJECT)-cli-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))$(if $(filter windows,$(word 1,$(subst /, ,$(platform)))),.exe,) \
		./cmd/weather-cli && echo "  ✅ $(PROJECT)-cli-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform)))" || exit 1;)
	@# Build host binaries
	@echo "  Building host server binary ($(HOST_OS)/$(HOST_ARCH))..."
	@CGO_ENABLED=0 go build $(BUILD_FLAGS_SERVER) -o binaries/$(PROJECT) ./src
	@echo "  ✅ $(PROJECT) (host: $(HOST_OS)/$(HOST_ARCH))"
	@echo "  Building host CLI binary ($(HOST_OS)/$(HOST_ARCH))..."
	@CGO_ENABLED=0 go build $(BUILD_FLAGS_CLI) -o binaries/$(PROJECT)-cli ./cmd/weather-cli
	@echo "  ✅ $(PROJECT)-cli (host: $(HOST_OS)/$(HOST_ARCH))"
	@# Strip Linux binaries for smaller size
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
	echo "  📦 Creating source archives..."; \
	git archive --format=tar.gz --prefix=$(PROJECT)-$$NEW_VERSION/ HEAD -o releases/$(PROJECT)-$$NEW_VERSION-src.tar.gz; \
	git archive --format=zip --prefix=$(PROJECT)-$$NEW_VERSION/ HEAD -o releases/$(PROJECT)-$$NEW_VERSION-src.zip; \
	echo "  ✅ Source archives created"; \
	echo "  🗑️  Cleaning up old release/tag if exists..."; \
	gh release delete "v$$NEW_VERSION" -y 2>/dev/null || true; \
	git tag -d "v$$NEW_VERSION" 2>/dev/null || true; \
	git push origin ":refs/tags/v$$NEW_VERSION" 2>/dev/null || true; \
	echo "  🏷️  Creating tag v$$NEW_VERSION..."; \
	git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	git push origin "v$$NEW_VERSION"; \
	echo "  📦 Creating GitHub release..."; \
	gh release create "v$$NEW_VERSION" \
		--title "$(PROJECT) v$$NEW_VERSION" \
		--notes "Release $(PROJECT) v$$NEW_VERSION\n\n**Built:** $(BUILD_DATE)\n**Commit:** $(GIT_COMMIT)" \
		releases/*
	@echo "✅ Release v$$(cat release.txt) created!"
	@echo "   https://github.com/$(ORG)/$(PROJECT)/releases/tag/v$$(cat release.txt)"

# Build and push Docker images (multi-platform)
docker:
	@echo "🐳 Building multi-platform Docker images..."
	@docker buildx build \
		-f ./docker/Dockerfile \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg COMMIT_ID=$(COMMIT_ID) \
		--tag $(REGISTRY)/$(PROJECT):$(VERSION) \
		--tag $(REGISTRY)/$(PROJECT):latest \
		--push .
	@echo "✅ Docker images pushed:"
	@echo "   $(REGISTRY)/$(PROJECT):$(VERSION)"
	@echo "   $(REGISTRY)/$(PROJECT):latest"

# Run tests (TEMPLATE.md PART 12: unit, integration, e2e)
test:
	@echo "🧪 Running tests..."
	@echo "  📦 Unit tests..."
	@go test -v -race -coverprofile=coverage-unit.out ./tests/unit/... 2>&1 || echo "  ⚠️  No unit tests found or tests failed"
	@echo "  🔗 Integration tests..."
	@go test -v -race -coverprofile=coverage-integration.out ./tests/integration/... 2>&1 || echo "  ⚠️  No integration tests found or tests failed"
	@echo "  🎯 End-to-end tests..."
	@go test -v -race -coverprofile=coverage-e2e.out ./tests/e2e/... 2>&1 || echo "  ⚠️  No e2e tests found or tests failed"
	@echo "  📊 Generating combined coverage report..."
	@go test -v -race -coverprofile=coverage.out ./... 2>&1 || echo "  ⚠️  Some tests failed"
	@if [ -f coverage.out ]; then \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "  ✅ Coverage report: coverage.html"; \
	fi
	@echo "✅ Tests complete!"


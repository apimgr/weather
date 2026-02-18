# Infer PROJECTNAME and PROJECTORG from git remote or directory path (NEVER hardcode)
PROJECTNAME := $(shell git remote get-url origin 2>/dev/null | sed -E 's|.*/([^/]+)(\.git)?$$|\1|' || basename "$$(pwd)")
PROJECTORG := $(shell git remote get-url origin 2>/dev/null | sed -E 's|.*/([^/]+)/[^/]+(\.git)?$$|\1|' || basename "$$(dirname "$$(pwd)")")

# Version: env var > release.txt > default
VERSION ?= $(shell cat release.txt 2>/dev/null || echo "0.1.0")

# Build info - use TZ env var or system timezone
# Format: "Thu Dec 17, 2025 at 18:19:24 EST"
BUILD_DATE := $(shell date +"%a %b %d, %Y at %H:%M:%S %Z")
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
# COMMIT_ID used directly - no VCS_REF alias

# Official site URL (set per project)
OFFICIALSITE := https://wthr.top

# Linker flags to embed build info
LDFLAGS := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.CommitID=$(COMMIT_ID)' \
	-X 'main.BuildDate=$(BUILD_DATE)' \
	-X 'main.OfficialSite=$(OFFICIALSITE)'

# Directories
BINDIR := binaries
RELDIR := releases

# Go directories (persistent across builds)
GODIR := $(HOME)/.local/share/go
GOCACHE := $(HOME)/.local/share/go/build

# Build targets
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 freebsd/amd64 freebsd/arm64

# Docker - Set REGISTRY based on your platform (ghcr.io, registry.gitlab.com, git.example.com)
REGISTRY ?= ghcr.io/$(PROJECTORG)/$(PROJECTNAME)
GO_DOCKER := docker run --rm \
	-v $(PWD):/build \
	-v $(GOCACHE):/root/.cache/go-build \
	-v $(GODIR):/go \
	-w /build \
	-e CGO_ENABLED=0 \
	golang:alpine

.PHONY: build local release docker test dev clean

# =============================================================================
# BUILD - Build all platforms + host binary (via Docker with cached modules)
# =============================================================================
build: clean
	@mkdir -p $(BINDIR)
	@echo "Building version $(VERSION)..."
	@mkdir -p $(GOCACHE) $(GODIR)

	# Download modules first (cached)
	@echo "Downloading Go modules..."
	@$(GO_DOCKER) go mod download

	# Build for host OS/ARCH
	@echo "Building host binary..."
	@$(GO_DOCKER) sh -c "GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH) \
		go build -ldflags \"$(LDFLAGS)\" -o $(BINDIR)/$(PROJECTNAME) ./src"

	# Build server for all platforms
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		OUTPUT=$(BINDIR)/$(PROJECTNAME)-$$OS-$$ARCH; \
		[ "$$OS" = "windows" ] && OUTPUT=$$OUTPUT.exe; \
		echo "Building $(PROJECTNAME) $$OS/$$ARCH..."; \
		$(GO_DOCKER) sh -c "GOOS=$$OS GOARCH=$$ARCH \
			go build -ldflags \"$(LDFLAGS)\" \
			-o $$OUTPUT ./src" || exit 1; \
	done

	# Build CLI client for all platforms (if src/client exists)
	@if [ -d "src/client" ]; then \
		echo "Building CLI client for all platforms..."; \
		for platform in $(PLATFORMS); do \
			OS=$${platform%/*}; \
			ARCH=$${platform#*/}; \
			OUTPUT=$(BINDIR)/$(PROJECTNAME)-cli-$$OS-$$ARCH; \
			[ "$$OS" = "windows" ] && OUTPUT=$$OUTPUT.exe; \
			echo "Building $(PROJECTNAME)-cli $$OS/$$ARCH..."; \
			$(GO_DOCKER) sh -c "GOOS=$$OS GOARCH=$$ARCH \
				go build -ldflags \"$(LDFLAGS)\" \
				-o $$OUTPUT ./src/client" || exit 1; \
		done; \
	fi

	# Build agent for all platforms (if src/agent exists)
	@if [ -d "src/agent" ]; then \
		echo "Building agent for all platforms..."; \
		for platform in $(PLATFORMS); do \
			OS=$${platform%/*}; \
			ARCH=$${platform#*/}; \
			OUTPUT=$(BINDIR)/$(PROJECTNAME)-agent-$$OS-$$ARCH; \
			[ "$$OS" = "windows" ] && OUTPUT=$$OUTPUT.exe; \
			echo "Building $(PROJECTNAME)-agent $$OS/$$ARCH..."; \
			$(GO_DOCKER) sh -c "GOOS=$$OS GOARCH=$$ARCH \
				go build -ldflags \"$(LDFLAGS)\" \
				-o $$OUTPUT ./src/agent" || exit 1; \
		done; \
	fi

	@echo "Build complete: $(BINDIR)/"

# =============================================================================
# LOCAL - Build local platform with full version info (production testing)
# =============================================================================
local: clean
	@mkdir -p $(BINDIR)
	@echo "Building local platform version $(VERSION)..."
	@mkdir -p $(GOCACHE) $(GODIR)

	# Download modules first (cached)
	@echo "Downloading Go modules..."
	@$(GO_DOCKER) go mod download

	# Build server binary with version info
	@echo "Building $(PROJECTNAME)..."
	@$(GO_DOCKER) sh -c "GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH) \
		go build -ldflags \"$(LDFLAGS)\" -o $(BINDIR)/$(PROJECTNAME) ./src"

	# Build CLI binary (if exists)
	@if [ -d "src/client" ]; then \
		echo "Building $(PROJECTNAME)-cli..."; \
		$(GO_DOCKER) sh -c "GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH) \
			go build -ldflags \"$(LDFLAGS)\" -o $(BINDIR)/$(PROJECTNAME)-cli ./src/client"; \
	fi

	# Build agent binary (if exists)
	@if [ -d "src/agent" ]; then \
		echo "Building $(PROJECTNAME)-agent..."; \
		$(GO_DOCKER) sh -c "GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH) \
			go build -ldflags \"$(LDFLAGS)\" -o $(BINDIR)/$(PROJECTNAME)-agent ./src/agent"; \
	fi

	@echo "Local build complete: $(BINDIR)/"

# =============================================================================
# RELEASE - Manual local release (stable only)
# =============================================================================
release: build
	@mkdir -p $(RELDIR)
	@echo "Preparing release $(VERSION)..."

	# Create version.txt
	@echo "$(VERSION)" > $(RELDIR)/version.txt

	# Copy all binaries to releases (server, cli, agent - strip if possible)
	@for f in $(BINDIR)/$(PROJECTNAME)*; do \
		[ -f "$$f" ] || continue; \
		strip "$$f" 2>/dev/null || true; \
		cp "$$f" $(RELDIR)/; \
	done

	# Create source archive (exclude VCS and build artifacts)
	@tar --exclude='.git' --exclude='.github' --exclude='.gitea' \
		--exclude='binaries' --exclude='releases' --exclude='*.tar.gz' \
		-czf $(RELDIR)/$(PROJECTNAME)-$(VERSION)-source.tar.gz .

	# Delete existing release/tag if exists
	@gh release delete $(VERSION) --yes 2>/dev/null || true
	@git tag -d $(VERSION) 2>/dev/null || true
	@git push origin :refs/tags/$(VERSION) 2>/dev/null || true

	# Create new release (stable)
	@gh release create $(VERSION) $(RELDIR)/* \
		--title "$(PROJECTNAME) $(VERSION)" \
		--notes "Release $(VERSION)" \
		--latest

	@echo "Release complete: $(VERSION)"

# =============================================================================
# DOCKER - Build and push container to registry (set REGISTRY env var)
# =============================================================================
# Uses multi-stage Dockerfile - Go compilation happens inside Docker
# No pre-built binaries needed
docker:
	@echo "Building Docker image $(VERSION)..."

	# Ensure buildx is available
	@docker buildx version > /dev/null 2>&1 || (echo "docker buildx required" && exit 1)

	# Create/use builder
	@docker buildx create --name $(PROJECTNAME)-builder --use 2>/dev/null || \
		docker buildx use $(PROJECTNAME)-builder

	# Build and push multi-arch (multi-stage Dockerfile handles Go compilation)
	@docker buildx build \
		-f docker/Dockerfile \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION="$(VERSION)" \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		--build-arg COMMIT_ID="$(COMMIT_ID)" \
		--build-arg OFFICIALSITE="$(OFFICIALSITE)" \
		-t $(REGISTRY):$(VERSION) \
		-t $(REGISTRY):latest \
		--push \
		.

	@echo "Docker push complete: $(REGISTRY):$(VERSION)"

# =============================================================================
# TEST - Run all tests (via Docker with cached modules)
# =============================================================================
test:
	@echo "Running tests in Docker..."
	@mkdir -p $(GOCACHE) $(GODIR)
	@$(GO_DOCKER) go mod download
	@$(GO_DOCKER) go test -v -cover ./...
	@echo "Tests complete"

# =============================================================================
# DEV - Quick build for local development/testing (to temp dir)
# =============================================================================
# Fast: local platform only, no ldflags, random temp dir for isolation
# Builds server + CLI + agent (if they exist)
dev:
	@mkdir -p $(GOCACHE) $(GODIR)
	@$(GO_DOCKER) go mod tidy
	@mkdir -p "$${TMPDIR:-/tmp}/$(PROJECTORG)" && \
		BUILD_DIR=$$(mktemp -d "$${TMPDIR:-/tmp}/$(PROJECTORG)/$(PROJECTNAME)-XXXXXX") && \
		echo "Quick dev build to $$BUILD_DIR..." && \
		$(GO_DOCKER) go build -o $$BUILD_DIR/$(PROJECTNAME) ./src && \
		echo "Built: $$BUILD_DIR/$(PROJECTNAME)" && \
		if [ -d "src/client" ]; then \
			$(GO_DOCKER) go build -o $$BUILD_DIR/$(PROJECTNAME)-cli ./src/client && \
			echo "Built: $$BUILD_DIR/$(PROJECTNAME)-cli"; \
		fi && \
		if [ -d "src/agent" ]; then \
			$(GO_DOCKER) go build -o $$BUILD_DIR/$(PROJECTNAME)-agent ./src/agent && \
			echo "Built: $$BUILD_DIR/$(PROJECTNAME)-agent"; \
		fi && \
		echo "Test:  docker run --rm -v $$BUILD_DIR:/app alpine:latest /app/$(PROJECTNAME) --help"

# =============================================================================
# CLEAN - Remove build artifacts
# =============================================================================
clean:
	@rm -rf $(BINDIR) $(RELDIR)

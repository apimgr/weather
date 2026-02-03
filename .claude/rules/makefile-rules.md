# Makefile Rules (PART 26)

⚠️ **Makefile is for LOCAL DEVELOPMENT ONLY. NEVER use in CI/CD.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use Makefile in CI/CD workflows
- ❌ Run `go` commands directly on local machine
- ❌ Hardcode version numbers
- ❌ Skip CGO_ENABLED=0

## REQUIRED - ALWAYS DO
- ✅ Makefile for local development only
- ✅ CI/CD uses explicit commands with all env vars
- ✅ Container-only builds (golang:alpine)
- ✅ CGO_ENABLED=0 for all builds
- ✅ LDFLAGS with Version, CommitID, BuildDate, OfficialSite

## MAKEFILE TARGETS
| Target | Purpose | Output |
|--------|---------|--------|
| make dev | Quick development build | /tmp/apimgr/weather-XXXXXX/ |
| make local | Production test build | binaries/ |
| make build | Full release (all 8 platforms) | binaries/ |
| make test | Run unit tests | Coverage report |
| make clean | Remove build artifacts | - |
| make docker | Build Docker image | - |

## BUILD VARIABLES
```make
PROJECTNAME := weather
PROJECTORG := apimgr
VERSION := $(shell cat release.txt)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
OFFICIALSITE := https://wthr.top
LDFLAGS := -s -w \
  -X 'main.Version=$(VERSION)' \
  -X 'main.CommitID=$(COMMIT)' \
  -X 'main.BuildDate=$(DATE)' \
  -X 'main.OfficialSite=$(OFFICIALSITE)'
```

## CONTAINER BUILD
```bash
# Makefile uses Docker internally
docker run --rm \
  -v "$PWD:/src" \
  -w /src \
  -e CGO_ENABLED=0 \
  golang:alpine \
  go build -ldflags "$(LDFLAGS)" -o binary ./src
```

## CI/CD DIFFERENCE
```yaml
# CI/CD: Explicit commands, no Makefile
- name: Build
  env:
    CGO_ENABLED: 0
    GOOS: linux
    GOARCH: amd64
  run: |
    go build -ldflags "-s -w -X 'main.Version=$VERSION'" -o weather-linux-amd64 ./src
```

---
**Full details: AI.md PART 26**

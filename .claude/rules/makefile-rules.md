# Makefile Rules (PART 26)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use Makefile in CI/CD (explicit commands only)
- ❌ Run `go` commands directly on local machine
- ❌ Skip any of the 8 platform builds
- ❌ Use CGO in builds (CGO_ENABLED=0 always)
- ❌ Hardcode Go version

## CRITICAL - ALWAYS DO
- ✅ Makefile is for LOCAL development ONLY
- ✅ All builds use Docker (`golang:alpine`)
- ✅ CI/CD uses explicit commands with all env vars
- ✅ Support GODIR/GOCACHE for faster rebuilds
- ✅ Include all required targets

## MAKEFILE TARGETS
| Target | Purpose | Output |
|--------|---------|--------|
| `make dev` | Quick development build | `${TMPDIR}/${PROJECTORG}/${PROJECTNAME}-XXXXXX/` |
| `make local` | Production test build | `binaries/` (with version) |
| `make build` | Full release build | `binaries/` (all 8 platforms) |
| `make test` | Run unit tests | Coverage report |
| `make docker` | Build Docker image | Local image |
| `make clean` | Clean build artifacts | - |

## BUILD WORKFLOW
```bash
# 1. Active development
make dev                # Quick build to temp dir

# 2. Debug in Docker (with tools)
BUILD_DIR=$(ls -td ${TMPDIR:-/tmp}/${PROJECTORG}/${PROJECTNAME}-*/ 2>/dev/null | head -1)
docker run --rm -it -v "$BUILD_DIR:/app" alpine:latest sh -c "
  apk add --no-cache curl bash file jq
  /app/weather --help
"

# 3. Unit tests
make test

# 4. Integration tests
./tests/run_tests.sh    # Auto-detects incus/docker

# 5. Production test
make local              # Build with version info
./tests/incus.sh        # Full systemd testing (PREFERRED)

# 6. Full release build
make build              # All 8 platforms
```

## LDFLAGS (VERSION INFO)
```makefile
LDFLAGS := -s -w \
    -X 'main.Version=$(VERSION)' \
    -X 'main.CommitID=$(COMMIT)' \
    -X 'main.BuildDate=$(DATE)' \
    -X 'main.OfficialSite=$(SITE)'
```

## 8 PLATFORM MATRIX
| OS | Architectures |
|----|---------------|
| linux | amd64, arm64 |
| darwin | amd64, arm64 |
| windows | amd64, arm64 |
| freebsd | amd64, arm64 |

## BINARY NAMING
```
weather-linux-amd64
weather-linux-arm64
weather-darwin-amd64
weather-darwin-arm64
weather-windows-amd64.exe
weather-windows-arm64.exe
weather-freebsd-amd64
weather-freebsd-arm64
```

CLI binaries use `weather-cli-{os}-{arch}` format.

---
For complete details, see AI.md PART 26

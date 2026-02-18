# Makefile Rules (PART 26)

⚠️ **Makefile is for LOCAL DEVELOPMENT ONLY. Never used in CI/CD.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use Makefile in CI/CD workflows
- ❌ Run Go commands locally → use make targets
- ❌ Skip CGO_ENABLED=0

## REQUIRED - ALWAYS DO
- ✅ Use `make dev` for quick builds
- ✅ Use `make test` for unit tests
- ✅ Use `make build` for release builds
- ✅ All builds via Docker (golang:alpine)

## MAKEFILE TARGETS
| Target | Purpose | Output |
|--------|---------|--------|
| `make dev` | Quick development build | `${TMPDIR}/apimgr/weather-XXXXXX/` |
| `make local` | Production test build | `binaries/` with version |
| `make build` | Full release (all 8) | `binaries/` all platforms |
| `make test` | Unit tests | Coverage report |
| `make clean` | Remove build artifacts | - |
| `make docker` | Build Docker image | Local image |

## BUILD ENVIRONMENT
```makefile
CGO_ENABLED=0
GOOS={target}
GOARCH={arch}
```

All builds use Docker `golang:alpine` internally.

## WHY NOT IN CI/CD?
- CI/CD needs explicit env vars for reproducibility
- CI/CD has different paths/permissions
- CI/CD needs matrix builds (parallel platforms)
- Makefile abstracts away details needed for debugging

---
**Full details: AI.md PART 26**

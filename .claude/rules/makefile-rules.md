# Makefile Rules

@AI.md PART 26: Makefile (Local Dev Only)

## CRITICAL: Makefile is for LOCAL development ONLY
- NEVER use Makefile in CI/CD
- CI/CD uses explicit commands with env vars

## Local Development Targets
| Command | Purpose | Output |
|---------|---------|--------|
| `make dev` | Quick dev build | temp dir |
| `make host` | Production test | binaries/ |
| `make build` | Full release | binaries/ (8 platforms) |
| `make test` | Unit tests | coverage report |
| `make docker` | Build Docker image | local image |
| `make clean` | Remove build artifacts | - |

## NEVER on Host Machine
- `go build` → use `make dev`
- `go test` → use `make test`
- `go run` → use `make dev` then Docker

## Caching
- GODIR: `~/.local/share/go`
- GOCACHE: `~/.local/share/go/build`

# CI/CD Rules (PART 28)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO

- Never use Makefile in CI/CD -- always use explicit commands
- Never hardcode Go version -- use stable in workflows
- Never include -musl suffix in released binary names
- Never push :dev or :test tags to production registry
- Never use CGO in CI/CD builds -- CGO_ENABLED=0 always
- Never require interactive input in CI/CD

## CRITICAL - ALWAYS DO

- Use explicit commands in CI/CD (never Makefile targets)
- Use go-version: stable in GitHub Actions workflows
- Set CGO_ENABLED=0 in all build environment variables
- Have a Jenkinsfile in the repository root
- Have GitHub Actions workflows for all required checks
- Support both amd64 and arm64 agent labels in Jenkins

## CI/CD vs Local Development

| Task | Local Development | CI/CD |
|------|------------------|-------|
| Build | make build | Explicit go build commands |
| Test | make test | Explicit go test commands |
| Docker | make docker | Explicit docker commands |

NEVER use Makefile targets in CI/CD pipelines.

## Required Environment Variables (All Workflows)

- CGO_ENABLED=0
- GOFLAGS=-mod=readonly

## GitHub Actions Rules

- Use go-version: stable (NEVER hardcode like 1.21)
- Workflows MUST NOT use Makefile targets
- Workflows MUST set CGO_ENABLED=0
- Workflows MUST NOT require interactive input

## Jenkinsfile Requirements

- MUST exist in repository root
- Agent labels: amd64 AND arm64 must be available
- Pipeline stages: build, test, docker (at minimum)

## CI/CD Workflow Requirements

Workflows MUST:
- Build all 8 platform binaries
- Run unit tests with coverage
- Build Docker image
- Verify binary names (no -musl suffix)
- Verify checksums

Workflows MUST NOT:
- Use Makefile
- Hardcode Go version
- Require .env files
- Use CGO

## Binary Naming in CI/CD

| Platform | Server | CLI |
|----------|--------|-----|
| linux/amd64 | weather-linux-amd64 | weather-cli-linux-amd64 |
| linux/arm64 | weather-linux-arm64 | weather-cli-linux-arm64 |
| darwin/amd64 | weather-darwin-amd64 | weather-cli-darwin-amd64 |
| darwin/arm64 | weather-darwin-arm64 | weather-cli-darwin-arm64 |
| windows/amd64 | weather-windows-amd64.exe | weather-cli-windows-amd64.exe |
| windows/arm64 | weather-windows-arm64.exe | weather-cli-windows-arm64.exe |
| freebsd/amd64 | weather-freebsd-amd64 | weather-cli-freebsd-amd64 |
| freebsd/arm64 | weather-freebsd-arm64 | weather-cli-freebsd-arm64 |

## Reference

For complete details, see AI.md PART 28 (lines 36223-39156)

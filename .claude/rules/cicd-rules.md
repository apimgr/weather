# CI/CD Rules (PART 28)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use Makefile in CI/CD (explicit commands only)
- ❌ Hardcode Go version (use `go-version: 'stable'`)
- ❌ Skip any of the 8 platforms
- ❌ Use v prefix in VERSION (strip it: v1.2.3 → 1.2.3)
- ❌ Forget OfficialSite in LDFLAGS

## CRITICAL - ALWAYS DO
- ✅ Explicit commands with all env vars
- ✅ Same logic across GitHub, Gitea, Jenkins
- ✅ Build all 8 platforms (linux/darwin/windows/freebsd × amd64/arm64)
- ✅ Docker builds on EVERY push
- ✅ Proper LDFLAGS: Version, CommitID, BuildDate, OfficialSite

## WORKFLOW FILES
| Platform | Location |
|----------|----------|
| GitHub | `.github/workflows/*.yml` |
| Gitea | `.gitea/workflows/*.yml` |
| GitLab | `.gitlab-ci.yml` |
| Jenkins | `Jenkinsfile` |

## REQUIRED WORKFLOWS
| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `release.yml` | Tag push (v*) | Stable releases |
| `beta.yml` | Branch: beta | Beta releases |
| `daily.yml` | Branch: main, schedule | Daily builds |
| `docker.yml` | Any push | Docker images |

## VERSION FROM TAG
```bash
# Strip 'v' prefix from semver tags only
VERSION=$(echo $TAG | sed 's/^v//')
# v1.2.3 → 1.2.3
# dev → dev (unchanged)
# beta → beta (unchanged)
```

## LDFLAGS
```bash
LDFLAGS="-s -w \
  -X 'main.Version=${VERSION}' \
  -X 'main.CommitID=${COMMIT}' \
  -X 'main.BuildDate=${DATE}' \
  -X 'main.OfficialSite=${SITE}'"
```

## 8 PLATFORM BUILD MATRIX
```yaml
strategy:
  matrix:
    goos: [linux, darwin, windows, freebsd]
    goarch: [amd64, arm64]
```

## DOCKER TAGS
| Trigger | Tags |
|---------|------|
| Any push | `devel`, `{commit}` |
| Beta branch | adds `beta` |
| Release tag | `{version}`, `latest`, `YYMM`, `{commit}` |

## CLI BINARY NAMING
```bash
# Server binary
weather-linux-amd64
weather-darwin-arm64
weather-windows-amd64.exe

# CLI binary (CORRECT)
weather-cli-linux-amd64
weather-cli-darwin-arm64
weather-cli-windows-amd64.exe

# WRONG
weather-linux-amd64-cli
```

## ENVIRONMENT VARIABLES
```yaml
env:
  CGO_ENABLED: 0
  GOOS: ${{ matrix.goos }}
  GOARCH: ${{ matrix.goarch }}
```

---
For complete details, see AI.md PART 28

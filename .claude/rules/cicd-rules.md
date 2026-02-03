# CI/CD Rules (PART 28)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use Makefile in CI/CD workflows
- ❌ Hardcode Go version (use 'stable')
- ❌ Skip any of the 8 platforms
- ❌ Different logic between GitHub/Gitea/Jenkins
- ❌ Skip .exe suffix for Windows
- ❌ Skip OfficialSite in LDFLAGS

## REQUIRED - ALWAYS DO
- ✅ Explicit commands with all env vars
- ✅ go-version: 'stable' (never hardcoded)
- ✅ Build all 8 platforms
- ✅ GitHub/Gitea/Jenkins must match
- ✅ VERSION from tag (strip v prefix for semver)
- ✅ Docker builds on EVERY push

## LDFLAGS (REQUIRED)
```bash
-s -w \
-X 'main.Version=$VERSION' \
-X 'main.CommitID=$COMMIT' \
-X 'main.BuildDate=$DATE' \
-X 'main.OfficialSite=$OFFICIALSITE'
```

## 8 PLATFORM BUILD MATRIX
```yaml
strategy:
  matrix:
    include:
      - goos: linux
        goarch: amd64
      - goos: linux
        goarch: arm64
      - goos: darwin
        goarch: amd64
      - goos: darwin
        goarch: arm64
      - goos: windows
        goarch: amd64
      - goos: windows
        goarch: arm64
      - goos: freebsd
        goarch: amd64
      - goos: freebsd
        goarch: arm64
```

## BINARY NAMING
```bash
# Server binary
weather-$GOOS-$GOARCH          # Linux/macOS/FreeBSD
weather-$GOOS-$GOARCH.exe      # Windows

# CLI binary (REQUIRED)
weather-cli-$GOOS-$GOARCH      # Linux/macOS/FreeBSD
weather-cli-$GOOS-$GOARCH.exe  # Windows

# Agent binary (OPTIONAL - only if src/agent/ exists)
weather-agent-$GOOS-$GOARCH
weather-agent-$GOOS-$GOARCH.exe
```

**Pattern:** `{project}-{type}-{os}-{arch}` where type is `cli` or `agent`

## WORKFLOW FILES
```
.github/workflows/
├── release.yml   # Stable releases (tags)
├── beta.yml      # Beta releases (beta tags)
└── daily.yml     # Daily builds (any push)

.gitea/workflows/
├── release.yml   # Same as GitHub
├── beta.yml
└── daily.yml
```

## DOCKER IMAGE TAGS
| Trigger | Tags |
|---------|------|
| Any push | devel, {commit-sha} |
| Beta tag | beta, {commit-sha} |
| Release | {version}, latest, YYMM, {commit-sha} |

---
**Full details: AI.md PART 28**

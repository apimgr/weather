# CI/CD Rules (PART 28)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use Makefile in CI/CD → explicit commands only
- ❌ Skip any of the 8 platforms
- ❌ Use different logic in GitHub vs Gitea
- ❌ Include 'v' prefix in VERSION → strip it

## REQUIRED - ALWAYS DO
- ✅ Explicit go build commands with all env vars
- ✅ All 8 platforms: linux/darwin/windows/freebsd × amd64/arm64
- ✅ GitHub/Gitea/Jenkins must match exactly
- ✅ VERSION from tag (strip 'v' prefix)
- ✅ Docker builds on EVERY push

## LDFLAGS (REQUIRED)
```bash
-ldflags "-s -w \
  -X 'main.Version=${VERSION}' \
  -X 'main.CommitID=${COMMIT}' \
  -X 'main.BuildDate=${DATE}' \
  -X 'main.OfficialSite=${SITE}'"
```

## VERSION HANDLING
| Git Ref | VERSION |
|---------|---------|
| `v1.2.3` tag | `1.2.3` (strip 'v') |
| `dev` branch | `dev` |
| `beta` branch | `beta` |

## DOCKER TAGS
| Trigger | Tags Applied |
|---------|--------------|
| Any push | `devel`, `{commit}` |
| Beta branch | + `beta` |
| Release tag | `{version}`, `latest`, `YYMM`, `{commit}` |

## BUILD MATRIX
```yaml
matrix:
  include:
    - { goos: linux, goarch: amd64 }
    - { goos: linux, goarch: arm64 }
    - { goos: darwin, goarch: amd64 }
    - { goos: darwin, goarch: arm64 }
    - { goos: windows, goarch: amd64 }
    - { goos: windows, goarch: arm64 }
    - { goos: freebsd, goarch: amd64 }
    - { goos: freebsd, goarch: arm64 }
```

---
**Full details: AI.md PART 28**

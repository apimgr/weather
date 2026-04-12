# Makefile Rules (PART 26)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO

- Never add more than 6 core Makefile targets
- Never hardcode PROJECTNAME (weather) or PROJECTORG (apimgr) -- infer from git remote or directory
- Never add `v` prefix to text versions (dev, beta, daily) or timestamps
- Never double the `v` prefix (vv0.3.0 -- wrong)
- Never copy or symlink binaries -- they stay in binaries/ until moved for release
- Never run Makefile in CI/CD -- use explicit commands instead
- Never create .env files -- hardcode sane defaults

## CRITICAL - ALWAYS DO

- Keep exactly 6 core targets
- Infer PROJECTNAME and PROJECTORG from git remote or directory path
- Use CGO_ENABLED=0 for all builds
- Build via Docker (golang:alpine) with GODIR/GOCACHE
- Embed version, commit, build date at compile time
- Add `v` prefix ONLY to numeric semantic versions (X.Y.Z)
- Use format_version_tag() shell function to prevent vv prefix

## Six Core Targets (DO NOT ADD MORE)

| Target | Purpose | Output |
|--------|---------|--------|
| `dev` | Quick development build | Temp directory |
| `local` | Production test build | binaries/ (with version) |
| `build` | Full release (all 8 platforms) | binaries/ |
| `test` | Run unit tests | Coverage report |
| `release` | Release with source archive | releases/ |
| `docker` | Build and push container | Registry |

## Version Tag Rules

| Tag Input | Type | Add v? | Result |
|-----------|------|--------|--------|
| `0.2.0` | Semver | YES | `v0.2.0` |
| `1.2.3` | Semver | YES | `v1.2.3` |
| `v1.2.0` | Already has v | NO | `v1.2.0` |
| `dev` | Text | NO | `dev` |
| `beta` | Text | NO | `beta` |
| `20251218` | Timestamp | NO | `20251218` |

Wrong examples: `vdev`, `vbeta`, `v20251218`, `vv0.3.0` -- NEVER do these.

## Version Sources

| File | Purpose |
|------|---------|
| `release.txt` | Stable version (SemVer X.Y.Z) -- source of truth |
| `site.txt` | Official site URL (https://wthr.top) |

## Binary Output Structure

```
binaries/
  weather                      # Local server binary
  weather-cli                  # Local CLI binary
  weather-linux-amd64          # Server: 8 platform builds
  weather-linux-arm64
  weather-darwin-amd64
  weather-darwin-arm64
  weather-windows-amd64.exe
  weather-windows-arm64.exe
  weather-freebsd-amd64
  weather-freebsd-arm64
  weather-cli-linux-amd64      # CLI: 8 platform builds
  weather-cli-linux-arm64
  ...
```

## GitHub Release Requirements

Every release MUST include:
- All 8 platform binaries for server
- All 8 platform binaries for CLI
- Source archive
- Checksums file

## CI/CD vs Local Development

| Task | Local | CI/CD |
|------|-------|-------|
| Build | Use Makefile targets | Explicit commands only |
| Test | Use Makefile targets | Explicit commands only |

NEVER use Makefile in CI/CD -- always use explicit commands.

## Reference

For complete details, see AI.md PART 26 (lines 33940-34716)

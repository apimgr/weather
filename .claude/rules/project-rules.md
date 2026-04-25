# Project Structure Rules (PART 2, 3, 4)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO

- Never use GPL/AGPL/LGPL licensed dependencies
- Never hardcode `weather` or `apimgr` — always infer from git remote or directory path
- Never place Dockerfile or docker-compose.yml in project root (use `docker/`)
- Never commit runtime volume data (config, data, logs)
- Never assume current directory is project root
- Never mix config_dir and data_dir purposes
- Never use `github.com/mattn/go-sqlite3` (CGO — forbidden)
- Never use `strconv.ParseBool()` — always use `config.ParseBool()`
- Never store plaintext passwords — always Argon2id
- Never store passwords in config files (`server.yml`)
- Never log passwords (even hashed)

## CRITICAL - ALWAYS DO

- Use MIT License (required for all projects)
- Include `LICENSE.md` in project root with third-party attributions
- Include README.md license badge
- Use `modernc.org/sqlite` (pure Go, CGO_ENABLED=0)
- Use Argon2id for all password hashing
- Use SHA-256 for API token hashing
- Use `config.ParseBool()` for all boolean parsing
- Build all 8 platforms (linux/darwin/windows/freebsd × amd64/arm64)
- Set CGO_ENABLED=0 always
- Use latest stable Go version (never hardcode specific versions)

## Project Identity

| Field | Value |
|-------|-------|
| Project name | `weather` |
| Organization | `apimgr` |
| Official site | `https://wthr.top` |
| Admin path | `admin` |
| API version | `v1` |
| License | MIT |

## Directory Structure

```
weather/
├── LICENSE.md          # Required
├── README.md           # Required
├── CLAUDE.md           # AI memory
├── AI.md               # Source of truth
├── Makefile            # Six core targets only
├── go.mod / go.sum
├── tools.go
├── release.txt         # Stable version (SemVer)
├── site.txt            # Official site URL
├── .claude/rules/      # Auto-loaded rule files
├── src/                # All source code
├── docker/             # ALL Docker files
├── scripts/            # run_tests.sh, docker.sh, incus.sh
├── tests/              # Test scripts
├── docs/               # ReadTheDocs (MkDocs) ONLY
└── binaries/           # Built binaries
```

## Library Rules

| Library | CGO | Rule |
|---------|-----|------|
| `modernc.org/sqlite` | NO | **ALWAYS USE** |
| `github.com/mattn/go-sqlite3` | YES | **NEVER USE** |

## Password & Token Security

| Item | Algorithm | Storage |
|------|-----------|---------|
| Passwords | Argon2id | Database only |
| API tokens | SHA-256 | Database only |
| Config file | — | NEVER store passwords here |

## Platforms (All 8 Required)

| OS | Architectures |
|----|--------------|
| Linux | amd64, arm64 |
| macOS (Darwin) | amd64, arm64 |
| Windows | amd64, arm64 |
| FreeBSD | amd64, arm64 |

## File Path Rules

- ALL paths are relative to project root
- NEVER assume current directory is project root
- Commands must `cd` to project root OR use absolute paths

## Required Scripts

| Script | Location | Purpose |
|--------|----------|---------|
| `run_tests.sh` | `scripts/` | Auto-detect and run tests |
| `docker.sh` | `scripts/` | Beta testing with Docker |
| `incus.sh` | `scripts/` | Beta testing with Incus |

## Reference

For complete details, see AI.md PART 2 (5026-5359), PART 3 (5360-6320), PART 4 (6321-6514)

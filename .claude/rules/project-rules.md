# Project Rules (PART 2, 3, 4)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Put Dockerfile in project root → `docker/Dockerfile`
- ❌ Create config/, data/, logs/ in project root
- ❌ Create TODO.md, CHANGELOG.md, SUMMARY.md, COMPLIANCE.md
- ❌ Put CONTRIBUTING.md in root → `.github/CONTRIBUTING.md`
- ❌ Use vendor/ directory (use Go modules)
- ❌ Create *.example.*, *.sample.* files
- ❌ Use plural directory names → singular: `handler/`, `model/`
- ❌ Use CamelCase for directories → lowercase only

## REQUIRED - ALWAYS DO
- ✅ Source code in `src/` directory
- ✅ Docker files in `docker/` directory
- ✅ Use `server.yml` (not .yaml)
- ✅ File naming: `lowercase_snake.go`
- ✅ Singular directory names: `handler/`, `model/`, `service/`
- ✅ MIT License in LICENSE.md
- ✅ All 8 platforms: linux/darwin/windows/freebsd × amd64/arm64

## DIRECTORY STRUCTURE
```
src/                        # Go source code (REQUIRED)
src/main.go                 # Entry point
src/config/                 # Configuration package
src/server/                 # HTTP server package
src/client/                 # CLI client (REQUIRED)
docker/                     # Docker files (REQUIRED)
docker/Dockerfile           # Multi-stage Dockerfile
docker/docker-compose.yml   # Production compose
binaries/                   # Build output (gitignored)
```

## OS-SPECIFIC PATHS

### Linux/BSD (Root)
| Type | Path |
|------|------|
| Config | `/etc/apimgr/weather/server.yml` |
| Data | `/var/lib/apimgr/weather/` |
| Logs | `/var/log/apimgr/weather/` |

### Linux/BSD (User)
| Type | Path |
|------|------|
| Config | `~/.config/apimgr/weather/server.yml` |
| Data | `~/.local/share/apimgr/weather/` |
| Logs | `~/.local/share/apimgr/weather/logs/` |

### Docker
| Type | Path |
|------|------|
| Config | `/config/server.yml` |
| Data | `/data/` |
| Database | `/data/db/sqlite/` |

---
**Full details: AI.md PART 2, PART 3, PART 4**

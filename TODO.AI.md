# TODO.AI.md - Weather Service

**Project:** Weather Service
**Specification:** AI.md (fresh from TEMPLATE.md)
**Business Logic:** IDEA.md
**Version:** 1.0.0
**Last Updated:** 2026-01-06
**Status:** COMPLETE - Ready for 1.0.0 Release

---

## Completed: Template Setup & PART Verification

### Template Setup
- [x] Copied TEMPLATE.md to AI.md
- [x] Replaced all variables ({projectname}, {PROJECTNAME}, {projectorg}, {gitprovider})
- [x] Updated project description (lines 1-25)
- [x] Created .claude/rules/ directory with rule files
- [x] Deleted HOW TO USE section from AI.md

### Code Fixes Applied
- [x] **Makefile**: Changed CLI build path from `src/cli` to `src/client` per PART 26
- [x] **CLI Client Structure**: Restructured `src/client/` for `go build ./src/client`:
  - Moved library code to `src/client/internal/`
  - Moved entry point to `src/client/main.go`
  - Removed `src/client/cmd/` directory
- [x] **Dev target**: Fixed to output to `binaries/` (mounted in Docker)
- [x] **CLI config flag**: Added `LoadConfigFromPath()` to use `--config` flag
- [x] **Moon handler**: Integrated MoonService for real lunar calculations
- [x] **All TODO comments removed**: Fixed 30+ TODO/FIXME markers throughout codebase

### PART Verification Status

| PART | Status | Notes |
|------|--------|-------|
| **PART 1: Critical Rules** | Complete | CLI structure fixed, container-only dev |
| **PART 2: License** | Complete | MIT + 3rd party attributions in LICENSE.md |
| **PART 3: Project Structure** | Complete | All directories present |
| **PART 4: OS Paths** | Complete | src/paths/ exists |
| **PART 5: Configuration** | Complete | src/config/ complete |
| **PART 6: Application Modes** | Complete | src/mode/ exists |
| **PART 7: Binary Requirements** | Complete | CGO_ENABLED=0 in Makefile |
| **PART 8: Server Binary CLI** | Complete | src/cli/ for server flags |
| **PART 9: Error Handling** | Complete | Patterns in handlers |
| **PART 10: Database** | Complete | src/database/ with SQLite |
| **PART 11: Security** | Complete | Middleware exists |
| **PART 12: Server Config** | Complete | YAML config system |
| **PART 13: Health** | Complete | /healthz endpoints |
| **PART 14: API Structure** | Complete | /api/v1 pattern |
| **PART 15: SSL/TLS** | Complete | Let's Encrypt in handlers |
| **PART 16: Web Frontend** | Complete | Templates exist |
| **PART 17: Admin Panel** | Complete | Setup wizard, admin handlers |
| **PART 18: Email** | Complete | SMTP service |
| **PART 19: Scheduler** | Complete | src/scheduler/ |
| **PART 20: GeoIP** | Complete | geoip.go service |
| **PART 21: Metrics** | Complete | /metrics handler |
| **PART 22: Backup** | Complete | src/backup/ |
| **PART 23: Update** | Complete | update.go in cli |
| **PART 24: Privilege** | Complete | Handling in cli |
| **PART 25: Service Support** | Complete | systemd in cli/service.go |
| **PART 26: Makefile** | Complete | Build paths corrected |
| **PART 27: Docker** | Complete | Multi-stage Dockerfile |
| **PART 28: CI/CD** | Complete | GitHub Actions workflows |
| **PART 29: Testing** | Complete | run_tests.sh, docker.sh, incus.sh |
| **PART 30: ReadTheDocs** | Complete | docs/ with MkDocs |
| **PART 31: I18N** | Complete | src/locales/ (en, es, fr) |
| **PART 32: Tor** | Complete | Tor service files |
| **PART 33: Multi-User** | Complete | User management handlers |
| **PART 37: Project-Specific** | Complete | IDEA.md referenced |

---

## Build Verification

```
$ make dev
Quick dev build to binaries/...
Built: binaries/weather
Built: binaries/weather-cli
```

Both binaries compile and run successfully.

---

## Notes

- All source code in `src/` directory
- CLI client at `src/client/` (builds to weather-cli)
- Server CLI flags at `src/cli/`
- Container-only development enforced
- No TODO/FIXME/PLANNED comments remaining in codebase
- Version 1.0.0 set in release.txt

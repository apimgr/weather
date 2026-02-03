# Project Rules (PART 2, 3, 4)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use GPL/AGPL/LGPL dependencies (copyleft)
- ❌ Put files in wrong locations (Dockerfile in root, etc.)
- ❌ Create forbidden directories (config/, data/, logs/, tmp/, vendor/)
- ❌ Create forbidden files (CHANGELOG.md, SUMMARY.md, COMPLIANCE.md)
- ❌ Use plural directory names (handlers/, models/ - use handler/, model/)
- ❌ Hardcode dev machine paths or values
- ❌ Put config files in repo (server.yml is runtime-generated)

## REQUIRED - ALWAYS DO
- ✅ MIT License for all projects
- ✅ LICENSE.md with embedded third-party licenses
- ✅ Source code in `src/` directory
- ✅ Docker files in `docker/` directory
- ✅ Build output to `binaries/` (gitignored)
- ✅ Follow OS-specific paths for config/data/logs/backup
- ✅ Use lowercase snake_case for Go files
- ✅ Use singular directory names (handler/, model/, service/)

## DIRECTORY STRUCTURE
```
src/              # All Go source code (REQUIRED)
docker/           # Dockerfile, compose files (REQUIRED)
docs/             # ReadTheDocs documentation only
tests/            # Test scripts
binaries/         # Build output (gitignored)
```

## OS-SPECIFIC PATHS
| OS | Config | Data | Logs |
|----|--------|------|------|
| Linux (root) | /etc/apimgr/weather/ | /var/lib/apimgr/weather/ | /var/log/apimgr/weather/ |
| Linux (user) | ~/.config/apimgr/weather/ | ~/.local/share/apimgr/weather/ | ~/.local/log/apimgr/weather/ |
| macOS (root) | /Library/Application Support/apimgr/weather/ | Same | /Library/Logs/apimgr/weather/ |
| macOS (user) | ~/Library/Application Support/apimgr/weather/ | Same | ~/Library/Logs/apimgr/weather/ |
| Windows (admin) | %ProgramData%\apimgr\weather\ | Same | Same\logs\ |
| Windows (user) | %AppData%\apimgr\weather\ | %LocalAppData%\apimgr\weather\ | Same\logs\ |
| Docker | /config/weather/ | /data/weather/ | /data/log/weather/ |

## FILE NAMING
| Type | Convention | Example |
|------|------------|---------|
| Go files | lowercase_snake.go | user_handler.go |
| Config | server.yml (NEVER .yaml) | server.yml |
| Docs | UPPERCASE.md | README.md |
| Binaries | weather-{os}-{arch} | weather-linux-amd64 |

---
**Full details: AI.md PART 2, PART 3, PART 4**

# Project Rules (PART 2, 3, 4)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Create files not in allowed list (see PART 3)
- ❌ Create directories not in allowed list
- ❌ Put Dockerfile in project root (must be `docker/Dockerfile`)
- ❌ Use `.yaml` extension (always `.yml`)
- ❌ Create SUMMARY.md, COMPLIANCE.md, NOTES.md, CHANGELOG.md
- ❌ Create config/, data/, logs/, tmp/ in project root
- ❌ Use vendor/ directory (use Go modules)
- ❌ Use GPL/AGPL/LGPL licensed dependencies
- ❌ Hardcode paths - use OS-specific path detection

## CRITICAL - ALWAYS DO
- ✅ All source code in `src/` directory
- ✅ Dockerfile in `docker/Dockerfile`
- ✅ Use MIT License with embedded third-party licenses
- ✅ Support all 4 OSes: Linux, macOS, Windows, FreeBSD
- ✅ Support both AMD64 and ARM64 architectures
- ✅ Use latest stable Go version
- ✅ Use OS-specific paths (see PART 4)
- ✅ Use `server.yml` for config files
- ✅ All paths relative to project root

## DIRECTORY STRUCTURE
```
./                        # Project root
├── src/                  # All Go source code
├── docker/               # Dockerfile, compose files
├── docs/                 # MkDocs documentation
├── tests/                # Test scripts
├── scripts/              # Production scripts
├── binaries/             # Build output (gitignored)
├── .claude/rules/        # AI rule files (this directory)
├── AI.md, IDEA.md        # Specifications
├── README.md, LICENSE.md # Documentation
└── Makefile, go.mod      # Build files
```

## OS PATH PATTERNS
| OS | Config | Data | Logs |
|----|--------|------|------|
| Linux (root) | `/etc/apimgr/weather/` | `/var/lib/apimgr/weather/` | `/var/log/apimgr/weather/` |
| Linux (user) | `~/.config/apimgr/weather/` | `~/.local/share/apimgr/weather/` | `~/.local/log/apimgr/weather/` |
| macOS | `/Library/Application Support/apimgr/weather/` | same | `/Library/Logs/apimgr/weather/` |
| Windows | `%ProgramData%\apimgr\weather\` | same | `%ProgramData%\apimgr\weather\logs\` |
| Docker | `/config/weather/` | `/data/weather/` | `/data/log/weather/` |

## LICENSE REQUIREMENTS
- MIT License for all projects
- LICENSE.md required in root
- All 3rd party licenses embedded in LICENSE.md
- NEVER use GPL/AGPL/LGPL dependencies

---
For complete details, see AI.md PART 2, 3, 4

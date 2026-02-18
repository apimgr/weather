# Binary Rules (PART 7, 8, 33)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use CGO (CGO_ENABLED=0 always)
- ❌ Run `go` commands locally (use `make dev/test/build`)
- ❌ Skip any of the 8 platform builds
- ❌ Use `-musl` suffix in binary names
- ❌ Add short flags except `-h` and `-v`
- ❌ Use `NO_COLOR` for anything except ANSI color output
- ❌ Ignore `TERM=dumb` (must disable ALL ANSI escapes)

## CRITICAL - ALWAYS DO
- ✅ CGO_ENABLED=0 for all builds
- ✅ Single static binary with embedded assets (`embed` package)
- ✅ Build all 8 platforms (linux/darwin/windows/freebsd × amd64/arm64)
- ✅ Binary naming: `weather-{os}-{arch}` (windows adds `.exe`)
- ✅ CLI binary naming: `weather-cli-{os}-{arch}`
- ✅ Support NO_COLOR environment variable (disable color only)
- ✅ Support TERM=dumb (disable ALL ANSI escapes, force CLI mode)
- ✅ Build source: always `./src` directory

## BINARY TYPES
| Binary | Name | Required | Purpose |
|--------|------|:--------:|---------|
| server | `weather` | ✅ | Main application, runs as service |
| client | `weather-cli` | ✅ | CLI/TUI/GUI companion |
| agent | `weather-agent` | ❌ | Optional, runs on remote machines |

## CLI FLAGS (PART 8)
```
--help, -h              # Show help
--version, -v           # Show version
--mode                  # production|development
--config                # Config directory
--data                  # Data directory
--log                   # Log directory
--pid                   # PID file path
--address               # Listen address
--port                  # Listen port
--baseurl               # URL path prefix
--debug                 # Enable debug mode
--status                # Show status and health
--service               # Service management
--daemon                # Daemonize
--maintenance           # Maintenance operations
--update                # Update binary
--color                 # auto|always|never
```

## DISPLAY ENVIRONMENT
| TERM=dumb | NO_COLOR | Result |
|:---------:|:--------:|--------|
| ✓ | - | No ANSI, no emoji, CLI mode forced |
| - | ✓ | No color, emoji allowed, TUI/GUI allowed |
| - | - | Full ANSI, color, auto-detect mode |

## CLIENT AUTO-DETECTION (PART 33)
- Detect display mode automatically (headless → CLI → TUI → GUI)
- NO command-line flags for mode selection
- Check: GUI available? → TUI supported? → CLI mode
- TERM=dumb forces CLI mode regardless of display capability

---
For complete details, see AI.md PART 7, 8, 33

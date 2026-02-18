# Binary Rules (PART 7, 8, 33)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use CGO → CGO_ENABLED=0 always
- ❌ Skip platforms → build all 8
- ❌ Use -musl suffix in binary names
- ❌ Run Go locally → use containers only
- ❌ Use short flags except -h and -v

## REQUIRED - ALWAYS DO
- ✅ CGO_ENABLED=0 for all builds
- ✅ All 8 platforms: linux/darwin/windows/freebsd × amd64/arm64
- ✅ Single static binary with embedded assets
- ✅ Support NO_COLOR and TERM=dumb
- ✅ Client binary for ALL projects

## BINARY NAMES
| Binary | Name | Purpose |
|--------|------|---------|
| Server | `weather` | Main service |
| Client | `weather-cli` | CLI/TUI/GUI interface |
| Agent | `weather-agent` | Optional, remote machines |

## BUILD NAMING
```
weather-linux-amd64
weather-linux-arm64
weather-darwin-amd64
weather-darwin-arm64
weather-windows-amd64.exe
weather-windows-arm64.exe
weather-freebsd-amd64
weather-freebsd-arm64
```

## CLI FLAGS (Server)
```
--help, -h              Show help
--version, -v           Show version
--mode production|development
--config {config_dir}
--data {data_dir}
--log {log_dir}
--pid {pid_file}
--address {listen}
--port {port}
--baseurl {path}
--debug                 Enable debug mode
--status                Show health status
--service {command}     Service management
--daemon                Daemonize
--maintenance {cmd}     Maintenance operations
--update [check|yes|branch]
```

## NO_COLOR / TERM=dumb SUPPORT
| Environment | Effect |
|-------------|--------|
| NO_COLOR set (any value) | Disable ANSI colors |
| TERM=dumb | Disable ALL ANSI escapes, force CLI mode |
| --color=never | Disable colors via flag |
| --color=always | Force colors on |

---
**Full details: AI.md PART 7, PART 8, PART 33**

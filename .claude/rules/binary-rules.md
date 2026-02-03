# Binary Rules (PART 7, 8, 33)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Enable CGO (ALWAYS CGO_ENABLED=0)
- ❌ Use -musl suffix in binary names
- ❌ Skip any of the 8 required platforms
- ❌ Run Go commands locally (use containers)
- ❌ Use mattn/go-sqlite3 (requires CGO)
- ❌ Add short flags except -h and -v

## REQUIRED - ALWAYS DO
- ✅ CGO_ENABLED=0 for all builds
- ✅ Single static binary with embedded assets
- ✅ Build all 8 platforms (4 OS × 2 arch)
- ✅ Use container-only development (make dev/test/build)
- ✅ Binary names: `weather-{os}-{arch}` (.exe for Windows)
- ✅ Client binary: `weather-cli-{os}-{arch}`

## 8 REQUIRED PLATFORMS
| OS | amd64 | arm64 |
|----|-------|-------|
| linux | weather-linux-amd64 | weather-linux-arm64 |
| darwin | weather-darwin-amd64 | weather-darwin-arm64 |
| windows | weather-windows-amd64.exe | weather-windows-arm64.exe |
| freebsd | weather-freebsd-amd64 | weather-freebsd-arm64 |

## CLI FLAGS (server binary)
```
--help, -h               Show help
--version, -v            Show version
--mode {production|development}
--config {config_dir}
--data {data_dir}
--log {log_dir}
--pid {pid_file}
--address {listen}
--port {port}
--debug                  Enable debug mode
--status                 Show status (exit 0=healthy, 1=unhealthy)
--service {start,restart,stop,reload,--install,--uninstall,--disable,--help}
--daemon                 Daemonize
--maintenance {backup,restore,update,mode,setup,--help}
--update [check|yes|branch {stable|beta|daily}]
```

## CLIENT (weather-cli)
- REQUIRED for all projects
- Auto-detect display mode (TUI vs CLI)
- Use --server/--token for connection
- Output formats: json, table, plain, yaml, csv

## BUILD COMMANDS
```bash
make dev      # Quick build to temp dir
make local    # Production build to binaries/
make build    # All 8 platforms
make test     # Unit tests
```

---
**Full details: AI.md PART 7, PART 8, PART 33**

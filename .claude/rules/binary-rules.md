# Binary Rules (PART 7, 8, 33)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO

- Never use CGO -- always CGO_ENABLED=0
- Never use mattn/go-sqlite3 (requires CGO)
- Never hardcode host, IP, or port in binary code
- Never display partial URLs (GET /api/) -- always full URL with proto/fqdn/port
- Never include :80 or :443 in displayed URLs
- Never embed security databases in binary
- Never run binaries directly on local machine -- always use containers
- Never include -musl suffix in released binary names
- Never use strconv.ParseBool() -- always config.ParseBool()

## CRITICAL - ALWAYS DO

- Set CGO_ENABLED=0 for all builds
- Build all 8 platform targets
- Embed build info at compile time (version, commit, build date)
- Handle TERM=dumb (no special terminal capabilities assumed)
- Respect NO_COLOR standard
- Handle all signals properly for graceful shutdown (platform-dependent)
- Detect display environment and adapt output (GUI/TUI/CLI/Headless)
- Create directories if they do not exist (all directory flags)
- Use platform-dependent signal handling (Windows differs from Unix)

## Build Targets (All 8 Required)

| Platform | Architectures |
|----------|--------------|
| linux | amd64, arm64 |
| darwin | amd64, arm64 |
| windows | amd64, arm64 |
| freebsd | amd64, arm64 |

## Binary Naming

| Type | Linux Example | Windows Example |
|------|--------------|-----------------|
| Server | weather-linux-amd64 | weather-windows-amd64.exe |
| CLI | weather-cli-linux-amd64 | weather-cli-windows-amd64.exe |
| Agent | weather-agent-linux-amd64 | weather-agent-linux-amd64 |

- Musl builds: strip binary; final name has NO -musl suffix
- Local build outputs: weather, weather-cli, weather-agent (no platform suffix)

## Display Mode Detection

| Mode | Trigger |
|------|---------|
| GUI | Native display available (X11/Wayland/Windows/macOS), CLI binary only |
| TUI | Terminal available, interactive (TTY, SSH, mosh, screen, tmux) |
| CLI | Command provided or piped output |
| Headless | No display, no TTY (daemon, service, cron) |

Smart detection only -- NEVER implement explicit --ui-mode flags.

## Build Info (Embedded at Compile Time)

Every binary MUST have these values embedded:
- Version -- from git tag or release.txt
- Commit -- git commit SHA
- BuildDate -- build timestamp
- ProjectName -- weather
- ProjectOrg -- apimgr

## Signal Handling

| Signal | Unix | Windows |
|--------|------|---------|
| SIGTERM | Graceful shutdown | Graceful shutdown |
| SIGINT | Graceful shutdown | Graceful shutdown |
| SIGHUP | Config reload | NOT SUPPORTED |
| SIGUSR1/SIGUSR2 | Custom handlers | NOT SUPPORTED |
| SIGQUIT | Core dump | NOT SUPPORTED |

Use build tags to separate platform signal code (signal_unix.go, signal_windows.go).

## Client Binary (weather-cli) -- REQUIRED

- REQUIRED for this project
- Supports --token for API authentication (PART 34 is implemented)
- Supports --user flag for user context
- Server auto-detects if target is user or org
- ALWAYS runs as unprivileged user -- NEVER as root/administrator
- NEVER uses OS system directories
- Supports automatic cluster failover

## PID File Rules

- Stale PID detection is REQUIRED
- Crash or signal 9 leaves stale PID files -- must handle gracefully

## Source Structure

src/
  shared/         # Shared code (version, config, utils)
  server/         # Server binary
  client/         # CLI binary
    cli/          # CLI mode
    tui/          # TUI mode (bubbletea)
    gui/          # GUI mode (native)
  agent/          # Agent binary (if implemented)

## Reference

For complete details, see AI.md PART 7 (lines 8983-9645), PART 8 (9646-12820), PART 33 (45096-49510)

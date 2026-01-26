# WEATHER - AI Quick Reference

## Binary Terminology
- **server** = `weather` (main binary, runs as service)
- **client** = `weather-cli` (REQUIRED companion, CLI/TUI/GUI)
- **agent** = `weather-agent` (optional, runs on remote machines)

## Key Placeholders
- `{projectname}` = weather
- `{projectorg}` = apimgr
- `{admin_path}` = admin (default)
- `{api_version}` = v1

## Account Types (CRITICAL)
- **Server Admin** = manages the app (NOT a privileged OS user)
- **Primary Admin** = first admin, cannot be deleted
- **Regular User** = end-user (PART 34, optional feature)
- Server Admins != Regular Users (separate DB tables)

## Cluster vs Managed Nodes (CRITICAL)
- **Cluster Node** = another instance of THIS app (horizontal scaling)
- **Managed Node** = EXTERNAL resource app controls/monitors (Docker hosts, etc.)
- Most apps only have cluster nodes

## NEVER Do (Top 13)
1. Use bcrypt -> Use Argon2id
2. Put Dockerfile in root -> `docker/Dockerfile`
3. Use CGO -> CGO_ENABLED=0 always
4. Hardcode dev values -> Detect at runtime
5. Use external cron -> Internal scheduler (PART 19)
6. Store passwords plaintext -> Argon2id (tokens use SHA-256)
7. Create premium tiers -> All features free
8. Use Makefile in CI/CD -> Explicit commands
9. Guess or assume -> Read spec or ask
10. Skip platforms -> Build all 8
11. Use inline comments -> Comments above code only
12. Use strconv.ParseBool() -> Use config.ParseBool()
13. Run Go locally -> Use containers only (make dev/test/build)

## MUST Do (Top 12)
1. Re-read spec before each task
2. Use containers for all builds/tests
3. Use config.ParseBool() for ALL boolean parsing
4. Use Argon2id for passwords, SHA-256 for tokens
5. Use path normalization/validation (prevent traversal)
6. Support all 4 OSes and 2 architectures (8 binaries)
7. Use server.yml (not .yaml)
8. Keep documentation in sync with code
9. Have admin WebUI for ALL settings
10. Have corresponding API endpoint for every web page
11. Use single static binary (all assets embedded)
12. Detect machine settings at runtime

## File Locations
- Config: `{config_dir}/server.yml`
- Data: `{data_dir}/`
- Logs: `{log_dir}/`
- Source: `src/`
- Docker: `docker/`

## Where to Find Details
- AI behavior: `rules/ai-rules.md` (PART 0, 1)
- Project structure: `rules/project-rules.md` (PART 2, 3, 4)
- Configuration: `rules/config-rules.md` (PART 5, 6, 12)
- Frontend/WebUI: `rules/frontend-rules.md` (PART 16, 17)
- Full spec: `AI.md` (~54k lines)

## Current Project State
- AI.md: Fresh copy from TEMPLATE.md (2026-01-26)
- Build: Compiles successfully (`make dev`)
- Tests: Some runtime test failures (not compilation)
- CSS: Inline styles migrated to dracula.css
- Scheduler: backup_hourly task implemented
- Service support: All platforms (systemd, launchd, runit, rc.d, Windows)
- Tor hidden service: Implemented (PART 32)
- Docker configuration: Per PART 27

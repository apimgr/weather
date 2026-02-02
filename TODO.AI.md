# Weather Service - AI.md Compliance TODO

## Project Info
- **Project Name**: weather
- **Organization**: apimgr
- **Template Version**: Fresh copy from TEMPLATE.md 2026-02-01
- **AI.md Location**: /root/Projects/github/apimgr/weather/AI.md

---

## CRITICAL RULES (Committed to Memory)

### NEVER Do These (15 Rules)
1. Use bcrypt → Use Argon2id
2. Put Dockerfile in root → `docker/Dockerfile`
3. Use CGO → CGO_ENABLED=0 always
4. Hardcode dev values → Detect at runtime
5. Use external cron → Internal scheduler (PART 19)
6. Store passwords plaintext → Argon2id (tokens use SHA-256)
7. Create premium tiers → All features free
8. Use Makefile in CI/CD → Explicit commands
9. Guess or assume → Read spec or ask
10. Skip platforms → Build all 8
11. Use strconv.ParseBool() → Use config.ParseBool()
12. Run Go locally → Use containers only (make dev/test/build)
13. Client-side rendering (React/Vue) → Server-side Go templates
14. Require JavaScript for core features → Progressive enhancement only
15. Let long strings break mobile → Use word-break CSS

### MUST Do These (12 Rules)
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

### COMMIT Rules
1. AI cannot run git add/commit/push - write to .git/COMMIT_MESS instead
2. COMMIT_MESS must accurately reflect actual git status
3. Use emoji prefixes for commit types (✨ feat, 🐛 fix, 📝 docs, etc.)
4. Title max 64 chars including emojis
5. Never commit without verification

---

## Completed Items

### Initial Setup (2026-02-01)
- [x] AI.md copied fresh from TEMPLATE.md
- [x] Placeholders replaced (weather/apimgr/WEATHER/APIMGR)
- [x] Read PART 0-5 of AI.md
- [x] IDEA.md verified - project features documented
- [x] TODO.AI.md updated with committed rules

### Previous Work
- [x] Fixed compilation errors
- [x] Build verification - `make dev` successful
- [x] PART 16: Inline CSS cleanup completed
- [x] PART 19: backup_hourly task implemented
- [x] Service support: All platforms
- [x] Tor hidden service: Implemented (PART 32)
- [x] Docker configuration: Per PART 27
- [x] PART 33: CLI client rewrite per specification
  - [x] Flags: --server, --token, --token-file, --user, --config, --output, --debug, --color
  - [x] Shell completions: --shell completions/init [SHELL]
  - [x] Nested YAML config: server.primary, auth.token, output.format, tui.theme
  - [x] Cross-platform paths: Unix XDG, Windows AppData
  - [x] TUI: Window-aware, 7 size modes, vim navigation, Dracula theme
  - [x] URL encoding: url.QueryEscape() for location params
  - [x] Official site: https://wthr.top

---

## Current Status

**READY FOR DEVELOPMENT**

AI.md has been freshly copied and configured. Rules committed to memory.

---

## Working Notes

- Container-only development - NEVER run go locally
- Use `make dev` for quick builds, `make test` for tests
- Test binaries in Docker/Incus, never on host
- AI.md is SOURCE OF TRUTH - always re-read relevant PART before implementing

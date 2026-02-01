# Weather Service - AI.md Compliance TODO

## Project Info
- **Project Name**: weather
- **Organization**: apimgr
- **Template Version**: Fresh copy from TEMPLATE.md 2026-01-29
- **AI.md Location**: /root/Projects/github/apimgr/weather/AI.md

---

## CRITICAL RULES (Committed to Memory)

### NEVER Do These (13 Rules)
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
11. Use inline comments → Comments above code only
12. Use strconv.ParseBool() → Use config.ParseBool()
13. Run Go locally → Use containers only (make dev/test/build)

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

---

## Completed Items

### Initial Setup (2026-01-29)
- [x] AI.md copied fresh from TEMPLATE.md
- [x] Placeholders replaced (weather/apimgr)
- [x] Read PART 0-5 of AI.md

### Previous Work
- [x] Fixed compilation errors
- [x] Build verification - `make dev` successful
- [x] PART 16: Inline CSS cleanup completed
- [x] PART 19: backup_hourly task implemented
- [x] Service support: All platforms
- [x] Tor hidden service: Implemented (PART 32)
- [x] Docker configuration: Per PART 27

### Audit (2026-01-29)
- [x] .gitignore: Added AI config dirs, IDE dirs, fixed rootfs/
- [x] .claude/: Untracked from git (now gitignored per spec)
- [x] .gitea/workflows/: Created with CI workflows
- [x] docs/Makefile: Deleted (misplaced)
- [x] IDEA.md: Verified all features implemented, added missing

### Verification (2026-01-29)
- [x] `make dev` - builds successfully
- [x] `make test` - all tests pass
- [x] Binary --help works in container
- [x] CLI --help works in container

---

## Current Status

**READY TO COMMIT**

All audit items complete. Build and tests pass.

---

## Working Notes

- Container-only development - NEVER run go locally
- Use `make dev` for quick builds, `make test` for tests
- Test binaries in Docker/Incus, never on host

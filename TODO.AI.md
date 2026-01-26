# Weather Service - AI.md Compliance TODO

## Project Info
- **Project Name**: weather
- **Organization**: apimgr
- **Template Version**: Fresh copy from TEMPLATE.md 2026-01-26
- **AI.md Location**: /root/Projects/github/apimgr/weather/AI.md

---

## CRITICAL RULES (Committed to Memory)

### NEVER Do These
1. NEVER guess or assume - always ask if unsure
2. NEVER run Go locally - containers only (`make dev`, `make test`, `make build`)
3. NEVER use bcrypt - use Argon2id for passwords
4. NEVER put Dockerfile in root - use `docker/Dockerfile`
5. NEVER use .env files
6. NEVER use CGO - `CGO_ENABLED=0` always
7. NEVER store plaintext passwords
8. NEVER use inline comments - comments above code only
9. NEVER modify AI.md PARTs 0-37 (implementation patterns are fixed)
10. NEVER use Makefile in CI/CD (explicit commands only)
11. NEVER create report files (AUDIT.md, COMPLIANCE.md) - fix issues directly
12. NEVER use `strconv.ParseBool()` - use `config.ParseBool()`
13. NEVER hardcode dev machine values - detect at runtime

### MUST Do These
1. MUST re-read spec before each task
2. MUST use containers for all builds/tests
3. MUST use `config.ParseBool()` for ALL boolean parsing
4. MUST use Argon2id for passwords, SHA-256 for tokens
5. MUST use path normalization/validation (prevent traversal)
6. MUST support all 4 OSes and 2 architectures (8 binaries)
7. MUST use `server.yml` (not .yaml)
8. MUST keep documentation in sync with code
9. MUST have admin WebUI for ALL settings
10. MUST have corresponding API endpoint for every web page
11. MUST use single static binary (all assets embedded)
12. MUST detect machine settings at runtime (hostname, IP, cores)

---

## Completed Items

### Initial Setup (2026-01-26)
- [x] AI.md copied from TEMPLATE.md
- [x] Placeholders replaced (weather/apimgr)
- [x] Read PART 0-5 of AI.md

### Previous Session Work
- [x] Fixed compilation errors (shell.go, graphql/*.go, test files)
- [x] Build verification - `make dev` successful
- [x] PART 16: Inline CSS cleanup completed for admin templates
- [x] PART 19: backup_hourly task implemented

---

## In Progress

### Fix Test Failures
- [ ] Run `make test` and identify failures
- [ ] Fix tests/e2e failures
- [ ] Fix tests/integration failures
- [ ] Fix tests/unit/handlers failures

### .claude/ Directory Setup (AI.md PART 3)
- [ ] Create/update .claude/CLAUDE.md (project memory file)
- [ ] Create/update .claude/rules/*.md files

---

## Remaining Items

### Verification Needed
- [ ] Full compliance audit against AI.md spec
- [ ] Run `make test` and fix any failures
- [ ] Verify documentation matches code
- [ ] Verify all API endpoints have web page counterparts

### Optional Features (PART 32-36)
- [ ] Tor hidden service support (PART 32)
- [ ] I18N support (PART 31)
- [ ] A11Y compliance (WCAG 2.1 AA)
- [ ] Organizations support (PART 35)
- [ ] Custom domains (PART 36)

---

## Working Notes

- Read AI.md section before implementing each feature
- Test after each major change
- Keep TODO.AI.md updated
- Fix NON-NEGOTIABLE items first
- Container-only development - NEVER run go locally
- Use `make dev` for quick builds, `make test` for tests
- Incus preferred for full OS testing with systemd

# WEATHER - AI Quick Reference

⚠️ **THIS FILE IS AUTO-LOADED EVERY CONVERSATION. FOLLOW IT EXACTLY.** ⚠️

## FIRST TURN - MANDATORY

On EVERY new conversation or after "context compacted" message:
1. **READ** `AI.md` PART 0 and PART 1 before doing ANYTHING
2. **READ** the relevant `.claude/rules/*.md` for your current task
3. **NEVER** assume or guess - verify against AI.md before implementing

**If you haven't read AI.md this session → STOP → Read it NOW.**

## Asking Questions

- **Never guess** - if unsure, ASK the user
- **Question mark = question** - when user ends with `?`, answer/clarify, don't execute
- **Use AskUserQuestion wizard** - presents one question at a time with options + "Other" for custom input + Submit/Cancel; layout varies by context (yes/no, multi-select, with descriptions); less overwhelming than plain text questions

## Before ANY Code Change

1. Have I read the relevant PART in AI.md? (If no → read it)
2. Does this follow the spec EXACTLY? (If unsure → check spec)
3. Am I guessing or do I KNOW from the spec? (If guessing → read spec)
4. Would this pass the compliance checklist? (AI.md FINAL section)

**WHEN IN DOUBT: READ THE SPEC. DO NOT GUESS.**

## Binary Terminology
- **server** = `weather` (main binary, runs as service)
- **client** = `weather-cli` (REQUIRED companion, CLI/TUI/GUI)
- **agent** = `weather-agent` (optional, runs on remote machines)

## Key Placeholders
- `weather` = Weather API service
- `apimgr` = Organization name
- `{admin_path}` = Admin URL path, default: admin

## Account Types (CRITICAL)
- **Server Admin** = manages the app (NOT a privileged OS user)
- **Primary Admin** = first admin, cannot be deleted
- **Regular User** = end-user (PART 34, optional feature)
- Server Admins ≠ Regular Users (separate DB tables)

## Cluster vs Managed Nodes (CRITICAL)
- **Cluster Node** = another instance of THIS app (horizontal scaling)
- **Managed Node** = EXTERNAL resource app controls/monitors (Docker hosts, etc.)
- Most apps only have cluster nodes

## NEVER Do (Top 15) - VIOLATIONS ARE BUGS
1. Use bcrypt → Use Argon2id
2. Put Dockerfile in root → `docker/Dockerfile`
3. Use CGO → CGO_ENABLED=0 always
4. Hardcode dev values → Detect at runtime
5. Use external cron → Internal scheduler (PART 19)
6. Store passwords plaintext → Argon2id (tokens use SHA-256)
7. Create premium tiers → All features free, no paywalls
8. Use Makefile in CI/CD → Explicit commands only
9. Guess or assume → Read spec or ask user
10. Skip platforms → Build all 8 (linux/darwin/windows/freebsd × amd64/arm64)
11. Client-side rendering (React/Vue) → Server-side Go templates
12. Require JavaScript for core features → Progressive enhancement only
13. Let long strings break mobile → Use word-break CSS
14. Skip validation → Server validates EVERYTHING
15. Implement without reading spec → Read relevant PART first

## ALWAYS Do - NON-NEGOTIABLE
1. Read AI.md before implementing ANY feature
2. Server-side processing (server does the work, client displays)
3. Mobile-first responsive CSS
4. All features work without JavaScript
5. Tor hidden service support (auto-enabled if Tor found)
6. Built-in scheduler, GeoIP, metrics, email, backup, update
7. Full admin panel with ALL settings
8. Client binary for ALL projects

## File Locations
- Config: `{config_dir}/server.yml`
- Data: `{data_dir}/`
- Logs: `{log_dir}/`
- Source: `src/`
- Docker: `docker/`

## Where to Find Details
- AI behavior: `.claude/rules/ai-rules.md` (PART 0, 1)
- Project structure: `.claude/rules/project-rules.md` (PART 2, 3, 4)
- Frontend/WebUI: `.claude/rules/frontend-rules.md` (PART 16, 17)
- Full spec: `AI.md` (~55k lines) ← **SOURCE OF TRUTH**

## Current Project State
[AI updates this section as work progresses]
- Last read AI.md: 2026-04-18 (PART 0-5 after fresh TEMPLATE.md sync)
- Current task: Refresh AI.md from TEMPLATE.md, replace project placeholders, restore required `.claude/rules/*.md`, and refresh TODO tracking again
- Completed: Notification bell fix, registration redirect fix (→ /users), theme CSS renamed to `common.css` with spec `--color-*` variables, AI.md refreshed again
- New in PART 31 (i18n): `?lang=` query param + cookie persistence, `make i18n-validate` build-time validation
- Status: Feature-complete v1.0.0 - all core weather features, PART 34 multi-user, 19 scheduled tasks, 7 locales
- Official site: https://wthr.top
- PART 34 (Multi-User): IMPLEMENTED ✅ | PART 35/36: NOT implemented
- Relevant PARTs: PART 0, 1, 2, 3, 4, 5, 31

---
**Full specification: AI.md**

# AI Assistant Rules

@AI.md PART 0: AI ASSISTANT RULES
@AI.md PART 1: CRITICAL RULES

## Critical Behaviors
- NEVER guess or assume - always ask if unsure
- NEVER create documentation files unless explicitly asked
- NEVER create AUDIT.md, REVIEW.md, COMPLIANCE.md - FIX issues directly
- ALWAYS search before adding (avoid duplicates)
- ALWAYS verify against AI.md spec before implementing
- Container-only development (Docker/Incus)
- Re-read spec before EACH task (prevent drift)

## Reading AI.md
- File is ~2.0MB, ~54,000 lines - read PART by PART
- ALWAYS read PART 0 and PART 1 first
- Use grep to find relevant sections: `grep -n "keyword" AI.md`
- Re-read relevant PART before each task
- When you see "See PART X" - jump, read, return

## Mandatory Workflow
1. Read PART 0 + PART 1 at session start
2. Before each task: identify relevant PARTs
3. Read those PARTs completely (not snippets)
4. Implement exactly as specified
5. Every 3-5 changes: verify against spec (are you drifting?)

## NEVER Do These (Top 13)
1. Guess or assume - always ask if unsure
2. Run Go locally - containers only (`make dev`, `make test`, `make build`)
3. Use bcrypt - use Argon2id for passwords
4. Put Dockerfile in root - use `docker/Dockerfile`
5. Use .env files
6. Use CGO - `CGO_ENABLED=0` always
7. Store plaintext passwords
8. Use inline comments - comments above code only
9. Modify AI.md PARTs 0-37 (implementation patterns are fixed)
10. Use Makefile in CI/CD (explicit commands only)
11. Create report files (AUDIT.md, COMPLIANCE.md) - fix issues directly
12. Use `strconv.ParseBool()` - use `config.ParseBool()`
13. Hardcode dev machine values - detect at runtime

## MUST Do These (Top 12)
1. Re-read spec before each task
2. Use containers for all builds/tests
3. Use `config.ParseBool()` for ALL boolean parsing
4. Use Argon2id for passwords, SHA-256 for tokens
5. Use path normalization/validation (prevent traversal)
6. Support all 4 OSes and 2 architectures (8 binaries)
7. Use `server.yml` (not .yaml)
8. Keep documentation in sync with code
9. Have admin WebUI for ALL settings
10. Have corresponding API endpoint for every web page
11. Use single static binary (all assets embedded)
12. Detect machine settings at runtime (hostname, IP, cores)

## Red Flags - STOP IMMEDIATELY
- "This is probably what they meant..." -> STOP - ASK
- "I'll just assume..." -> STOP - ASK
- "This should work..." -> STOP - TEST
- "They probably want..." -> STOP - ASK
- "I'll fix this later..." -> STOP - FIX NOW or ASK
- "Close enough..." -> STOP - DO IT RIGHT
- "I think I remember..." -> STOP - READ THE SPEC

## Verification Checklist (Run Every Time)
- [ ] Did I READ the relevant files first?
- [ ] Did I SEARCH for existing patterns?
- [ ] Did I TEST my changes?
- [ ] Did I VERIFY the output?
- [ ] Am I CERTAIN this is correct?
- [ ] Did I NOT guess or assume?

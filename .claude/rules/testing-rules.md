# Testing Rules (PART 29, 30, 31)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO

- Never run binaries directly on local machine -- always use containers
- Never use project directory for test/runtime data
- Never use /tmp directly -- always use /tmp/apimgr/weather-XXXXXX structure
- Never use bare mktemp -d without org prefix
- Never bypass authentication in tests -- TEST that it works
- Never skip any route or endpoint from testing
- Never use debug bypass mode in automated tests
- Never skip content negotiation testing

## CRITICAL - ALWAYS DO

- Run ALL builds, tests, and binaries inside containers (Docker/Incus)
- Use /tmp/apimgr/weather-XXXXXX/ for all test temp directories
- Test every route with ALL applicable Accept headers
- Test authentication -- verify it works, never bypass it
- Use trap for cleanup in all test scripts
- Achieve 100% test coverage
- Run both unit tests and integration tests

## Temp Directory Structure (REQUIRED)

All temp dirs MUST be under /tmp/apimgr/weather-XXXXXX/:

  /tmp/apimgr/weather-XXXXXX/
    rootfs/
      config/
      data/
      logs/

NEVER use:
- /tmp/ directly (too broad -- prohibited)
- ./rootfs/ (pollutes project directory)
- ./docker/rootfs/ (pollutes project directory)

## Container-Only Execution Rule

ALL builds, tests, and binary execution MUST use containers.
Local machine is for ORCHESTRATION ONLY.

Build inside golang:alpine container. Run tests inside container.
Execute binaries inside container.

## Two Types of Tests (Both REQUIRED)

| Type | Purpose |
|------|---------|
| Unit tests | Test individual functions/packages |
| Integration tests | Test full request/response cycles |

## Content Negotiation Testing (REQUIRED)

Every route MUST be tested with ALL applicable Accept headers:

Frontend routes (browser and CLI detection):
- Accept: text/html  (renders HTML for browser)
- Accept: text/plain (renders formatted text for CLI)

API routes:
- Accept: application/json  (JSON response)
- Accept: text/plain        (plain text response)

Text extension endpoints:
- GET /api/v1/resource.txt

## Test Coverage (100% Required)

ALL code MUST have 100% test coverage. No exceptions.
Coverage report is required in CI/CD and checked automatically.

## Required Test Scripts

| Script | Location | Purpose |
|--------|----------|---------|
| docker.sh | scripts/ | Beta testing with Docker (HUMAN USE ONLY) |
| incus.sh | scripts/ | Beta testing with Incus (HUMAN USE ONLY) |
| run_tests.sh | scripts/ | Auto-detect and run tests |

docker.sh and incus.sh MUST:
- Copy to temp dir before use (NEVER run from project directory)
- Test ALL endpoints from IDEA.md (both API and frontend)
- Test authentication flows
- Clean up with trap

## Authentication Testing

Admin routes (/admin/**) MUST be tested properly:
- Verify that unauthenticated requests are rejected (401 or 302)
- Log in with valid credentials and verify access
- NEVER use debug bypass or mock auth in automated tests

## Process Safety (Container Management)

When managing containers and processes during testing:
- ALWAYS identify the exact container name or process before stopping it
- ONLY stop/remove containers created by this project (apimgr/weather:*)
- NEVER remove base images (golang, alpine, ubuntu, etc.)
- NEVER remove volumes unless they belong to this test run
- Be specific: use exact names, IDs, or paths -- never broad patterns

## Container Tools Required

Inside test containers, always install:
  apk add --no-cache curl bash file jq

## Internationalization (PART 31)

- Every human-readable string MUST be translatable
- Every key in en.json MUST exist in all language files
- Build-time validation enforces key parity
- Missing key falls back to English -- never shows raw key name

## Language Fallback Chain

1. ?lang= query param (sets lang cookie + uses immediately)
2. lang cookie
3. Accept-Language HTTP header
4. Default: en

NEVER use URL path prefixes (/es/dashboard) -- use ?lang=es + cookie only.

## Accessibility (PART 31)

- WCAG 2.1 AA compliance required
- Touch targets minimum 44x44px
- Screen reader compatibility (aria-* attributes)
- NEVER convey information by color alone
- Focus management on modal open/close

## Documentation (PART 30)

- docs/ directory is ONLY for ReadTheDocs/MkDocs files
- NEVER put source code or scripts in docs/
- Every project MUST have documentation on ReadTheDocs
- Required pages: index, installation, configuration, api, cli, admin

## Reference

For complete details, see AI.md PART 29 (39157-40977), PART 30 (40978-41709), PART 31 (41710-43316)

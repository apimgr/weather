# Testing Rules (PART 29, 30, 31)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Run tests on local machine (use containers)
- ❌ Run `go test` directly (use `make test`)
- ❌ Skip integration tests
- ❌ Test with hardcoded hostnames/IPs
- ❌ Commit code without testing

## CRITICAL - ALWAYS DO
- ✅ All testing in containers (Docker or Incus)
- ✅ Use `make test` for unit tests
- ✅ Use `./tests/run_tests.sh` for integration tests
- ✅ Prefer Incus for full OS testing (systemd support)
- ✅ Test content negotiation (HTML, JSON, text)
- ✅ Test all 8 platforms before release

## TEST SCRIPTS
| Script | Purpose | Container |
|--------|---------|-----------|
| `make test` | Unit tests | Docker golang:alpine |
| `tests/run_tests.sh` | Auto-detect and run | Incus or Docker |
| `tests/docker.sh` | Quick Docker tests | Docker alpine |
| `tests/incus.sh` | Full OS tests | Incus debian/ubuntu |

## TESTING PREFERENCE
```
Incus (PREFERRED) > Docker
```

- **Incus**: Full OS, systemd, persistent, SSH-able
- **Docker**: Ephemeral, fast, limited environment

## CONTAINER TESTING WORKFLOW
```bash
# 1. Build
make dev

# 2. Quick test (Docker)
./tests/docker.sh

# 3. Full test (Incus - PREFERRED)
./tests/incus.sh

# 4. Unit tests
make test
```

## AI AS BETA TESTER
| Role | Description |
|------|-------------|
| Find bugs | Try edge cases, invalid inputs |
| Break it | Stress test, race conditions |
| Fix it | Implement the fix immediately |
| Verify | Re-test to confirm fix |

## CONTENT NEGOTIATION TESTING
```bash
# HTML (browser)
curl -H "Accept: text/html" http://localhost/healthz

# JSON (API)
curl -H "Accept: application/json" http://localhost/healthz

# Plain text (CLI)
curl -H "Accept: text/plain" http://localhost/healthz
```

## DOCUMENTATION (PART 30)
- MkDocs for ReadTheDocs
- Keep docs/ in sync with code
- Required pages: index, installation, configuration, api, cli, admin

## I18N & A11Y (PART 31)
- Support multiple locales (AI.md + IDEA.md languages)
- WCAG 2.1 AA compliance
- Touch targets minimum 44x44px
- Screen reader compatibility

## DEBUG TOOLS IN CONTAINERS
```bash
apk add --no-cache curl bash file jq
```

---
For complete details, see AI.md PART 29, 30, 31

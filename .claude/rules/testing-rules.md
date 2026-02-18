# Testing Rules (PART 29, 30, 31)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Run Go locally → use containers
- ❌ Test on host machine → use Docker/Incus
- ❌ Skip content negotiation tests
- ❌ Skip accessibility testing

## REQUIRED - ALWAYS DO
- ✅ Container-only development
- ✅ Test all response formats (HTML, JSON, text)
- ✅ WCAG 2.1 AA compliance
- ✅ ReadTheDocs documentation

## TESTING WORKFLOW
```bash
# Unit tests (via Makefile)
make test

# Integration tests
./tests/run_tests.sh        # Auto-detects incus/docker

# Full OS testing (PREFERRED)
./tests/incus.sh            # Systemd, real services
```

## CONTAINER PREFERENCE
| Container | Use For |
|-----------|---------|
| **Incus** (preferred) | Full OS, systemd, SSH |
| Docker | Quick checks, ephemeral |

## CONTENT NEGOTIATION TESTING
Test all Accept headers:
- `text/html` → HTML response
- `application/json` → JSON response
- `text/plain` → Plain text
- `*/*` (browsers) → HTML
- No Accept (curl) → Plain text

## READTHEDOCS (PART 30)
- MkDocs with Material theme
- mkdocs.yml in project root
- docs/ directory structure
- .readthedocs.yaml config

## I18N & A11Y (PART 31)
- WCAG 2.1 AA compliance
- 7 locales minimum: en, es, fr, de, ar, ja, zh
- RTL support for Arabic
- Semantic HTML
- ARIA labels where needed
- 44x44px touch targets

---
**Full details: AI.md PART 29, PART 30, PART 31**

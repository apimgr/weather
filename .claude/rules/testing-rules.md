# Testing Rules (PART 29, 30, 31)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Run tests on local machine (use containers)
- ❌ Skip integration testing
- ❌ Leave TODO/FIXME in code
- ❌ Commit without testing
- ❌ Skip accessibility testing

## REQUIRED - ALWAYS DO
- ✅ Container-only testing (Docker/Incus)
- ✅ Unit tests with make test
- ✅ Integration tests in tests/ directory
- ✅ ReadTheDocs documentation in docs/
- ✅ i18n support (locale files)
- ✅ WCAG 2.1 AA accessibility

## TESTING WORKFLOW
```bash
# Unit tests
make test

# Integration tests (auto-detects container runtime)
./tests/run_tests.sh

# Docker-based testing
./tests/docker.sh

# Full OS testing (PREFERRED)
./tests/incus.sh
```

## CONTAINER PREFERENCE
| Container | Best For |
|-----------|----------|
| Incus (PREFERRED) | Full integration, systemd, persistent |
| Docker (fallback) | Quick checks, ephemeral |

## AI AS BETA TESTER
When testing, AI should:
- Find bugs (edge cases, invalid inputs)
- Break it (stress test, race conditions)
- Fix it (implement the fix)
- Verify fix (re-test to confirm)

## READTHEDOCS (PART 30)
```
docs/
├── index.md           # Homepage
├── installation.md    # Install guide
├── configuration.md   # Config reference
├── api.md            # API docs
├── cli.md            # CLI reference
├── admin.md          # Admin guide
├── development.md    # Dev guide
├── stylesheets/      # Theme customization
└── requirements.txt  # Python deps
```

## I18N (PART 31)
- Locale files in src/locales/*.json
- I18nService with T() method
- ParseAcceptLanguage() for detection
- Template integration: {{ t .Lang "key" }}

## A11Y (PART 31)
- WCAG 2.1 AA compliance
- Touch targets 44x44px minimum
- Keyboard navigation
- Screen reader support
- Color contrast ratios

---
**Full details: AI.md PART 29, PART 30, PART 31**

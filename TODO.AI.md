# Weather Project - Task Tracking

## Current Sprint: TEMPLATE.md Compliance & Code Cleanup

**Started:** 2025-12-06
**Goal:** Ensure full TEMPLATE.md compliance, remove legacy references, verify metric/imperial support

---

## Tasks

### ✅ Completed
- [x] Create TODO.AI.md for task tracking
- [x] Remove wttr.in code references
  - [x] Renamed `ParseWttrParams` → `ParseQueryParams` in `src/utils/params.go`
  - [x] Updated comments in `src/renderers/ascii.go`
  - [x] Updated all function calls in `src/handlers/weather.go`
  - [x] Kept LICENSE.md attribution (acknowledgment only)
- [x] Verify metric/imperial support
  - [x] Confirmed `weather.html` template handles metric/imperial conditionals
  - [x] Verified all renderers (ascii.go, oneline.go) use unit conversion functions
  - [x] Confirmed unit helpers: getTemperatureUnit(), getSpeedUnit(), getPrecipitationUnit(), getPressureUnit()
- [x] Migrate CLAUDE.md → AI.md
  - [x] Created AI.md with TEMPLATE.md header
  - [x] Copied full CLAUDE.md content
  - [x] Updated all SPEC.md → TEMPLATE.md references
  - [x] Added "Recent Changes" section
  - [x] Deleted CLAUDE.md
- [x] Update days query parameter format
  - [x] Changed from `?0`, `?1`, `?2`, `?3` to `?days=N`
  - [x] Added `MaxForecastDays = 16` constant
  - [x] Implemented max capping logic (days > 16 → 16)
  - [x] Negative values handled (days < 0 → 0)
  - [x] Removed digit cases from combined flags

### 📋 Next Steps

- Docker build/test (requires Docker permissions setup)
- Consider adding unit tests for ParseQueryParams function
- Update any documentation that might reference the old function name

---

## Notes

- **TEMPLATE.md is the foundation** - weather project builds on it, doesn't extend it
- **Docker-first development** - all build/test/debug must use Docker
- **No AI attribution** in code, commits, or docs (per TEMPLATE.md line 1688-1693)

---

## Decision Log

**2025-12-06:** Keep wttr.in attribution in LICENSE.md as acknowledgment (inspiration for format design)

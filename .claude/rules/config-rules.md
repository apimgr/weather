# Configuration Rules (PART 5, 6, 12)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use `strconv.ParseBool()` - use `config.ParseBool()` for ALL booleans
- ❌ Inline YAML comments (must be above the setting)
- ❌ Use `.yaml` extension (always `.yml`)
- ❌ Hardcode dev machine values (hostname, IP, etc.)
- ❌ Store passwords in config files (database only)
- ❌ Skip path validation (traversal attacks)
- ❌ Use external cron/scheduler

## CRITICAL - ALWAYS DO
- ✅ All booleans via `config.ParseBool()` - accepts yes/no, true/false, 1/0, etc.
- ✅ YAML comments above settings, never inline
- ✅ Use `server.yml` for config file
- ✅ Detect machine settings at runtime (hostname, IP, CPU, memory)
- ✅ Validate and normalize ALL paths (prevent traversal)
- ✅ Support both single instance (SQLite) and cluster (remote DB) modes
- ✅ Implement self-healing for critical errors

## BOOLEAN VALUES (config.ParseBool)
| Truthy | Falsy |
|--------|-------|
| 1, yes, true, on, enable, enabled | 0, no, false, off, disable, disabled |
| y, yep, yup, yeah, affirmative | n, nope, nah, nay, negative |
| aye, si, oui, da, hai | nein, non, niet, iie |

## YAML COMMENT STYLE
```yaml
# CORRECT - comment above
enabled: true

# WRONG - inline comment
enabled: true  # this is wrong
```

## APPLICATION MODES (PART 6)
| Mode | Value | Purpose |
|------|-------|---------|
| Production | `production` | Default, no debug endpoints |
| Development | `development` | Debug endpoints enabled |

## PATH SECURITY
- Use `SafePath()` for all user-provided paths
- Block path traversal (`..`)
- Normalize paths (`//` → `/`)
- Validate path segments (lowercase alphanumeric, hyphens, underscores)

## CONFIG STORAGE
| Mode | Source of Truth | server.yml Role |
|------|-----------------|-----------------|
| Single Instance | server.yml | Primary configuration |
| Cluster Mode | Database | Bootstrap + cache/backup |

---
For complete details, see AI.md PART 5, 6, 12

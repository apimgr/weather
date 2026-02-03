# Configuration Rules (PART 5, 6, 12)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use .yaml extension (ALWAYS .yml)
- ❌ Use strconv.ParseBool() (ALWAYS config.ParseBool())
- ❌ Put inline comments in YAML (comments ABOVE code only)
- ❌ Hardcode config values from dev machine
- ❌ Store passwords in config files (database only)
- ❌ Store config files in git repo (runtime-generated)
- ❌ Allow path traversal (../) in config paths

## REQUIRED - ALWAYS DO
- ✅ Config file is ALWAYS `server.yml`
- ✅ All comments ABOVE the setting, NEVER inline
- ✅ Use config.ParseBool() for ALL boolean parsing
- ✅ Detect machine settings at runtime
- ✅ Normalize and validate all paths
- ✅ Support both production and development modes
- ✅ Admin WebUI for ALL settings (no SSH required)
- ✅ Live reload for config changes (no restart)

## BOOLEAN PARSING
Accept 40+ variations via config.ParseBool():
- Truthy: yes, true, 1, on, enable, enabled, allow, active, y, t
- Falsy: no, false, 0, off, disable, disabled, deny, inactive, n, f

## APPLICATION MODES
| Mode | Detection | Behavior |
|------|-----------|----------|
| production | Default, or `--mode production` | Strict security, no debug |
| development | `--mode development` | Verbose logging, debug endpoints |

## PATH NORMALIZATION
```go
// ALWAYS normalize and validate paths
safe, err := SafePath(input)
if err != nil {
    return err // Reject path traversal
}
```

| Input | Result |
|-------|--------|
| `/admin/` | `admin` |
| `//admin//test` | `admin/test` |
| `/../admin` | REJECTED |
| `/admin/..` | REJECTED |

## YAML COMMENT STYLE
```yaml
# CORRECT: Comment above
enabled: true

# WRONG: enabled: true  # inline comment
```

---
**Full details: AI.md PART 5, PART 6, PART 12**

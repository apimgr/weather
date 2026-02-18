# Configuration Rules (PART 5, 6, 12)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use strconv.ParseBool() → use config.ParseBool()
- ❌ Inline YAML comments → comments go ABOVE the setting
- ❌ Hardcode dev machine values → detect at runtime
- ❌ Accept path traversal (../) in any path input
- ❌ Use server.yaml → use server.yml

## REQUIRED - ALWAYS DO
- ✅ Use config.ParseBool() for ALL boolean parsing
- ✅ Use path.Clean() and validate all paths
- ✅ Detect hostname/IP/cores/memory at runtime
- ✅ YAML comments ABOVE settings, never inline
- ✅ Support MODE env var (production/development)
- ✅ Auto-migrate server.yaml → server.yml

## BOOLEAN HANDLING
Accept ALL of these (case-insensitive):
| Truthy | Falsy |
|--------|-------|
| 1, yes, true, on, enable, enabled | 0, no, false, off, disable, disabled |
| yep, yup, yeah, aye, si, oui, da | nope, nah, nay, nein, non, niet |
| affirmative, accept, allow, grant | negative, reject, block, revoke |

## APPLICATION MODES
| Mode | Detection | Behavior |
|------|-----------|----------|
| **production** | Default, `MODE=production` | Secure defaults, minimal logging |
| **development** | `MODE=development` | Debug endpoints, verbose logging |

## PATH SECURITY
```go
// REQUIRED for ALL path inputs
safe, err := SafePath(input)
if err != nil {
    return fmt.Errorf("invalid path: %w", err)
}
```

Block: `..`, `%2e`, uppercase letters, special chars

## ENV VARIABLES
| Variable | Description |
|----------|-------------|
| MODE | production (default) or development |
| NO_COLOR | Disable ANSI colors when set |
| TERM=dumb | Disable ALL ANSI escapes |
| DOMAIN | FQDN override |
| DATABASE_URL | Database connection string |

---
**Full details: AI.md PART 5, PART 6, PART 12**

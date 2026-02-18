# Backend Rules (PART 9, 10, 11, 32)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use bcrypt → use Argon2id for passwords
- ❌ Store passwords/tokens in plaintext
- ❌ Use string concatenation for SQL → parameterized queries
- ❌ Expose stack traces to users
- ❌ Use mattn/go-sqlite3 → use modernc.org/sqlite

## REQUIRED - ALWAYS DO
- ✅ Argon2id for password hashing
- ✅ SHA-256 for token hashing
- ✅ Parameterized SQL queries
- ✅ Rate limiting on all endpoints
- ✅ Tor hidden service auto-enabled when Tor found
- ✅ Pure Go SQLite (CGO_ENABLED=0)

## PASSWORD HASHING (Argon2id)
```go
// CORRECT - Use Argon2id
import "golang.org/x/crypto/argon2"

hash := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)

// NEVER - No bcrypt
import "golang.org/x/crypto/bcrypt" // ❌ FORBIDDEN
```

## TOKEN STORAGE
| Type | Storage | Example |
|------|---------|---------|
| Passwords | Argon2id hash | Admin/user passwords |
| API tokens | SHA-256 hash | `adm_`, `usr_`, `org_` prefixed |
| Sessions | Random + SHA-256 | Session IDs |

## DATABASE
| Driver | Use Case |
|--------|----------|
| SQLite | Default, single instance |
| PostgreSQL | Cluster mode |
| MySQL/MariaDB | Cluster mode |

## ERROR MESSAGES
| Audience | Detail Level |
|----------|--------------|
| User (WebUI/API) | Minimal, helpful |
| Admin (Panel) | Actionable, no stack traces |
| Console/Logs | Full detail with context |

## TOR HIDDEN SERVICE
- Auto-enabled when Tor binary found at startup
- Binary controls Tor lifecycle
- .onion address generated and stored
- All routes accessible via .onion

---
**Full details: AI.md PART 9, PART 10, PART 11, PART 32**

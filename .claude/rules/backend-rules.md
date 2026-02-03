# Backend Rules (PART 9, 10, 11, 32)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use bcrypt for passwords (ALWAYS Argon2id)
- ❌ Store passwords in plaintext
- ❌ Log passwords, tokens, or secrets
- ❌ Expose internal errors to users
- ❌ Use SQL string concatenation (parameterized only)
- ❌ Skip input validation
- ❌ Use mattn/go-sqlite3 (use modernc.org/sqlite)

## REQUIRED - ALWAYS DO
- ✅ Argon2id for password hashing
- ✅ SHA-256 for API token hashing
- ✅ Parameterized SQL queries ONLY
- ✅ Validate and sanitize ALL input
- ✅ SQLite default: server.db and users.db
- ✅ Support Valkey/Redis for caching/clustering
- ✅ Tor hidden service (auto-enabled when Tor found)
- ✅ Structured JSON logging in production

## PASSWORD HASHING (Argon2id)
```go
// OWASP 2023 parameters
ArgonTime    = 3
ArgonMemory  = 64 * 1024  // 64 MB
ArgonThreads = 4
ArgonKeyLen  = 32
ArgonSaltLen = 16
```

## TOKEN HASHING
```go
// API tokens use SHA-256 (fast lookup needed)
hash := sha256.Sum256([]byte(token))
```

## DATABASE DRIVERS
| Database | Library | Driver |
|----------|---------|--------|
| SQLite | modernc.org/sqlite | sqlite |
| PostgreSQL | github.com/jackc/pgx/v5 | pgx |
| MySQL | github.com/go-sql-driver/mysql | mysql |
| Valkey/Redis | github.com/redis/go-redis/v9 | - |

## ERROR MESSAGES
| Context | Detail Level |
|---------|--------------|
| User | Minimal, helpful |
| Admin | Actionable, no stack traces |
| Logs | Full with context |

## TOR HIDDEN SERVICE (PART 32)
- Auto-enabled when Tor binary found
- Server controls startup (not container)
- Health monitoring via scheduler

---
**Full details: AI.md PART 9, PART 10, PART 11, PART 32**

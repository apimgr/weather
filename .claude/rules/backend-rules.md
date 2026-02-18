# Backend Rules (PART 9, 10, 11, 32)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Use bcrypt for passwords - use Argon2id
- ❌ Store plaintext passwords anywhere
- ❌ Use `mattn/go-sqlite3` (requires CGO) - use `modernc.org/sqlite`
- ❌ Use `lib/pq` for PostgreSQL - use `jackc/pgx/v5`
- ❌ Log passwords or tokens (even hashed)
- ❌ Reveal internal errors to users
- ❌ Use external cron/scheduler

## CRITICAL - ALWAYS DO
- ✅ Argon2id for password hashing (OWASP 2023 params)
- ✅ SHA-256 for API token hashing (fast lookup needed)
- ✅ `modernc.org/sqlite` for SQLite (pure Go, no CGO)
- ✅ `jackc/pgx/v5` for PostgreSQL
- ✅ Parameterized queries only (never string concatenation)
- ✅ Separate error messages by audience (user, admin, console, log)
- ✅ Built-in scheduler for all background tasks
- ✅ Support Valkey/Redis for caching and clustering

## PASSWORD HASHING (Argon2id)
```go
const (
    ArgonTime    = 3        // Iterations
    ArgonMemory  = 64*1024  // 64 MB
    ArgonThreads = 4        // Parallelism
    ArgonKeyLen  = 32       // Output length
    ArgonSaltLen = 16       // Salt length
)
```

## DATABASE DRIVERS
| Database | Driver | Config Aliases |
|----------|--------|----------------|
| SQLite | `modernc.org/sqlite` | sqlite, sqlite2, sqlite3 |
| PostgreSQL | `jackc/pgx/v5` | postgres, pgsql, postgresql |
| MySQL | `go-sql-driver/mysql` | mysql, mariadb |
| libSQL | `tursodatabase/libsql-client-go` | libsql, turso |

## ERROR MESSAGES BY CONTEXT
| Audience | Detail Level | Example |
|----------|--------------|---------|
| User | Minimal, helpful | "Invalid credentials" |
| Admin | Actionable | "Login failed for user@example.com" |
| Console | Full | Stack traces, full context |
| Log | Structured | JSON with request_id, timestamps |

## TOR HIDDEN SERVICE (PART 32)
- Auto-enabled when Tor binary found on system
- Binary controls Tor startup (not Docker)
- Generate .onion address on first run
- Store keys in `{data_dir}/tor/`

## RATE LIMITING
| Endpoint | Default | Window |
|----------|---------|--------|
| Login | 5 attempts | 15 min |
| Password reset | 3 attempts | 1 hour |
| API (auth) | Configurable | 1 min |
| API (anon) | Configurable | 1 min |

---
For complete details, see AI.md PART 9, 10, 11, 32

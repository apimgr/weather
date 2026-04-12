# Configuration Rules (PART 5, 6, 12)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO

- Never use `strconv.ParseBool()` — always use `config.ParseBool()`
- Never place YAML comments inline — always above the setting
- Never hardcode host, IP, or port in project code
- Never display `GET /api/` without full URL (`{proto}://{fqdn}:{port}/path`)
- Never include `:80` or `:443` in displayed URLs
- Never store passwords in `server.yml`
- Never mix `config_dir` and `data_dir` purposes
- Never use `strconv.ParseBool()` anywhere in the codebase

## CRITICAL - ALWAYS DO

- Use `config.ParseBool()` or `config.IsTruthy()` for all boolean parsing
- Config file is ALWAYS `server.yml` (never `.yaml`)
- Normalize and validate all file paths from user input
- Apply path traversal middleware FIRST in the chain (before auth, before routing)
- Detect `{proto}`, `{fqdn}`, `{port}` dynamically from request context
- Strip `:80` for HTTP and `:443` for HTTPS in all URLs
- Support Valkey/Redis for cluster and mixed-mode deployments

## Config File Rules

| Rule | Detail |
|-------|-------|
| Filename | `server.yml` (NEVER `.yaml`) |
| YAML comments | ALWAYS above the setting (NEVER inline) |
| Passwords | NEVER stored here |
| Single instance | `server.yml` is source of truth |
| Cluster mode | `server.yml` is cache+backup; database is source of truth |

## Directory Purposes

| Directory | Variable | Contents |
|-----------|----------|---------|
| Config | `{config_dir}` | `server.yml`, TLS certs, security data |
| Data | `{data_dir}` | SQLite databases, user uploads |
| Logs | `{log_dir}` | Log files |

## Boolean Parsing (CRITICAL)

```go
// ALWAYS use this:
config.ParseBool(value)
config.IsTruthy(value)

// NEVER use this:
strconv.ParseBool(value)  // Only handles true/false/1/0/t/f — too restrictive
```

`config.ParseBool()` accepts: `true`, `false`, `yes`, `no`, `on`, `off`, `1`, `0`, `enabled`, `disabled`

## URL Construction Rules

| Rule | Example |
|------|---------|
| Always full URL | `https://example.com:8080/api/v1/users` |
| Strip :80 (HTTP) | `http://example.com/path` (not `:80`) |
| Strip :443 (HTTPS) | `https://example.com/path` (not `:443`) |
| Never hardcode host | Detect from request context |

## Path Security (Middleware Order)

1. Path normalization/validation middleware (FIRST)
2. Authentication middleware
3. Routing

```go
// When constructing file paths from user input, always validate:
// result stays within bounds of allowed directory
```

## Cluster Mode

| Mode | Source of Truth | Valkey/Redis |
|------|----------------|-------------|
| Single Instance (SQLite) | `server.yml` | Optional |
| Cluster (Remote DB) | Database | REQUIRED |
| Mixed Mode | Database | REQUIRED |

## Port Behavior

- Default port is a **random unused port in 64000-64999 range**
- Selected ONCE on first run when no port configured; saved to `server.yml`
- Port persists across restarts — random selection ONLY on first run
- Ports <1024 require elevated privileges (service mode only, drops after binding)

## Security Data

- Security databases are NEVER embedded in the binary
- Downloaded on first run to `{config_dir}/security/`
- Updated via built-in scheduler

## Reference

For complete details, see AI.md PART 5 (lines 6463-8374), PART 6 (8375-8982), PART 12 (15636-16762)

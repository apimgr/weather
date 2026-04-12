# Backend Rules (PART 9, 10, 11, 32)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO

- Never expose sensitive data in /healthz response
- Never log passwords (even hashed)
- Never redirect admin login to user routes (or vice versa)
- Never keep legacy/old API endpoints -- delete them
- Never use bcrypt for passwords -- use Argon2id
- Never use strconv.ParseBool() -- use config.ParseBool()

## CRITICAL - ALWAYS DO

- Support cluster mode with config sync
- Use connection pooling for all database connections
- Set timeouts on all database queries
- Include Request ID in every request for tracing
- Log all errors with context
- Serve a valid security.txt file
- Use raw text only in log FILES (no icons, no ASCII art)
- Support Valkey/Redis for cluster and mixed-mode deployments

## Error Code Standards

| Code | HTTP | Message |
|------|------|---------|
| UNAUTHORIZED | 401 | "Unauthorized" |
| TOKEN_EXPIRED | 401 | "Token expired" |
| TOKEN_INVALID | 401 | "Invalid token" |
| 2FA_REQUIRED | 401 | "Two-factor authentication required" |
| 2FA_INVALID | 401 | "Invalid 2FA code" |

## Healthz Endpoint Rules

- /healthz is PUBLIC -- never expose sensitive data
- Database/cache checks MUST be vague (pass/fail only)

| Category | NEVER Include |
|----------|--------------|
| Internal IPs | Internal network addresses |
| File paths | Server filesystem paths |
| Credentials | Any secrets or tokens |
| DB queries | Query details or schemas |

## Logging Rules

**ALWAYS Log:**
- Request ID, method, path, status, duration
- Auth events (login, logout, failures)
- Admin actions with username and IP
- Error events with context

**NEVER Log:**
- Passwords (even hashed)
- API tokens (even hashed)
- PII beyond what is necessary

**Log Format:** Raw text only. No emoji, no icons, no special characters in log files.

## Authentication Routing

- Admin login NEVER redirects to user routes
- User login NEVER redirects to admin routes
- Auth routes under /auth/ -- NEVER at root (/login, /register)

## Security.txt

ALL projects MUST serve a valid security.txt at /.well-known/security.txt

## Cluster Mode

- ALL projects MUST support cluster mode with config sync
- Valkey/Redis is REQUIRED for cluster or mixed-mode state sync
- All databases MUST have identical schema across nodes

## Tor Hidden Service (PART 32)

- Hidden service is ALWAYS enabled if Tor binary is found
- NEVER use default Tor ports (9050, 9051)
- The application MUST start its OWN dedicated Tor process -- NEVER use system Tor
- Server NEVER fails to start due to Tor issues
- Tor runs as the same user the server runs as (after privilege drop)
- Ports are NEVER hardcoded -- always detected at runtime
- Tor dirs: {config_dir}/tor/, {data_dir}/tor/, {log_dir}/tor.log

## Reference

For complete details, see AI.md PART 9 (12821-13197), PART 10 (13198-13743), PART 11 (13744-15635), PART 32 (43317-45095)

# API Rules (PART 13, 14, 15)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Different response format for same endpoint
- ❌ Expose internal paths or IPs in errors
- ❌ Return stack traces to users
- ❌ Skip content negotiation
- ❌ Use inconsistent error format

## REQUIRED - ALWAYS DO
- ✅ /healthz endpoint (web) + /api/v1/healthz (API)
- ✅ Content negotiation (HTML for browsers, JSON for API)
- ✅ OpenAPI/Swagger at /openapi, /openapi.json
- ✅ GraphQL at /graphql
- ✅ Every web page has corresponding API endpoint
- ✅ SSL/TLS with Let's Encrypt support
- ✅ Trailing newline on all responses

## ENDPOINT PATTERN
| Web Route (HTML) | API Route (JSON) |
|------------------|------------------|
| / | /api/v1/ |
| /healthz | /api/v1/healthz |
| /{admin_path}/dashboard | /api/v1/{admin_path}/dashboard |
| /openapi | /openapi.json |

## HEALTH CHECK RESPONSE
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "24h30m15s",
  "checks": {
    "database": "ok",
    "cache": "ok"
  }
}
```

## ERROR RESPONSE FORMAT
```json
{
  "error": "User-friendly message",
  "code": "ERROR_CODE",
  "details": {}
}
```

## SSL/TLS (PART 15)
- Let's Encrypt auto-renewal
- Self-signed fallback for development
- Certificate paths: ssl/letsencrypt/, ssl/local/
- 7-day expiry warning

## RATE LIMITING
| Endpoint | Limit |
|----------|-------|
| Login | 5/15min |
| API (auth) | Configurable/min |
| API (anon) | Configurable/min |
| Registration | 5/hour |

---
**Full details: AI.md PART 13, PART 14, PART 15**

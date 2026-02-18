# API Rules (PART 13, 14, 15)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Missing corresponding API for web pages
- ❌ Inconsistent JSON response format
- ❌ Skip CSRF tokens on forms
- ❌ Missing rate limiting

## REQUIRED - ALWAYS DO
- ✅ Every web page has corresponding API endpoint
- ✅ Content negotiation (HTML for browsers, JSON for API)
- ✅ Standard error response format
- ✅ /healthz and /api/v1/healthz endpoints
- ✅ SSL/TLS with Let's Encrypt support

## ENDPOINT PATTERN
| Web Route (HTML) | API Route (JSON) |
|------------------|------------------|
| `/` | `/api/v1/` |
| `/healthz` | `/api/v1/healthz` |
| `/admin/dashboard` | `/api/v1/admin/dashboard` |
| `/quotes` | `/api/v1/quotes` |

## HEALTH ENDPOINT
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "mode": "production",
  "uptime": "2d 5h 30m",
  "checks": {
    "database": "ok",
    "disk": "ok"
  }
}
```

## ERROR RESPONSE FORMAT
```json
{
  "error": "Brief error message",
  "code": "ERROR_CODE",
  "status": 400,
  "message": "Detailed explanation"
}
```

## CONTENT NEGOTIATION
| Accept Header | Response |
|---------------|----------|
| text/html | HTML page |
| application/json | JSON |
| text/plain | Plain text |
| */* (browsers) | HTML |
| curl/wget (no Accept) | Plain text |

## SSL/TLS
- Let's Encrypt auto-provisioning
- HTTP-01 challenge on port 80
- TLS-ALPN-01 challenge on port 443
- Auto-renewal via scheduler

---
**Full details: AI.md PART 13, PART 14, PART 15**

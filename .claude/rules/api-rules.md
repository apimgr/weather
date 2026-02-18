# API Rules (PART 13, 14, 15)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Expose /metrics publicly (internal only)
- ❌ Return detailed errors to users (internal info leak)
- ❌ Use client-side routing (SPA)
- ❌ Skip content negotiation
- ❌ Hardcode API version in code

## CRITICAL - ALWAYS DO
- ✅ Every web page has corresponding API endpoint
- ✅ Every API endpoint's data can be displayed in web page
- ✅ Content negotiation: HTML for browsers, JSON for API clients, text for CLI
- ✅ /healthz for public status (limited info)
- ✅ /metrics for Prometheus (internal only)
- ✅ OpenAPI/Swagger documentation at /openapi
- ✅ GraphQL endpoint at /api/v1/graphql
- ✅ SSL/TLS with Let's Encrypt support

## ENDPOINT PATTERN
| Web Route (HTML) | API Route (JSON) | Purpose |
|------------------|------------------|---------|
| `/` | `/api/v1/` | Homepage / API root |
| `/healthz` | `/api/v1/healthz` | Health status |
| `/{admin_path}/dashboard` | `/api/v1/{admin_path}/dashboard` | Admin dashboard |
| `/openapi` | `/openapi.json` | API documentation |
| `/graphql` | `/api/v1/graphql` | GraphQL endpoint |

## HEALTH ENDPOINTS
| Endpoint | Access | Format | Content |
|----------|--------|--------|---------|
| `/healthz` | PUBLIC | HTML/JSON/text | Status, version, uptime |
| `/api/v1/healthz` | PUBLIC | JSON only | Same as above |
| `/metrics` | INTERNAL | Prometheus | All metrics |

## CONTENT NEGOTIATION
| Accept Header | Response Format |
|---------------|-----------------|
| text/html | HTML page |
| application/json | JSON response |
| text/plain | Plain text (for curl, CLI) |
| */* (default) | Based on User-Agent |

## SSL/TLS (PART 15)
- Let's Encrypt auto-renewal
- HTTP-01 challenge on port 80
- TLS-ALPN-01 challenge on port 443
- Fallback to self-signed for localhost
- Certificate storage in `{config_dir}/ssl/`

## API RESPONSE FORMAT
```json
{
  "data": {},
  "meta": {
    "request_id": "abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

## ERROR RESPONSE FORMAT
```json
{
  "error": "Brief user message",
  "code": "ERROR_CODE",
  "status": 400
}
```

---
For complete details, see AI.md PART 13, 14, 15

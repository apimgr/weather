# API Rules

@AI.md PART 13, 14, 15: Health, API Structure, SSL/TLS

## Health Endpoint (PART 13)
- `/healthz` - HTML/JSON/text (content negotiation)
- `/api/v1/healthz` - Always JSON
- Status: "healthy", "degraded", "unhealthy"

## API Structure (PART 14)
- Base: `/api/v1/`
- GraphQL: `/api/v1/graphql`
- OpenAPI: `/openapi` and `/openapi.json`
- Plural routes: `/users`, `/items`
- Hyphens for multi-word: `/api-tokens`

## Content Negotiation
| Accept Header | Response |
|---------------|----------|
| text/html | HTML page |
| application/json | JSON |
| text/plain | Plain text |

## Response Format
```json
// Success
{"success": true, "message": "...", "id": "..."}

// Error
{"error": "message", "code": "CODE", "status": 400}

// Pagination
{"data": [...], "pagination": {...}}
```

## SSL/TLS (PART 15)
- Let's Encrypt auto-renewal
- Renew 7 days before expiry
- Support HTTP-01, TLS-ALPN-01, DNS-01

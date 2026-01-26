# Backend Rules

@AI.md PART 9, 10, 11: Error Handling, Database, Security

## Error Handling (PART 9)
- Always return structured errors
- Log errors with context
- Never expose internal errors to users

## Caching
- Use in-memory cache for frequently accessed data
- Cache invalidation on updates
- Configurable TTL

## Database (PART 10)
- SQLite for single instance
- PostgreSQL/MySQL for cluster mode
- Query timeouts required
- Connection pooling configured

## Security (PART 11)
- Argon2id for passwords (NEVER bcrypt)
- SHA-256 for token hashing
- Rate limiting on all endpoints
- CSRF protection on forms
- Security headers required

## Audit Logging
- JSON Lines format
- ULID for event IDs
- Log all auth events

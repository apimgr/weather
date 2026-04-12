# API Rules (PART 13, 14, 15)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO

- Never keep legacy or removed endpoints -- delete them
- Never use unversioned API routes (/api/users -- wrong, use /api/v1/users)
- Never use singular resource names (/api/v1/user -- wrong, use /api/v1/users)
- Never use uppercase in routes (/api/v1/Users -- wrong)
- Never add trailing slash to routes (/api/v1/users/ -- wrong)
- Never manually edit OpenAPI JSON or GraphQL schema
- Never put Swagger/GraphQL files in project root (use src/swagger/ and src/graphql/)
- Never guess external API compatibility -- look up actual documentation first
- Never expose admin_path in API responses or JSON outputs

## CRITICAL - ALWAYS DO

- Version all API routes: /api/v1/...
- Use plural nouns for all resources
- Use lowercase for all routes
- Keep Swagger and GraphQL in sync with each other AND the project API
- Use standardized locations: src/swagger/ and src/graphql/
- Follow all relevant RFCs completely if implementing a protocol
- End all non-HTML responses with a single newline
- Indent all JSON responses
- Test every route with ALL applicable Accept headers

## Route Standards

| Rule | Wrong | Correct |
|------|-------|---------|
| Versioning | /api/users | /api/v1/users |
| Plural nouns | /api/v1/user | /api/v1/users |
| Lowercase | /api/v1/Users | /api/v1/users |
| No trailing slash | /api/v1/users/ | /api/v1/users |

## API Version

- Current API version: v1
- All routes: /api/v1/...

## Response Formats

- JSON: indented, ends with newline
- Text: ends with single newline
- HTML: valid HTML5, indented with 2 spaces, ends with newline

## Content Negotiation (Required)

Every API route MUST be tested with ALL applicable Accept headers:
- application/json
- text/plain
- text/html
- .txt extension on applicable endpoints

## Swagger / OpenAPI

- Location: src/swagger/
- NEVER manually edit -- always auto-generate
- MUST match current project API at all times
- Swagger UI MUST match project-wide theme system

## GraphQL

- Location: src/graphql/
- NEVER manually edit schema manually
- MUST stay in sync with Swagger/OpenAPI
- GraphiQL MUST match project-wide theme system

## Versioning (SemVer)

All stable releases MUST follow semantic versioning (MAJOR.MINOR.PATCH).

## Healthz & Versioning

- /healthz: public, never expose sensitive data
- /api/v1/version: returns build info (all fields must be public-safe)

## RFC Compliance (CRITICAL)

If the application implements an RFC-defined protocol, it MUST follow ALL relevant RFCs completely. This is NOT optional -- non-compliance means the application is fundamentally broken.

## External API Compatibility

When implementing compatibility with external services:
- Research actual API documentation -- NEVER guess
- Focus on creation endpoints and response formats
- Do NOT replicate their entire API surface

## Reference

For complete details, see AI.md PART 13 (16763-17513), PART 14 (17514-19146), PART 15 (19147-20117)

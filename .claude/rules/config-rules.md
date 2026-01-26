# Configuration Rules

@AI.md PART 5, 6, 12: Configuration, Application Modes, Server Settings

## Configuration (PART 5)
- Config file: `server.yml` (NOT .yaml)
- YAML comments ABOVE settings, NEVER inline
- Use `config.ParseBool()` for ALL boolean parsing
- Path normalization required for ALL paths

## Path Security
- Validate ALL paths (config, HTTP, file, API params)
- Block path traversal (`..`, `%2e%2e`)
- PathSecurityMiddleware MUST be first in chain

## Application Modes (PART 6)
| Mode | Detection |
|------|-----------|
| Production | Default, no DEBUG env |
| Development | DEBUG=true or --debug flag |

## Server Settings (PART 12)
- Default port: 64948
- Bind address: `[::]` (all interfaces)
- server.yml is source of truth (single instance)
- Database is source of truth (cluster mode)

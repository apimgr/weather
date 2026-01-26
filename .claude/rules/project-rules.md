# Project Rules

@AI.md PART 2, 3, 4: License, Project Structure, OS-Specific Paths

## License (PART 2)
- MIT License in LICENSE.md
- Embedded 3rd-party licenses at bottom of LICENSE.md

## Project Structure (PART 3)
- `src/` - ALL Go source code
- `docker/` - Dockerfile, docker-compose, file_system/
- `docs/` - MkDocs documentation only
- `tests/` - Test scripts (run_tests.sh, docker.sh, incus.sh)
- `scripts/` - Production/install scripts
- `binaries/` - Build output (gitignored)

## OS-Specific Paths (PART 4)
| OS | Config | Data | Logs |
|----|--------|------|------|
| Linux | `/etc/{org}/{project}` | `/var/lib/{org}/{project}` | `/var/log/{org}/{project}` |
| macOS | `/Library/Application Support/{org}/{project}` | Same | `/Library/Logs/{org}/{project}` |
| Windows | `%PROGRAMDATA%\{org}\{project}` | Same | Same + `\logs` |

## Forbidden
- NO `cmd/` directory
- NO `internal/` directory
- NO `pkg/` directory
- NO root-level `data/`, `config/`, `logs/`

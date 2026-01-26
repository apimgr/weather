# CI/CD Rules

@AI.md PART 28: CI/CD Workflows

## CRITICAL: NO Makefile in CI/CD
- Use explicit commands with env vars
- Makefile is for LOCAL dev only

## Workflow Files
GitHub: `.github/workflows/`
Gitea: `.gitea/workflows/`

## Required Workflows
| Workflow | Trigger | Purpose |
|----------|---------|---------|
| release.yml | Tag v*.*.* | Stable releases |
| beta.yml | Tag beta-* | Beta releases |
| daily.yml | Schedule/manual | Daily builds |
| docker.yml | Release/manual | Docker images |

## Build Matrix
- OS: linux, darwin, windows, freebsd
- Arch: amd64, arm64
- Total: 8 binaries

## Artifacts
- Upload all 8 binaries
- Create checksums
- Sign releases (optional)

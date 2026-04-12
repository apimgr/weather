# Docker Rules (PART 27)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO

- Never place Dockerfile or docker-compose.yml in project root -- always use docker/
- Never modify ENTRYPOINT or CMD -- all customization goes in entrypoint.sh
- Never include build: or version: in docker-compose.yml
- Never use .env files -- hardcode sane defaults directly
- Never push :dev or :test tags to production registry
- Never run docker compose from the project directory -- always use temp dir workflow
- Never create runtime rootfs/ in the project repo
- Never put pre-built binaries in Docker images -- always multi-stage build

## CRITICAL - ALWAYS DO

- Place ALL Docker files in docker/ directory
- Use multi-stage build (no pre-built binaries needed)
- Use entrypoint.sh for all container startup logic
- Include required OCI labels in Dockerfile
- Use x-logging anchor in every docker-compose.yml service
- Hardcode environment variables with sane defaults
- Database names are ALWAYS server.db and users.db (globally consistent)
- Use temp directory workflow for all Docker testing

## Directory Structure

docker/
  Dockerfile
  docker-compose.yml
  docker-compose.dev.yml
  entrypoint.sh
  file_system/            # Build-time rootfs (COPY into image)

## Required OCI Labels (Dockerfile)

- org.opencontainers.image.title
- org.opencontainers.image.description
- org.opencontainers.image.url
- org.opencontainers.image.source
- org.opencontainers.image.licenses
- org.opencontainers.image.version
- org.opencontainers.image.created
- org.opencontainers.image.revision

For multi-arch images, OCI labels MUST also be set as manifest annotations.

## docker-compose.yml Rules

| Setting | Rule |
|---------|------|
| Environment variables | Hardcode with sane defaults (NEVER use .env files) |
| build: section | NEVER include |
| version: field | NEVER include |
| x-logging | ALWAYS include anchor, ALWAYS use in every service |
| rootfs/ volume | NEVER from project dir -- use temp dir |

## Two rootfs/ Contexts

| Context | Location | Purpose |
|---------|----------|---------|
| Build-time | docker/file_system/ | Files COPYd into image (in repo) |
| Runtime | TEMP_DIR/rootfs/ | Volume mounts (config, data) -- NEVER in repo |

## Registry Rules

| Build Type | Registry |
|------------|---------|
| Release builds | PLATFORM_CONTAINER_REGISTRY/apimgr/weather |
| Development builds | Local-only tags (no registry prefix) |
| :dev or :test tags | NEVER push to production registry |

## Environment Variables

ALWAYS hardcode -- NEVER require .env files:
- SERVER_HOST=0.0.0.0
- SERVER_PORT=8080
- DATA_DIR=/data

NEVER use dollar-sign VAR or dollar-sign{VAR:-default} syntax requiring .env.

## AI Testing Workflow

Scripts docker.sh and incus.sh are FOR HUMAN USE ONLY.
AI must use the AI/Automated Testing Workflow (temp dir, containers).
AI must NEVER run docker.sh or incus.sh directly from the project directory.

## Database Names

- Server database: server.db (ALWAYS this name)
- Users database: users.db (ALWAYS this name)

## Reference

For complete details, see AI.md PART 27 (lines 34717-36222)

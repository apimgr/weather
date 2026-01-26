# Docker Rules

@AI.md PART 27: Docker, Containers

## Dockerfile Location
- `docker/Dockerfile` (NEVER in root)
- Multi-stage build required
- Alpine base for production

## Docker Compose Files
- `docker/docker-compose.yml` - Production
- `docker/docker-compose.dev.yml` - Development
- `docker/docker-compose.test.yml` - Testing (DEBUG=true)

## Filesystem Overlay
- `docker/file_system/` - NOT rootfs/
- `docker/file_system/usr/local/bin/entrypoint.sh`

## Container Rules
- NEVER copy/symlink binaries in Dockerfile
- Build from source inside container
- Use temp directory workflow
- STOPSIGNAL: SIGRTMIN+3 (signal 37)

## Environment Variables
- DEBUG=true for development
- No .env files (hardcode defaults)

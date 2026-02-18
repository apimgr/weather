# Docker Rules (PART 27)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Put Dockerfile in project root → `docker/Dockerfile`
- ❌ Use .env files → hardcode defaults in compose
- ❌ Run docker-compose in project directory
- ❌ Copy/symlink binaries in container
- ❌ Add USER directive → binary handles privilege drop

## REQUIRED - ALWAYS DO
- ✅ Multi-stage Dockerfile (builder + runtime)
- ✅ `docker/Dockerfile` location
- ✅ Use temp directory for compose
- ✅ `tini` as init process
- ✅ Include `tor` package (binary controls startup)

## DOCKERFILE STRUCTURE
```dockerfile
# Stage 1: Build
FROM golang:alpine AS builder
# Build binary...

# Stage 2: Runtime
FROM alpine:latest
RUN apk add --no-cache git curl bash tini tor
COPY --from=builder /app/weather /usr/local/bin/weather
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["tini", "-p", "SIGTERM", "--", "/usr/local/bin/entrypoint.sh"]
```

## PORTS
| Type | Port |
|------|------|
| Internal default | 80 |
| External mapping | Random 64xxx → 80 |

## VOLUMES
```yaml
volumes:
  - './rootfs/config:/config:z'
  - './rootfs/data:/data:z'
```

## TEMP DIRECTORY WORKFLOW
1. Create temp dir: `/tmp/apimgr/weather-XXXXXX/`
2. Copy docker-compose.yml to temp
3. Run docker-compose from temp
4. Volumes created in temp (not project dir)

---
**Full details: AI.md PART 27**

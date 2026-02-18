# Docker Rules (PART 27)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Put Dockerfile in project root (must be `docker/Dockerfile`)
- ❌ Copy or symlink binaries into container (build from source)
- ❌ Use .env files (hardcode sane defaults)
- ❌ Run docker-compose in project directory (use temp dir)
- ❌ Add USER directive (binary handles privilege drop)
- ❌ Modify ENTRYPOINT or CMD (use entrypoint.sh)

## CRITICAL - ALWAYS DO
- ✅ Dockerfile location: `docker/Dockerfile`
- ✅ Multi-stage build: golang:alpine → alpine:latest
- ✅ Build from source inside container
- ✅ STOPSIGNAL: SIGRTMIN+3
- ✅ ENTRYPOINT: tini with entrypoint.sh
- ✅ Required packages: git, curl, bash, tini, tor
- ✅ Default timezone: America/New_York (override with TZ)
- ✅ Default internal port: 80

## DOCKERFILE STRUCTURE
```dockerfile
# Stage 1: Build
FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY src/ ./src/
RUN CGO_ENABLED=0 go build -ldflags="..." -o weather ./src

# Stage 2: Runtime
FROM alpine:latest
RUN apk add --no-cache bash curl tini tor tzdata ca-certificates
COPY --from=builder /build/weather /usr/local/bin/
COPY docker/file_system/ /
EXPOSE 80
STOPSIGNAL SIGRTMIN+3
ENTRYPOINT ["tini", "-p", "SIGTERM", "--", "/usr/local/bin/entrypoint.sh"]
```

## DOCKER COMPOSE FILES
| File | Purpose | DEBUG |
|------|---------|-------|
| `docker-compose.yml` | Production | No |
| `docker-compose.dev.yml` | Development | Yes |
| `docker-compose.test.yml` | Testing | Yes |

## PORT MAPPING
```yaml
ports:
  - "64580:80"  # Random 64xxx:internal 80
```

| Context | Address | Port |
|---------|---------|------|
| Container internal | 0.0.0.0 | 80 |
| Container custom | 0.0.0.0 | PORT env |
| Host mapping | - | 64xxx |

## VOLUME MOUNTS
```yaml
volumes:
  - './rootfs/config:/config:z'
  - './rootfs/data:/data:z'
```

## CONTAINER PATHS
| Type | Path |
|------|------|
| Binary | `/usr/local/bin/weather` |
| Config | `/config/weather/server.yml` |
| Data | `/data/weather/` |
| SQLite | `/data/db/sqlite/` |
| Logs | `/data/log/weather/` |

## OCI LABELS (REQUIRED)
```dockerfile
LABEL org.opencontainers.image.title="Weather"
LABEL org.opencontainers.image.source="https://github.com/apimgr/weather"
LABEL org.opencontainers.image.licenses="MIT"
```

---
For complete details, see AI.md PART 27

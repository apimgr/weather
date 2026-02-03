# Docker Rules (PART 27)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- ❌ Put Dockerfile in project root (ALWAYS docker/)
- ❌ Use .env files
- ❌ Run docker-compose in project directory
- ❌ Modify ENTRYPOINT/CMD directly
- ❌ Include rootfs/ in Docker image

## REQUIRED - ALWAYS DO
- ✅ Dockerfile in docker/Dockerfile
- ✅ Multi-stage build (golang:alpine → alpine:latest)
- ✅ STOPSIGNAL SIGRTMIN+3
- ✅ ENTRYPOINT with tini
- ✅ Required packages: git, curl, bash, tini, tor
- ✅ Customization via entrypoint.sh only
- ✅ docker-compose.yml for each environment

## DOCKERFILE LOCATION
```
docker/
├── Dockerfile
├── docker-compose.yml       # Production (NO debug)
├── docker-compose.dev.yml   # Development
├── docker-compose.test.yml  # Testing (DEBUG=true)
└── file_system/             # Build-time overlay
    └── usr/local/bin/entrypoint.sh
```

## CONTAINER PORTS
| Context | Internal | External |
|---------|----------|----------|
| Default | 80 | 64xxx (random) |
| Override | PORT env | -p mapping |

## VOLUME MOUNTS
```yaml
volumes:
  - './rootfs/config:/config:z'
  - './rootfs/data:/data:z'
```

## DOCKER LABELS (REQUIRED)
```dockerfile
LABEL org.opencontainers.image.source="https://github.com/apimgr/weather"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.title="weather"
LABEL org.opencontainers.image.description="Weather API service"
```

## TEMP DIRECTORY WORKFLOW
```bash
# Create temp dir for testing
TMPDIR=$(mktemp -d)
cp docker/docker-compose.yml "$TMPDIR/"
cd "$TMPDIR"
docker compose up -d
# Test...
docker compose down
rm -rf "$TMPDIR"
```

## DOCKER TAGS (CI/CD)
| Event | Tags |
|-------|------|
| Any push | devel, {commit} |
| Beta tag | beta, {commit} |
| Release tag | {version}, latest, YYMM, {commit} |

---
**Full details: AI.md PART 27**

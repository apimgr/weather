# =============================================================================
# Multi-Stage Dockerfile for Weather Service
# AI.md PART 13 Compliant
# =============================================================================

# -----------------------------------------------------------------------------
# Build Stage
# -----------------------------------------------------------------------------
FROM golang:alpine AS builder

# Build arguments (used in ldflags and labels)
ARG TARGETARCH
ARG VERSION=dev
ARG BUILD_DATE
ARG COMMIT_ID

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies (cached layer)
RUN go mod download

# Copy source code
COPY src/ ./src/

# Build static binary for target architecture
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w \
        -X 'main.Version=${VERSION}' \
        -X 'main.CommitID=${COMMIT_ID}' \
        -X 'main.BuildDate=${BUILD_DATE}'" \
    -trimpath \
    -o /build/binary/weather \
    src

# -----------------------------------------------------------------------------
# Runtime Stage
# -----------------------------------------------------------------------------
FROM alpine:latest

# Runtime arguments for labels
ARG VERSION=dev
ARG BUILD_DATE
ARG COMMIT_ID
ARG TARGETARCH

# OCI Standard Labels (NON-NEGOTIABLE)
LABEL org.opencontainers.image.title="weather" \
      org.opencontainers.image.description="Production-grade weather API providing global forecasts, severe weather alerts, earthquake data, moon phases, and hurricane tracking" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${COMMIT_ID}" \
      org.opencontainers.image.vendor="apimgr" \
      org.opencontainers.image.authors="Weather Service Contributors <noreply@github.com>" \
      org.opencontainers.image.url="https://github.com/apimgr/weather" \
      org.opencontainers.image.source="https://github.com/apimgr/weather" \
      org.opencontainers.image.documentation="https://apimgr-weather.readthedocs.io" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.ref.name="${VERSION}" \
      org.opencontainers.image.base.name="alpine:latest"

# Multi-arch manifest annotations
LABEL org.opencontainers.image.platform="linux/${TARGETARCH}"

# Install required runtime packages
# curl - for health checks
# bash - for entrypoint script
# tini - for proper init (PID 1) and signal handling
# tor - for hidden service (auto-enabled when detected)
RUN apk add --no-cache \
    curl \
    bash \
    tini \
    tor

# Create required directory structure
RUN mkdir -p \
    /config \
    /config/security \
    /data/db \
    /data/logs \
    /data/tor \
    /data/backup \
    /data/cache

# Copy binary from builder
COPY --from=builder /build/binary/weather /usr/local/bin/weather

# Copy container overlay (entrypoint.sh)
# Build context is docker/, so COPY rootfs/ not docker/rootfs/
COPY rootfs/ /

# Make binaries executable
RUN chmod 755 /usr/local/bin/weather /usr/local/bin/entrypoint.sh

# Environment variables with sane defaults
# MODE: development (default for containers per AI.md PART 14 - allows localhost, .local, .test, etc.)
# TZ: Timezone for logs and timestamps (default: America/New_York per AI.md)
# NOTE: Tor is auto-enabled (no ENABLE_TOR flag) when tor binary is installed
ENV MODE=development \
    TZ=America/New_York

# Expose internal port (always 80)
EXPOSE 80

# Graceful shutdown signal
# SIGRTMIN+3 (signal 37) for systemd/container compatibility
STOPSIGNAL SIGRTMIN+3

# Health check with long start period for service initialization
HEALTHCHECK --start-period=10m --interval=5m --timeout=15s --retries=3 \
    CMD /usr/local/bin/weather --status || exit 1

# Use tini as init with signal propagation
# -p SIGTERM: propagate SIGTERM to child processes
# Handles SIGRTMIN+3 from Docker STOPSIGNAL
ENTRYPOINT ["tini", "-p", "SIGTERM", "--", "/usr/local/bin/entrypoint.sh"]

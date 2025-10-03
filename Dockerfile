# Multi-stage build for Go weather service
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies (bash, curl, git, ca-certificates, tzdata)
RUN apk add --no-cache bash curl git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build args for version information
ARG VERSION=dev
ARG BUILD_DATE
ARG GIT_COMMIT

# Build static binary with optimization and version info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}" \
    -o weather . && \
    chmod +x weather

# Final stage - minimal runtime image
FROM alpine:latest as base
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/weather /usr/local/bin/weather

RUN apk update --no-cache && apk add --no-cache curl bash

FROM scratch

# Build args for labels
ARG VERSION=dev
ARG BUILD_DATE
ARG GIT_COMMIT

# OCI Standard Labels
LABEL org.opencontainers.image.title="Weather Service" \
      org.opencontainers.image.description="Beautiful weather forecasts, moon phases, earthquakes, and hurricane tracking with authentication and admin dashboard" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.authors="Weather Service Contributors" \
      org.opencontainers.image.url="https://github.com/apimgr/weather" \
      org.opencontainers.image.source="https://github.com/apimgr/weather" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.documentation="https://github.com/apimgr/weather/blob/main/README.md"

# Application-specific labels
LABEL app.weather.features="authentication,database,admin-ui,saved-locations,api-tokens,scheduler,notifications,weather-alerts,hurricane-tracking,pwa" \
      app.weather.database="sqlite" \
      app.weather.framework="gin" \
      app.weather.theme="dracula"

# Copy base to scratch
COPY --from=base / /

# Set working directory
WORKDIR /config

# Environment variables with defaults
ENV PORT=80 \
    GIN_MODE=release \
    SESSION_SECRET="" \
    TZ=${TZ:-America/New_York}

# Expose port
EXPOSE 80

# Create data and config directories
# Note: In production, mount volumes to /data and /config for persistence
VOLUME ["/data", "/config"]

# Health check - uses built-in healthcheck endpoint
HEALTHCHECK --interval=120s --timeout=5s --start-period=90s --retries=3 CMD ["/usr/local/bin/weather", "--healthcheck"] || exit 1

# Start the application with directory-based CLI flags
ENTRYPOINT ["/usr/local/bin/weather"]
CMD ["--data", "/data/weather", "--config", "/config/weather"]

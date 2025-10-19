# Multi-stage build for Go weather service
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
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
    -o weather ./src && \
    chmod +x weather

# Download GeoIP databases (for Docker images only)
# This embeds the databases in the image for faster first startup
# Binary installations still download on first run
RUN mkdir -p /tmp/geoip && \
    echo "📥 Downloading GeoIP databases for Docker image..." && \
    curl -L -o /tmp/geoip/geolite2-city-ipv4.mmdb \
      https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city-mmdb/geolite2-city-ipv4.mmdb && \
    echo "✅ geolite2-city-ipv4.mmdb downloaded" && \
    curl -L -o /tmp/geoip/geolite2-city-ipv6.mmdb \
      https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city-mmdb/geolite2-city-ipv6.mmdb && \
    echo "✅ geolite2-city-ipv6.mmdb downloaded" && \
    curl -L -o /tmp/geoip/geo-whois-asn-country.mmdb \
      https://cdn.jsdelivr.net/npm/@ip-location-db/geo-whois-asn-country-mmdb/geo-whois-asn-country.mmdb && \
    echo "✅ geo-whois-asn-country.mmdb downloaded" && \
    curl -L -o /tmp/geoip/asn.mmdb \
      https://cdn.jsdelivr.net/npm/@ip-location-db/asn-mmdb/asn.mmdb && \
    echo "✅ asn.mmdb downloaded" && \
    echo "✅ All 4 GeoIP databases ready (~103MB)"

# Final stage - Alpine with curl and bash
FROM alpine:latest

# Build args for labels
ARG VERSION=dev
ARG BUILD_DATE
ARG GIT_COMMIT

# OCI Standard Labels
LABEL org.opencontainers.image.title="Weather Service" \
      org.opencontainers.image.description="Beautiful weather forecasts, moon phases, earthquakes, and severe weather tracking with GeoIP, authentication and admin dashboard" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.authors="Weather Service Contributors" \
      org.opencontainers.image.url="https://github.com/apimgr/weather" \
      org.opencontainers.image.source="https://github.com/apimgr/weather" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.documentation="https://github.com/apimgr/weather/blob/main/README.md"

# Application-specific labels
LABEL app.weather.features="authentication,database,admin-ui,saved-locations,api-tokens,scheduler,notifications,weather-alerts,severe-weather,geoip,pwa" \
      app.weather.database="sqlite" \
      app.weather.framework="gin" \
      app.weather.theme="dracula" \
      app.weather.geoip="maxmind-geolite2"

# Copy timezone data and certificates
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Install curl and bash
RUN apk add --no-cache curl bash

# Copy binary to /usr/local/bin
COPY --from=builder /app/weather /usr/local/bin/weather

# Copy GeoIP databases to /config/geoip (pre-downloaded for faster startup)
COPY --from=builder /tmp/geoip /config/geoip

# Set working directory
WORKDIR /config

# Environment variables with defaults
ENV PORT=80 \
    ENV=production \
    TZ=America/New_York \
    DATA_DIR=/data \
    CONFIG_DIR=/config \
    LOG_DIR=/var/log/weather

# Create required directories
RUN mkdir -p /data /config /var/log/weather /data/db && \
    echo "✅ GeoIP databases embedded in image (4 files, ~103MB)"

# Expose port
EXPOSE 80

# Volumes for persistence
VOLUME ["/data", "/config", "/var/log/weather"]

# Health check
HEALTHCHECK --interval=120s --timeout=5s --start-period=90s --retries=3 \
    CMD ["/usr/local/bin/weather", "--healthcheck"] || exit 1

# Start the application
ENTRYPOINT ["/usr/local/bin/weather"]
CMD ["--data", "/data", "--config", "/config"]

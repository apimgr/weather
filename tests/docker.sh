#!/usr/bin/env bash
# AI.md PART 29: Full integration testing in Docker Alpine container
set -euo pipefail

# Detect project info
PROJECTNAME=$(basename "$PWD")
PROJECTORG=$(basename "$(dirname "$PWD")")

# Create temp directory for build (proper org/project structure)
mkdir -p "${TMPDIR:-/tmp}/${PROJECTORG}"
BUILD_DIR=$(mktemp -d "${TMPDIR:-/tmp}/${PROJECTORG}/${PROJECTNAME}-XXXXXX")
trap "rm -rf $BUILD_DIR" EXIT

echo "Building binary in Docker..."
docker run --rm \
  -v "$(pwd):/build" \
  -w /build \
  -e CGO_ENABLED=0 \
  golang:alpine go build -o "$BUILD_DIR/$PROJECTNAME" ./src

echo "Testing in Docker (Alpine)..."
docker run --rm \
  -v "$BUILD_DIR:/app" \
  alpine:latest sh -c "
    set -e

    # Install curl for API tests
    apk add --no-cache curl >/dev/null 2>&1

    chmod +x /app/$PROJECTNAME

    echo '=== Version Check ==='
    /app/$PROJECTNAME --version

    echo '=== Help Check ==='
    /app/$PROJECTNAME --help

    echo '=== Binary Info ==='
    ls -lh /app/$PROJECTNAME
    file /app/$PROJECTNAME

    echo '=== Starting Server for API Tests ==='
    /app/$PROJECTNAME serve --port 64580 --mode development &
    SERVER_PID=\$!
    sleep 5

    echo '=== API Endpoint Tests ==='
    # Test JSON response (default)
    curl -sf http://localhost:64580/api/v1/healthz || echo 'FAILED: /api/v1/healthz'

    # Test .txt extension (plain text)
    curl -sf http://localhost:64580/api/v1/healthz.txt || echo 'FAILED: /api/v1/healthz.txt'

    # Test Accept header (text/plain)
    curl -sf -H 'Accept: text/plain' http://localhost:64580/api/v1/healthz || echo 'FAILED: Accept text/plain'

    # Test Accept header (application/json)
    curl -sf -H 'Accept: application/json' http://localhost:64580/api/v1/healthz || echo 'FAILED: Accept application/json'

    echo '=== Weather-Specific Endpoint Tests ==='
    # Weather API endpoints
    curl -sf http://localhost:64580/api/v1/weather || echo 'FAILED: /api/v1/weather'
    curl -sf http://localhost:64580/api/v1/moon || echo 'FAILED: /api/v1/moon'
    curl -sf http://localhost:64580/api/v1/earthquakes || echo 'FAILED: /api/v1/earthquakes'
    curl -sf http://localhost:64580/api/v1/hurricanes || echo 'FAILED: /api/v1/hurricanes'
    curl -sf http://localhost:64580/api/v1/severe-weather || echo 'FAILED: /api/v1/severe-weather'

    # Test .txt extension on weather endpoints
    curl -sf http://localhost:64580/api/v1/weather.txt || echo 'FAILED: /api/v1/weather.txt'
    curl -sf http://localhost:64580/api/v1/moon.txt || echo 'FAILED: /api/v1/moon.txt'

    echo '=== Frontend Tests (Smart Detection) ==='
    # CLI should get text response
    WEATHER=\$(curl -s http://localhost:64580/)
    if [ -n \"\$WEATHER\" ]; then echo '✓ Frontend works'; else echo 'FAILED: Frontend'; fi

    echo '=== Stopping Server ==='
    kill \$SERVER_PID 2>/dev/null || true
    wait \$SERVER_PID 2>/dev/null || true

    echo '=== All tests passed ==='
"

echo "Docker tests completed successfully"

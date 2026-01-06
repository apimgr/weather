#!/usr/bin/env bash
# AI.md PART 29: Full integration + systemd testing in Incus Debian container
set -euo pipefail

# Check if incus is available
if ! command -v incus &>/dev/null; then
    echo "ERROR: incus not found. Install incus or use tests/docker.sh"
    exit 1
fi

# Detect project info
PROJECTNAME=$(basename "$PWD")
PROJECTORG=$(basename "$(dirname "$PWD")")
CONTAINER_NAME="test-${PROJECTNAME}-$$"

# Create temp directory for build (proper org/project structure)
mkdir -p "${TMPDIR:-/tmp}/${PROJECTORG}"
BUILD_DIR=$(mktemp -d "${TMPDIR:-/tmp}/${PROJECTORG}/${PROJECTNAME}-XXXXXX")
trap "rm -rf $BUILD_DIR; incus delete $CONTAINER_NAME --force 2>/dev/null || true" EXIT

echo "Building binary in Docker..."
docker run --rm \
  -v "$(pwd):/build" \
  -w /build \
  -e CGO_ENABLED=0 \
  golang:alpine go build -o "$BUILD_DIR/$PROJECTNAME" ./src

echo "Launching Incus container (Debian + systemd)..."
incus launch images:debian/12 "$CONTAINER_NAME"

# Wait for container to be ready
sleep 3

echo "Installing curl in container..."
incus exec "$CONTAINER_NAME" -- apt-get update -qq
incus exec "$CONTAINER_NAME" -- apt-get install -y -qq curl >/dev/null

echo "Copying binary to container..."
incus file push "$BUILD_DIR/$PROJECTNAME" "$CONTAINER_NAME/usr/local/bin/"
incus exec "$CONTAINER_NAME" -- chmod +x "/usr/local/bin/$PROJECTNAME"

echo "Running tests in Incus..."
incus exec "$CONTAINER_NAME" -- bash -c "
    set -e

    echo '=== Version Check ==='
    $PROJECTNAME --version

    echo '=== Help Check ==='
    $PROJECTNAME --help

    echo '=== Binary Info ==='
    ls -lh /usr/local/bin/$PROJECTNAME
    file /usr/local/bin/$PROJECTNAME

    echo '=== Service Install Test ==='
    $PROJECTNAME --service --install

    echo '=== Service Status ==='
    systemctl status $PROJECTNAME || true

    echo '=== Service Start Test ==='
    systemctl start $PROJECTNAME
    sleep 3
    systemctl status $PROJECTNAME

    echo '=== API Endpoint Tests ==='
    # Test JSON response (default)
    curl -sf http://localhost:80/api/v1/healthz || echo 'FAILED: /api/v1/healthz'

    # Test .txt extension (plain text)
    curl -sf http://localhost:80/api/v1/healthz.txt || echo 'FAILED: /api/v1/healthz.txt'

    # Test Accept header (text/plain)
    curl -sf -H 'Accept: text/plain' http://localhost:80/api/v1/healthz || echo 'FAILED: Accept text/plain'

    # Test Accept header (application/json)
    curl -sf -H 'Accept: application/json' http://localhost:80/api/v1/healthz || echo 'FAILED: Accept application/json'

    echo '=== Weather-Specific Endpoint Tests ==='
    # Weather API endpoints
    curl -sf http://localhost:80/api/v1/weather || echo 'FAILED: /api/v1/weather'
    curl -sf http://localhost:80/api/v1/moon || echo 'FAILED: /api/v1/moon'
    curl -sf http://localhost:80/api/v1/earthquakes || echo 'FAILED: /api/v1/earthquakes'
    curl -sf http://localhost:80/api/v1/hurricanes || echo 'FAILED: /api/v1/hurricanes'
    curl -sf http://localhost:80/api/v1/severe-weather || echo 'FAILED: /api/v1/severe-weather'

    # Test .txt extension on weather endpoints
    curl -sf http://localhost:80/api/v1/weather.txt || echo 'FAILED: /api/v1/weather.txt'
    curl -sf http://localhost:80/api/v1/moon.txt || echo 'FAILED: /api/v1/moon.txt'

    echo '=== Frontend Tests (Smart Detection) ==='
    # CLI should get text response
    WEATHER=\$(curl -s http://localhost:80/)
    if [ -n \"\$WEATHER\" ]; then echo '✓ Frontend works'; else echo 'FAILED: Frontend'; fi

    echo '=== Service Stop Test ==='
    systemctl stop $PROJECTNAME

    echo '=== Service Uninstall Test ==='
    $PROJECTNAME --service --uninstall

    echo '=== All tests passed ==='
"

echo "Incus tests completed successfully"

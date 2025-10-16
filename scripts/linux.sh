#!/usr/bin/env bash
# Weather Service - Linux Installer with systemd service

set -e

VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
SERVICE_USER="${SERVICE_USER:-weather}"
DATA_DIR="${DATA_DIR:-/var/lib/weather}"
CONFIG_DIR="${CONFIG_DIR:-/etc/weather}"
REPO="apimgr/weather"
BINARY_NAME="weather"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🌤️  Weather Service - Linux Installer"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}❌ This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Detect architecture
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64)     ARCH_TYPE=amd64;;
    amd64)      ARCH_TYPE=amd64;;
    arm64)      ARCH_TYPE=arm64;;
    aarch64)    ARCH_TYPE=arm64;;
    *)          echo -e "${RED}❌ Unsupported architecture: ${ARCH}${NC}"; exit 1;;
esac

echo -e "${GREEN}✓${NC} Detected: linux/${ARCH_TYPE}"

# Get latest version
if [ "${VERSION}" = "latest" ]; then
    echo "🔍 Fetching latest version..."
    VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi

BINARY_FILE="${BINARY_NAME}-linux-${ARCH_TYPE}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_FILE}"

echo "📥 Downloading ${VERSION}..."
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

curl -L -o "${TMP_DIR}/${BINARY_FILE}" "${DOWNLOAD_URL}"
chmod +x "${TMP_DIR}/${BINARY_FILE}"

# Install binary
echo "📦 Installing binary..."
mv "${TMP_DIR}/${BINARY_FILE}" "${INSTALL_DIR}/${BINARY_NAME}"

# Create user
echo "👤 Creating service user..."
if ! id "${SERVICE_USER}" &>/dev/null; then
    useradd -r -s /bin/false -d "${DATA_DIR}" "${SERVICE_USER}"
fi

# Create directories
echo "📁 Creating directories..."
mkdir -p "${DATA_DIR}/db"
mkdir -p "${DATA_DIR}/backups"
mkdir -p "${CONFIG_DIR}/certs"
mkdir -p "${CONFIG_DIR}/databases"
mkdir -p "/var/log/weather"
mkdir -p "/var/cache/weather/weather"
chown -R "${SERVICE_USER}:${SERVICE_USER}" "${DATA_DIR}"
chown -R "${SERVICE_USER}:${SERVICE_USER}" "${CONFIG_DIR}"
chown -R "${SERVICE_USER}:${SERVICE_USER}" "/var/log/weather"
chown -R "${SERVICE_USER}:${SERVICE_USER}" "/var/cache/weather"

# Create systemd service
echo "⚙️  Creating systemd service..."
cat > /etc/systemd/system/weather.service <<EOF
[Unit]
Description=Weather Service
After=network.target
Documentation=https://github.com/${REPO}

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_USER}
Environment="PORT=3000"
Environment="GIN_MODE=release"
Environment="TZ=UTC"
ExecStart=${INSTALL_DIR}/${BINARY_NAME}
Restart=on-failure
RestartSec=5s
StandardOutput=journal
StandardError=journal
SyslogIdentifier=weather

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${DATA_DIR} ${CONFIG_DIR} /var/log/weather /var/cache/weather
ProtectKernelTunables=true
ProtectControlGroups=true
RestrictRealtime=true
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
echo "🔄 Reloading systemd..."
systemctl daemon-reload

echo ""
echo -e "${GREEN}✅ Installation complete!${NC}"
echo ""
echo "Next steps:"
echo "  sudo systemctl start weather    # Start service"
echo "  sudo systemctl enable weather   # Enable on boot"
echo "  sudo systemctl status weather   # Check status"
echo ""
echo "  journalctl -u weather -f        # View logs"
echo ""
echo "Service will run on: http://localhost:3000"

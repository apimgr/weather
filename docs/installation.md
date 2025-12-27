# Installation

This guide covers installing Weather Service on various platforms.

## System Requirements

- **OS**: Linux, macOS, Windows, FreeBSD
- **Architecture**: amd64 (x86_64) or arm64 (aarch64)
- **Memory**: Minimum 512MB RAM, recommended 1GB+
- **Disk**: 100MB for binary + data storage
- **Ports**: Port 80 (HTTP) or custom port

## Docker Installation (Recommended)

Docker is the recommended deployment method for production environments.

### Quick Start

```bash
docker run -d \
  --name weather \
  -p 80:80 \
  -v weather-data:/var/lib/weather \
  -v weather-config:/etc/weather \
  ghcr.io/apimgr/weather:latest
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  weather:
    image: ghcr.io/apimgr/weather:latest
    container_name: weather
    ports:
      - "80:80"
    volumes:
      - ./data:/var/lib/weather
      - ./config:/etc/weather
    environment:
      - MODE=production
    restart: unless-stopped
```

Start the service:

```bash
docker-compose up -d
```

### Available Tags

| Tag | Description |
|-----|-------------|
| `latest` | Latest stable release |
| `beta` | Beta releases from beta branch |
| `devel` | Development builds from main branch |
| `v1.2.3` | Specific version tag |
| `2501` | Monthly release (YYMM format) |

## Binary Installation

### Download Pre-built Binary

Download the appropriate binary for your platform from [GitHub Releases](https://github.com/apimgr/weather/releases/latest):

=== "Linux (amd64)"

    ```bash
    wget https://github.com/apimgr/weather/releases/latest/download/weather-linux-amd64
    chmod +x weather-linux-amd64
    sudo mv weather-linux-amd64 /usr/local/bin/weather
    ```

=== "Linux (arm64)"

    ```bash
    wget https://github.com/apimgr/weather/releases/latest/download/weather-linux-arm64
    chmod +x weather-linux-arm64
    sudo mv weather-linux-arm64 /usr/local/bin/weather
    ```

=== "macOS (amd64)"

    ```bash
    wget https://github.com/apimgr/weather/releases/latest/download/weather-darwin-amd64
    chmod +x weather-darwin-amd64
    sudo mv weather-darwin-amd64 /usr/local/bin/weather
    ```

=== "macOS (arm64)"

    ```bash
    wget https://github.com/apimgr/weather/releases/latest/download/weather-darwin-arm64
    chmod +x weather-darwin-arm64
    sudo mv weather-darwin-arm64 /usr/local/bin/weather
    ```

=== "Windows (amd64)"

    ```powershell
    # Download weather-windows-amd64.exe
    # Move to C:\Program Files\weather\weather.exe
    ```

=== "FreeBSD (amd64)"

    ```bash
    wget https://github.com/apimgr/weather/releases/latest/download/weather-freebsd-amd64
    chmod +x weather-freebsd-amd64
    sudo mv weather-freebsd-amd64 /usr/local/bin/weather
    ```

### Run Directly

```bash
weather
```

The server will start on port 80 and open the setup wizard in your browser.

## Linux Systemd Service

For production deployments on Linux, use systemd for automatic startup and management.

### Create Service User

```bash
sudo useradd -r -s /bin/false weather
```

### Create Directories

```bash
sudo mkdir -p /var/lib/weather/{data,config}
sudo mkdir -p /var/log/weather
sudo chown -R weather:weather /var/lib/weather /var/log/weather
```

### Create Systemd Service

Create `/etc/systemd/system/weather.service`:

```ini
[Unit]
Description=Weather API Service
Documentation=https://apimgr-weather.readthedocs.io
After=network.target

[Service]
Type=simple
User=weather
Group=weather
WorkingDirectory=/var/lib/weather

ExecStart=/usr/local/bin/weather \
  --mode production \
  --data /var/lib/weather/data \
  --config /var/lib/weather/config \
  --log /var/log/weather \
  --address :80

Restart=always
RestartSec=10

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/weather /var/log/weather

[Install]
WantedBy=multi-user.target
```

### Enable and Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable weather
sudo systemctl start weather
sudo systemctl status weather
```

### Service Management

```bash
# Start service
sudo systemctl start weather

# Stop service
sudo systemctl stop weather

# Restart service
sudo systemctl restart weather

# View logs
sudo journalctl -u weather -f
```

## macOS LaunchDaemon

For macOS, use launchd for automatic startup.

Create `/Library/LaunchDaemons/com.apimgr.weather.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.apimgr.weather</string>
    
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/weather</string>
        <string>--mode</string>
        <string>production</string>
    </array>
    
    <key>RunAtLoad</key>
    <true/>
    
    <key>KeepAlive</key>
    <true/>
    
    <key>StandardOutPath</key>
    <string>/var/log/weather.log</string>
    
    <key>StandardErrorPath</key>
    <string>/var/log/weather.error.log</string>
</dict>
</plist>
```

Load the service:

```bash
sudo launchctl load /Library/LaunchDaemons/com.apimgr.weather.plist
sudo launchctl start com.apimgr.weather
```

## Windows Service

Install as a Windows service:

```powershell
# Install service
weather --service install

# Start service
weather --service start

# Stop service
weather --service stop

# Uninstall service
weather --service uninstall
```

## Building from Source

Requirements:
- Go 1.24 or later
- Docker (for containerized builds)

### Using Makefile

```bash
# Clone repository
git clone https://github.com/apimgr/weather.git
cd weather

# Build for current platform
make build

# Build for all platforms
make release

# Build Docker image
make docker

# Run development server
make dev
```

### Manual Build

```bash
# Build for Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o weather-linux-amd64 \
  -ldflags="-s -w -X 'main.Version=dev' -X 'main.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
  ./src

# Run the binary
./weather-linux-amd64
```

## First Run Setup

On first run, Weather Service will:

1. Create necessary directories and databases
2. Download GeoIP databases (may take a few minutes)
3. Open the setup wizard at `http://localhost`
4. Guide you through creating the admin account

!!! note "Setup Wizard"
    The setup wizard is a one-time process. After completion, you'll have:
    
    - Admin account with username and password
    - Configured server settings
    - Active weather service ready to use

## Verification

Verify the installation:

```bash
# Check version
weather --version

# Check health
curl http://localhost/healthz

# View status
weather --status
```

## Next Steps

- [Configuration Guide](configuration.md) - Configure weather sources, GeoIP, notifications
- [Admin Panel](admin.md) - Manage users, settings, and domains
- [API Reference](api.md) - Use the RESTful API

# Weather Service - Production Scripts

Scripts for installing, running, and managing the Weather Service in production environments.

## Quick Links

- **[Service Installation Guide](SERVICE_INSTALL.md)** - Complete guide for Linux, macOS, and Windows
- **[Windows Service Setup](WINDOWS_SERVICE.md)** - Detailed Windows service options

## Installation Scripts

### `install.sh`
Main installation script that auto-detects your OS and runs the appropriate installer.

```bash
# Auto-detect OS and install
sudo ./scripts/install.sh

# Or run directly
curl -fsSL https://raw.githubusercontent.com/apimgr/weather/main/scripts/install.sh | sudo bash
```

### `linux.sh`
Linux-specific installation (Ubuntu, Debian, RHEL, CentOS, Fedora, Arch).

**Features:**
- Creates system user `weather`
- Installs to `/usr/local/bin/weather`
- Sets up systemd service
- Creates directories:
  - `/etc/weather/` - Configuration (certs, databases)
  - `/var/lib/weather/` - Data (SQLite, backups)
  - `/var/log/weather/` - Logs (access, error, audit)
  - `/var/cache/weather/` - Cache

**Usage:**
```bash
sudo ./scripts/linux.sh
```

### `macos.sh`
macOS-specific installation.

**Features:**
- Installs to `/usr/local/bin/weather`
- Sets up LaunchAgent (user) or LaunchDaemon (system)
- Creates directories:
  - `~/Library/Application Support/weather/` - Config and data
  - `~/Library/Logs/weather/` - Logs
  - `~/Library/Caches/weather/` - Cache

**Usage:**
```bash
./scripts/macos.sh
```

### `windows.ps1`
Windows-specific installation (PowerShell).

**Features:**
- Installs to `C:\Program Files\Weather\`
- Sets up Windows Service
- Creates directories:
  - `%PROGRAMDATA%\Weather\` - Data
  - `%PROGRAMDATA%\Weather\logs\` - Logs

**Usage:**
```powershell
# Run as Administrator
.\scripts\windows.ps1
```

## Production Deployment

### Systemd Service (Linux)

```bash
# Install
sudo ./scripts/linux.sh

# Start service
sudo systemctl start weather

# Enable on boot
sudo systemctl enable weather

# Check status
sudo systemctl status weather

# View logs
sudo journalctl -u weather -f
```

### LaunchDaemon (macOS)

```bash
# Install
./scripts/macos.sh

# Start service
sudo launchctl load /Library/LaunchDaemons/com.apimgr.weather.plist

# Stop service
sudo launchctl unload /Library/LaunchDaemons/com.apimgr.weather.plist

# View logs
tail -f ~/Library/Logs/weather/weather.log
```

### Windows Service

```powershell
# Install (as Administrator)
.\scripts\windows.ps1

# Start service
Start-Service Weather

# Stop service
Stop-Service Weather

# Set to start automatically
Set-Service Weather -StartupType Automatic

# View logs
Get-Content C:\ProgramData\Weather\logs\weather.log -Tail 50 -Wait
```

## Manual Production Deployment

### Running with Custom Settings

```bash
# Specific ports
weather --port "80,443"

# Custom data directory
weather --data /var/lib/weather

# Custom config directory
weather --config /etc/weather

# All together
weather --port "80,443" --data /var/lib/weather --address 0.0.0.0
```

### Environment Variables

```bash
# Database configuration (first run only)
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=weather
export DB_USER=weather
export DB_PASSWORD=secure_password

# Server configuration
export PORT=3000
export SERVER_ADDRESS=0.0.0.0

# Run
weather
```

## Docker Deployment

```bash
# Build image
docker build -t apimgr/weather:latest .

# Run container
docker run -d \
  --name weather \
  -p 80:80 \
  -v /var/lib/weather:/data \
  -v /etc/weather:/config \
  apimgr/weather:latest

# With docker-compose
docker-compose up -d
```

## Production Checklist

- [ ] Use `--port "80,443"` for standard HTTP/HTTPS
- [ ] Configure external database (PostgreSQL/MySQL recommended)
- [ ] Set up SSL certificates (Let's Encrypt auto-configured on ports 80/443)
- [ ] Configure firewall rules
- [ ] Set up log rotation
- [ ] Configure backups
- [ ] Set `SESSION_SECRET` environment variable
- [ ] Enable systemd/launchd/Windows Service for auto-start
- [ ] Monitor with `/healthz` endpoint
- [ ] Set up reverse proxy (nginx/traefik) if needed

## Notes

- Installation scripts require root/administrator privileges
- Production deployments should use external databases (PostgreSQL/MySQL)
- SQLite is fine for development and small deployments
- All configuration is stored in the database (no config files)
- For testing and development, use scripts in `tests/` directory

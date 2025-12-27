# CLI Reference

Weather Service provides a comprehensive command-line interface for server management and operations.

## Installation

The server binary includes all CLI functionality:

```bash
# Download and install
wget https://github.com/apimgr/weather/releases/latest/download/weather-linux-amd64
chmod +x weather-linux-amd64
sudo mv weather-linux-amd64 /usr/local/bin/weather
```

## Basic Usage

```bash
weather [flags] [command]
```

## Flags

### Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--help` | - | Show help message |
| `--version` | - | Show version information |
| `--mode` | `production` | Server mode (production or development) |
| `--config` | `/etc/weather` | Configuration directory |
| `--data` | `/var/lib/weather` | Data directory |
| `--log` | `/var/log/weather` | Log directory |
| `--pid` | `/var/run/weather.pid` | PID file location |
| `--address` | `:80` | Listen address |
| `--port` | `80` | Listen port (deprecated, use --address) |

### Mode-Specific Behavior

**Production Mode** (`--mode production`):
- JSON logging format
- Optimized caching
- Rate limiting enabled
- Debug endpoints disabled

**Development Mode** (`--mode development`):
- Text logging format
- Reduced caching
- No rate limiting
- Debug endpoints enabled
- Auto-reload on file changes

## Commands

### Server Operations

#### Start Server

Start the weather service (default command):

```bash
weather
weather --mode production --address :8080
```

#### Status

Check server status and health:

```bash
weather --status
```

Output includes:
- Server running/stopped status
- Process ID (if running)
- Uptime
- Memory usage
- Active connections
- Health check results

#### Daemon Mode

Run server in background (detached from terminal):

```bash
weather --daemon
```

!!! note "Systemd Preferred"
    For production deployments, use systemd instead of daemon mode.
    See [Installation Guide](installation.md#linux-systemd-service).

### Service Management

Manage Weather Service as a system service.

#### Install Service

Install as a system service (Linux/macOS/Windows):

```bash
# Linux (systemd)
sudo weather --service install

# macOS (launchd)
sudo weather --service install

# Windows
weather --service install
```

#### Start Service

```bash
sudo weather --service start
```

#### Stop Service

```bash
sudo weather --service stop
```

#### Restart Service

```bash
sudo weather --service restart
```

#### Reload Configuration

Reload configuration without stopping the service:

```bash
sudo weather --service reload
```

#### Uninstall Service

```bash
sudo weather --service uninstall
```

#### Disable Service

Disable auto-start:

```bash
sudo weather --service disable
```

#### Service Help

```bash
weather --service help
```

### Maintenance Commands

Perform maintenance operations.

#### Backup

Create a backup of databases and configuration:

```bash
# Backup to default location
weather --maintenance backup

# Backup to specific file
weather --maintenance backup /path/to/backup.tar.gz
```

Backup includes:
- Server database (`server.db`)
- User database (`users.db`)
- Configuration files
- SSL certificates (if present)

#### Restore

Restore from a backup file:

```bash
weather --maintenance restore /path/to/backup.tar.gz
```

!!! warning "Service Must Be Stopped"
    Stop the service before restoring:
    ```bash
    sudo systemctl stop weather
    weather --maintenance restore backup.tar.gz
    sudo systemctl start weather
    ```

#### Update

Check for and install updates:

```bash
# Check for updates
weather --update check

# Update to latest stable
weather --update yes

# Update to specific branch
weather --update branch stable
weather --update branch beta
weather --update branch daily
```

Update branches:
- `stable` - Latest stable release
- `beta` - Beta releases
- `daily` - Daily development builds

#### Maintenance Mode

Enable/disable maintenance mode:

```bash
# Enable maintenance mode
weather --maintenance mode on

# Disable maintenance mode
weather --maintenance mode off

# Check status
weather --maintenance mode status
```

When maintenance mode is enabled:
- All requests return 503 Service Unavailable
- Admin panel shows maintenance notice
- Safe to perform database migrations

#### Setup Wizard

Re-run the initial setup wizard:

```bash
weather --maintenance setup
```

Use this to:
- Reset admin credentials
- Reconfigure server settings
- Fix corrupted configuration

### Database Operations

#### Verify Database Integrity

```bash
weather --maintenance verify
```

Checks:
- Database file integrity
- Schema version
- Foreign key constraints
- Index consistency

#### Vacuum Database

Optimize database files:

```bash
weather --maintenance vacuum
```

Benefits:
- Reclaim unused space
- Improve query performance
- Rebuild indexes

### Configuration Management

#### Show Configuration

Display effective configuration (with secrets redacted):

```bash
weather --show-config
```

#### Validate Configuration

Check configuration syntax and values:

```bash
weather --validate-config
```

#### Export Configuration

```bash
weather --export-config > config-backup.yml
```

#### Import Configuration

```bash
weather --import-config config-backup.yml
```

## Examples

### Production Deployment

```bash
# Install as systemd service
sudo weather --service install

# Configure via file
sudo nano /etc/weather/server.yml

# Start service
sudo systemctl start weather

# Check status
weather --status

# View logs
sudo journalctl -u weather -f
```

### Development Setup

```bash
# Run in development mode on custom port
weather --mode development --address :8080

# Or with environment variables
export WEATHER_MODE=development
export WEATHER_ADDRESS=:8080
weather
```

### Backup and Restore

```bash
# Create backup before update
weather --maintenance backup /backup/weather-$(date +%Y%m%d).tar.gz

# Stop service
sudo systemctl stop weather

# Update to latest
weather --update yes

# If issues occur, restore backup
weather --maintenance restore /backup/weather-20250115.tar.gz

# Start service
sudo systemctl start weather
```

### Maintenance Window

```bash
# Enable maintenance mode
weather --maintenance mode on

# Perform updates
weather --update yes

# Run vacuum
weather --maintenance vacuum

# Verify integrity
weather --maintenance verify

# Disable maintenance mode
weather --maintenance mode off
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Database error |
| 4 | Network error |
| 5 | Permission denied |
| 10 | Invalid arguments |
| 20 | Service not running |
| 30 | Update failed |

## Environment Variables

All flags can be set via environment variables:

```bash
export WEATHER_MODE=production
export WEATHER_ADDRESS=:8080
export WEATHER_DATA_DIR=/custom/data
export WEATHER_LOG_LEVEL=debug

# Then run without flags
weather
```

Format: `WEATHER_<FLAG_NAME>`

## Output Formats

### JSON Output

Use `--format json` for machine-readable output:

```bash
weather --status --format json
```

```json
{
  "status": "running",
  "pid": 12345,
  "uptime": 86400,
  "version": "1.0.0"
}
```

### Quiet Mode

Suppress non-error output:

```bash
weather --quiet --maintenance backup
```

### Verbose Mode

Show detailed output:

```bash
weather --verbose
```

## Scripting

### Check if Server is Running

```bash
#!/bin/bash
if weather --status --quiet; then
    echo "Server is running"
else
    echo "Server is stopped"
    exit 1
fi
```

### Automated Backups

```bash
#!/bin/bash
BACKUP_DIR="/backup/weather"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/weather-${TIMESTAMP}.tar.gz"

mkdir -p "${BACKUP_DIR}"
weather --maintenance backup "${BACKUP_FILE}"

# Keep only last 7 backups
find "${BACKUP_DIR}" -name "weather-*.tar.gz" -mtime +7 -delete
```

### Health Check Script

```bash
#!/bin/bash
if ! weather --status --quiet; then
    echo "Server is down, attempting restart..."
    sudo systemctl restart weather
    sleep 5
    if weather --status --quiet; then
        echo "Server restarted successfully"
    else
        echo "Failed to restart server" >&2
        exit 1
    fi
fi
```

## Next Steps

- [Configuration](configuration.md) - Configure the weather service
- [Admin Panel](admin.md) - Manage via web UI
- [API Reference](api.md) - Use the RESTful API
- [Installation](installation.md) - Deployment guides

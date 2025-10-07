# Weather Application Directory Structure

This document describes the directory layout used by the weather application across different environments.

## Directory Layout by Environment

### Linux/BSD (Running as Root)

```
/etc/weather/                    # Configuration directory
├── certs/                       # SSL/TLS certificates
│   ├── cert.pem
│   ├── key.pem
│   └── chain.pem
└── databases/                   # Configuration databases
    ├── geoip.mmdb              # GeoIP database
    └── *.json                  # Other JSON configuration files

/var/lib/weather/               # Data directory
├── db/                         # Database files
│   └── weather.db              # SQLite database
└── backups/                    # Backup files
    ├── weather_db_YYYY-MM-DD_HH-MM-SS.sql.gz
    └── weather_data_YYYY-MM-DD_HH-MM-SS.tar.gz

/var/log/weather/               # Log directory
├── access.log                  # HTTP access logs (Apache Combined format)
├── access.log.YYYY-MM-DD       # Rotated access logs
├── error.log                   # Application error logs
├── error.log.YYYY-MM-DD        # Rotated error logs
├── audit.log                   # Audit logs (JSON format)
└── audit.log.YYYY-MM-DD        # Rotated audit logs

/var/cache/weather/             # Cache directory
└── weather/                    # Weather API cache
    └── (cached weather data)
```

### Linux/BSD (Running as Regular User)

```
~/.config/weather/              # Configuration directory
├── certs/                      # SSL/TLS certificates
└── databases/                  # Configuration databases
    └── geoip.mmdb

~/.local/share/weather/         # Data directory
├── db/                         # Database files
│   └── weather.db              # SQLite database
├── backups/                    # Backup files
└── logs/                       # Log directory (user-specific)
    ├── access.log
    ├── error.log
    └── audit.log

~/.cache/weather/               # Cache directory
└── weather/                    # Weather API cache
```

### macOS (Running as Root)

```
/Library/Application Support/weather/
├── certs/
├── databases/
└── data/
    ├── db/
    │   └── weather.db
    └── backups/

/Library/Logs/weather/
├── access.log
├── error.log
└── audit.log

/Library/Caches/weather/
└── weather/
```

### macOS (Running as Regular User)

```
~/Library/Application Support/weather/
├── certs/
├── databases/
└── data/
    ├── db/
    │   └── weather.db
    └── backups/

~/Library/Logs/weather/
├── access.log
├── error.log
└── audit.log

~/Library/Caches/weather/
└── weather/
```

### Windows

```
%APPDATA%\weather\              # Configuration directory
├── certs\
└── databases\

%LOCALAPPDATA%\weather\
├── data\
│   ├── db\
│   │   └── weather.db
│   └── backups\
├── logs\
│   ├── access.log
│   ├── error.log
│   └── audit.log
└── cache\
    └── weather\
```

### Docker Container

```
/config/                        # Mounted volume
├── certs/
└── databases/

/data/                          # Mounted volume
├── db/                         # Database files
│   └── weather.db
└── backups/

/tmp/                           # Temporary files
└── (runtime logs - also accessible via docker logs)
```

### Test Environment

```
$TMPDIR/weather-test/           # Temporary test directory
├── config/
│   ├── certs/
│   └── databases/
│       ├── geoip.mmdb
│       └── airports.json
├── data/
│   ├── db/
│   │   └── weather.db
│   └── backups/
├── logs/
│   ├── access.log
│   ├── error.log
│   └── audit.log
└── cache/
    └── weather/
```

## Directory Usage

### Configuration Directory (`config/`)

**Purpose**: Stores configuration files and non-volatile data needed for operation.

**Contents**:
- `certs/` - SSL/TLS certificates (Let's Encrypt or self-signed)
- `databases/` - GeoIP database, JSON configuration files, static data

**Permissions**: 0755 (directories), 0644 (files), 0600 (private keys)

### Data Directory (`data/`)

**Purpose**: Stores application data and user-generated content.

**Contents**:
- `db/` - Database files directory
  - `weather.db` - SQLite database (all user data, settings, sessions)
- `backups/` - Automated backups (database dumps, user data archives)

**Permissions**: 0755 (directories), 0644 (database), 0644 (backups)

### Log Directory (`logs/`)

**Purpose**: Stores application logs for debugging and audit.

**Contents**:
- `access.log` - HTTP access logs (Apache Combined format)
- `error.log` - Application errors and warnings
- `audit.log` - Security audit trail (JSON format)

**Rotation**: Daily rotation, 30-day retention, automatic compression

**Permissions**: 0755 (directory), 0644 (log files)

### Cache Directory (`cache/`)

**Purpose**: Stores persistent cache data that improves performance.

**Contents**:
- `weather/` - Cached weather API responses

**Note**: Cache can be safely deleted - data will be re-fetched as needed.

**Permissions**: 0755 (directory), 0644 (files)

## CLI Overrides

You can override default directories using command-line flags:

```bash
weather --data /custom/data --config /custom/config
```

Environment variables can also override paths:
- `DATA_DIR` - Override data directory
- `CONFIG_DIR` - Override config directory
- `LOG_DIR` - Override log directory
- `CERTS_DIR` - Override certificates directory

## Utility Functions

The application provides helper functions to get correct paths:

```go
import "weather-go/src/utils"

// Get OS-appropriate paths
paths, err := utils.GetDirectoryPaths()

// Get specific paths
dbPath := utils.GetDatabasePath(paths)           // SQLite database
backupPath := utils.GetBackupPath(paths)         // Backup directory
certsPath := utils.GetCertsPath(paths)           // SSL certificates
configDBPath := utils.GetConfigDatabasesPath(paths) // Config databases
geoipPath := utils.GetGeoIPPath(paths)           // GeoIP database
airportPath := utils.GetAirportDataPath(paths)   // Airport JSON data
cachePath := utils.GetWeatherCachePath(paths)    // Weather cache
tempPath := utils.GetTempPath()                  // System temp (use sparingly)

// For tests - use temp directory
testPaths, err := utils.GetTestDirectoryPaths()
defer utils.CleanupTestDirectories(testPaths)
```

## Docker Volumes

For Docker deployments, mount these volumes:

```yaml
volumes:
  - ./rootfs/config/weather:/config
  - ./rootfs/data/weather:/data
```

The application will automatically use `/config` and `/data` when running in Docker.

## Backup Recommendations

**What to backup**:
- `data/db/weather.db` - Main database (critical)
- `data/backups/` - Backup history
- `config/databases/` - GeoIP and config files (if customized)
- `config/certs/` - SSL certificates (if custom)

**What NOT to backup**:
- `logs/` - Can be regenerated, large size
- `cache/` - Temporary data only

**Automated backups**: The application creates daily backups in `data/backups/` automatically.

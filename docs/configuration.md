# Configuration

Weather Service is configured via YAML files and environment variables.

## Configuration Files

Weather Service uses two configuration files:

| File | Location | Purpose |
|------|----------|---------|
| `server.yml` | `{configdir}/server.yml` | Server-wide settings |
| CLI config | `~/.config/weather/cli.yml` | CLI client settings (if using CLI) |

Default config directory locations:

- **Linux/FreeBSD**: `/etc/weather/`
- **macOS**: `/usr/local/etc/weather/`
- **Windows**: `C:\ProgramData\weather\`
- **Docker**: `/etc/weather/` (mount as volume)

## Command-Line Flags

Override configuration with CLI flags:

```bash
weather --mode production --address :8080 --data /custom/path
```

| Flag | Default | Description |
|------|---------|-------------|
| `--help` | - | Show help message |
| `--version` | - | Show version information |
| `--mode` | `production` | Mode: `production` or `development` |
| `--config` | `/etc/weather` | Configuration directory |
| `--data` | `/var/lib/weather` | Data directory |
| `--log` | `/var/log/weather` | Log directory |
| `--pid` | `/var/run/weather.pid` | PID file location |
| `--address` | `:80` | Listen address (`:port` or `host:port`) |
| `--port` | `80` | Listen port (deprecated, use --address) |

## server.yml Reference

Complete configuration reference with all available settings.

### Server Section

```yaml
server:
  # Server mode (production or development)
  mode: production
  # Listen address (:80 or 0.0.0.0:80 or specific IP)
  address: ":80"
  # Server name (for logging and identification)
  name: "Weather Service"
  # Enable debug mode (verbose logging, no caching)
  debug: false
  # Graceful shutdown timeout
  shutdown_timeout: 30s
```

### Weather Settings

```yaml
weather:
  # Weather data providers
  providers:
    openmeteo:
      enabled: true
      base_url: "https://api.open-meteo.com/v1"
      timeout: 10s

  # Severe weather alerts
  alerts:
    us_nws:
      enabled: true
      base_url: "https://api.weather.gov"
    canada:
      enabled: true
    uk:
      enabled: true
    australia:
      enabled: true
    japan:
      enabled: true
    mexico:
      enabled: true

  # Earthquake data
  earthquakes:
    enabled: true
    usgs_url: "https://earthquake.usgs.gov/earthquakes/feed/v1.0"
    min_magnitude: 2.5
    max_age_days: 7

  # Hurricane tracking
  hurricanes:
    enabled: true
    nhc_url: "https://www.nhc.noaa.gov"
    check_interval: 30m

  # Moon phase data
  moon:
    enabled: true

  # Cache settings
  cache:
    weather_ttl: 10m
    forecast_ttl: 1h
    earthquake_ttl: 5m
    hurricane_ttl: 15m
```

### GeoIP Settings

```yaml
geoip:
  # Enable GeoIP location detection
  enabled: true

  # Auto-update databases
  auto_update: true

  # Update interval (7 days recommended)
  update_interval: "168h"

  # Data source
  data_source: "sapics/ip-location-db"

  # Database URLs (CDN-hosted, free)
  databases:
    - type: "city-ipv4"
      url: "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city/geolite2-city-ipv4.csv"
    - type: "city-ipv6"
      url: "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city/geolite2-city-ipv6.csv"
    - type: "country-ipv4"
      url: "https://cdn.jsdelivr.net/npm/@ip-location-db/dbip-country/dbip-country-ipv4.csv"
    - type: "country-ipv6"
      url: "https://cdn.jsdelivr.net/npm/@ip-location-db/dbip-country/dbip-country-ipv6.csv"
    - type: "asn-ipv4"
      url: "https://cdn.jsdelivr.net/npm/@ip-location-db/asn/asn-ipv4.tsv"
    - type: "asn-ipv6"
      url: "https://cdn.jsdelivr.net/npm/@ip-location-db/asn/asn-ipv6.tsv"
```

### Location Settings

```yaml
locations:
  # Cookie duration for location persistence (30 days)
  cookie_duration: "720h"

  # Default location (auto = GeoIP, or specify coords)
  default_location: "auto"
```

### Database Settings

```yaml
database:
  # Server infrastructure database
  server:
    type: "sqlite"
    path: "/var/lib/weather/db/server.db"

  # User accounts database
  users:
    type: "sqlite"
    path: "/var/lib/weather/db/users.db"

  # Connection pool settings
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: "5m"
```

### Authentication Settings

```yaml
auth:
  # Session settings
  session:
    # Session duration (7 days)
    duration: "168h"
    cookie_name: "weather_session"
    secure: true
    http_only: true
    same_site: "lax"

  # Password requirements
  password:
    min_length: 12
    require_uppercase: true
    require_lowercase: true
    require_numbers: true
    require_special: true

  # Rate limiting for authentication
  rate_limit:
    login_attempts: 5
    login_window: "15m"
    password_reset_attempts: 3
    password_reset_window: "1h"
```

### Email/SMTP Settings

```yaml
email:
  # Enable email notifications
  enabled: false

  # SMTP configuration
  smtp:
    host: "smtp.example.com"
    port: 587
    username: "weather@example.com"
    password: "your-password"
    from: "Weather Service <weather@example.com>"
    use_tls: true

  # Email templates
  templates:
    welcome: "email/welcome.tmpl"
    password_reset: "email/password_reset.tmpl"
    2fa_enabled: "email/2fa_enabled.tmpl"
```

### Notification Settings

```yaml
notifications:
  # WebSocket notifications
  websocket:
    enabled: true
    ping_interval: "30s"
    pong_timeout: "10s"

  # Admin notifications
  admin:
    enable_toast: true
    enable_banner: true
    enable_center: true
    toast_duration_success: 5
    toast_duration_info: 5
    toast_duration_warning: 10

  # User notifications
  user:
    enable_toast: true
    enable_banner: true
    enable_center: true
    enable_sound: false
```

### SSL/TLS Settings

```yaml
ssl:
  # Enable HTTPS
  enabled: false

  # Certificate paths
  cert_file: "/etc/weather/ssl/cert.pem"
  key_file: "/etc/weather/ssl/key.pem"

  # Let's Encrypt automatic SSL
  letsencrypt:
    enabled: false
    email: "admin@example.com"
    domains:
      - "weather.example.com"
    cache_dir: "/var/lib/weather/letsencrypt"
```

### Logging Settings

```yaml
logging:
  # Log level (debug, info, warn, error)
  level: "info"

  # Log format (json, text)
  format: "json"

  # Log output (stdout, file, both)
  output: "both"

  # File logging
  file:
    path: "/var/log/weather/weather.log"
    # Maximum log file size in MB
    max_size: 100
    max_backups: 10
    # Maximum log age in days
    max_age: 30
    compress: true

  # Access log
  access_log:
    enabled: true
    path: "/var/log/weather/access.log"
```

### Security Settings

```yaml
security:
  # CSRF protection
  csrf:
    enabled: true
    field_name: "csrf_token"
    cookie_name: "csrf_cookie"

  # CORS settings
  cors:
    enabled: false
    allowed_origins:
      - "https://example.com"
    allowed_methods:
      - "GET"
      - "POST"
    allowed_headers:
      - "Content-Type"
      - "Authorization"

  # Rate limiting
  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst: 100

  # IP blocking
  blocklist:
    enabled: false
    # ISO country codes to block
    countries: []
    # IP addresses to block
    ips: []
    # CIDR ranges to block
    cidrs: []
```

### Tor Hidden Service

```yaml
tor:
  # Enable Tor hidden service
  enabled: false

  # Tor data directory
  data_dir: "/var/lib/weather/tor"

  # Hidden service port
  port: 80
```

## Environment Variables

Override settings with environment variables:

```bash
export WEATHER_MODE=production
export WEATHER_ADDRESS=:8080
export WEATHER_DATA_DIR=/custom/data
export WEATHER_LOG_LEVEL=debug
```

Environment variable format: `WEATHER_<SECTION>_<KEY>`

Examples:
- `WEATHER_SERVER_MODE=production` → `server.mode`
- `WEATHER_DATABASE_SERVER_PATH=/db/server.db` → `database.server.path`
- `WEATHER_LOGGING_LEVEL=debug` → `logging.level`

## Live Configuration Updates

Most settings can be updated via the admin panel without restart:

1. Go to **Admin Panel** → **Settings**
2. Modify settings in the web UI
3. Click **Save Settings**
4. Changes take effect immediately

Settings requiring restart:
- `server.address` (listen address)
- `database.*` (database connections)
- `ssl.*` (SSL certificates)

## Configuration Validation

Validate your configuration:

```bash
# Check configuration syntax
weather --config /etc/weather --validate

# Show effective configuration
weather --config /etc/weather --show-config
```

## Example Configurations

### Development Setup

```yaml
server:
  mode: development
  address: ":8080"
  debug: true

logging:
  level: debug
  format: text
  output: stdout

weather:
  cache:
    # Short cache for testing
    weather_ttl: 1m
```

### Production Setup

```yaml
server:
  mode: production
  address: ":443"
  name: "Weather Production"

ssl:
  enabled: true
  letsencrypt:
    enabled: true
    email: "admin@example.com"
    domains:
      - "weather.example.com"

logging:
  level: info
  format: json
  output: both

security:
  csrf:
    enabled: true
  rate_limit:
    enabled: true
    requests_per_minute: 100
```

### High-Performance Setup

```yaml
database:
  max_open_conns: 100
  max_idle_conns: 25
  conn_max_lifetime: "30m"

weather:
  cache:
    weather_ttl: 30m
    forecast_ttl: 6h

security:
  rate_limit:
    enabled: true
    requests_per_minute: 1000
    burst: 2000
```

## Next Steps

- [API Reference](api.md) - Use the RESTful API
- [Admin Panel](admin.md) - Manage settings via web UI
- [CLI Reference](cli.md) - Use command-line tools

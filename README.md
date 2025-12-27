# Weather Service

> A production-grade weather API service providing global weather forecasts, severe weather alerts, earthquake data, moon phase information, and hurricane tracking.

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows%20%7C%20FreeBSD-lightgrey.svg)]()
[![Docker](https://img.shields.io/badge/docker-ghcr.io-blue.svg)](https://ghcr.io/apimgr/weather)

## About

Weather Service is a comprehensive weather platform built with Go, providing:

- **Global Weather Forecasts** - 16-day forecasts for any location worldwide via Open-Meteo
- **Severe Weather Alerts** - Real-time alerts for hurricanes, tornadoes, storms, floods, and winter weather
- **International Support** - Weather alerts from US, Canada, UK, Australia, Japan, and Mexico
- **Earthquake Tracking** - Real-time earthquake data from USGS with interactive maps
- **Moon Phases** - Detailed lunar information including phases, illumination, rise/set times
- **Hurricane Tracking** - Active storm tracking with advisories and forecasts from NOAA
- **Real-Time Notifications** - WebSocket-powered notification system with toast, banner, and notification center
- **GeoIP Location** - Automatic location detection via IP address (IPv4 and IPv6)
- **Location Persistence** - Cookie-based location memory (30 days)
- **Mobile Responsive** - Optimized for desktop, tablet, and mobile devices
- **Dracula Theme** - Beautiful dark theme with cyan accents

### Key Features

- ✅ **Single Static Binary** - No external dependencies, all assets embedded
- ✅ **Multi-Platform** - Linux, macOS, Windows, FreeBSD (amd64/arm64)
- ✅ **Docker Ready** - Official images on GitHub Container Registry
- ✅ **Production Grade** - Rate limiting, caching, health checks, graceful shutdown
- ✅ **RESTful JSON API** - Clean, well-documented API endpoints
- ✅ **Web Interface** - Modern, responsive UI with Dracula theme
- ✅ **Real-Time Notifications** - WebSocket-powered notifications with customizable preferences
- ✅ **Lightweight** - <50MB memory, <200ms response times
- ✅ **Zero Configuration** - Works out of the box with sensible defaults

### Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: Gin
- **Weather Data**: Open-Meteo (global), NOAA NWS/NHC (US), international APIs
- **GeoIP**: sapics/ip-location-db with 4 databases (IPv4/IPv6 city, country, ASN)
- **Frontend**: HTML templates, vanilla CSS/JS, Dracula theme
- **Deployment**: Docker, systemd, launchd, Windows service

## Official Links

- **Documentation**: [https://apimgr-weather.readthedocs.io](https://apimgr-weather.readthedocs.io)
- **GitHub**: [https://github.com/apimgr/weather](https://github.com/apimgr/weather)
- **Docker Hub**: [https://ghcr.io/apimgr/weather](https://ghcr.io/apimgr/weather)
- **Issues**: [https://github.com/apimgr/weather/issues](https://github.com/apimgr/weather/issues)
- **Discussions**: [https://github.com/apimgr/weather/discussions](https://github.com/apimgr/weather/discussions)

---

## Production Deployment

### Binary Installation

Download pre-built binaries for your platform from [GitHub Releases](https://github.com/apimgr/weather/releases/latest).

#### Linux (systemd)

```bash
# Download binary
wget https://github.com/apimgr/weather/releases/latest/download/weather-linux-amd64
chmod +x weather-linux-amd64
sudo mv weather-linux-amd64 /usr/local/bin/weather

# Create service user
sudo useradd -r -s /bin/false weather

# Create directories
sudo mkdir -p /var/lib/weather/{data,config}
sudo mkdir -p /var/log/weather
sudo chown -R weather:weather /var/lib/weather /var/log/weather

# Create systemd service
sudo tee /etc/systemd/system/weather.service > /dev/null <<EOF
[Unit]
Description=Weather API Service
After=network.target

[Service]
Type=simple
User=weather
Group=weather
WorkingDirectory=/var/lib/weather
ExecStart=/usr/local/bin/weather
Restart=always
RestartSec=10
Environment="PORT=80"
Environment="DATA_DIR=/var/lib/weather/data"
Environment="CONFIG_DIR=/var/lib/weather/config"
Environment="LOG_DIR=/var/log/weather"

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable weather
sudo systemctl start weather
sudo systemctl status weather

# View logs
sudo journalctl -u weather -f
```

#### macOS (launchd)

```bash
# Download binary (Apple Silicon)
wget https://github.com/apimgr/weather/releases/latest/download/weather-darwin-arm64
chmod +x weather-darwin-arm64
sudo mv weather-darwin-arm64 /usr/local/bin/weather

# Create directories
sudo mkdir -p /usr/local/var/weather/{data,config}
sudo mkdir -p /usr/local/var/log/weather

# Create launchd service
sudo tee /Library/LaunchDaemons/com.apimgr.weather.plist > /dev/null <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.apimgr.weather</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/weather</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>/usr/local/var/log/weather/error.log</string>
    <key>StandardOutPath</key>
    <string>/usr/local/var/log/weather/output.log</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PORT</key>
        <string>80</string>
    </dict>
</dict>
</plist>
EOF

# Load and start
sudo launchctl load /Library/LaunchDaemons/com.apimgr.weather.plist
```

#### Windows (NSSM Service)

```powershell
# Download binary
Invoke-WebRequest -Uri "https://github.com/apimgr/weather/releases/latest/download/weather-windows-amd64.exe" -OutFile "weather.exe"

# Create directories
New-Item -ItemType Directory -Force -Path "C:\ProgramData\Weather\data"
New-Item -ItemType Directory -Force -Path "C:\ProgramData\Weather\config"
New-Item -ItemType Directory -Force -Path "C:\ProgramData\Weather\logs"

# Move binary
Move-Item weather.exe "C:\Program Files\Weather\weather.exe"

# Install as Windows Service using NSSM
# Download NSSM from https://nssm.cc/download
nssm install Weather "C:\Program Files\Weather\weather.exe"
nssm set Weather AppDirectory "C:\ProgramData\Weather"
nssm set Weather AppStdout "C:\ProgramData\Weather\logs\output.log"
nssm set Weather AppStderr "C:\ProgramData\Weather\logs\error.log"
nssm set Weather AppEnvironmentExtra "PORT=80"
nssm start Weather
```

#### FreeBSD (rc.d)

```bash
# Download binary
fetch https://github.com/apimgr/weather/releases/latest/download/weather-bsd-amd64
chmod +x weather-bsd-amd64
sudo mv weather-bsd-amd64 /usr/local/bin/weather

# Create service user
sudo pw useradd weather -d /var/db/weather -s /usr/sbin/nologin

# Create directories
sudo mkdir -p /var/db/weather/{data,config}
sudo mkdir -p /var/log/weather
sudo chown -R weather:weather /var/db/weather /var/log/weather

# Create rc.d script
sudo tee /usr/local/etc/rc.d/weather > /dev/null <<'EOF'
#!/bin/sh
# PROVIDE: weather
# REQUIRE: NETWORKING
# KEYWORD: shutdown

. /etc/rc.subr

name="weather"
rcvar="weather_enable"
command="/usr/local/bin/weather"
weather_user="weather"
pidfile="/var/run/${name}.pid"
command_args="&"

load_rc_config $name
run_rc_command "$1"
EOF

sudo chmod +x /usr/local/etc/rc.d/weather

# Enable and start
sudo sysrc weather_enable="YES"
sudo service weather start
```

### Environment Variables

Configure the service using environment variables:

```bash
# Server Configuration
PORT=80                              # HTTP server port (default: 80)
DOMAIN=weather.example.com           # Public domain name (for URL generation)
HOSTNAME=server.local                # Server hostname (fallback for DOMAIN)

# Directory Paths (OS-specific defaults if not set)
DATA_DIR=/var/lib/weather/data       # Data directory
CONFIG_DIR=/var/lib/weather/config   # Configuration directory
LOG_DIR=/var/log/weather             # Log directory
CACHE_DIR=/var/cache/weather         # Cache directory (optional)
TEMP_DIR=/tmp/weather                # Temporary directory (optional)
```

### Configuration File

Optional configuration file at `$CONFIG_DIR/server.yml`:

```yaml
server:
  port: 80
  domain: weather.example.com

logging:
  level: info  # debug, info, warn, error
  format: json  # json, text

geoip:
  update_interval: 168h  # Update weekly (hours)
  databases:
    - geolite2-city-ipv4
    - geolite2-city-ipv6
    - geo-whois-asn-country
    - asn

cache:
  weather_ttl: 15m
  severe_ttl: 5m
  moon_ttl: 1h
  earthquake_ttl: 10m

rate_limit:
  requests_per_window: 100
  window_duration: 15m
  burst: 20
```

---

## Docker Deployment

### Quick Start

```bash
# Pull and run
docker pull ghcr.io/apimgr/weather:latest
docker run -d \
  --name weather \
  -p 172.17.0.1:64080:80 \
  -v ./rootfs/data/weather:/data \
  -v ./rootfs/config/weather:/config \
  -v ./rootfs/logs/weather:/var/log/weather \
  -e DOMAIN=weather.example.com \
  --restart unless-stopped \
  ghcr.io/apimgr/weather:latest

# View logs
docker logs -f weather

# Check status
curl http://172.17.0.1:64080/api/health
```

### Docker Compose (Production)

Create `docker-compose.yml`:

```yaml
services:
  weather:
    image: ghcr.io/apimgr/weather:latest
    container_name: weather
    restart: unless-stopped

    environment:
      - CONFIG_DIR=/config
      - DATA_DIR=/data
      - LOGS_DIR=/var/log/weather
      - PORT=80
      - DOMAIN=weather.example.com

    volumes:
      - ./rootfs/config/weather:/config
      - ./rootfs/data/weather:/data
      - ./rootfs/logs/weather:/var/log/weather

    ports:
      - "172.17.0.1:64080:80"

    networks:
      - weather

    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost/api/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 10s

networks:
  weather:
    name: weather
    external: false
    driver: bridge
```

Start the service:

```bash
docker compose up -d
docker compose logs -f
docker compose ps
```

### Docker Compose (Development)

For development/testing with ephemeral storage:

```yaml
services:
  weather:
    image: weather:dev
    container_name: weather-test
    restart: "no"

    environment:
      - CONFIG_DIR=/config
      - DATA_DIR=/data
      - LOGS_DIR=/var/log/weather
      - PORT=80
      - DEV=true

    volumes:
      - /tmp/weather/rootfs/config/weather:/config
      - /tmp/weather/rootfs/data/weather:/data
      - /tmp/weather/rootfs/logs/weather:/var/log/weather

    ports:
      - "64181:80"

    networks:
      - weather

networks:
  weather:
    name: weather
    external: false
    driver: bridge
```

### Building Custom Images

```bash
# Development image (local testing)
make docker-dev

# Production image (multi-platform push to registry)
make docker
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: weather
spec:
  replicas: 3
  selector:
    matchLabels:
      app: weather
  template:
    metadata:
      labels:
        app: weather
    spec:
      containers:
      - name: weather
        image: ghcr.io/apimgr/weather:latest
        ports:
        - containerPort: 80
        env:
        - name: PORT
          value: "80"
        - name: DOMAIN
          value: "weather.example.com"
        volumeMounts:
        - name: data
          mountPath: /data
        - name: config
          mountPath: /config
        livenessProbe:
          httpGet:
            path: /api/health
            port: 80
          initialDelaySeconds: 60
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/health
            port: 80
          initialDelaySeconds: 30
          periodSeconds: 5
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: weather-data-pvc
      - name: config
        persistentVolumeClaim:
          claimName: weather-config-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: weather
spec:
  selector:
    app: weather
  ports:
  - port: 80
    targetPort: 80
  type: LoadBalancer
```

---

## API Usage

### Quick Examples

```bash
# Weather forecast (IP-based location)
curl http://localhost/

# Weather for specific location
curl http://localhost/weather/Brooklyn,NY

# Severe weather alerts
curl http://localhost/severe-weather

# Moon phase
curl http://localhost/moon

# Earthquakes
curl http://localhost/earthquake

# API endpoints (JSON)
curl http://localhost/api/v1/weather?location=London
curl http://localhost/api/v1/severe-weather?location=Miami,FL&distance=50
curl http://localhost/api/v1/moon?location=Tokyo
```

### API Endpoints

#### Weather

| Endpoint | Method | Description |
|----------|--------|-------------|
| `GET /` | GET | Weather for your location (IP-based) |
| `GET /weather/{location}` | GET | Weather for specific location |
| `GET /{location}` | GET | Weather (backwards compatible) |
| `GET /api/weather?location={loc}` | GET | JSON weather data |

#### Severe Weather

| Endpoint | Method | Description |
|----------|--------|-------------|
| `GET /severe-weather` | GET | Alerts for your location |
| `GET /severe-weather/{location}` | GET | Alerts for specific location |
| `GET /severe/{type}` | GET | Filter by type (hurricanes, tornadoes, storms, winter, floods) |
| `GET /severe/{type}/{location}` | GET | Filtered alerts for location |
| `GET /api/v1/severe-weather?location={loc}&distance={mi}&type={t}` | GET | JSON alerts |

#### Other Data

| Endpoint | Method | Description |
|----------|--------|-------------|
| `GET /moon` | GET | Moon phase for your location |
| `GET /moon/{location}` | GET | Moon phase for specific location |
| `GET /api/moon?location={loc}` | GET | JSON moon data |
| `GET /earthquake` | GET | Recent earthquakes near you |
| `GET /earthquake/{location}` | GET | Earthquakes near location |
| `GET /api/v1/earthquakes?location={loc}&radius={km}` | GET | JSON earthquake data |
| `GET /api/health` | GET | Health check |

#### Notifications

| Endpoint | Method | Description |
|----------|--------|-------------|
| `GET /ws/notifications` | WebSocket | Real-time notification stream |
| `GET /api/v1/user/notifications` | GET | List user notifications (paginated) |
| `GET /api/v1/user/notifications/count` | GET | Get unread notification count |
| `PATCH /api/v1/user/notifications/:id/read` | PATCH | Mark notification as read |
| `PATCH /api/v1/user/notifications/read` | PATCH | Mark all notifications as read |
| `DELETE /api/v1/user/notifications/:id` | DELETE | Delete notification |
| `GET /api/v1/user/notifications/preferences` | GET | Get notification preferences |
| `PATCH /api/v1/user/notifications/preferences` | PATCH | Update notification preferences |

See [docs/API_NOTIFICATIONS.md](docs/API_NOTIFICATIONS.md) for complete notification API documentation.

### Location Formats

The service accepts multiple location formats:

- **City, State**: `Brooklyn, NY` or `Brooklyn,NY`
- **City, Country**: `London, UK` or `Tokyo, JP`
- **ZIP Code**: `10001` (US only)
- **Coordinates**: `40.7128,-74.0060` (lat,lon)
- **Airport Code**: `JFK` or `LAX`

### Response Example

```json
{
  "location": "Brooklyn, NY",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "timezone": "America/New_York",
  "current": {
    "time": "2025-10-16T12:00:00-04:00",
    "temperature": 72.5,
    "humidity": 65,
    "wind_speed": 10.5,
    "weather_code": 2
  },
  "daily": [
    {
      "date": "2025-10-16",
      "temperature_max": 75.2,
      "temperature_min": 62.1,
      "sunrise": "06:42:00",
      "sunset": "18:15:00",
      "uv_index_max": 5.2
    }
  ]
}
```

### Rate Limiting

- **Default**: 100 requests per 15 minutes per IP
- HTTP 429 response when rate limit exceeded
- Burst allowance: Up to 20 requests in rapid succession

### Health Check

```bash
curl http://localhost/api/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-16T12:00:00Z"
}
```

---

## Reverse Proxy Setup

### Nginx

```nginx
server {
    listen 80;
    server_name weather.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name weather.example.com;

    ssl_certificate /etc/letsencrypt/live/weather.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/weather.example.com/privkey.pem;

    location / {
        proxy_pass http://172.17.0.1:64080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Caddy

```caddy
weather.example.com {
    reverse_proxy 172.17.0.1:64080

    header {
        Strict-Transport-Security "max-age=31536000"
    }

    encode gzip
}
```

---

## Data Sources

- **Weather**: [Open-Meteo](https://open-meteo.com/) - Global weather forecasts
- **US Severe Weather**: [NOAA NWS](https://www.weather.gov/) - National Weather Service
- **Hurricanes**: [NOAA NHC](https://www.nhc.noaa.gov/) - National Hurricane Center
- **Canadian Alerts**: [Environment Canada](https://weather.gc.ca/)
- **UK Alerts**: [Met Office](https://www.metoffice.gov.uk/)
- **Australian Alerts**: [Bureau of Meteorology](http://www.bom.gov.au/)
- **Japanese Alerts**: [JMA](https://www.jma.go.jp/)
- **Mexican Alerts**: [CONAGUA](https://www.gob.mx/conagua)
- **Earthquakes**: [USGS](https://earthquake.usgs.gov/)
- **GeoIP**: [sapics/ip-location-db](https://github.com/sapics/ip-location-db)

---

## Security

### Production Checklist

- ✅ **Use HTTPS** - Deploy behind reverse proxy with TLS certificates
- ✅ **Firewall Rules** - Restrict access to necessary ports only
- ✅ **Regular Updates** - Keep service and dependencies updated
- ✅ **Monitor Logs** - Set up log monitoring and alerting
- ✅ **Backup Data** - Regular backups of data directory
- ✅ **Rate Limiting** - Configured automatically (100 req/15min per IP)

### Firewall Configuration

```bash
# UFW (Ubuntu/Debian)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# firewalld (RHEL/CentOS)
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

---

## Monitoring

### Health Checks

```bash
# Check service health
curl http://localhost/api/health

# Expected response
{"status":"healthy","timestamp":"2025-10-16T12:00:00Z"}
```

### Logs

```bash
# Docker
docker logs -f weather

# systemd
sudo journalctl -u weather -f

# Direct file
tail -f /var/log/weather/weather.log
```

### Metrics

Monitor these key metrics:

- Request rate (requests per second)
- Response time (average latency)
- Error rate (5xx errors per minute)
- Memory usage (RSS in MB)
- GeoIP database age (days since update)

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker logs weather
# or
sudo journalctl -u weather -n 50

# Check port availability
sudo netstat -tulpn | grep :80
```

### High Memory Usage

```bash
# Check memory usage
docker stats weather
# or
ps aux | grep weather
```

### GeoIP Not Working

```bash
# Check database files
ls -lh /data/geoip/

# Force database update
rm -rf /data/geoip/*.mmdb
docker restart weather
```

---

## Development

### Prerequisites

- Go 1.24+
- Git
- Make
- Docker (optional, for containerized builds)

### Clone and Build

```bash
# Clone repository
git clone https://github.com/apimgr/weather.git
cd weather

# Install dependencies
go mod download

# Build all platforms
make build

# Build for host platform only
CGO_ENABLED=0 go build -o binaries/weather ./src

# Run
./binaries/weather
```

### Build System

```bash
# Available targets
make help

# Build for all platforms (8 binaries)
make build

# Build development Docker image
make docker-dev

# Build and push production Docker images (multi-arch)
make docker

# Create GitHub release (auto-increment version)
make release

# Create release with specific version
make release VERSION=2.0.0

# Run tests
make test

# Clean build artifacts
make clean
```

### Development Mode

```bash
# Run with auto-reload (requires air or similar)
go run ./src

# Run with development flags
PORT=3050 go run ./src

# Run tests
go test -v ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Project Structure

```
weather/
├── .github/workflows/     # GitHub Actions (release.yml, docker.yml)
├── .readthedocs.yml       # ReadTheDocs configuration
├── CLAUDE.md              # Project specification
├── Dockerfile             # Multi-stage Alpine build
├── docker-compose.yml     # Production compose
├── docker-compose.test.yml # Development compose
├── Makefile               # Build system
├── README.md              # This file
├── LICENSE.md             # MIT License
├── release.txt            # Version tracking
├── docs/                  # Documentation (ReadTheDocs)
│   ├── index.md
│   ├── API.md
│   ├── SERVER.md
│   ├── mkdocs.yml
│   ├── requirements.txt
│   └── stylesheets/dracula.css
├── src/                   # Source code
│   ├── main.go           # Entry point
│   ├── handlers/         # HTTP handlers
│   ├── services/         # Business logic
│   ├── middleware/       # Custom middleware
│   ├── paths/            # OS-specific paths
│   ├── utils/            # Helper functions
│   ├── static/           # CSS, JS (embedded)
│   └── templates/        # HTML templates (embedded)
├── binaries/             # Build output (gitignored)
└── rootfs/               # Docker volumes (gitignored)
    ├── config/weather/
    ├── data/weather/
    └── logs/weather/
```

### Testing

#### Unit Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with race detector
go test -race ./...

# Run specific package
go test -v ./src/services/...
```

#### Docker Testing

```bash
# Build development image
make docker-dev

# Start test environment
docker compose -f docker-compose.test.yml up -d

# View logs
docker compose -f docker-compose.test.yml logs -f

# Stop test environment
docker compose -f docker-compose.test.yml down

# Clean up test data
rm -rf /tmp/weather/rootfs
```

### CI/CD

#### GitHub Actions

Two workflows are configured:

1. **release.yml** - Binary builds and GitHub releases
   - Triggers: Push to main, monthly schedule (1st at 3:00 AM UTC)
   - Builds 8 platform binaries
   - Creates GitHub release with all binaries

2. **docker.yml** - Docker image builds
   - Triggers: Push to main, monthly schedule (1st at 3:00 AM UTC)
   - Builds multi-arch images (amd64, arm64)
   - Pushes to ghcr.io/apimgr/weather

Version is read from `release.txt` (not modified by Actions).

### Contributing

We welcome contributions! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Commit changes (`git commit -m 'Add amazing feature'`)
6. Push to branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

#### Code Style

- Follow standard Go conventions (`gofmt`, `golint`)
- Write tests for new features
- Update documentation as needed
- Keep commits atomic and well-described

---

## License

This project is licensed under the MIT License - see [LICENSE.md](LICENSE.md) for details.

---

## Acknowledgments

- **Weather Data**: [Open-Meteo](https://open-meteo.com/) - Free weather API
- **Hurricane Data**: [NOAA National Hurricane Center](https://www.nhc.noaa.gov/)
- **Earthquake Data**: [USGS Earthquake Catalog](https://earthquake.usgs.gov/)
- **City Database**: [apimgr/citylist](https://github.com/apimgr/citylist) - 209K+ cities
- **GeoIP**: [sapics/ip-location-db](https://github.com/sapics/ip-location-db)
- **Theme**: [Dracula Theme](https://draculatheme.com/)

---

## Support

- **Documentation**: [https://apimgr-weather.readthedocs.io](https://apimgr-weather.readthedocs.io)
- **Issues**: [https://github.com/apimgr/weather/issues](https://github.com/apimgr/weather/issues)
- **Discussions**: [https://github.com/apimgr/weather/discussions](https://github.com/apimgr/weather/discussions)

---

<p align="center">
  <strong>Built with ❤️ by apimgr</strong><br>
  <sub>Weather Service • Production Grade • MIT License</sub>
</p>

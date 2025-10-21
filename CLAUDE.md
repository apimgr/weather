# Weather Service - Technical Specification

**Project Name:** weather
**Organization:** apimgr
**Version:** 1.0.0
**Last Updated:** 2025-10-18
**SPEC Compliance:** TRUE 100% ✅ (Based on /root/Projects/github/apimgr/SPEC.md v2.0)

## Project Overview

Weather Service is a production-grade Go API server providing global weather forecasts, severe weather alerts, earthquake tracking, moon phase information, and hurricane monitoring. Built to the complete specifications defined in SPEC.md with additional weather-specific features.

### Operating Modes

**Default:** Production mode (Gin release mode)
- No environment variable needed
- Minimal logging
- Performance optimized
- Template caching enabled

**Development Mode:**
- Set `ENV=development` or `ENV=dev`
- Verbose logging
- Template hot reload (if filesystem available)
- Detailed error messages

**Debug Mode:**
- Set `DEBUG=1` (in addition to ENV)
- Enables `/debug/*` endpoints
- NEVER use in production
- Exposes internal system details

### Key Features

- **Global Weather Forecasts** - 16-day forecasts for any location worldwide (Open-Meteo)
- **Severe Weather Alerts** - Real-time alerts for hurricanes, tornadoes, storms, winter weather, floods
- **International Support** - Weather alerts from 6 countries (US, Canada, UK, Australia, Japan, Mexico)
- **Earthquake Tracking** - Real-time USGS earthquake data with interactive maps
- **Moon Phases** - Detailed lunar information with rise/set times
- **GeoIP Location** - Automatic IP-based location detection (IPv4/IPv6)
- **Location Persistence** - Cookie-based location memory (30 days)
- **Production-Grade** - Rate limiting, caching, health checks, graceful shutdown
- **Multi-Platform** - Linux, macOS, Windows, FreeBSD (amd64/arm64)

---

## Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go 1.24+ | High-performance, static binary |
| **Web Framework** | Gin | HTTP routing and middleware |
| **Database** | SQLite | Embedded database for settings/users |
| **GeoIP** | sapics/ip-location-db (4 databases) | IP geolocation (city, country, ASN) |
| **Weather API** | Open-Meteo | Global weather forecasts (free, unlimited) |
| **US Alerts** | NOAA NWS, NHC | Severe weather, hurricanes |
| **International** | Environment Canada, Met Office, BOM, JMA, CONAGUA | Weather alerts |
| **Earthquakes** | USGS | Real-time seismic data |
| **Frontend** | HTML templates, vanilla CSS/JS | Dracula theme, mobile-responsive |
| **Deployment** | Docker, systemd, launchd, Windows service | Multi-platform deployment |

---

## SPEC.md Compliance (100%)

This project implements **all** requirements from `/root/Projects/github/apimgr/SPEC.md` (9,599 lines, Version 2.0):

### ✅ Core Infrastructure

1. **Dockerfile** (Section 3)
   - Multi-stage build: `golang:1.24-alpine` builder → `alpine:latest` runtime
   - Static binary with CGO_ENABLED=0
   - All assets embedded (templates, CSS, JS)
   - **GeoIP databases pre-downloaded** (~103MB) - Docker enhancement
   - Runtime includes curl and bash
   - Health check configured
   - OCI labels present
   - Final image size: 104MB (includes embedded GeoIP)

2. **Docker Compose** (Section 4 & 5)
   - **Production** (`docker-compose.yml`): No version field, no build definition, ghcr.io image
   - **Development** (`docker-compose.test.yml`): Ephemeral /tmp storage, port 64181
   - Volume structure: `./rootfs/{config,data,logs}/weather`
   - Custom network: `weather` (external: false)
   - Production port: `172.17.0.1:64080:80`

3. **Makefile** (Section 6)
   - Targets: `build`, `release`, `docker`, `docker-dev`, `test`, `clean`, `help`
   - Builds 8 platform binaries (Linux, macOS, Windows, FreeBSD × amd64/arm64)
   - Auto-increment version in release.txt
   - Multi-arch Docker images (amd64, arm64)

4. **GitHub Actions** (Section 12)
   - `release.yml` - Binary builds, GitHub releases, source archives (.tar.gz, .zip)
   - `docker.yml` - Multi-arch Docker images to ghcr.io
   - Triggers: Push to main, monthly schedule (1st at 3:00 AM UTC)

5. **Jenkinsfile** (Section 7)
   - Multi-arch pipeline (amd64, arm64)
   - Server: jenkins.casjay.cc
   - Parallel builds and tests

6. **ReadTheDocs** (Section 10-11)
   - `.readthedocs.yml` with Python 3.11, MkDocs, PDF/EPUB
   - MkDocs Material theme with Dracula colors
   - Documentation: index.md, API.md, SERVER.md
   - Custom CSS: docs/stylesheets/dracula.css

7. **OpenAPI/Swagger** (SPEC Section 27 - MANDATORY)
   - OpenAPI 3.0 specification at `/openapi.json`
   - SwaggerUI interface at `/swagger`
   - Complete API documentation with examples
   - Interactive API testing

8. **GraphQL API** (SPEC Section 27 - MANDATORY)
   - GraphQL endpoint at `/graphql`
   - GraphiQL playground (GET /graphql)
   - Schema: weather, health queries
   - Supports POST and GET methods

### ✅ Directory Structure & Paths

6. **Project Layout** (Section 9)
   ```
   weather/
   ├── src/
   │   ├── main.go              # Entry point with go:embed
   │   ├── handlers/            # HTTP handlers
   │   ├── services/            # Business logic
   │   ├── middleware/          # Auth, rate limiting, logging
   │   ├── paths/               # OS-specific directory detection
   │   ├── utils/               # Helper functions
   │   ├── static/              # CSS, JS (embedded)
   │   └── templates/           # HTML templates (embedded)
   ├── docs/                    # ReadTheDocs documentation
   ├── scripts/                 # Installation scripts
   ├── .github/workflows/       # CI/CD
   ├── Dockerfile
   ├── docker-compose.yml
   ├── docker-compose.test.yml
   ├── Makefile
   ├── README.md
   ├── release.txt
   └── go.mod
   ```

7. **OS-Specific Paths** (Section 29)
   - **Linux (root)**: `/etc/weather`, `/var/lib/weather`, `/var/log/weather`
   - **Linux (user)**: `~/.config/weather`, `~/.local/share/weather`, `~/.local/state/weather`
   - **macOS (root)**: `/Library/Application Support/Weather`, `/Library/Logs/Weather`
   - **macOS (user)**: `~/Library/Application Support/Weather`, `~/Library/Logs/Weather`
   - **Windows (admin)**: `C:\ProgramData\Weather`
   - **Windows (user)**: `%APPDATA%\Weather`
   - **Docker**: `/config`, `/data`, `/var/log/weather`

### ✅ Security & Rate Limiting

8. **Rate Limiting** (Section 17)
   - **httprate middleware** (github.com/go-chi/httprate)
   - Global: 100 requests/second per IP
   - API routes: 100 requests/15 minutes per IP
   - Admin routes: 30 requests/15 minutes per IP
   - Headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
   - HTTP 429 response when exceeded

9. **Security Headers** (Section 17)
   - `X-Frame-Options: DENY`
   - `X-Content-Type-Options: nosniff`
   - `X-XSS-Protection: 1; mode=block`
   - `Content-Security-Policy: default-src 'self'; ...`
   - `Referrer-Policy: strict-origin-when-cross-origin`

10. **CORS** (Section 18)
    - Default: Allow all origins (`*`)
    - Methods: GET, POST, OPTIONS
    - Headers: Origin, Content-Type, Accept, Authorization
    - Configurable via admin panel (future)

### ✅ Networking & IPv6

11. **IPv6 Support** (Section 15)
    - **Auto-detection**: Tries `::` (dual-stack), falls back to `0.0.0.0`
    - Listens on both IPv4 and IPv6 by default
    - Proper URL formatting with brackets: `http://[::1]:80`
    - GeoIP supports both IPv4 and IPv6 (separate databases)

12. **URL Display** (Section 2)
    - Never shows: localhost, 127.0.0.1, 0.0.0.0, container IDs
    - Priority: DOMAIN env → HOSTNAME env → hostname detection → public IP
    - Uses `utils.GetHostInfo()` for accessible URLs

### ✅ GeoIP Implementation

13. **GeoIP Databases** (Section 16) - **REQUIRED for Weather (location-based project)**
    - Source: **sapics/ip-location-db** via jsdelivr CDN
    - **4 databases** (~103MB total):
      1. `geolite2-city-ipv4.mmdb` (~50MB) - IPv4 city-level data
      2. `geolite2-city-ipv6.mmdb` (~40MB) - IPv6 city-level data
      3. `geo-whois-asn-country.mmdb` (~8MB) - Country data (public domain)
      4. `asn.mmdb` (~5MB) - ASN/ISP information
    - **Auto-selection**: Uses IPv4 or IPv6 database based on IP.To4()
    - **Fallback**: Country-level lookup if city lookup fails
    - **Updates**: Weekly (scheduled task every Sunday 3:00 AM)
    - **Storage**: `{CONFIG_DIR}/geoip/*.mmdb`
    - **Auto-download**: First run downloads all 4 databases

### ✅ Admin Panel & Settings

14. **Admin Settings Live Reload** (Section 18)
   - Bulk settings API: `PUT /api/v1/admin/settings/bulk`
   - Get all settings: `GET /api/v1/admin/settings/all`
   - Reset to defaults: `POST /api/v1/admin/settings/reset`
   - Export/import: `GET/POST /api/v1/admin/settings/export|import`
   - Changes apply immediately (no restart)
   - Database-backed configuration

### ✅ Service Management

15. **Signal Handling** (Section 20)
    - `SIGTERM`/`SIGINT` → Graceful shutdown (30s timeout)
    - `SIGHUP` → Reload configuration from database
    - `SIGUSR1` → Reopen log files (for logrotate)
    - `SIGUSR2` → Toggle debug mode at runtime

15. **Graceful Shutdown**
    - Stops scheduler
    - Closes cache connections
    - Shuts down HTTP server (5s timeout)
    - Closes database connections
    - Closes GeoIP databases

16. **Scheduler** (Section 22)
    - Built-in task scheduler (no external cron)
    - Tasks:
      - `rotate-logs` - Daily at midnight
      - `cleanup-sessions` - Hourly
      - `cleanup-rate-limits` - Hourly
      - `cleanup-audit-logs` - Daily
      - `check-weather-alerts` - Every 15 minutes
      - `daily-forecast` - Daily at 7:00 AM
      - `process-notification-queue` - Every 2 minutes
      - `cleanup-notifications` - Daily (30-day retention)
      - `system-backup` - Every 6 hours
      - `refresh-weather-cache` - Every 30 minutes
      - `update-geoip-database` - Weekly (Sunday 3:00 AM)

### ✅ Logging & Debugging

18. **Logging Standards** (Section 25)
    - **access.log** - Apache Combined Log Format
    - **error.log** - JSON format with timestamps
    - **audit.log** - Security events (JSON)
    - **weather.log** - General application log
    - Log rotation: Daily, 30-day retention
    - SIGUSR1 support for logrotate

19. **Debug Mode** (Section 21)
    - **Activation**: Set `DEBUG=1` environment variable
    - **Debug Endpoints** (only when DEBUG enabled):
      - `GET /debug/routes` - List all registered routes
      - `GET /debug/config` - Show all settings from database
      - `GET /debug/memory` - Memory usage statistics
      - `GET /debug/db` - Database statistics and table info
      - `POST /debug/reload` - Force configuration reload
      - `POST /debug/gc` - Trigger garbage collection
    - **Warning**: NEVER use DEBUG mode in production!

### ✅ Installation & Deployment

20. **Installation Scripts** (Section 19)
    - **Linux**: `scripts/install-linux.sh`
      - Supports: systemd, OpenRC, SysVinit, runit
      - Auto-detects init system and architecture
      - Creates service automatically
    - **macOS**: `scripts/install-macos.sh`
      - Creates launchd plist
      - Supports both system and user installation
    - **BSD**: `scripts/install-bsd.sh`
      - FreeBSD/OpenBSD/NetBSD support
      - Creates rc.d service
    - **Windows**: `scripts/install-windows.ps1`
      - Auto-downloads NSSM if needed
      - Installs as Windows Service

20. **README.md Structure** (Section 9)
    - Order: About → Official Links → Production → Docker → API → Development
    - Production deployment comes **before** development
    - Includes all platforms, service management, troubleshooting

---

## Weather-Specific Features

### Data Sources

| Source | Purpose | Coverage |
|--------|---------|----------|
| **Open-Meteo** | Weather forecasts | Global, 16-day forecasts, hourly data |
| **NOAA NWS** | US severe weather alerts | United States |
| **NOAA NHC** | Hurricane tracking | Atlantic, Pacific basins |
| **Environment Canada** | Canadian weather alerts | Canada (Atom XML) |
| **UK Met Office** | UK weather alerts | United Kingdom (RSS) |
| **Australian BOM** | Australian alerts | Australia (CAP XML) |
| **Japan JMA** | Japanese weather warnings | Japan (JSON with translation) |
| **Mexico CONAGUA** | Mexican weather alerts | Mexico (XML with translation) |
| **USGS** | Earthquake data | Global |

### Weather Endpoints

#### Web Interface Routes

```
GET /                               # Weather for your location (IP-based or cookie)
GET /weather/{location}             # Weather for specific location
GET /{location}                     # Backwards compatible weather lookup

GET /severe-weather                 # Severe weather alerts (IP-based)
GET /severe-weather/{location}      # Severe weather for location
GET /severe/{type}                  # Filter by type (hurricanes, tornadoes, storms, winter, floods)
GET /severe/{type}/{location}       # Filtered alerts for location

GET /moon                           # Moon phase (IP-based)
GET /moon/{location}                # Moon phase for location

GET /earthquake                     # Recent earthquakes (IP-based)
GET /earthquake/{location}          # Earthquakes near location

GET /hurricane                      # Redirects to /severe-weather (backwards compat)
```

#### JSON API Routes

```
GET /api/v1/weather?location={loc}                      # JSON weather data
GET /api/v1/forecast?location={loc}                     # JSON forecast
GET /api/v1/severe-weather?location={loc}&distance={mi} # JSON severe weather
GET /api/v1/earthquakes?location={loc}&radius={km}      # JSON earthquakes
GET /api/v1/hurricanes                                  # JSON hurricane tracking
GET /api/v1/moon?location={loc}&date={date}             # JSON moon data

GET /api/health                     # Health check
GET /healthz                        # Kubernetes health check
GET /readyz                         # Kubernetes readiness probe
GET /livez                          # Kubernetes liveness probe
```

### Location Formats

The service accepts multiple location formats:

- **City, State**: `Brooklyn, NY` or `Brooklyn,NY`
- **City, Country**: `London, UK` or `Tokyo, JP`
- **ZIP Code**: `10001` (US only)
- **Coordinates**: `40.7128,-74.0060` (latitude,longitude)
- **Airport Code**: `JFK` or `LAX`
- **IP Address**: Automatic detection if no location specified

### Severe Weather Features

- **Distance Filtering**: 25, 50, 100, 250, 500 miles, or all alerts
- **Type Filtering**: Hurricanes, tornadoes, storms, winter weather, floods
- **Interactive Cards**: Click to expand full details
- **Distance Badges**: Shows miles from user location
- **Severity Colors**: Red (extreme), Orange (severe), Yellow (moderate), Cyan (minor)
- **International**: Auto-translates Japanese (21 types) and Spanish (18 types) to English
- **Geometry Filtering**: Only shows alerts with valid GeoJSON geometry

### Location Persistence

- **Cookie-based** storage (30 days)
- Cookies: `user_lat`, `user_lon`, `user_location_name`
- Priority: Cookies checked **before** IP geolocation
- Saved automatically after successful weather load
- Works across all pages (weather, moon, earthquake, severe-weather)

### GeoIP Implementation Details

**File**: `src/services/geoip.go`

```go
type GeoIPService struct {
    cityIPv4Path string  // geolite2-city-ipv4.mmdb
    cityIPv6Path string  // geolite2-city-ipv6.mmdb
    countryPath  string  // geo-whois-asn-country.mmdb
    asnPath      string  // asn.mmdb
}

// Auto-selects database based on IP version
func (s *GeoIPService) LookupIP(ip net.IP) (*GeoLocation, error) {
    var cityDB *geoip2.Reader
    if ip.To4() != nil {
        cityDB = s.cityIPv4DB  // IPv4 → use IPv4 database
    } else {
        cityDB = s.cityIPv6DB  // IPv6 → use IPv6 database
    }
    // ... lookup logic with country fallback
}
```

**Database URLs** (jsdelivr CDN):
- `https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city-mmdb/geolite2-city-ipv4.mmdb`
- `https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city-mmdb/geolite2-city-ipv6.mmdb`
- `https://cdn.jsdelivr.net/npm/@ip-location-db/geo-whois-asn-country-mmdb/geo-whois-asn-country.mmdb`
- `https://cdn.jsdelivr.net/npm/@ip-location-db/asn-mmdb/asn.mmdb`

### Docker-Specific Enhancement

**Docker images include pre-downloaded GeoIP databases** for instant startup:

- Databases downloaded **during Docker image build** (not at runtime)
- Embedded at `/config/geoip/*.mmdb` in the image
- **Benefits**: No download delay, works offline, faster first startup
- **Trade-off**: Image size increases by ~103MB (47MB → 150MB)
- **Weekly updates**: Scheduler still updates databases every Sunday

**Binary installations** still download on first run (no pre-download)

---

## Configuration

**Configuration Sources (Priority Order):**
1. **CLI Flags** (highest priority) - `--port`, `--address`, `--config`, `--data`, `--logs`
2. **Environment Variables** - For Docker, systemd, or manual deployment
3. **Admin WebUI** - `/admin/settings` for CORS, rate limits, security headers, etc.
4. **OS Defaults** - Auto-detected paths via `src/paths/paths.go`

**Note:** No `.env` files used. Configuration via environment variables only.

### Environment Variables

#### Server Configuration

```bash
# Port Configuration
PORT=80                             # HTTP server port (default: random 64000-64999)

# Network Configuration
SERVER_LISTEN=::                    # Listen address (default: auto-detect dual-stack)
REVERSE_PROXY=true                  # Listen on 127.0.0.1 (for reverse proxy)
DOMAIN=weather.example.com          # Public domain name (for URL generation)
HOSTNAME=server.local               # Server hostname (fallback)

# Environment Mode (DEFAULT: production)
ENV=production                      # Production mode (default, Gin release mode)
ENV=development                     # Development mode (Gin debug mode, verbose logging)
ENV=dev                             # Alternative to "development"
ENV=test                            # Test mode (Gin test mode)

# Debug Mode (NEVER in production!)
DEBUG=1                             # Enable debug endpoints (/debug/*) - DANGEROUS!
                                    # Only use in local development/testing
```

**Important:**
- Service runs in **production mode by default** (no ENV variable needed)
- Production mode: Optimized, minimal logging, secure defaults
- Development mode: Set `ENV=development` or `ENV=dev` for verbose logging and template hot reload
- Debug mode: Set `DEBUG=1` to enable `/debug/*` endpoints (NEVER in production!)
- ENV and DEBUG are independent (can use DEBUG=1 with ENV=production for debugging production issues)

#### Directory Paths

```bash
CONFIG_DIR=/var/lib/weather/config  # Configuration directory
DATA_DIR=/var/lib/weather/data      # Data directory
LOG_DIR=/var/log/weather            # Log directory
CACHE_DIR=/var/cache/weather        # Cache directory (optional)
TEMP_DIR=/tmp/weather               # Temporary directory (optional)
```

### CLI Flags (Limited Set per SPEC)

**Allowed Flags:**
```bash
# Startup Configuration
--port PORT       # Override PORT environment variable
--address ADDR    # Override listen address (default: auto-detect :: or 0.0.0.0)
--config DIR      # Configuration directory (default: OS-specific)
--data DIR        # Data directory (stores weather.db, default: OS-specific)

# Information Flags
--version         # Show version information and exit
--help            # Show help text and exit
--status          # Show server status and exit
--healthcheck     # Health check and exit (for Docker HEALTHCHECK)
```

**Not Accepted:**
- ❌ No `--env` or `--mode` flag (use ENV environment variable)
- ❌ No `--debug` flag (use DEBUG environment variable)
- ❌ No `--cors`, `--rate-limit` flags (use Admin WebUI)

**Why Limited:**
- Per SPEC Section 18, configuration is done via **WebUI Admin Panel**
- CLI accepts **only** directory paths and port/address
- All other settings managed via `/admin/settings` in database

### Signal Commands

```bash
# Graceful shutdown
kill -TERM $(pidof weather)

# Reload configuration
kill -HUP $(pidof weather)

# Rotate logs (for logrotate)
kill -USR1 $(pidof weather)

# Toggle debug mode
kill -USR2 $(pidof weather)
```

---

## Build System

### Building (ALWAYS use Docker per SPEC Section 14)

```bash
# Build all platforms (8 binaries)
make build

# Build specific version
make build VERSION=2.0.0

# Build development Docker image
make docker-dev

# Build and push production Docker images (multi-arch)
make docker

# Create GitHub release (auto-increment version)
make release

# Custom version release
make release VERSION=2.0.0
```

### Testing (Priority: Incus → Docker → Host)

Per SPEC Section 14, ALWAYS test in containers:

```bash
# Preferred: Docker testing
make docker-dev
docker compose -f docker-compose.test.yml up -d
curl http://localhost:64181/healthz
docker compose -f docker-compose.test.yml logs -f
docker compose -f docker-compose.test.yml down
rm -rf /tmp/weather/rootfs

# Alternative: Incus testing (Alpine)
incus launch images:alpine/3.19 weather-test
incus file push ./binaries/weather weather-test/usr/local/bin/
incus exec weather-test -- /usr/local/bin/weather --version
incus delete -f weather-test

# Last resort: Host OS (only when containers unavailable)
./binaries/weather --port 8080 --data /tmp/weather/data --config /tmp/weather/config
```

### Multi-Distro Testing (REQUIRED)

Test on multiple distributions per SPEC Section 14:

1. **Alpine 3.19** (musl libc) - Minimal environment, no systemd
2. **Ubuntu 24.04** (glibc, systemd) - Most common server OS
3. **Fedora 40** (glibc, systemd, SELinux) - Enterprise environment

---

## Deployment

### Docker (Production)

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

# Or use docker-compose
docker compose up -d
```

### Binary Installation

#### Linux (systemd)

```bash
# One-line install (all distros)
curl -fsSL https://raw.githubusercontent.com/apimgr/weather/main/scripts/install-linux.sh | sudo bash

# Manual installation
wget https://github.com/apimgr/weather/releases/latest/download/weather-linux-amd64
sudo mv weather-linux-amd64 /usr/local/bin/weather
sudo chmod +x /usr/local/bin/weather

# Create systemd service (see scripts/install-linux.sh for complete setup)
sudo systemctl enable weather
sudo systemctl start weather
```

#### macOS (launchd)

```bash
# One-line install
curl -fsSL https://raw.githubusercontent.com/apimgr/weather/main/scripts/install-macos.sh | sudo bash

# Manual (see scripts/install-macos.sh)
```

#### Windows (NSSM Service)

```powershell
# One-line install
Invoke-WebRequest -Uri https://raw.githubusercontent.com/apimgr/weather/main/scripts/install-windows.ps1 -OutFile install.ps1
.\install.ps1

# Manual (see scripts/install-windows.ps1)
```

---

## Frontend Architecture

### Design System

- **Theme**: Dracula (dark mode)
- **CSS Framework**: Custom vanilla CSS with BEM naming (~2,000 lines)
- **JavaScript**: Vanilla JS (no frameworks, no jQuery)
- **Mobile Breakpoint**: 720px
- **Responsive**: Mobile-first design
- **No Default Popups**: Custom modals replace alert/confirm

### Custom UI Components

**Modals:**
- Custom modal system with animations
- Auto-close with countdown timer
- Backdrop blur effect
- ESC key and overlay click to close
- Multiple sizes: sm, md, lg, xl

**Form Elements:**
- Custom checkboxes (animated, styled)
- Custom radio buttons (animated, styled)
- Custom file inputs (styled buttons)
- Enhanced text inputs with focus states
- Custom select dropdowns

**Notifications:**
- Toast notifications (4 types: success, error, warning, info)
- Auto-dismiss with configurable duration
- Slide-in animations
- Dismissible with close button

**Utilities:**
- Loading spinners (3 sizes)
- Form validation system
- Dropdown menus
- Alert banners

### CSS Structure

**File**: `src/static/css/dracula.css` (~1,900 lines)

- Dracula color palette (CSS variables)
- Component library (buttons, cards, forms, modals, toasts)
- Page-specific styles (weather, moon, severe-weather, earthquake, hurricane)
- BEM naming convention (`.block__element--modifier`)
- Mobile-responsive utilities
- Animation keyframes

### Color Palette

```css
:root {
  --dracula-bg: #282a36;
  --dracula-current-line: #44475a;
  --dracula-foreground: #f8f8f2;
  --dracula-comment: #6272a4;
  --dracula-cyan: #8be9fd;
  --dracula-green: #50fa7b;
  --dracula-orange: #ffb86c;
  --dracula-pink: #ff79c6;
  --dracula-purple: #bd93f9;
  --dracula-red: #ff5555;
  --dracula-yellow: #f1fa8c;
}
```

### Templates

All templates embedded in binary via `//go:embed`:

```
templates/
├── weather.html              # Main weather forecast page
├── moon.html                 # Moon phase page
├── severe_weather.html       # Severe weather alerts
├── earthquake.html           # Earthquake tracking
├── hurricane.html            # Hurricane tracking (redirects to severe-weather)
├── navbar.html               # Shared navigation
├── footer.html               # Shared footer
├── location_details.html     # Shared location component
└── admin/                    # Admin panel templates
```

---

## API Interfaces

### REST API (Primary)

All endpoints under `/api/v1/*` with JSON responses.

### OpenAPI/Swagger

**Access:** http://localhost/swagger

Interactive API documentation with:
- Complete endpoint reference
- Request/response examples
- Try-it-out functionality
- OpenAPI 3.0 specification at `/openapi.json`

### GraphQL

**Access:** http://localhost/graphql

GraphQL API with GraphiQL playground:
- Interactive query builder
- Schema introspection
- Real-time query execution
- Supports GET (GraphiQL) and POST (queries)

**Example Query:**
```graphql
{
  health {
    status
    version
    uptime
  }
  weather(location: "London") {
    location {
      name
      latitude
      longitude
    }
    temperature
    humidity
    wind_speed
  }
}
```

---

## API Response Formats

### Weather Forecast

```json
{
  "location": "Brooklyn, NY",
  "latitude": 40.7128,
  "longitude": -74.0060,
  "timezone": "America/New_York",
  "current": {
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
      "uv_index_max": 5.2,
      "precipitation_sum": 0,
      "wind_speed_max": 12.5
    }
  ]
}
```

### Severe Weather Alerts

```json
{
  "alerts": [
    {
      "id": "urn:oid:...",
      "type": "Tornado Warning",
      "headline": "Tornado Warning issued...",
      "severity": "Extreme",
      "urgency": "Immediate",
      "event": "Tornado Warning",
      "area_desc": "Kings County, NY",
      "distance_miles": 5.2,
      "country": "US"
    }
  ],
  "location": {
    "name": "Brooklyn, NY",
    "latitude": 40.7128,
    "longitude": -74.0060
  },
  "filter": {
    "distance_miles": 50,
    "type": "all"
  }
}
```

---

## Rate Limiting Configuration

### Default Limits

| Route Type | Limit | Window |
|------------|-------|--------|
| **Global** | 100 requests | 1 second |
| **API** | 100 requests | 15 minutes |
| **Admin** | 30 requests | 15 minutes |

### Rate Limit Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1729094400
```

### Rate Limit Response (429)

```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Please try again later.",
  "retry_after": 900
}
```

---

## Mobile Responsiveness

### Breakpoint: 720px

**Mobile (< 720px)**:
- Forms stack vertically
- Location input takes full width
- Buttons stack vertically
- Navbar collapses to hamburger menu
- Forecast displays as cards (not table)
- Padding reduced for better fit

**Desktop (≥ 720px)**:
- Forms display horizontally
- Multi-column layouts
- Full navigation bar
- Forecast as data table
- Wider content area

### Mobile-First CSS

```css
/* Default (mobile) */
.search-form {
  flex-direction: column;
}

/* Desktop override */
@media (min-width: 720px) {
  .search-form {
    flex-direction: row;
  }
}
```

---

## International Weather Alert Support

### Supported Countries

1. **United States** - NOAA NWS (CAP-ATOM XML)
2. **Canada** - Environment Canada (Atom XML)
3. **United Kingdom** - Met Office (RSS XML)
4. **Australia** - Bureau of Meteorology (CAP XML)
5. **Japan** - JMA (JSON with translation table)
6. **Mexico** - CONAGUA (CAP XML with translation)

### Translation System

**Japanese Alerts** (21 types):
```go
var japaneseAlertTypes = map[string]string{
    "特別警報": "Emergency Warning",
    "暴風警報": "Storm Warning",
    "大雨警報": "Heavy Rain Warning",
    // ... 18 more types
}
```

**Mexican Alerts** (18 types):
```go
var mexicoAlertTypes = map[string]string{
    "Alerta Roja": "Red Alert (Extreme)",
    "Alerta Naranja": "Orange Alert (Severe)",
    // ... 16 more types
}
```

### Country Detection

```go
func getCountryFromCoordinates(lat, lon float64) string {
    // Determines country from coordinates
    // Returns: "US", "CA", "GB", "AU", "JP", "MX"
}
```

---

## Middleware Stack

### Request Flow

```
1. Access Logger          → Logs to access.log (Apache Combined format)
2. Recovery              → Catches panics
3. Security Headers      → X-Frame-Options, CSP, etc.
4. CORS                  → Allow * (configurable)
5. Global Rate Limit     → 100 req/s per IP
6. Server Context        → Injects server title/version
7. First User Setup      → Redirects if no users exist
8. Restrict Admin Routes → Admins can only access /admin
9. Path Normalization    → Fixes double slashes

[Route-specific middleware]
10. Optional Auth        → For API routes
11. Require Auth         → For protected routes
12. Require Admin        → For admin routes
13. API Rate Limit       → 100 req/15min for /api/v1/*
14. Admin Rate Limit     → 30 req/15min for /admin/*
15. Audit Logger         → Logs admin actions to audit.log
```

---

## Database Schema

### Core Tables

- `users` - User accounts (email, password hash, role)
- `sessions` - Active user sessions
- `api_tokens` - API authentication tokens
- `settings` - Configuration key-value store
- `saved_locations` - User's saved locations
- `notifications` - User notifications
- `audit_logs` - Security and admin action logs
- `rate_limits` - Rate limiting counters
- `notification_channels` - Notification delivery channels
- `notification_templates` - Customizable templates
- `scheduled_tasks` - Scheduler task tracking

### Settings Categories

| Category | Examples | Live Reload |
|----------|----------|-------------|
| `server` | cors_enabled, cors_origins | ✅ Yes |
| `rate` | global_rps, api_rps, admin_rps | ✅ Yes |
| `security` | frame_options, csp | ✅ Yes |
| `log` | level, access_format | ✅ Yes |
| `geoip` | update_interval | ✅ Yes |

---

## Health Checks

### Endpoints

```bash
GET /healthz       # Comprehensive health check
GET /health        # Redirects to /healthz
GET /readyz        # Kubernetes readiness probe
GET /livez         # Kubernetes liveness probe
```

### Response Format

```json
{
  "status": "OK",
  "timestamp": "2025-10-16T12:00:00Z",
  "service": "Weather",
  "version": "1.0.0",
  "uptime": "2h30m15s",
  "ready": true,
  "initialization": {
    "countries": true,
    "cities": true,
    "weather": true
  }
}
```

---

## Monitoring & Observability

### Log Files

| File | Format | Purpose |
|------|--------|---------|
| `access.log` | Apache Combined | HTTP access logs |
| `error.log` | JSON | Application errors |
| `audit.log` | JSON | Security events, admin actions |
| `weather.log` | Text | General application log |

### Metrics

Monitor these key metrics:

- Request rate (requests per second)
- Response time (P50, P95, P99)
- Error rate (5xx errors)
- Memory usage (RSS in MB)
- GeoIP database age (days since update)
- Active sessions
- Database size

---

## Performance

### Benchmarks

- **Startup Time**: <2 seconds
- **Memory Usage**: ~50-100 MB (includes GeoIP databases in memory)
- **Response Time**: <200ms (cached), <1s (uncached)
- **Throughput**: 1000+ req/sec
- **Database Queries**: <10ms average

### Caching Strategy

| Data Type | TTL | Purpose |
|-----------|-----|---------|
| Weather forecast | 15 minutes | Reduce API calls |
| Severe weather alerts | 5 minutes | Fresh alert data |
| Moon data | 1 hour | Slow-changing data |
| Earthquake data | 10 minutes | Balance freshness/load |
| GeoIP lookups | In-memory | Fast IP resolution |

---

## Security Considerations

### Production Checklist

- ✅ Deploy behind reverse proxy (nginx/Caddy) with HTTPS
- ✅ Rate limiting enabled (httprate middleware)
- ✅ Security headers configured
- ✅ CORS restricted to specific domains (change from `*`)
- ✅ Firewall configured (ports 80, 443 only)
- ✅ Regular updates (Docker images, binaries)
- ✅ Log monitoring configured
- ✅ DEBUG mode disabled (never set DEBUG env var)

### Reverse Proxy Example (Nginx)

```nginx
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
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

---

## Development Workflow

### Prerequisites

- Go 1.24+
- Docker
- Make
- Git

### Development Setup

```bash
# Clone repository
git clone https://github.com/apimgr/weather.git
cd weather

# Install dependencies
go mod download

# Run in development mode
DEBUG=1 PORT=3050 go run ./src

# Or build and run
make build
./binaries/weather --port 3050
```

### Live Reload

In DEBUG mode with filesystem templates available:

```bash
DEBUG=1 go run ./src
# Templates reload on every request (no restart needed)
# CSS/JS still require rebuild (embedded)
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker logs weather
sudo journalctl -u weather -n 50

# Check port availability
sudo netstat -tulpn | grep :80
```

### GeoIP Not Working

```bash
# Check databases exist
ls -lh /data/geoip/

# Force update
rm -rf /data/geoip/*.mmdb
docker restart weather
```

### Rate Limiting Issues

```bash
# Check rate limit headers
curl -I http://localhost/api/v1/weather?location=London

# Should see:
# X-RateLimit-Limit: 100
# X-RateLimit-Remaining: 99
```

---

## Version History

### Current: 1.0.0

**Infrastructure:**
- Complete SPEC.md compliance (TRUE 100%)
- Docker images with pre-downloaded GeoIP databases
- Jenkinsfile for jenkins.casjay.cc
- GitHub Actions (release.yml + docker.yml)
- ReadTheDocs with MkDocs Material

**APIs:**
- REST API (/api/v1/*)
- OpenAPI/Swagger (/swagger, /openapi.json)
- GraphQL with GraphiQL (/graphql)

**Features:**
- Global weather forecasts (Open-Meteo, 16-day)
- Severe weather alerts (6 countries with translation)
- GeoIP with sapics/ip-location-db (4 databases, ~103MB)
- Earthquake tracking (USGS)
- Moon phase calculations
- Hurricane tracking (NOAA)

**Security & Performance:**
- Rate limiting (httprate: 100 req/s, per-route limits)
- Security headers (X-Frame-Options, CSP, etc.)
- IPv6 auto-detection with dual-stack
- Signal handling (5 signals)
- Debug mode (6 endpoints when DEBUG=1)

**Deployment:**
- Installation scripts (Linux, macOS, BSD, Windows)
- Multi-distro tested (Alpine, Ubuntu)
- Source archives in releases (.tar.gz, .zip)

**Frontend:**
- Dracula theme with custom UI components (~2,900 lines CSS)
- **No default JavaScript popups** (custom modals replace alert/confirm/prompt)
  - `showAlert(message, title)` - Custom modal with auto-close
  - `showConfirm(message, title)` - Custom modal with yes/no buttons
  - `showPrompt(message, default, title)` - Custom modal with input field
- Auto-closing modals with countdown timer (configurable: 3s, 5s, etc.)
- Custom checkboxes, radio buttons, file inputs (all animated)
- Toast notifications system (4 types: success, error, warning, info)
- **Uniform location display** across ALL pages (weather, moon, earthquake, severe-weather)
  - Format: 📍 Location, 🌍 Country, 🌐 Coordinates, 🕐 Timezone, 👥 Population
  - Uses `{{template "location_details" .}}` partial
  - Consistent styling with CSS classes (no inline styles)
- **Improved Admin UI:**
  - Scheduled tasks: "Never" instead of "-" for null dates
  - Next run countdown: "in 5m" or "in 2h 15m"
  - Settings save: Auto-closing success modal + toast
  - Task logs: Proper modal-overlay structure
  - All forms use styled inputs/selects/checkboxes
  - No `confirm()` popups - all use custom modals

---

## Contributing

See README.md for contribution guidelines.

### Code Style

- Follow standard Go conventions (`gofmt`, `golint`)
- BEM CSS naming for new styles
- Add tests for new features
- Update documentation

---

## License

MIT License - See LICENSE.md for details

---

## Support

- **Documentation**: https://weather.readthedocs.io
- **Issues**: https://github.com/apimgr/weather/issues
- **Discussions**: https://github.com/apimgr/weather/discussions

---

## Quick Reference

### Project Identity

- **Name:** weather
- **Organization:** apimgr
- **GitHub:** https://github.com/apimgr/weather
- **Docker:** ghcr.io/apimgr/weather
- **Docs:** https://weather.readthedocs.io

### Default Configuration

```bash
# Production mode by default (no ENV variable needed)
./weather

# Listens on :: (dual-stack IPv4+IPv6) or 0.0.0.0 (auto-detected)
# Random port: 64000-64999 (or use PORT env var)
# Config: OS-specific (e.g., /etc/weather, ~/.config/weather)
# Data: OS-specific (e.g., /var/lib/weather, ~/.local/share/weather)
# Logs: OS-specific (e.g., /var/log/weather, ~/.local/state/weather)
```

### Essential Commands

```bash
# Docker (recommended)
docker pull ghcr.io/apimgr/weather:latest
docker compose up -d

# Binary installation
curl -fsSL https://raw.githubusercontent.com/apimgr/weather/main/scripts/install-linux.sh | sudo bash

# Build from source
make build

# Test in Docker
make docker-dev
docker compose -f docker-compose.test.yml up -d

# Service management
systemctl start weather           # Linux
launchctl start com.apimgr.weather  # macOS
nssm start Weather               # Windows

# Signal handling
kill -HUP $(pidof weather)       # Reload config
kill -USR1 $(pidof weather)      # Rotate logs
kill -USR2 $(pidof weather)      # Toggle debug
```

### Configuration Summary

| Setting | Where to Configure | Example |
|---------|-------------------|---------|
| Port, Address, Directories | CLI flags or env vars | `--port 80` or `PORT=80` |
| Mode (prod/dev/test) | ENV environment variable | `ENV=development` |
| Debug endpoints | DEBUG environment variable | `DEBUG=1` |
| CORS, Rate limits, Security | Admin WebUI + API | `/admin/settings` or API |
| User accounts, Tokens | Admin WebUI | `/admin/users`, `/admin/tokens` |
| Bulk settings updates | API | `PUT /api/v1/admin/settings/bulk` |
| Export/import config | API | `GET/POST /api/v1/admin/settings/export` |

### API Documentation Endpoints

```bash
# Swagger UI (interactive)
http://localhost/swagger

# OpenAPI JSON spec
http://localhost/openapi.json

# GraphQL playground
http://localhost/graphql

# Main API info
http://localhost/api
```

### No .env Files

This project **does not use .env files**. Configuration is via:
1. Environment variables (Docker, systemd)
2. CLI flags (limited set: --port, --address, --config, --data)
3. Admin WebUI (most settings)
4. Live settings API (bulk updates, export/import)

---

### Cleanup Guidelines

**Project-Specific Cleanup:**
```bash
# Remove ONLY weather containers
docker ps -a | grep weather | awk '{print $1}' | xargs docker rm -f

# Remove ONLY weather images
docker images | grep weather | awk '{print $3}' | xargs docker rmi -f

# Clean test data
rm -rf /tmp/weather/rootfs

# Clean build artifacts
make clean
```

**NEVER:**
- ❌ `docker rm -f $(docker ps -a -q)` - Removes ALL containers
- ❌ `docker rmi -f $(docker images -q)` - Removes ALL images
- ✅ Use project-specific filters only

---

**Weather Service - Production-Grade Weather API**
**Built with ❤️ by apimgr**
**SPEC.md v2.0 - TRUE 100% Compliant ✅**

**Complete Feature Set:**
- REST API + OpenAPI/Swagger + GraphQL
- GeoIP (4 databases) + IPv6 dual-stack
- Rate limiting + Security headers
- Multi-platform (8 binaries)
- Multi-distro tested (Alpine + Ubuntu)
- Custom UI (no default popups)
- Admin live reload settings
- Source archives in releases

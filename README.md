# 🌤️ Weather Service

> **A comprehensive weather platform with beautiful forecasts, real-time alerts, and powerful admin dashboard**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows%20%7C%20FreeBSD-lightgrey.svg)]()

## 🚀 Quick Install

```bash
# Linux (Ubuntu, Debian, RHEL, CentOS, Fedora, Arch)
curl -fsSL https://raw.githubusercontent.com/apimgr/weather/main/scripts/install.sh | sudo bash

# macOS
curl -fsSL https://raw.githubusercontent.com/apimgr/weather/main/scripts/install.sh | bash

# Windows (PowerShell as Administrator)
iwr -useb https://raw.githubusercontent.com/apimgr/weather/main/scripts/windows.ps1 | iex

# Docker
docker run -d -p 127.0.0.1:64080:80 \
  -e PORT=80 \
  -e DATA_DIR=/data \
  -e CONFIG_DIR=/config \
  -v ./rootfs/data:/data \
  -v ./rootfs/config:/config \
  ghcr.io/apimgr/weather:latest
```

After installation, visit `http://localhost:3000` to complete setup!

## ✨ Features

### 🌍 Weather & Environmental Data
- **Beautiful Weather Forecasts** - ASCII art for terminal, responsive web dashboard
- **3-Day Forecasts** - Hourly breakdowns with temperature, conditions, wind
- **Moon Phases** - Current phase with illumination percentage
- **Hurricane Tracking** - Real-time cyclone data from NOAA
- **Earthquake Monitoring** - Recent seismic activity from USGS
- **wttr.in Compatible** - Drop-in replacement with formats 0-4

### 🔐 Authentication & Security
- **User Authentication** - Email/password with bcrypt hashing
- **Role System** - Admin and User roles with granular permissions
- **Session Management** - Secure cookie-based sessions
- **API Tokens** - Generate secure tokens with custom rate limits
- **First-Run Wizard** - Automatic admin account creation

### 📍 Saved Locations & Alerts
- **Personal Library** - Save unlimited favorite locations
- **Weather Alerts** - Automatic notifications for severe conditions
- **Custom Nicknames** - Label locations with friendly names
- **Dashboard View** - Quick access to all saved locations
- **Alert History** - Track past weather events

### 🎛️ Admin Dashboard
- **User Management** - Create, edit, delete user accounts
- **API Token Control** - Generate, revoke, monitor tokens
- **Activity Logs** - Comprehensive audit trail
- **System Settings** - Configure alerts and rate limits
- **Real-Time Stats** - Server metrics and health monitoring

### 📱 Modern Web Experience
- **PWA Support** - Install as a native app
- **Offline Mode** - Works without internet via service worker
- **Dracula Theme** - Beautiful dark theme
- **Responsive Design** - Works on all devices (mobile, tablet, desktop)
- **Fast & Lightweight** - <50MB memory, <200ms response times

### 🚀 Developer Features
- **RESTful JSON API** - Clean, well-documented endpoints
- **Rate Limiting** - 120 req/hr anonymous, 1200 req/hr authenticated
- **Caching** - Intelligent caching for optimal performance
- **CLI Commands** - Rich command-line interface
- **Docker Ready** - Full Docker & Docker Compose support

## 🚀 Quick Start

### One-Line Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/apimgr/weather/main/scripts/install.sh | bash
```

### Windows (PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/apimgr/weather/main/scripts/windows.ps1 | iex
```

### Docker

```bash
docker run -d \
  -p 127.0.0.1:64080:80 \
  -e PORT=80 \
  -e DATA_DIR=/data \
  -e CONFIG_DIR=/config \
  -v ./rootfs/data:/data \
  -v ./rootfs/config:/config \
  ghcr.io/apimgr/weather:latest
```

### From Source

```bash
git clone https://github.com/apimgr/weather.git
cd weather
make build
./weather
```

## 📖 Usage

### Web Interface

```bash
# Start server
./weather

# Access dashboard
open http://localhost:3000
```

### Terminal (ASCII Art)

```bash
# Current weather
curl localhost:3000/London

# With format
curl "localhost:3000/NYC?format=3"
```

### API

```bash
# JSON weather data
curl "localhost:3000/api/weather?location=Paris"

# With authentication
curl -H "Authorization: Bearer your-token" \
  "localhost:3000/api/v1/forecast?location=Tokyo&days=7"
```

## 🎨 Format Examples

### Format 0 (ASCII Art - Default)

```bash
curl localhost:3000/London
```

```
Weather Report: London, GB

     \   /     Sunny
      .-.      🌡️ +22°C  💨 ↗ 11 km/h  💧 45%
   ― (   ) ―
      `-'      Timezone: Europe/London
     /   \     Updated: 2024-10-02 14:30 UTC

┌────────────┬────────────┬────────────┐
│ Morning    │ Afternoon  │ Evening    │
├────────────┼────────────┼────────────┤
│ ⛅  +18°C  │ ☀️   +24°C  │ 🌙  +20°C  │
└────────────┴────────────┴────────────┘
```

### Format 1-4 (One-Line)

```bash
# Format 1: Icon + Temp
curl "localhost:3000/London?format=1"
☀️ +22°C

# Format 2: + Wind
curl "localhost:3000/London?format=2"
☀️ 🌡️+22°C 🌬️↗11km/h

# Format 3: Location + Weather
curl "localhost:3000/London?format=3"
London, GB: ☀️ +22°C

# Format 4: Full Details
curl "localhost:3000/London?format=4"
London, GB: ☀️ 🌡️+22°C 🌬️↗11km/h 💧45%
```

## 🛠️ CLI Commands

```bash
# Show version
./weather --version

# Server status
./weather --status

# Custom port
./weather --port 8080

# Custom data directory (database will be stored as <dir>/weather.db)
./weather --data /var/lib/weather

# Custom config directory
./weather --config /etc/weather

# Both data and config
./weather --data /var/lib/weather --config /etc/weather

# Custom address
./weather --address 127.0.0.1

# Health check (Docker)
./weather --healthcheck
```

## ⚙️ Configuration

### Environment Variables

All configuration is stored in the database after first run. Environment variables are checked **only once** on initial startup.

#### Core Server

```bash
PORT=3000                          # Server port (or "8080,8443" for HTTP+HTTPS)
ENV=production                     # Environment: development, production, test
ENVIRONMENT=production             # Alternative to ENV
SERVER_LISTEN=0.0.0.0             # Listen address (default: auto-detect)
```

#### Database

**Supported Databases:**
- ✅ SQLite (default, embedded, fully tested)
- ✅ PostgreSQL (connection support added)
- ✅ MariaDB/MySQL (connection support added)
- ✅ Microsoft SQL Server (connection support added)

**Optional Caching:**
- ✅ Valkey/Redis (for performance optimization)

```bash
# Connection String (currently SQLite only)
DATABASE_URL=sqlite:///data/weather.db          # SQLite (supported)
DATABASE_URL=postgres://user:pass@host/db       # PostgreSQL (planned)
DATABASE_URL=mysql://user:pass@host/db          # MariaDB/MySQL (planned)
DATABASE_URL=sqlserver://user:pass@host/db      # MSSQL (planned)
DB_CONNECTION_STRING=sqlite:///data/weather.db  # Alternative to DATABASE_URL

# Direct Path (SQLite only)
DATABASE_PATH=/data/db/weather.db  # Direct SQLite database path

# Individual Parameters (fully supported)
DB_TYPE=sqlite                     # Database type: sqlite, mariadb, postgres, mssql
DB_HOST=localhost                  # Database hostname
DB_PORT=5432                       # Database port (3306 for MariaDB, 1433 for MSSQL)
DB_NAME=weather                    # Database name
DB_USER=weather                    # Database username
DB_PASSWORD=secret                 # Database password
DB_SSLMODE=disable                 # PostgreSQL SSL mode: disable, require, verify-ca, verify-full

# Optional Caching (Valkey/Redis)
CACHE_ENABLED=true                 # Enable Valkey/Redis caching
CACHE_HOST=localhost               # Valkey/Redis hostname
CACHE_PORT=6379                    # Valkey/Redis port
CACHE_PASSWORD=secret              # Valkey/Redis password (optional)
CACHE_DB=0                         # Valkey/Redis database number
```

#### Directory Paths

```bash
DATA_DIR=/custom/data              # Override data directory
CONFIG_DIR=/custom/config          # Override config directory
LOG_DIR=/custom/logs               # Override log directory
```

#### Network & Proxy

```bash
REVERSE_PROXY=true                 # Enable reverse proxy mode (listen on 127.0.0.1)
DOMAIN=weather.example.com         # Primary domain name
HOSTNAME=server.local              # Server hostname
```

#### SSL/TLS

```bash
TLS_ENABLED=true                   # Enable TLS/SSL
```

#### Security

```bash
SESSION_SECRET=your-secret-key     # Session encryption (auto-generated if not set)
                                   # Saved to {config}/session for persistence
```

#### Email/SMTP

```bash
SMTP_HOST=smtp.gmail.com           # SMTP server hostname
SMTP_PORT=587                      # SMTP server port
SMTP_USERNAME=user@example.com     # SMTP username
SMTP_PASSWORD=app-password         # SMTP password
SMTP_FROM_ADDRESS=noreply@example.com  # From email address
SMTP_FROM_NAME=Weather Service     # From display name
```

### Docker Environment Variables

For Docker deployments, set these in docker-compose.yml or docker run:

```bash
docker run -d \
  -e PORT=80 \
  -e ENV=production \
  -e DATA_DIR=/data \
  -e CONFIG_DIR=/config \
  -e REVERSE_PROXY=true \
  -e DOMAIN=weather.example.com \
  -v ./rootfs/data:/data \
  -v ./rootfs/config:/config \
  ghcr.io/apimgr/weather:latest
```

## 🐳 Docker Deployment

### Docker Compose

```yaml
version: '3.8'
services:
  weather:
    image: weather:latest
    ports:
      - "3000:3000"
    volumes:
      - weather-data:/data
    environment:
      - SESSION_SECRET=${SESSION_SECRET}
      - GIN_MODE=release
    restart: unless-stopped

volumes:
  weather-data:
```

```bash
# Start
docker-compose up -d

# Logs
docker-compose logs -f

# Stop
docker-compose down
```

## 📊 API Reference

### Weather Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/weather?location={loc}` | GET | Current weather |
| `/api/forecast?location={loc}&days={n}` | GET | Forecast (1-7 days) |
| `/api/moon` | GET | Moon phase |
| `/api/hurricanes` | GET | Active hurricanes |
| `/api/earthquakes?mag={m}&days={d}` | GET | Recent earthquakes |

### Response Example

```json
{
  "location": {
    "name": "New York, NY",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "timezone": "America/New_York"
  },
  "current": {
    "temperature": 72,
    "condition": "Partly Cloudy",
    "icon": "⛅",
    "wind_speed": 8,
    "wind_direction": "NE",
    "humidity": 65,
    "pressure": 1013,
    "visibility": 10,
    "uv_index": 5
  },
  "forecast": [
    {
      "date": "2024-10-02",
      "high": 75,
      "low": 62,
      "condition": "Sunny",
      "icon": "☀️",
      "precipitation": 0,
      "hourly": [...]
    }
  ]
}
```

### Authentication

```bash
# Generate API token in admin panel or:
curl -X POST http://localhost:3000/api/tokens \
  -H "Content-Type: application/json" \
  -d '{"name":"My API Token","scopes":"read,write"}'

# Use token
curl -H "Authorization: Bearer your-token" \
  http://localhost:3000/api/v1/forecast?location=NYC
```

### Rate Limits

- **Anonymous**: 120 requests/hour
- **Authenticated**: 1200 requests/hour
- **Custom**: Per-token limits via admin panel

## 🏗️ System Architecture

```
┌─────────────────────────────────────────┐
│           Client Applications            │
│  (Browser, Mobile PWA, Terminal, API)   │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│         Gin HTTP Server (Go)            │
│  ┌─────────────────────────────────┐   │
│  │ Middleware                       │   │
│  │ - Logging, CORS, Security       │   │
│  │ - Authentication, Rate Limiting  │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │ Handlers & Services              │   │
│  │ - Weather, Hurricane, Earthquake │   │
│  │ - Auth, Admin, Dashboard         │   │
│  └─────────────────────────────────┘   │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│            Data Layer                    │
│  ┌──────────┐  ┌───────────────────┐   │
│  │ SQLite   │  │ External APIs      │   │
│  │ Database │  │ - Open-Meteo       │   │
│  │          │  │ - NOAA, USGS       │   │
│  └──────────┘  └───────────────────┘   │
└──────────────────────────────────────────┘
```

## 📁 Project Structure

```
weather/
├── main.go              # Application entry
├── handlers/            # HTTP handlers
├── services/            # Business logic
├── src/
│   ├── database/       # Database layer
│   ├── handlers/       # Auth handlers
│   ├── middleware/     # Custom middleware
│   └── scheduler/      # Background tasks
├── templates/          # HTML templates
├── static/             # CSS, JS, images
├── scripts/            # Install scripts
└── data/               # SQLite database
```

## 🔧 Development

### Prerequisites

- Go 1.23+
- Git
- Make (optional)

### Build

```bash
# Clone repository
git clone https://github.com/apimgr/weather.git
cd weather

# Install dependencies
go mod download

# Build
make build

# Run
./weather
```

### Development Mode

```bash
# With auto-reload
GIN_MODE=debug go run main.go

# Or use make
make dev
```

### Testing

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...
```

## 🌟 Features in Detail

### Saved Locations

1. **Save Your Favorites**
   - Click "Save Location" on any weather page
   - Add custom nicknames
   - Enable/disable alerts per location

2. **Weather Alerts**
   - Automatic checks every 15 minutes
   - Notifications for:
     - Severe weather (storms, hurricanes)
     - Temperature extremes
     - High winds, heavy precipitation

3. **Dashboard Access**
   - View all locations at once
   - Quick weather overview
   - One-click detailed forecasts

### Admin Panel

Access at `/admin` (admin role required)

1. **User Management**
   - Create/edit/delete users
   - Assign roles (admin/user)
   - View login history

2. **API Tokens**
   - Generate secure tokens
   - Set custom rate limits
   - Monitor usage statistics

3. **System Settings**
   - Configure alert thresholds
   - Adjust rate limits
   - Database management

4. **Activity Logs**
   - Track all user actions
   - API usage metrics
   - Security events

### PWA Installation

**Desktop (Chrome/Edge)**
1. Visit http://localhost:3000
2. Click install icon in address bar
3. Follow prompts

**Mobile**
1. Open in Safari/Chrome
2. Tap "Add to Home Screen"
3. Confirm installation

**Features**
- Works offline
- App shortcuts
- Push notifications (coming soon)

## 🌐 Deployment

### Linux (systemd)

```bash
# Install
sudo ./scripts/linux.sh

# Service commands
sudo systemctl start weather
sudo systemctl enable weather
sudo systemctl status weather
sudo systemctl stop weather

# Logs
sudo journalctl -u weather -f
```

### macOS (LaunchAgent)

```bash
# Install
./scripts/macos.sh

# Service commands
launchctl load ~/Library/LaunchAgents/com.apimgr.weather.plist
launchctl unload ~/Library/LaunchAgents/com.apimgr.weather.plist

# Logs
tail -f ~/Library/Application\ Support/Weather/stdout.log
```

### Windows (Service)

```powershell
# Install NSSM first
choco install nssm

# Install service
.\scripts\windows.ps1 -InstallService

# Service commands
nssm start Weather
nssm stop Weather
nssm restart Weather
```

### Kubernetes

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
        image: weather:latest
        ports:
        - containerPort: 3000
        env:
        - name: SESSION_SECRET
          valueFrom:
            secretKeyRef:
              name: weather-secrets
              key: session-secret
        # Note: --data and --config flags are set in Dockerfile CMD
        volumeMounts:
        - name: data
          mountPath: /data
        - name: config
          mountPath: /config
        livenessProbe:
          httpGet:
            path: /livez
            port: 3000
          initialDelaySeconds: 60
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 3000
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
    targetPort: 3000
  type: LoadBalancer
```

## 🔒 Security

### Production Checklist

- [ ] **Set SESSION_SECRET**
  ```bash
  export SESSION_SECRET=$(openssl rand -base64 32)
  ```

- [ ] **Use HTTPS** (with Nginx/Caddy reverse proxy)
  ```nginx
  server {
      listen 443 ssl http2;
      server_name weather.example.com;

      ssl_certificate /path/to/cert.pem;
      ssl_certificate_key /path/to/key.pem;

      location / {
          proxy_pass http://localhost:3000;
          proxy_set_header Host $host;
          proxy_set_header X-Real-IP $remote_addr;
      }
  }
  ```

- [ ] **Restrict Database Permissions**
  ```bash
  chmod 600 /data/weather.db
  chown weather:weather /data/weather.db
  ```

- [ ] **Enable Firewall**
  ```bash
  # UFW (Ubuntu/Debian)
  sudo ufw allow 3000/tcp
  sudo ufw enable

  # firewalld (RHEL/CentOS)
  sudo firewall-cmd --add-port=3000/tcp --permanent
  sudo firewall-cmd --reload
  ```

- [ ] **Regular Backups**
  ```bash
  # Automated backup script
  #!/bin/bash
  DATE=$(date +%Y%m%d-%H%M%S)
  sqlite3 /data/weather.db ".backup /backups/weather-$DATE.db"
  find /backups -name "weather-*.db" -mtime +7 -delete
  ```

## 📈 Performance

### Benchmarks

- **Startup Time**: <1 second
- **Memory Usage**: ~50MB
- **Response Time**: <200ms (cached)
- **Request Handling**: 1000+ req/sec
- **Database Queries**: <10ms average

### Optimization Tips

1. **Enable Caching**
   - Weather data cached for 10 minutes
   - Geocoding cached for 1 hour
   - Reduces external API calls

2. **Use Connection Pooling**
   - Database connections reused
   - Configurable pool size

3. **Rate Limiting**
   - Protects against abuse
   - Configurable per endpoint

## 🐛 Troubleshooting

### Database Locked Error

```bash
# Check for stale connections
./weather --status

# Restart service
systemctl restart weather
```

### Port Already in Use

```bash
# Find process using port
lsof -i :3000

# Kill process
kill -9 <PID>

# Or use different port
./weather --port 8080
```

### High Memory Usage

```bash
# Check database size
du -h /data/weather.db

# Clean old cache
sqlite3 /data/weather.db "DELETE FROM weather_cache WHERE expires_at < datetime('now');"
```

### API Rate Limit Exceeded

```bash
# Check current limits
./weather --status

# Adjust in .env or admin panel
RATE_LIMIT_AUTH=5000
```

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📝 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Weather Data**: [Open-Meteo](https://open-meteo.com) - Free weather API
- **Hurricane Data**: [NOAA National Hurricane Center](https://www.nhc.noaa.gov/)
- **Earthquake Data**: [USGS Earthquake Catalog](https://earthquake.usgs.gov/)
- **Geocoding**: [Nominatim (OpenStreetMap)](https://nominatim.org/)
- **City Database**: [apimgr/citylist](https://github.com/apimgr/citylist) - 209K+ cities
- **Theme**: [Dracula Theme](https://draculatheme.com/)

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/apimgr/weather/issues)
- **Discussions**: [GitHub Discussions](https://github.com/apimgr/weather/discussions)
- **Documentation**: [CLAUDE.md](CLAUDE.md)

## 🗺️ Roadmap

- [ ] Multi-language support (i18n)
- [ ] Push notifications for alerts
- [ ] Historical weather data
- [ ] Weather maps integration
- [ ] Mobile apps (iOS/Android)
- [ ] GraphQL API
- [ ] Webhook support
- [ ] Custom themes
- [ ] Plugin system

## 📊 Stats

- **Lines of Code**: ~15,000
- **Supported Platforms**: 8 (Linux, macOS, Windows, FreeBSD × amd64/arm64)
- **Cities Supported**: 209,579
- **Countries**: 247
- **Database Tables**: 11
- **API Endpoints**: 20+
- **Built with**: Go 1.23, Gin, SQLite

---

<p align="center">
  <strong>Built with ❤️</strong><br>
  <sub>Weather Service • Version 3.0.0 • 2024</sub>
</p>

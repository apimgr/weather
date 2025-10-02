# 🌤️ Weather Service - Complete Platform Documentation

## 📚 Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Features](#features)
4. [Technology Stack](#technology-stack)
5. [Directory Structure](#directory-structure)
6. [Database Schema](#database-schema)
7. [Authentication System](#authentication-system)
8. [API Documentation](#api-documentation)
9. [CLI Commands](#cli-commands)
10. [Configuration](#configuration)
11. [Deployment](#deployment)
12. [Development](#development)

---

## Overview

Weather Service is a comprehensive weather platform built in Go that provides:
- **Beautiful weather forecasts** with ASCII art for terminals
- **Web dashboard** with Dracula theme
- **User authentication** with admin and user roles
- **SQLite database** for data persistence
- **Saved locations** with weather alerts
- **API token system** with rate limiting
- **Hurricane tracking** using NOAA data
- **Earthquake monitoring** using USGS data
- **Moon phase calculations**
- **PWA support** for offline functionality
- **Built-in scheduler** for periodic tasks
- **Notification system** with in-app alerts

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Client Layer                          │
│  ┌────────────┐  ┌────────────┐  ┌────────────────┐    │
│  │ Web Browser│  │   CLI/curl │  │  Mobile (PWA)  │    │
│  └─────┬──────┘  └─────┬──────┘  └────────┬───────┘    │
└────────┼───────────────┼──────────────────┼─────────────┘
         │               │                  │
         └───────────────┼──────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────┐
│                  Gin HTTP Server                          │
│  ┌─────────────────────────────────────────────────┐    │
│  │ Middleware Stack                                  │    │
│  │  - Apache Combined Logging                        │    │
│  │  - CORS                                           │    │
│  │  - Security Headers (CSP, X-Frame-Options)        │    │
│  │  - Session Authentication                         │    │
│  │  - Rate Limiting (120/hr anon, 1200/hr auth)      │    │
│  └─────────────────────────────────────────────────┘    │
│                                                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Handlers   │  │   Services   │  │  Renderers   │  │
│  │              │  │              │  │              │  │
│  │ • Weather    │  │ • Weather    │  │ • ASCII      │  │
│  │ • Auth       │  │ • Hurricane  │  │ • HTML       │  │
│  │ • Admin      │  │ • Earthquake │  │ • JSON       │  │
│  │ • API        │  │ • Location   │  │ • OneLineFmt │  │
│  │ • Health     │  │ • Geocoding  │  │              │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└───────────────────────────────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────┐
│                 Data Layer                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   SQLite DB  │  │ External APIs│  │    Caches    │  │
│  │              │  │              │  │              │  │
│  │ • users      │  │ • Open-Meteo │  │ • Weather    │  │
│  │ • locations  │  │ • NOAA NHC   │  │ • Geocoding  │  │
│  │ • api_tokens │  │ • USGS       │  │ • Hurricane  │  │
│  │ • sessions   │  │ • Nominatim  │  │              │  │
│  │ • alerts     │  │ • City DB    │  │              │  │
│  │ • logs       │  │              │  │              │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└───────────────────────────────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────┐
│               Background Tasks (Scheduler)                │
│  • Session cleanup (every 6 hours)                        │
│  • Weather alerts check (every 15 minutes)                │
│  • Database backup (daily at 3 AM)                        │
│  • Token expiration cleanup (hourly)                      │
└───────────────────────────────────────────────────────────┘
```

## Features

### Core Weather Features
- **Current Weather**: Temperature, conditions, wind, humidity
- **3-Day Forecast**: Hourly breakdowns with highs/lows
- **Moon Phases**: Current phase, illumination percentage
- **Hurricanes**: Active cyclone tracking from NOAA
- **Earthquakes**: Recent seismic activity from USGS
- **wttr.in Compatibility**: Formats 0-4 supported

### Authentication & Users
- **Email/Password Auth**: bcrypt hashed passwords
- **Session Management**: Secure cookie-based sessions
- **Role System**: Admin and User roles
- **First-Run Wizard**: Automatic admin account creation
- **Password Reset**: Secure reset flow
- **User Profiles**: Customizable display names

### Admin Dashboard
- **User Management**: Create, edit, delete users
- **API Token Management**: Generate, revoke, monitor tokens
- **System Settings**: Configure alerts, rate limits
- **Activity Logs**: Track user actions and API usage
- **Server Status**: Real-time metrics and health
- **Database Stats**: User count, locations, tokens

### Saved Locations
- **Personal Library**: Save favorite locations
- **Weather Alerts**: Automatic notifications for severe weather
- **Quick Access**: Dashboard with all saved locations
- **Customizable**: Add nicknames to locations

### API System
- **RESTful JSON API**: `/api/weather`, `/api/forecast`
- **Token Authentication**: Show-once secure tokens
- **Rate Limiting**: Per-token and per-IP limits
- **API Documentation**: Interactive `/api/docs`

### PWA Support
- **Offline Mode**: Service worker caching
- **Install Prompt**: Add to home screen
- **App Shortcuts**: Quick access to features
- **Push Notifications**: Weather alerts (future)

## Technology Stack

### Backend
- **Language**: Go 1.23+
- **Framework**: Gin (HTTP router)
- **Database**: SQLite (modernc.org/sqlite - pure Go)
- **Sessions**: gorilla/sessions
- **Password**: bcrypt

### Frontend
- **CSS Framework**: Custom Dracula theme
- **JavaScript**: Vanilla JS (no framework)
- **PWA**: Service Worker + Manifest
- **Icons**: Emoji-based weather icons

### External APIs
- **Weather**: Open-Meteo (free, no API key)
- **Geocoding**: Open-Meteo + Nominatim (fallback)
- **Hurricanes**: NOAA National Hurricane Center
- **Earthquakes**: USGS Earthquake Catalog
- **Location Data**: 209K+ cities from apimgr/citylist

### Infrastructure
- **Build**: Go build with CGO_ENABLED=0
- **Container**: Docker multi-stage (scratch base)
- **Orchestration**: Docker Compose
- **CI/CD**: GitHub Actions (8 platforms)

## Directory Structure

```
weather/
├── main.go                    # Application entry point
├── go.mod                     # Go dependencies
├── go.sum                     # Dependency checksums
├── Makefile                   # Build automation
├── Dockerfile                 # Container image
├── docker-compose.yml         # Docker orchestration
├── .env.example              # Environment template
├── CLAUDE.md                  # This file
├── README.md                  # User documentation
│
├── handlers/                  # HTTP request handlers
│   ├── weather.go            # Weather endpoints
│   ├── api.go                # JSON API
│   ├── health.go             # Health checks
│   └── web.go                # Web interface
│
├── services/                  # Business logic
│   ├── weather.go            # Weather service
│   ├── hurricane.go          # Hurricane tracking
│   ├── earthquake.go         # Earthquake monitoring
│   └── location_enhancer.go  # Geocoding enhancement
│
├── renderers/                 # Output formatters
│   ├── ascii.go              # Terminal ASCII art
│   ├── oneline.go            # wttr.in formats
│   └── html.go               # Web templates
│
├── utils/                     # Utilities
│   ├── types.go              # Data structures
│   ├── params.go             # Request parsing
│   └── helpers.go            # Helper functions
│
├── src/                       # Application modules
│   ├── database/             # Database layer
│   │   ├── schema.go         # Schema + migrations
│   │   └── queries.go        # Database queries
│   │
│   ├── handlers/             # Auth handlers
│   │   ├── auth.go           # Login/logout
│   │   ├── register.go       # Registration
│   │   ├── admin.go          # Admin panel
│   │   └── dashboard.go      # User dashboard
│   │
│   ├── middleware/           # Custom middleware
│   │   ├── auth.go           # Auth checking
│   │   └── ratelimit.go      # Rate limiting
│   │
│   └── scheduler/            # Background tasks
│       └── scheduler.go      # Task scheduler
│
├── templates/                 # HTML templates
│   ├── weather.html          # Main weather page
│   ├── loading.html          # Initialization page
│   ├── login.html            # Login form
│   ├── register.html         # Registration
│   ├── dashboard.html        # User dashboard
│   ├── admin.html            # Admin panel
│   └── api-docs.html         # API documentation
│
├── static/                    # Static assets
│   ├── css/
│   │   └── dracula.css       # Dracula theme
│   ├── js/
│   │   └── sw.js             # Service worker
│   ├── manifest.json         # PWA manifest
│   └── images/               # Icons
│
├── scripts/                   # Installation scripts
│   ├── install.sh            # Universal installer
│   ├── linux.sh              # Linux + systemd
│   ├── macos.sh              # macOS + LaunchAgent
│   └── windows.ps1           # Windows + NSSM
│
└── data/                      # Runtime data
    └── weather.db            # SQLite database
```

## Database Schema

### Tables (11 total)

#### 1. `users`
Stores user accounts with authentication

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    display_name TEXT,
    role TEXT DEFAULT 'user',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME,
    is_active BOOLEAN DEFAULT 1
);
```

#### 2. `sessions`
Manages user sessions

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

#### 3. `locations`
User's saved locations

```sql
CREATE TABLE locations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    nickname TEXT,
    alerts_enabled BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

#### 4. `api_tokens`
API authentication tokens

```sql
CREATE TABLE api_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    scopes TEXT,
    rate_limit INTEGER DEFAULT 1200,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    last_used DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

#### 5. `alerts`
Weather alert notifications

```sql
CREATE TABLE alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    location_id INTEGER NOT NULL,
    alert_type TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    acknowledged BOOLEAN DEFAULT 0,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE CASCADE
);
```

#### 6. `notifications`
In-app notifications

```sql
CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    type TEXT DEFAULT 'info',
    read BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

#### 7. `settings`
System configuration

```sql
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### 8. `activity_logs`
User action tracking

```sql
CREATE TABLE activity_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    action TEXT NOT NULL,
    details TEXT,
    ip_address TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);
```

#### 9. `api_usage`
API request metrics

```sql
CREATE TABLE api_usage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token_id INTEGER,
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    status_code INTEGER,
    response_time INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (token_id) REFERENCES api_tokens(id) ON DELETE SET NULL
);
```

#### 10. `weather_cache`
Weather data caching

```sql
CREATE TABLE weather_cache (
    location_key TEXT PRIMARY KEY,
    data TEXT NOT NULL,
    cached_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL
);
```

#### 11. `scheduled_tasks`
Scheduler task tracking

```sql
CREATE TABLE scheduled_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_name TEXT NOT NULL,
    last_run DATETIME,
    next_run DATETIME,
    status TEXT DEFAULT 'pending',
    error_message TEXT
);
```

## Authentication System

### Password Security
- **Hashing**: bcrypt with cost factor 10
- **Validation**: Min 8 chars, complexity checks
- **Storage**: Never store plaintext passwords

### Session Management
- **Storage**: SQLite + gorilla/sessions
- **Duration**: 7 days default
- **Security**: HTTP-only cookies, SameSite=Lax
- **Cleanup**: Automatic expiration via scheduler

### Role System
- **Admin**: Full access (user mgmt, settings, logs)
- **User**: Limited access (own data, saved locations)

### First-Run Setup
1. Check for existing users on startup
2. If none found, show registration wizard
3. First user automatically gets admin role
4. Subsequent users default to 'user' role

## API Documentation

### Weather Endpoints

#### Get Current Weather
```
GET /api/weather?location={location}
```

**Response**:
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
    "humidity": 65,
    "timestamp": "2024-10-02T14:30:00Z"
  }
}
```

#### Get Forecast
```
GET /api/forecast?location={location}&days=3
```

#### Get Moon Phase
```
GET /api/moon
```

#### Get Hurricanes
```
GET /api/hurricanes
```

#### Get Earthquakes
```
GET /api/earthquakes?mag=2.5&days=7
```

### Authentication Required

All `/api/v1/*` endpoints require token:
```
Authorization: Bearer {token}
```

### Rate Limits
- Anonymous: 120 requests/hour
- Authenticated: 1200 requests/hour
- Token-based: Custom per token

## CLI Commands

### Basic Commands

```bash
# Show version
./weather --version

# Show server status
./weather --status

# Run health check (Docker)
./weather --healthcheck
```

### Configuration Overrides

```bash
# Override port
./weather --port 8080

# Override database path
./weather --data /path/to/db

# Override listen address
./weather --address 127.0.0.1
```

### Examples

```bash
# Start on port 8080 with custom database
./weather --port 8080 --data /var/lib/weather/data.db

# Check status with custom config
./weather --data /custom/path.db --status
```

## Configuration

### Environment Variables

```bash
# Server
PORT=3000
GIN_MODE=release
SERVER_ADDRESS=0.0.0.0

# Database
DATABASE_PATH=/data/weather.db

# Security
SESSION_SECRET=your-secure-random-string

# Logging
LOG_LEVEL=info

# Features
RATE_LIMIT_ANON=120
RATE_LIMIT_AUTH=1200
```

### Configuration Files

Create `.env` from `.env.example`:
```bash
cp .env.example .env
# Edit .env with your values
```

### Security Best Practices

1. **Always set SESSION_SECRET** in production
   ```bash
   SESSION_SECRET=$(openssl rand -base64 32)
   ```

2. **Use HTTPS** with reverse proxy (Nginx/Caddy)

3. **Restrict database permissions**
   ```bash
   chmod 600 /data/weather.db
   ```

4. **Enable firewall** and limit exposed ports

## Deployment

### Docker

```bash
# Build image
docker build -t weather:latest .

# Run container
docker run -d \
  -p 3000:3000 \
  -v weather-data:/data \
  -e SESSION_SECRET=$(openssl rand -base64 32) \
  weather:latest
```

### Docker Compose

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Systemd (Linux)

```bash
# Install
sudo ./scripts/linux.sh

# Manage service
sudo systemctl start weather
sudo systemctl enable weather
sudo systemctl status weather
```

### LaunchAgent (macOS)

```bash
# Install
./scripts/macos.sh

# Manage service
launchctl load ~/Library/LaunchAgents/com.apimgr.weather.plist
launchctl unload ~/Library/LaunchAgents/com.apimgr.weather.plist
```

### Windows Service

```powershell
# Install (requires NSSM)
.\scripts\windows.ps1 -InstallService

# Manage service
nssm start Weather
nssm stop Weather
```

## Development

### Prerequisites

- Go 1.23+
- Make (optional)
- Docker (optional)

### Build

```bash
# Development build
make dev

# Production build
make build

# All platforms
make build-all
```

### Run Locally

```bash
# Development mode (auto-reload)
GIN_MODE=debug go run main.go

# Production mode
make run
```

### Testing

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./services/...
```

### Database Migrations

Migrations run automatically on startup. Manual execution:

```bash
# Run migrations
./weather --data ./test.db
```

### Adding Features

1. **New Endpoint**:
   - Add handler in `handlers/`
   - Register route in `main.go`
   - Update API docs

2. **New Database Table**:
   - Add schema to `src/database/schema.go`
   - Increment schema version
   - Add migration logic

3. **New External API**:
   - Create service in `services/`
   - Add caching layer
   - Handle errors gracefully

### Code Style

```bash
# Format code
make fmt

# Lint code
make lint

# Run pre-commit checks
make test
```

## Production Checklist

- [ ] Set SESSION_SECRET environment variable
- [ ] Use HTTPS with valid SSL certificate
- [ ] Configure firewall rules
- [ ] Set up regular database backups
- [ ] Enable systemd/launchd service
- [ ] Configure log rotation
- [ ] Set up monitoring (health checks)
- [ ] Review rate limits
- [ ] Restrict admin access
- [ ] Test disaster recovery

## Troubleshooting

### Database Locked

```bash
# Check for stale connections
./weather --status

# Restart service
systemctl restart weather
```

### High Memory Usage

```bash
# Check cache size
sqlite3 data/weather.db "SELECT COUNT(*) FROM weather_cache;"

# Clear old cache
sqlite3 data/weather.db "DELETE FROM weather_cache WHERE expires_at < datetime('now');"
```

### Rate Limit Issues

```bash
# Check current limits
./weather --status

# Adjust in .env
RATE_LIMIT_AUTH=5000
```

## API Client Examples

### cURL

```bash
# Get weather
curl "http://localhost:3000/api/weather?location=NYC"

# With API token
curl -H "Authorization: Bearer your-token" \
  "http://localhost:3000/api/v1/forecast?location=NYC&days=7"
```

### Python

```python
import requests

# Anonymous
r = requests.get('http://localhost:3000/api/weather?location=NYC')
print(r.json())

# Authenticated
headers = {'Authorization': 'Bearer your-token'}
r = requests.get('http://localhost:3000/api/v1/forecast?location=NYC', headers=headers)
print(r.json())
```

### JavaScript

```javascript
// Fetch weather
fetch('http://localhost:3000/api/weather?location=NYC')
  .then(r => r.json())
  .then(data => console.log(data));

// With token
fetch('http://localhost:3000/api/v1/forecast?location=NYC', {
  headers: { 'Authorization': 'Bearer your-token' }
})
  .then(r => r.json())
  .then(data => console.log(data));
```

## Performance Optimization

### Caching Strategy

- **Weather Data**: 10 minutes
- **Geocoding**: 1 hour
- **Hurricane Data**: 15 minutes
- **Static Assets**: Browser cache 1 day

### Database Optimization

```sql
-- Indexes for common queries
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_locations_user_id ON locations(user_id);
CREATE INDEX idx_tokens_token ON api_tokens(token);
CREATE INDEX idx_notifications_user_read ON notifications(user_id, read);
```

### Resource Limits

```yaml
# docker-compose.yml
deploy:
  resources:
    limits:
      cpus: '1'
      memory: 256M
    reservations:
      cpus: '0.5'
      memory: 128M
```

## Contributing

1. Fork the repository
2. Create feature branch
3. Make changes
4. Add tests
5. Update documentation
6. Submit pull request

## License

MIT License - see LICENSE file for details

## Support

- **Issues**: https://github.com/apimgr/weather/issues
- **Discussions**: https://github.com/apimgr/weather/discussions
- **Email**: support@example.com

---

**Built with ❤️**

*Last Updated: 2024-10-02*
*Version: 3.0.0*

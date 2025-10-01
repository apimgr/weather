# 🤖 Console Weather Service - Complete Implementation Guide

This document provides **complete instructions** for recreating this project from scratch using Claude Code. Follow these steps in order to build an identical weather service.

## 🎯 Project Goal

Build a production-ready weather service in Go that:
- Provides beautiful ASCII art weather displays for terminals
- Offers a responsive web interface with Dracula theming
- Supports wttr.in format compatibility (format=0,1,2,3,4)
- Features intelligent location detection with 209K+ cities
- Includes reverse geocoding for GPS coordinates with US state abbreviations
- Deploys as a single static binary across 8 platforms
- Initializes in ~3 seconds with automatic readiness detection

## 📋 Prerequisites

Before starting, ensure you have:
- Go 1.23 or later installed
- Git for version control
- Basic understanding of HTTP servers
- Optional: Docker for containerization

## 🏗️ Step-by-Step Implementation

### Step 1: Initialize Go Project

```bash
# Create project directory
mkdir weather && cd weather

# Initialize Go module
go mod init weather-go

# Create directory structure
mkdir -p handlers services renderers utils templates static/css static/js
```

### Step 2: Install Dependencies

```bash
# Add core dependencies
go get github.com/gin-gonic/gin@latest
go get github.com/gin-contrib/cors@latest
go get github.com/gin-contrib/secure@latest
```

Your `go.mod` should look like:
```go
module weather-go

go 1.23

require (
    github.com/gin-contrib/cors v1.7.2
    github.com/gin-contrib/secure v1.0.3
    github.com/gin-gonic/gin v1.10.0
)
```

### Step 3: Create Main Application Entry Point

Create `main.go` with:
- Gin HTTP server setup
- Middleware (logging, CORS, security headers)
- Route definitions
- Health check endpoints
- Initialization logic with 2-minute timeout fallback

Key features to implement:
- Apache2 combined log format
- Path normalization for double slashes
- Service initialization callbacks
- Template loading from `templates/`
- Static file serving from `static/`

### Step 4: Implement Type Definitions

Create `utils/types.go` with these core types:

```go
// Coordinates represents a geographic location
type Coordinates struct {
    Latitude    float64
    Longitude   float64
    Name        string
    Country     string
    CountryCode string
    Timezone    string
    Admin1      string  // State/Province
    Admin2      string  // County/Region
    Population  int
    FullName    string  // "Summerville, South Carolina, United States"
    ShortName   string  // "Summerville, SC"
}

// WeatherParams contains request parameters
type WeatherParams struct {
    Location     string
    Format       int     // 0-4 for wttr.in compatibility
    Units        string  // "metric" or "imperial"
    Lang         string
    NoFooter     bool
    Quiet        bool
    OneLineOnly  bool
}

// InitializationStatus tracks service readiness
type InitializationStatus struct {
    Countries bool
    Cities    bool
    Weather   bool
    Started   time.Time
}
```

### Step 5: Build Weather Service

Create `services/weather.go` with:

**Core functionality:**
- Open-Meteo API integration for weather data
- Geocoding (city name → coordinates)
- Reverse geocoding (coordinates → city name with Nominatim fallback)
- Weather code to emoji/description mapping
- Wind direction calculations
- Moon phase calculations

**API endpoints used:**
- `https://geocoding-api.open-meteo.com/v1/search` - Forward geocoding
- `https://api.open-meteo.com/v1/forecast` - Weather data
- `https://nominatim.openstreetmap.org/reverse` - Reverse geocoding (fallback)

**Caching:**
- 10-minute cache for weather data
- Thread-safe map with RWMutex

### Step 6: Implement Location Enhancer

Create `services/location_enhancer.go` with:

**Features:**
- Parallel loading of countries and cities databases
- Coordinate-based matching using Haversine distance
- US state abbreviation lookup (all 50 states)
- Canadian province abbreviation lookup
- Population and timezone enrichment
- Timing logs for initialization

**External databases:**
```go
countriesURL := "https://github.com/apimgr/countries/raw/refs/heads/main/countries.json"
citiesURL := "https://github.com/apimgr/citylist/raw/refs/heads/master/api/city.list.json"
```

**Key algorithms:**
- Haversine distance for coordinate matching
- Case-insensitive state name normalization
- Building full vs short location names

### Step 7: Create ASCII Renderers

Create `renderers/ascii.go` for full weather display:
- Weather header with location name
- ASCII art weather icons (sun, clouds, rain, etc.)
- Current conditions table
- 3-day forecast with hourly details
- Footer with attribution

Create `renderers/oneline.go` for wttr.in formats:
- Format 1: `🌙  +14°C`
- Format 2: `🌙  🌡️+14°C 🌬️↓4km/h`
- Format 3: `London, GB: 🌙  +14°C`
- Format 4: `London, GB: 🌙  🌡️+14°C 🌬️↓4km/h`

**ANSI color codes** (Dracula theme):
```go
cyan    := "\x1b[38;2;139;233;253m"  // #8BE9FD
green   := "\x1b[38;2;80;250;123m"   // #50FA7B
yellow  := "\x1b[38;2;241;250;140m"  // #F1FA8C
pink    := "\x1b[38;2;255;121;198m"  // #FF79C6
purple  := "\x1b[38;2;189;147;249m"  // #BD93F9
reset   := "\x1b[0m"
```

### Step 8: Build HTTP Handlers

Create `handlers/weather.go`:
- Root handler (/) with IP-based location detection
- Location handler (/:location) with format detection
- Browser vs console detection via User-Agent
- Parameter parsing (format, units, options)

Create `handlers/api.go`:
- JSON weather endpoints
- Forecast endpoints
- Location search
- IP detection endpoint

Create `handlers/health.go`:
- `/healthz` - Main health check with initialization status
- `/readyz` - Kubernetes readiness probe
- `/livez` - Kubernetes liveness probe
- Initialization state management with callbacks

Create `handlers/web.go`:
- HTML interface renderer
- Loading page with auto-refresh

### Step 9: Create HTML Templates

Create `templates/weather.html`:
- Dracula-themed responsive design
- Location search with autocomplete
- Unit toggle (°F/°C)
- Geolocation button
- Weather display with forecast
- Mobile-first responsive (320px+)

Create `templates/loading.html`:
- Coffee cup spinner animation
- Database loading status indicators
- 30-second initial delay, then 15-second auto-checks
- Auto-redirect when service is ready
- Console-friendly ASCII fallback

Create `templates/api-docs.html`:
- Interactive API documentation
- Usage examples
- Endpoint descriptions

### Step 10: Add Static Assets

Create `static/css/dracula.css`:
- Dracula color palette
- CSS variables for theming
- Responsive utilities
- Form styling
- Animation keyframes

Color scheme:
```css
--dracula-bg: #282a36
--dracula-current-line: #44475a
--dracula-foreground: #f8f8f2
--dracula-comment: #6272a4
--dracula-cyan: #8be9fd
--dracula-green: #50fa7b
--dracula-orange: #ffb86c
--dracula-pink: #ff79c6
--dracula-purple: #bd93f9
--dracula-red: #ff5555
--dracula-yellow: #f1fa8c
```

### Step 11: Implement Dockerfile

Create multi-stage `Dockerfile`:
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o weather main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/weather .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/healthz || exit 1
CMD ["./weather"]
```

### Step 12: Create GitHub Actions Workflow

Create `.github/workflows/release.yml`:
- Multi-platform build matrix (Linux, macOS, Windows, FreeBSD × AMD64/ARM64)
- Static binary builds with `CGO_ENABLED=0`
- Auto-delete existing releases
- Upload 8 platform binaries
- Comprehensive release notes

Supported platforms:
- weather-linux-amd64
- weather-linux-arm64
- weather-darwin-amd64
- weather-darwin-arm64
- weather-windows-amd64.exe
- weather-windows-arm64.exe
- weather-freebsd-amd64
- weather-freebsd-arm64

### Step 13: Add Build Automation

Create `Makefile`:
```makefile
.PHONY: build build-all run dev clean

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o weather main.go

build-all:
	# Linux
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/weather-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/weather-linux-arm64 main.go
	# macOS
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/weather-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/weather-darwin-arm64 main.go
	# Windows
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/weather-windows-amd64.exe main.go
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/weather-windows-arm64.exe main.go
	# FreeBSD
	GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/weather-freebsd-amd64 main.go
	GOOS=freebsd GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/weather-freebsd-arm64 main.go

run:
	PORT=3000 ./weather

dev:
	GIN_MODE=debug go run main.go

clean:
	rm -f weather
	rm -rf dist/
```

### Step 14: Documentation

Create comprehensive `README.md` with:
- Quick start for all 8 platforms
- Usage examples (curl, web browser, JSON API)
- Format parameters table
- Location format examples
- Configuration options
- Deployment guides (Docker, Kubernetes, systemd)
- Performance comparison
- Architecture diagram
- Contributing guidelines

Create `LICENSE.md` with:
- MIT License text
- Third-party attributions (Gin, Open-Meteo, Nominatim, etc.)
- Data source credits
- Privacy statement
- Disclaimer

### Step 15: Configuration Files

Create `.gitignore`:
```
# Binaries
weather
weather-*
*.exe
dist/

# Go
vendor/
*.out
*.test

# IDE
.vscode/
.idea/
*.swp

# OS
.DS_Store
Thumbs.db

# Development
*.log
.env
```

Create `docker-compose.yml`:
```yaml
version: '3.8'
services:
  weather:
    build: .
    ports:
      - "3000:3000"
    environment:
      - GIN_MODE=release
      - PORT=3000
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:3000/healthz"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s
```

## 🔑 Critical Implementation Details

### Database Initialization
```go
// In location_enhancer.go
func (le *LocationEnhancer) loadData() {
    startTime := time.Now()
    fmt.Printf("⏱️  Starting database initialization at %s\n", startTime.Format("15:04:05.000"))

    // Load in parallel with timing
    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        countryStart := time.Now()
        fmt.Println("📥 Loading countries database...")
        err := le.loadCountriesData()
        elapsed := time.Since(countryStart)
        if err != nil {
            fmt.Printf("❌ Countries database failed after %s: %v\n", elapsed, err)
        } else {
            fmt.Printf("✅ Countries database loaded in %s (%d countries)\n", elapsed, len(le.countriesData))
        }
    }()

    go func() {
        defer wg.Done()
        cityStart := time.Now()
        fmt.Println("📥 Loading cities database...")
        err := le.loadCitiesData()
        elapsed := time.Since(cityStart)
        if err != nil {
            fmt.Printf("❌ Cities database failed after %s: %v\n", elapsed, err)
        } else {
            fmt.Printf("✅ Cities database loaded in %s (%d cities)\n", elapsed, len(le.citiesData))
        }
    }()

    wg.Wait()
    totalElapsed := time.Since(startTime)
    fmt.Printf("🎉 Database initialization complete in %s\n", totalElapsed)

    // Notify main that initialization is complete
    if le.onInitComplete != nil {
        le.onInitComplete(countriesLoaded, citiesLoaded)
    }
}
```

### State Abbreviation Handling
```go
// Must handle both full names and abbreviations
func (le *LocationEnhancer) getStateAbbreviation(stateName string) string {
    normalizedState := strings.ToLower(strings.TrimSpace(stateName))

    states := map[string]string{
        "alabama": "AL", "alaska": "AK", "arizona": "AZ",
        // ... all 50 states
    }

    if abbrev, ok := states[normalizedState]; ok {
        return abbrev
    }
    return ""
}
```

### Reverse Geocoding with Nominatim Fallback
```go
// In weather.go ReverseGeocode function
// First try Open-Meteo, then fallback to Nominatim
if len(geocodeResult.Name) == 0 || geocodeResult.Admin1 == "" {
    nominatimURL := fmt.Sprintf(
        "https://nominatim.openstreetmap.org/reverse?lat=%.4f&lon=%.4f&format=json",
        latitude, longitude,
    )
    req, _ := http.NewRequest("GET", nominatimURL, nil)
    req.Header.Set("User-Agent", "WeatherService/2.0")
    // Parse and use Nominatim data
}
```

## 🎨 Design Principles

1. **Initialization Performance**: Load databases in parallel, show timing logs
2. **Graceful Degradation**: Service works even if databases fail to load
3. **Auto-Ready Detection**: Loading page checks every 15s and auto-redirects
4. **State Abbreviations**: GPS coordinates show "SC" not "US" for Summerville
5. **Multi-Platform**: Static binaries for 8 platforms (no dependencies)
6. **Health Checks**: Proper HTTP status codes (503 when initializing, 200 when ready)
7. **Logging**: Detailed initialization timing and Apache2 combined format

## ✅ Testing Checklist

After implementation, test these scenarios:

```bash
# Build and run
CGO_ENABLED=0 go build -ldflags="-s -w" -o weather main.go
PORT=3000 ./weather

# Wait for initialization (watch logs)
# Should show:
# ⏱️  Starting database initialization at HH:MM:SS.mmm
# ✅ Countries database loaded in ~250ms (247 countries)
# ✅ Cities database loaded in ~2.6s (209579 cities)
# 🎉 Database initialization complete in ~2.6s

# Test GPS coordinates with state
curl "http://localhost:3000/33.0285,-80.1551?format=3"
# Should return: Summerville, SC: 🌙  +69°F

# Test city with state
curl "http://localhost:3000/Albany,NY?format=3"
# Should return: Albany, NY: 🌙  +53°F

# Test health check
curl http://localhost:3000/healthz | jq '.'
# Should show: "ready": true

# Test loading page
curl http://localhost:3000/
# Should redirect to weather page or show loading

# Test formats
curl "http://localhost:3000/London?format=1"  # Icon + temp
curl "http://localhost:3000/London?format=2"  # + wind
curl "http://localhost:3000/London?format=3"  # Location + weather
curl "http://localhost:3000/London?format=4"  # Full details
curl "http://localhost:3000/London"           # ASCII art (format=0)
```

## 📊 Expected Performance

After implementation, you should see:
- **Startup**: <1 second
- **DB Init**: ~3 seconds (247 countries + 209,579 cities)
- **Memory**: ~50MB
- **Binary Size**: ~20MB
- **Response Time**: <200ms

## 🚀 Deployment

```bash
# Tag and release
git add .
git commit -m "Complete weather service implementation"
git tag v2.0.0
git push origin main
git push origin v2.0.0

# GitHub Actions will automatically:
# - Build 8 platform binaries
# - Create release with binaries
# - Generate release notes
```

## 📚 Key Files Summary

| File | Purpose | Lines |
|------|---------|-------|
| `main.go` | HTTP server, middleware, routes | ~340 |
| `handlers/weather.go` | Weather request handlers | ~200 |
| `handlers/api.go` | JSON API endpoints | ~150 |
| `handlers/health.go` | Health checks | ~230 |
| `services/weather.go` | Open-Meteo integration | ~500 |
| `services/location_enhancer.go` | Database enhancement | ~440 |
| `renderers/ascii.go` | Full ASCII display | ~400 |
| `renderers/oneline.go` | Format 1-4 rendering | ~150 |
| `utils/types.go` | Type definitions | ~100 |
| `templates/weather.html` | Main web interface | ~300 |
| `templates/loading.html` | Loading page | ~245 |
| `.github/workflows/release.yml` | CI/CD builds | ~200 |

**Total**: ~3,000 lines of production-ready Go code

## 🎯 Final Result

Following this guide creates a production-ready weather service that:
- ✅ Starts in <1 second
- ✅ Initializes databases in ~3 seconds
- ✅ Serves weather for 209K+ cities
- ✅ Shows correct US state abbreviations
- ✅ Provides 8 platform binaries
- ✅ Auto-detects when ready
- ✅ Handles 100+ requests/second
- ✅ Uses 50MB memory
- ✅ Deploys as single binary

**Developed with Claude Code** - This entire system was created through AI-assisted development, demonstrating the power of collaborative programming with Claude.

---

*Last Updated: 2024 | Version: 2.0.0 | Go Edition*

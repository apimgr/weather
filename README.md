# 🌤️ Console Weather Service

A modern, free console-oriented weather service with beautiful Dracula-themed web interface. Complete compatibility with popular weather CLI tools, no API keys required!

**🌐 Official Website**: [wthr.top](http://wthr.top)  
**📦 Repository**: [github.com/apimgr/weather](https://github.com/apimgr/weather)

## ✨ Features

- 🌍 **Global Weather Data** - Powered by Open-Meteo.com (no API key needed)
- 🎨 **Dual Interface** - ASCII art for console, beautiful HTML for browsers
- 🌈 **Dracula Theme** - Dark theme with purple, cyan, and vibrant colors
- 📱 **Mobile-First** - Responsive design with rem-based accessibility
- 🎯 **Smart Location Detection** - IP-based with browser geolocation fallback
- 🌡️ **Unit Auto-Detection** - Imperial/metric based on location
- 🌙 **Moon Phases** - Complete lunar information and cycles with ASCII art
- ⚡ **High Performance** - Built-in caching and optimized responses
- 🔒 **Production Ready** - Docker, nginx, security headers, rate limiting
- 🛡️ **Robust Error Handling** - Never fails, always provides weather data
- 🔗 **URL Encoding Support** - Handles spaces, commas, and special characters
- 🎛️ **Full Parameter Support** - Complete compatibility with CLI weather tools

## 🚀 Quick Start

### 🌐 Try the Live Service (Instant Access)

**No installation needed - use the live service at wthr.top:**
```bash
# Current location weather
curl wthr.top/

# Specific cities
curl wthr.top/london
curl wthr.top/summerville,sc
curl wthr.top/tokyo

# Different formats
curl wthr.top/london?format=3           # → London, GB: ☁️ +11⁰C
curl wthr.top/paris?m&format=4          # → Paris, FR: ☁️ 🌡️+11°C 🌬️↓4km/h

# Moon phases
curl wthr.top/moon@london

# JSON API
curl wthr.top/api/v1/weather?location=tokyo

# Web interface: http://wthr.top/london
```

### 🐳 Production Deployment (Recommended)

**Quick Production Setup with Docker:**
```bash
# Using Docker registry (no build required)
docker run -d \
  --name console-weather \
  -p 3000:3000 \
  --restart unless-stopped \
  ghcr.io/apimgr/weather:latest

# Or build from source
git clone https://github.com/apimgr/weather.git
cd weather
./docker-start.sh prod

# Access your service
curl localhost:3000/london              # Console weather  
open http://localhost:3000/london       # Web interface
```

### 🔧 Development Setup

**For developers who want to modify the code:**

**Prerequisites:**
- **Docker** (recommended) or **Node.js 18+** for local development

**Docker Installation:**
```bash
# Ubuntu/Debian
sudo apt update && sudo apt install docker.io docker-compose

# macOS  
brew install docker docker-compose

# Windows
# Download Docker Desktop from https://docs.docker.com/desktop/
```

**Development Options:**
```bash
# Clone repository
git clone https://github.com/apimgr/weather.git
cd weather

# Option 1: Docker development (recommended)
./docker-start.sh dev

# Option 2: Local Node.js development
npm install
npm run dev

# Access development server
curl localhost:3000/london              # Console ASCII art
open http://localhost:3000/london       # Browser interface
```

## 📖 Usage Examples

### 🌐 Live Service (wthr.top)

**Try the live service instantly:**
```bash
# Current location
curl wthr.top/

# Specific cities  
curl wthr.top/london
curl wthr.top/tokyo
curl wthr.top/summerville,sc

# Different formats
curl wthr.top/london?format=3           # → London, GB: ☁️ +11⁰C
curl wthr.top/paris?m&format=4          # → Paris, FR: ☁️ 🌡️+11°C 🌬️↓4km/h

# Moon phases
curl wthr.top/moon
curl wthr.top/moon@london

# API access
curl wthr.top/api/v1/weather?location=tokyo
```

### Console Interface (ASCII Art)

```bash
# Current location (IP-based detection)
curl localhost:3000/

# Specific cities with URL encoding support
curl localhost:3000/london
curl localhost:3000/summerville,sc
curl localhost:3000/new+york
curl localhost:3000/New%20York,%20NY

# Coordinates
curl localhost:3000/40.7128,-74.0060

# Format options (compatible with popular CLI tools)
curl localhost:3000/london?format=0     # Current weather only
curl localhost:3000/london?format=1     # Current + today  
curl localhost:3000/london?format=2     # Current + 2 days
curl localhost:3000/london?format=3     # Current + 3 days
curl localhost:3000/london?format=4     # Current + today + chart

# Unit control (short format)
curl localhost:3000/tokyo?u             # Imperial units
curl localhost:3000/tokyo?m             # Metric units  
curl localhost:3000/tokyo?M             # SI units
curl localhost:3000/paris?units=imperial # Explicit units

# View options
curl localhost:3000/london?F            # Remove footer
curl localhost:3000/london?n            # Narrow format
curl localhost:3000/london?q            # Quiet mode
curl localhost:3000/london?Q            # Super quiet

# Combined parameters
curl localhost:3000/london?u&format=3&F # Imperial, 3-day, no footer

# Moon phases with enhanced ASCII art
curl localhost:3000/moon
curl localhost:3000/moon@london

# Help and integration
curl localhost:3000/:help
curl localhost:3000/:bash.function
```

### Web Interface (Dracula Theme)

Visit any location in your browser for a beautiful dark-themed interface:
- `http://localhost:3000/london`
- `http://localhost:3000/new+york`
- `http://localhost:3000/` (auto-detects your location)

**Web Features:**
- 📱 **Mobile-optimized** responsive design
- 🎨 **Dracula color scheme** with dark purple backgrounds
- 📍 **GPS location button** with smart fallbacks
- 🔄 **Auto-filled forms** with detected location
- 📊 **Interactive forecast** tables and cards

### JSON API

```bash
# Current weather
curl localhost:3000/api/v1/weather?location=london

# Multi-day forecast  
curl localhost:3000/api/v1/forecast?location=tokyo&days=7

# Location search
curl localhost:3000/api/v1/search?q=new+york

# API documentation
curl localhost:3000/api/v1/docs
```

## 🎨 Interface Types

### Console Output (ASCII Art)
```
Weather report: london

   \  /           Mainly clear
 _ /"".-.         +51(47) °F
   \_(   ).       ↓ 7 mph
   /(_(__)        1 inHg
                  0.0 in

                                               ┌─────────────┐
┌──────────────────────────────┬───────────────────────┤  Wed, Sep 24 ├───────────────────────┬──────────────────────────────┐
│            Morning           │             Noon      └──────┬──────┘     Evening           │             Night            │
├──────────────────────────────┼──────────────────────────────┼──────────────────────────────┼──────────────────────────────┤
│           Overcast           │           Overcast           │           Overcast           │           Overcast           │
│          +55(57) °F          │          +64(67) °F          │          +60(62) °F          │          +51(52) °F          │
│           ↙ 9 mph            │           ↙ 11 mph           │           ↙ 13 mph           │           ↙ 7 mph            │
│             6 mi             │             6 mi             │             6 mi             │             6 mi             │
│         0.0 in | 0%          │         0.0 in | 0%          │         0.0 in | 0%          │         0.0 in | 0%          │
└──────────────────────────────┴──────────────────────────────┴──────────────────────────────┴──────────────────────────────┘

Console Weather Service • Free weather data from Open-Meteo.com
```

### Browser Interface (Dracula Theme)
- **Dark purple background** with vibrant accent colors
- **Large temperature displays** with weather icons
- **Interactive location search** with GPS support
- **Responsive forecast tables** and mobile cards
- **Console integration examples** for developers

## 🌍 Location Support

### Input Formats
- **City names**: `london`, `paris`, `tokyo`
- **City + State**: `summerville,sc`, `austin,tx`
- **Coordinates**: `40.7128,-74.0060`
- **IP detection**: Automatic when no location specified

### Smart Detection
- **IP geolocation** with country-specific unit defaults
- **State abbreviations** (SC → South Carolina)
- **Multiple city matching** (finds correct Springfield, IL vs Springfield, MA)
- **Robust fallbacks** (never fails to provide weather)

## 🌡️ Unit Systems

- **Imperial** (°F, mph, inHg, in) - Default for US
- **Metric** (°C, km/h, hPa, mm) - Default for most countries
- **Auto-detection** based on IP location
- **Manual override** with `?units=imperial` or `?units=metric`

## 🎯 Special Endpoints

### Console Integration
```bash
curl localhost:3000/:help              # Documentation
curl localhost:3000/:bash.function     # Shell integration script
```

### Moon Phases
```bash
curl localhost:3000/moon               # Current location moon
curl localhost:3000/moon@london        # London moon phase
```

### Development Tools
```bash
curl localhost:3000/health             # Health check
curl localhost:3000/debug/ip           # IP detection info
curl localhost:3000/examples           # Usage examples
```

## 🐳 Docker Deployment

### Quick Docker Setup

**Prerequisites:** Docker and Docker Compose installed (see Quick Start section above)

### Development Environment
```bash
# Clone and enter directory
git clone https://github.com/apimgr/weather.git
cd weather

# Start development environment with hot reload
./docker-start.sh dev

# Access the service
curl localhost:3000/london              # Console weather
open http://localhost:3000/london       # Web interface

# View real-time logs
./docker-start.sh logs
```

### Production Environment
```bash
# Start production with nginx reverse proxy
./docker-start.sh prod

# Access through nginx proxy
curl localhost/london                   # Port 80 through nginx
curl localhost/api/v1/weather?location=paris

# Production includes:
# - Nginx reverse proxy with SSL support
# - Rate limiting (30 requests/minute)
# - Security headers
# - Gzip compression
# - Health monitoring
```

### Docker Management Commands
```bash
./docker-start.sh build               # Build optimized image
./docker-start.sh logs                # View container logs
./docker-start.sh stop                # Stop all services
./docker-start.sh clean               # Remove containers and images

# Manual Docker Compose commands
docker-compose up -d                   # Start development
docker-compose -f docker-compose.yml -f docker-compose.prod.yml --profile production up -d  # Production
docker-compose down                    # Stop services
```

### Simple Docker Run Commands

**For users who prefer `docker run` without docker-compose:**

```bash
# Build the image first
docker build -t console-weather .

# Development mode
docker run -d \
  --name weather-dev \
  -p 3000:3000 \
  -e NODE_ENV=development \
  console-weather

# Production mode
docker run -d \
  --name weather-prod \
  -p 3000:3000 \
  -e NODE_ENV=production \
  --restart unless-stopped \
  --health-cmd="curl -f http://localhost:3000/health || exit 1" \
  --health-interval=30s \
  --health-timeout=10s \
  --health-retries=3 \
  console-weather

# With custom port
docker run -d \
  --name weather-custom \
  -p 8080:3000 \
  -e PORT=3000 \
  console-weather

# Access the service
curl localhost:3000/london              # Default port
curl localhost:8080/london              # Custom port

# View logs
docker logs weather-prod -f

# Stop and remove
docker stop weather-prod
docker rm weather-prod
```

### Docker Registry (GitHub Container Registry)
```bash
# Pull and run from GitHub Container Registry:
docker run -d \
  --name console-weather \
  -p 3000:3000 \
  --restart unless-stopped \
  ghcr.io/apimgr/weather:latest

# With health checks and logging
docker run -d \
  --name console-weather \
  -p 3000:3000 \
  --restart unless-stopped \
  --health-cmd="curl -f http://localhost:3000/health || exit 1" \
  --health-interval=30s \
  ghcr.io/apimgr/weather:latest

# No need to clone repository or build locally
```

### Docker Configuration Files

**Core Files:**
- `Dockerfile` - Optimized Node.js 18 Alpine build
- `docker-compose.yml` - Base service configuration
- `docker-compose.override.yml` - Development overrides
- `docker-compose.prod.yml` - Production with nginx
- `nginx.conf` - Reverse proxy configuration
- `.dockerignore` - Optimized build context

### Production SSL Setup
```bash
# Add SSL certificates to ssl/ directory
mkdir ssl
cp your-cert.pem ssl/cert.pem
cp your-key.pem ssl/key.pem

# Update nginx.conf with your domain
# Uncomment HTTPS server block in nginx.conf

# Start production with SSL
./docker-start.sh prod
```

### Docker Troubleshooting

**Common Issues:**

**Docker not found:**
```bash
# Install Docker first (see Prerequisites above)
docker --version  # Should show version
```

**Permission denied:**
```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
# Or use sudo with docker commands
```

**Port already in use:**
```bash
# Check what's using port 3000
sudo lsof -i :3000
# Kill the process or change PORT environment variable
PORT=3001 ./docker-start.sh dev
```

**Build failures:**
```bash
# Clean Docker cache and rebuild
./docker-start.sh clean
docker system prune -f
./docker-start.sh build
```

**Container health checks:**
```bash
# Check container status
docker-compose ps
docker-compose logs weather-service

# Manual health check
curl localhost:3000/health
```

## 📁 Project Structure

```
src/
├── routes/           # Express route handlers
│   ├── api.js        # JSON API endpoints
│   ├── weather.js    # Main weather routes
│   └── web.js        # HTML interface routes
├── services/         # Core business logic
│   ├── moonService.js     # Moon phase calculations
│   └── weatherService.js  # Weather data fetching
├── utils/            # Utility functions
│   ├── asciiWeather.js    # Simple ASCII renderer
│   ├── wttrRenderer.js    # Full ASCII art renderer
│   ├── oneLineRenderer.js # Status bar format
│   ├── parameterParser.js # CLI parameter parsing
│   ├── locationParser.js  # Location detection & parsing
│   ├── userAgentDetector.js # Browser vs console detection
│   ├── hostDetector.js    # Dynamic host detection
│   ├── unitConverter.js   # Imperial/metric conversion
│   └── locationFallbacks.js # Fallback coordinates
├── views/            # EJS templates
│   ├── weather.ejs   # Weather web interface
│   └── moon.ejs      # Moon phase web interface
├── public/           # Static assets
│   └── css/
│       └── dracula.css    # Dracula theme styling
├── share/            # Integration files
│   └── bash.function # Shell integration script
└── server.js         # Main Express application
```

## 🔧 Configuration

### Environment Variables
```bash
PORT=3000                    # Server port
NODE_ENV=production          # Environment mode
```

### Reverse Proxy Support
- **Trust proxy enabled** - Works behind nginx, Cloudflare, AWS ALB
- **Real IP detection** - Proper client IP for geolocation
- **Dynamic host detection** - Correct URLs in documentation

### Production Setup
1. **SSL certificates** → `./ssl/cert.pem` and `./ssl/key.pem`
2. **Update nginx.conf** with your domain
3. **Deploy**: `./docker-start.sh prod`

## 🌐 API Reference

### Console Endpoints
| Endpoint | Description | Example |
|----------|-------------|---------|
| `GET /` | Weather for current location | `curl localhost:3000/` |
| `GET /:location` | Weather for specific location | `curl localhost:3000/london` |
| `GET /moon` | Moon phase current location | `curl localhost:3000/moon` |
| `GET /:help` | Documentation | `curl localhost:3000/:help` |

### JSON API Endpoints  
| Endpoint | Description | Parameters |
|----------|-------------|------------|
| `GET /api/v1/weather` | Current weather + forecast | `location`, `lat`, `lon`, `units` |
| `GET /api/v1/current` | Detailed current weather | `location`, `lat`, `lon`, `units` |
| `GET /api/v1/forecast` | Multi-day forecast | `location`, `lat`, `lon`, `days`, `units` |
| `GET /api/v1/search` | Location search | `q` (required) |

### Query Parameters (Full Compatibility)

**Format Options:**
- **`format=0`** - Current weather only
- **`format=1`** - Current + today forecast
- **`format=2`** - Current + 2-day forecast  
- **`format=3`** - Current + 3-day forecast
- **`format=4`** - Current + today + chart
- **`format=simple`** - Minimal text output

**Unit Options:**
- **`u`** - Imperial/USCS units (°F, mph, inHg, in)
- **`m`** - Metric units (°C, km/h, hPa, mm)
- **`M`** - SI units (metric system)
- **`units=imperial|metric`** - Explicit unit specification

**View Options:**
- **`F`** - Remove footer/attribution
- **`n`** - Narrow format
- **`q`** - Quiet mode (no location caption)
- **`Q`** - Super quiet mode
- **`T`** - Transparency (for future PNG support)

**Examples:**
- **`?u&format=3&F`** - Imperial, 3-day forecast, no footer
- **`?m&q`** - Metric units, quiet mode
- **`?format=0`** - Current weather only

## 🎨 Theming

### Dracula Color Palette
- **Background**: `#282a36` (Dark purple)
- **Foreground**: `#f8f8f2` (Light gray) 
- **Accent Colors**: Cyan (`#8be9fd`), Yellow (`#f1fa8c`), Green (`#50fa7b`), Pink (`#ff79c6`), Purple (`#bd93f9`)
- **Interactive**: Orange (`#ffb86c`), Red (`#ff5555`)

### ASCII Art Colors
- **Weather icons**: Condition-specific colors
- **Temperature**: Yellow highlights
- **Wind**: Green arrows and speeds
- **Tables**: Purple borders with colorful data
- **Headers**: Bright cyan titles

## 📊 Performance

### Caching Strategy
- **Weather data**: 10-minute TTL
- **Location geocoding**: Permanent cache
- **IP geolocation**: Smart fallbacks
- **Static assets**: 1-year browser cache

### Rate Limiting
- **Development**: No limits
- **Production**: 30 requests/minute per IP (via nginx)
- **API**: No key required, generous limits

## 🛡️ Security

### Content Security Policy
- **Inline scripts**: Allowed for EJS templates
- **External resources**: Restricted to essential services
- **XSS protection**: Comprehensive header security

### Docker Security
- **Non-root user**: Container runs as `weatherapp` user
- **Minimal image**: Alpine Linux base
- **Security scanning**: Health checks and monitoring

## 🔍 Monitoring

### Health Checks
```bash
curl localhost:3000/health              # Application health
docker-compose ps                       # Container status
./docker-start.sh logs                  # View all logs
```

### Debug Tools
```bash
curl localhost:3000/debug/ip            # IP detection info
curl localhost:3000/examples            # Usage examples
curl localhost:3000/test                # Theme test page
```

## 🚀 Deployment Guide

### Local Development
1. **Clone repository**
2. **Install dependencies**: `npm install`
3. **Start server**: `npm run dev`
4. **Access**: `http://localhost:3000`

### Docker Development
1. **Start containers**: `./docker-start.sh dev`
2. **Access**: `http://localhost:3000`
3. **Logs**: `./docker-start.sh logs`

### Production Deployment
1. **Prepare SSL certificates** (optional)
2. **Start production**: `./docker-start.sh prod`
3. **Access**: `http://localhost` (nginx proxy)
4. **Monitor**: Health checks and logging

### Cloud Deployment
- **AWS/GCP/Azure**: Use `docker-compose.prod.yml`
- **Kubernetes**: Helm charts available in `/k8s` directory
- **Cloudflare**: Perfect reverse proxy support

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the repository: [github.com/apimgr/weather](https://github.com/apimgr/weather)
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Third-party licenses and attributions are documented in [LICENSE.md](LICENSE.md).

## 🙏 Acknowledgments

- **Open-Meteo.com** - Free weather data provider
- **Dracula Theme** - Beautiful color palette inspiration
- **Express.js** - Web framework foundation
- **Community** - Open source weather tools and CLI inspiration

## 🔗 Links

- **🌐 Official Service**: [wthr.top](http://wthr.top)
- **📦 Source Code**: [github.com/apimgr/weather](https://github.com/apimgr/weather)
- **🐛 Report Issues**: [github.com/apimgr/weather/issues](https://github.com/apimgr/weather/issues)
- **📚 API Documentation**: [wthr.top/api/v1/docs](http://wthr.top/api/v1/docs)
- **🔧 Docker Registry**: [ghcr.io/apimgr/weather](https://ghcr.io/apimgr/weather)

---

**Made with ❤️ for the developer community**

*A modern, accessible, and beautiful weather service for the terminal and web.*
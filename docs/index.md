# Weather API Documentation

A production-grade weather API service providing global weather forecasts, severe weather alerts, earthquake data, moon phase information, and hurricane tracking.

## Features

- **Global Weather Forecasts**: 16-day forecasts for any location worldwide
- **Severe Weather Alerts**: Real-time alerts for hurricanes, tornadoes, storms, floods, and winter weather
- **International Support**: Weather alerts from US, Canada, UK, Australia, Japan, and Mexico
- **Earthquake Tracking**: Real-time earthquake data with interactive maps
- **Moon Phases**: Detailed lunar information including phases, illumination, rise/set times
- **Hurricane Tracking**: Active storm tracking with advisories and forecasts
- **GeoIP Location**: Automatic location detection via IP address (IPv4 and IPv6)
- **Location Persistence**: Cookie-based location memory across sessions
- **Mobile Responsive**: Optimized for desktop, tablet, and mobile devices
- **Dracula Theme**: Beautiful dark theme with cyan accents

## Quick Start

### Docker (Production)

```bash
# Pull and run the latest image
docker pull ghcr.io/apimgr/weather:latest
docker run -d -p 80:80 \
  -v ./data:/data \
  -v ./config:/config \
  -v ./logs:/var/log/weather \
  --name weather \
  ghcr.io/apimgr/weather:latest
```

Visit http://localhost

### Docker Compose (Recommended)

```bash
# Clone repository
git clone https://github.com/apimgr/weather.git
cd weather

# Start service
docker compose up -d

# View logs
docker compose logs -f
```

Service runs on http://172.17.0.1:64080

### Binary Installation

Download the latest release for your platform from [GitHub Releases](https://github.com/apimgr/weather/releases):

```bash
# Linux AMD64
wget https://github.com/apimgr/weather/releases/download/v1.0.0/weather-linux-amd64
chmod +x weather-linux-amd64
./weather-linux-amd64

# macOS ARM64 (Apple Silicon)
wget https://github.com/apimgr/weather/releases/download/v1.0.0/weather-darwin-arm64
chmod +x weather-darwin-arm64
./weather-darwin-arm64

# Windows
# Download weather-windows-amd64.exe and run
```

## API Endpoints

### Weather
- `GET /` - Weather forecast for your location (IP-based)
- `GET /weather/{location}` - Weather forecast for specific location
- `GET /{location}` - Weather forecast (backwards compatible)

### Severe Weather
- `GET /severe-weather` - Severe weather alerts for your location
- `GET /severe-weather/{location}` - Severe weather alerts for specific location
- `GET /severe/{type}` - Filter by alert type (hurricanes, tornadoes, storms, winter, floods)
- `GET /severe/{type}/{location}` - Filtered alerts for specific location

### Specialized Data
- `GET /moon` - Moon phase for your location
- `GET /moon/{location}` - Moon phase for specific location
- `GET /earthquake` - Recent earthquakes near your location
- `GET /earthquake/{location}` - Recent earthquakes near specific location
- `GET /hurricane` - Active hurricane tracking (redirects to /severe-weather)

### API Endpoints (JSON)
- `GET /api/weather?location={location}` - JSON weather data
- `GET /api/moon?location={location}` - JSON moon data
- `GET /api/earthquake?location={location}` - JSON earthquake data
- `GET /api/severe?location={location}&distance={miles}&type={type}` - JSON severe weather data

## Location Formats

The service accepts multiple location formats:

- **City, State**: `Brooklyn, NY` or `Brooklyn,NY`
- **City, Country**: `London, UK` or `Tokyo, JP`
- **ZIP Code**: `10001` (US only)
- **Coordinates**: `40.7128,-74.0060` (lat,lon)
- **Airport Code**: `JFK` or `LAX`
- **IP Address**: Automatic detection if no location specified

## Data Sources

- **Weather**: [Open-Meteo](https://open-meteo.com/) - Global weather forecasts
- **US Severe Weather**: [NOAA NWS](https://www.weather.gov/) - National Weather Service alerts
- **Hurricanes**: [NOAA NHC](https://www.nhc.noaa.gov/) - National Hurricane Center
- **Canadian Alerts**: [Environment Canada](https://weather.gc.ca/)
- **UK Alerts**: [Met Office](https://www.metoffice.gov.uk/)
- **Australian Alerts**: [Bureau of Meteorology](http://www.bom.gov.au/)
- **Japanese Alerts**: [JMA](https://www.jma.go.jp/)
- **Mexican Alerts**: [CONAGUA](https://www.gob.mx/conagua)
- **Earthquakes**: [USGS](https://earthquake.usgs.gov/) - US Geological Survey
- **GeoIP**: [sapics/ip-location-db](https://github.com/sapics/ip-location-db) - MaxMind GeoLite2 databases

## Configuration

### Environment Variables

- `PORT` - HTTP server port (default: 80)
- `DOMAIN` - Public domain name (for URL generation)
- `HOSTNAME` - Server hostname (fallback for DOMAIN)
- `DATA_DIR` - Data directory (default: ./data)
- `CONFIG_DIR` - Configuration directory (default: ./config)
- `LOG_DIR` - Log directory (default: /var/log/weather)

### Example Configuration

```bash
PORT=3050
DOMAIN=weather.example.com
DATA_DIR=/var/lib/weather/data
CONFIG_DIR=/etc/weather
LOG_DIR=/var/log/weather
```

## Advanced Usage

### Custom Distance Filtering

Severe weather alerts can be filtered by distance:

```bash
# Show alerts within 25 miles
curl "http://localhost/api/severe?location=Brooklyn,NY&distance=25"

# Show alerts within 100 miles
curl "http://localhost/api/severe?location=Miami,FL&distance=100"
```

### Type Filtering

Filter severe weather by type:

- `hurricanes` - Tropical storms and hurricanes
- `tornadoes` - Tornado warnings and watches
- `storms` - Severe thunderstorm warnings
- `winter` - Winter storm warnings and advisories
- `floods` - Flood warnings and watches

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

## Support

- **Documentation**: [https://weather.readthedocs.io](https://weather.readthedocs.io)
- **Issues**: [https://github.com/apimgr/weather/issues](https://github.com/apimgr/weather/issues)
- **Docker Hub**: [https://ghcr.io/apimgr/weather](https://ghcr.io/apimgr/weather)

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.

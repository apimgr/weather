# Weather Service

[![Build](https://github.com/apimgr/weather/actions/workflows/build.yml/badge.svg)](https://github.com/apimgr/weather/actions/workflows/build.yml)
[![Release](https://img.shields.io/github/v/release/apimgr/weather)](https://github.com/apimgr/weather/releases)
[![Documentation](https://readthedocs.org/projects/apimgr-weather/badge/?version=latest)](https://apimgr-weather.readthedocs.io)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)

## About

A unified weather information platform that aggregates global meteorological data from authoritative sources into a single, accessible service for developers, businesses, and end users.

Weather Service combines forecasts, severe weather alerts, earthquakes, hurricanes, and lunar data from 10+ government agencies worldwide through one consistent API and web interface.

## Official Site

https://weather.apimgr.us

## Features

- Global weather forecasts (16-day) via Open-Meteo
- Severe weather alerts from 6 countries (NOAA, Environment Canada, UK Met Office, BOM, JMA, CONAGUA)
- Real-time earthquake monitoring via USGS
- Hurricane/tropical storm tracking via NOAA NHC
- Moon phase calculations
- Automatic IP-based location detection
- WebSocket real-time notifications
- Single static binary with all assets embedded
- Multi-platform (Linux, macOS, Windows, FreeBSD - amd64/arm64)
- Production grade with rate limiting, caching, health checks

## Production

### Docker (Recommended)

```bash
docker run -d \
  --name weather \
  -p 64580:80 \
  -v ./rootfs/config:/config:z \
  -v ./rootfs/data:/data:z \
  ghcr.io/apimgr/weather:latest
```

### Docker Compose

```bash
curl -O https://raw.githubusercontent.com/apimgr/weather/main/docker/docker-compose.yml
docker compose up -d
```

### Binary

```bash
# Download latest release
curl -LO https://github.com/apimgr/weather/releases/latest/download/weather-linux-amd64

# Make executable and run
chmod +x weather-linux-amd64
./weather-linux-amd64
```

### Systemd Service

```bash
# Download and install
sudo mv weather-linux-amd64 /usr/local/bin/weather
sudo useradd -r -s /bin/false weather
sudo mkdir -p /var/lib/weather/{data,config}
sudo chown -R weather:weather /var/lib/weather

# Create service
sudo tee /etc/systemd/system/weather.service > /dev/null <<EOF
[Unit]
Description=Weather API Service
After=network.target

[Service]
Type=simple
User=weather
ExecStart=/usr/local/bin/weather
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now weather
```

## CLI Client

A companion CLI client is available for interacting with the server API.

### Install

```bash
# Download latest release
curl -LO https://github.com/apimgr/weather/releases/latest/download/weather-cli-linux-amd64
chmod +x weather-cli-linux-amd64
sudo mv weather-cli-linux-amd64 /usr/local/bin/weather-cli
```

### Configure

```bash
# Connect to server (creates ~/.config/apimgr/weather/cli.yml)
weather-cli --server https://api.example.com --token YOUR_API_TOKEN
```

### Usage

```bash
weather-cli --help
weather-cli weather Brooklyn,NY
weather-cli severe-weather
weather-cli moon
```

## Configuration

Configuration is auto-generated on first run. Edit via admin panel at `http://{fqdn}:{port}/admin`.

Key settings:
- `server.address` - Listen address (default: :80)
- `server.mode` - production or development

Configuration file: `$CONFIG_DIR/server.yml`

```yaml
server:
  address: ":80"
  mode: production

logging:
  level: info
  format: json

geoip:
  update_interval: 168h

cache:
  weather_ttl: 15m
  severe_ttl: 5m
```

## API

API documentation available at `/api/v1/` when running.

| Endpoint | Description |
|----------|-------------|
| `GET /healthz` | Health check |
| `GET /api/v1/weather/:location` | Weather forecast |
| `GET /api/v1/severe-weather` | Severe weather alerts |
| `GET /api/v1/earthquakes` | Earthquake data |
| `GET /api/v1/hurricanes` | Hurricane tracking |
| `GET /api/v1/moon` | Moon phase |

### Examples

```bash
# Weather for location
curl http://localhost/api/v1/weather/Brooklyn,NY

# Severe weather
curl http://localhost/api/v1/severe-weather

# Moon phase
curl http://localhost/api/v1/moon
```

## Other

### Data Sources

- **Weather**: [Open-Meteo](https://open-meteo.com/)
- **US Alerts**: [NOAA NWS](https://www.weather.gov/)
- **Hurricanes**: [NOAA NHC](https://www.nhc.noaa.gov/)
- **Earthquakes**: [USGS](https://earthquake.usgs.gov/)
- **GeoIP**: [sapics/ip-location-db](https://github.com/sapics/ip-location-db)

### Troubleshooting

- Check logs: `docker logs weather`
- Health check: `curl http://{fqdn}:{port}/healthz`
- Port in use: `netstat -tulpn | grep :80`

### Support

- Documentation: https://apimgr-weather.readthedocs.io
- Issues: https://github.com/apimgr/weather/issues

## Development

**Development instructions are for contributors only.**

### Prerequisites

- Docker (for containerized builds)
- Make

### Build

```bash
# Clone
git clone https://github.com/apimgr/weather
cd weather

# Quick dev build (outputs to binaries/)
make dev

# Full build (all platforms, outputs to binaries/)
make build

# Test
make test
```

### Project Structure

```
src/           # Source code
tests/         # Test files
docker/        # Docker files
binaries/      # Built binaries (gitignored)
```

## License

MIT - See [LICENSE.md](LICENSE.md)

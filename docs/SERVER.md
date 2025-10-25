# Server Administration Guide

Complete guide for deploying, configuring, and maintaining the Weather service.

## System Requirements

### Minimum Requirements

- **CPU**: 1 core (2+ recommended)
- **RAM**: 512 MB (1 GB+ recommended)
- **Disk**: 500 MB free space (includes GeoIP databases ~150 MB)
- **OS**: Linux, macOS, Windows, FreeBSD
- **Network**: Internet access for external APIs

### Recommended Production Setup

- **CPU**: 2+ cores
- **RAM**: 2 GB+
- **Disk**: 2 GB+ free space (includes logs, cache)
- **OS**: Linux (Alpine, Ubuntu, Debian, RHEL)
- **Network**: Static IP, domain name, reverse proxy

## Installation

### Docker Installation (Recommended)

#### Using Docker Run

```bash
# Create directories
mkdir -p ~/weather/{data,config,logs}

# Run container
docker run -d \
  --name weather \
  --restart unless-stopped \
  -p 80:80 \
  -v ~/weather/data:/data \
  -v ~/weather/config:/config \
  -v ~/weather/logs:/var/log/weather \
  -e DOMAIN=weather.example.com \
  ghcr.io/apimgr/weather:latest

# View logs
docker logs -f weather
```

#### Using Docker Compose

1. **Create directory structure:**

```bash
mkdir -p ~/weather
cd ~/weather
```

2. **Create docker-compose.yml:**

```yaml
services:
  weather:
    image: ghcr.io/apimgr/weather:latest
    container_name: weather
    restart: unless-stopped
    ports:
      - "172.17.0.1:64080:80"
    volumes:
      - ./rootfs/data/weather:/data
      - ./rootfs/config/weather:/config
      - ./rootfs/logs/weather:/var/log/weather
    environment:
      - DOMAIN=weather.example.com
      - PORT=80
    networks:
      - weather

networks:
  weather:
    driver: bridge
```

3. **Start service:**

```bash
docker compose up -d
docker compose logs -f
```

### Binary Installation

#### Linux

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

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable weather
sudo systemctl start weather
sudo systemctl status weather
```

#### macOS

```bash
# Download binary
wget https://github.com/apimgr/weather/releases/latest/download/weather-darwin-arm64
chmod +x weather-darwin-arm64
sudo mv weather-darwin-arm64 /usr/local/bin/weather

# Create directories
sudo mkdir -p /usr/local/var/weather/{data,config}
sudo mkdir -p /usr/local/var/log/weather

# Run manually
weather

# Or create launchd service
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

sudo launchctl load /Library/LaunchDaemons/com.apimgr.weather.plist
```

#### Windows

```powershell
# Download binary
Invoke-WebRequest -Uri "https://github.com/apimgr/weather/releases/latest/download/weather-windows-amd64.exe" -OutFile "weather.exe"

# Create directories
New-Item -ItemType Directory -Force -Path "C:\ProgramData\Weather\data"
New-Item -ItemType Directory -Force -Path "C:\ProgramData\Weather\config"
New-Item -ItemType Directory -Force -Path "C:\ProgramData\Weather\logs"

# Move binary
Move-Item weather.exe "C:\Program Files\Weather\weather.exe"

# Run manually
& "C:\Program Files\Weather\weather.exe"

# Or install as Windows Service using NSSM
# Download NSSM from https://nssm.cc/download
nssm install Weather "C:\Program Files\Weather\weather.exe"
nssm set Weather AppDirectory "C:\ProgramData\Weather"
nssm set Weather AppStdout "C:\ProgramData\Weather\logs\output.log"
nssm set Weather AppStderr "C:\ProgramData\Weather\logs\error.log"
nssm set Weather AppEnvironmentExtra "PORT=80"
nssm start Weather
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | HTTP server port | `80` | No |
| `DOMAIN` | Public domain name | Auto-detected | No |
| `HOSTNAME` | Server hostname | Auto-detected | No |
| `DATA_DIR` | Data directory | `./data` | No |
| `CONFIG_DIR` | Configuration directory | `./config` | No |
| `LOG_DIR` | Log directory | `/var/log/weather` | No |

### Configuration File

Create `config/weather.yaml`:

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

## Reverse Proxy Setup

### Nginx

```nginx
server {
    listen 80;
    server_name weather.example.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name weather.example.com;

    ssl_certificate /etc/letsencrypt/live/weather.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/weather.example.com/privkey.pem;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    # Proxy settings
    location / {
        proxy_pass http://172.17.0.1:64080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # Buffering
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }

    # Cache static assets
    location ~* \.(css|js|jpg|jpeg|png|gif|ico|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://172.17.0.1:64080;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
```

### Apache

```apache
<VirtualHost *:80>
    ServerName weather.example.com
    Redirect permanent / https://weather.example.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName weather.example.com

    SSLEngine on
    SSLCertificateFile /etc/letsencrypt/live/weather.example.com/fullchain.pem
    SSLCertificateKeyFile /etc/letsencrypt/live/weather.example.com/privkey.pem

    # Security headers
    Header always set Strict-Transport-Security "max-age=31536000; includeSubDomains"
    Header always set X-Frame-Options "SAMEORIGIN"
    Header always set X-Content-Type-Options "nosniff"

    # Proxy settings
    ProxyPreserveHost On
    ProxyPass / http://172.17.0.1:64080/
    ProxyPassReverse / http://172.17.0.1:64080/

    RequestHeader set X-Forwarded-Proto "https"
    RequestHeader set X-Forwarded-For %{REMOTE_ADDR}s
</VirtualHost>
```

### Caddy

```caddy
weather.example.com {
    reverse_proxy 172.17.0.1:64080

    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
        X-Frame-Options "SAMEORIGIN"
        X-Content-Type-Options "nosniff"
    }

    encode gzip

    log {
        output file /var/log/caddy/weather.log
    }
}
```

## Monitoring

### Health Checks

Check service health:

```bash
curl http://localhost/api/health
```

Expected response:

```json
{
  "status": "healthy",
  "timestamp": "2025-10-16T12:00:00Z",
  "version": "1.0.0",
  "uptime_seconds": 86400
}
```

### Logs

View logs:

```bash
# Docker
docker logs -f weather

# Systemd
sudo journalctl -u weather -f

# Direct file
tail -f /var/log/weather/weather.log
```

### Metrics

Monitor key metrics:

- **Request rate**: requests per second
- **Response time**: average latency
- **Error rate**: 5xx errors per minute
- **Memory usage**: RSS in MB
- **Disk usage**: data directory size
- **GeoIP database age**: days since update

### Alerts

Set up alerts for:

- Service down (health check fails)
- High error rate (>5% 5xx responses)
- High latency (>2 seconds average)
- High memory (>80% of available)
- Disk full (>90% used)

## Maintenance

### Update Service

#### Docker

```bash
# Pull latest image
docker pull ghcr.io/apimgr/weather:latest

# Restart container
docker compose down
docker compose up -d

# Or with docker run
docker stop weather
docker rm weather
docker run -d [same options as before] ghcr.io/apimgr/weather:latest
```

#### Binary

```bash
# Download new binary
wget https://github.com/apimgr/weather/releases/latest/download/weather-linux-amd64
chmod +x weather-linux-amd64

# Stop service
sudo systemctl stop weather

# Replace binary
sudo mv weather-linux-amd64 /usr/local/bin/weather

# Start service
sudo systemctl start weather
sudo systemctl status weather
```

### Update GeoIP Databases

GeoIP databases update automatically every week. To force an update:

```bash
# Delete existing databases
rm -rf data/geoip/*.mmdb

# Restart service (will download fresh databases)
docker compose restart
# or
sudo systemctl restart weather
```

### Backup

Backup important directories:

```bash
# Create backup
tar -czf weather-backup-$(date +%Y%m%d).tar.gz \
  /var/lib/weather/config \
  /var/lib/weather/data

# Restore backup
tar -xzf weather-backup-20251016.tar.gz -C /
sudo chown -R weather:weather /var/lib/weather
sudo systemctl restart weather
```

### Log Rotation

Configure logrotate for systemd deployments:

```bash
sudo tee /etc/logrotate.d/weather > /dev/null <<EOF
/var/log/weather/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0640 weather weather
    sharedscripts
    postrotate
        systemctl reload weather > /dev/null 2>&1 || true
    endscript
}
EOF
```

For Docker, logs are managed by Docker's logging driver.

## Troubleshooting

### Service Won't Start

1. **Check logs**:

```bash
docker logs weather
# or
sudo journalctl -u weather -n 50
```

2. **Check port availability**:

```bash
sudo netstat -tulpn | grep :80
# or
sudo ss -tulpn | grep :80
```

3. **Check permissions**:

```bash
ls -la /var/lib/weather
ls -la /var/log/weather
```

### GeoIP Not Working

1. **Check database files exist**:

```bash
ls -lh data/geoip/
```

2. **Check database age**:

```bash
find data/geoip/ -name "*.mmdb" -mtime +7
```

3. **Force database update**:

```bash
rm -rf data/geoip/*.mmdb
docker compose restart
```

### High Memory Usage

1. **Check memory usage**:

```bash
docker stats weather
# or
ps aux | grep weather
```

2. **Limit Docker memory**:

```yaml
services:
  weather:
    image: ghcr.io/apimgr/weather:latest
    deploy:
      resources:
        limits:
          memory: 512M
```

3. **Review cache settings** in config/weather.yaml

### Slow Response Times

1. **Check external API status**:

```bash
curl -w "@curl-format.txt" -o /dev/null -s "https://api.open-meteo.com/v1/forecast?latitude=40.7128&longitude=-74.0060&current=temperature_2m"
```

2. **Check cache hit rate** in logs

3. **Consider adding CDN** for static assets

4. **Increase reverse proxy timeouts**

### SSL/TLS Issues

1. **Verify certificate expiration**:

```bash
echo | openssl s_client -servername weather.example.com -connect weather.example.com:443 2>/dev/null | openssl x509 -noout -dates
```

2. **Renew Let's Encrypt certificates**:

```bash
sudo certbot renew
sudo nginx -t && sudo systemctl reload nginx
```

## Security

### Firewall Configuration

#### UFW (Ubuntu/Debian)

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

#### firewalld (RHEL/CentOS)

```bash
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

### Rate Limiting

Rate limiting is built-in (100 requests per 15 minutes per IP). Adjust in config:

```yaml
rate_limit:
  requests_per_window: 100
  window_duration: 15m
  burst: 20
```

### DDoS Protection

Consider using:

- **Cloudflare**: CDN + DDoS protection
- **fail2ban**: Ban IPs with excessive failed requests
- **nginx limit_req**: Additional rate limiting at reverse proxy

### Security Updates

Subscribe to security advisories:

- GitHub Security Advisories: https://github.com/apimgr/weather/security/advisories
- Docker Hub: Watch for updated images
- OS security updates: Regular patching

## Performance Tuning

### Optimize Docker

```yaml
services:
  weather:
    image: ghcr.io/apimgr/weather:latest
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '1'
          memory: 512M
```

### Optimize Go Runtime

Set environment variables:

```bash
GOMAXPROCS=2  # Number of CPU cores
GOGC=100      # GC percentage (lower = more frequent GC)
```

### Database Optimization

GeoIP databases are memory-mapped for fast lookups. Ensure sufficient RAM:

- City databases: ~90 MB in memory
- Country/ASN: ~13 MB in memory
- Total: ~103 MB minimum

## Scaling

### Horizontal Scaling

Run multiple instances behind a load balancer:

```yaml
services:
  weather1:
    image: ghcr.io/apimgr/weather:latest
    # ... config ...

  weather2:
    image: ghcr.io/apimgr/weather:latest
    # ... config ...

  loadbalancer:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
```

Load balancer config:

```nginx
upstream weather_backend {
    least_conn;
    server weather1:80;
    server weather2:80;
}

server {
    listen 80;
    location / {
        proxy_pass http://weather_backend;
    }
}
```

### Vertical Scaling

Increase resources for single instance:

- **CPU**: 4+ cores for high traffic
- **RAM**: 2-4 GB for large cache
- **Disk**: SSD for faster database access

## Support

- **Documentation**: https://apimgr-weather.readthedocs.io
- **Issues**: https://github.com/apimgr/weather/issues
- **Discussions**: https://github.com/apimgr/weather/discussions
- **Docker Hub**: https://ghcr.io/apimgr/weather

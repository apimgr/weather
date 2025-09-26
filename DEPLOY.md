# 🚀 Deployment Guide for wthr.top

## Production Deployment Steps

### 1. Server Setup
```bash
# On production server (Ubuntu/Debian)
sudo apt update
sudo apt install docker.io docker-compose nginx certbot

# Clone repository
git clone https://github.com/apimgr/weather.git
cd weather
```

### 2. SSL Certificate Setup
```bash
# Get Let's Encrypt certificate for wthr.top
sudo certbot certonly --nginx -d wthr.top -d www.wthr.top

# Copy certificates to project
sudo cp /etc/letsencrypt/live/wthr.top/fullchain.pem ssl/cert.pem
sudo cp /etc/letsencrypt/live/wthr.top/privkey.pem ssl/key.pem
sudo chown $USER:$USER ssl/*.pem
```

### 3. Production Configuration
```bash
# Update nginx.conf with wthr.top domain
sed -i 's/server_name _;/server_name wthr.top www.wthr.top;/' nginx.conf

# Uncomment HTTPS server block in nginx.conf
# Update SSL certificate paths if needed
```

### 4. Deploy Service
```bash
# Start production environment
./docker-start.sh prod

# Or use direct docker run
docker run -d \
  --name console-weather \
  -p 3000:3000 \
  --restart unless-stopped \
  -e NODE_ENV=production \
  ghcr.io/apimgr/weather:latest
```

### 5. Configure Reverse Proxy
```bash
# Update main nginx config to proxy to container
sudo tee /etc/nginx/sites-available/wthr.top <<EOF
server {
    listen 80;
    server_name wthr.top www.wthr.top;
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name wthr.top www.wthr.top;
    
    ssl_certificate /etc/letsencrypt/live/wthr.top/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/wthr.top/privkey.pem;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

# Enable site
sudo ln -sf /etc/nginx/sites-available/wthr.top /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

### 6. Monitoring & Maintenance
```bash
# Check service health
curl wthr.top/health

# View logs
docker logs console-weather -f

# Update deployment
git pull
docker pull ghcr.io/apimgr/weather:latest
docker stop console-weather
docker rm console-weather
./docker-start.sh prod

# Renew SSL certificates (add to crontab)
# 0 3 * * * certbot renew --quiet && systemctl reload nginx
```

## Environment Variables

```bash
# Production environment
NODE_ENV=production
PORT=3000

# Optional: Custom settings
WTTR_CACHE_TTL=600          # Weather cache time (seconds)
WTTR_LOCATION_CACHE=604800  # Location cache time (seconds)
```

## Security Considerations

- **Rate limiting** - Built into nginx configuration
- **SSL/TLS** - Let's Encrypt certificates
- **Security headers** - Helmet.js configured
- **Non-root container** - Security best practices
- **Regular updates** - Automated via GitHub Actions

## Performance Optimization

- **Caching** - Multiple layers (weather, location, static)
- **Compression** - Gzip enabled in nginx
- **Health checks** - Container monitoring
- **Multi-platform** - ARM64 and AMD64 support

## Backup & Recovery

```bash
# Backup configuration
tar -czf weather-backup.tar.gz nginx.conf ssl/ docker-compose*.yml

# Quick recovery
tar -xzf weather-backup.tar.gz
./docker-start.sh prod
```

This guide ensures wthr.top runs reliably with proper SSL, monitoring, and security.
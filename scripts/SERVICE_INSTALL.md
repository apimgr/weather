# Weather Service Installation Guide

This guide covers installing Weather as a system service on Linux, macOS, and Windows.

## Quick Start

```bash
# Linux
sudo make install              # Installs binary and systemd service
sudo systemctl enable weather
sudo systemctl start weather

# macOS
make install                   # Installs binary and launchd agent
launchctl load ~/Library/LaunchAgents/com.apimgr.weather.plist

# Windows (PowerShell as Administrator)
.\scripts\windows.ps1 -InstallService
```

## Linux (systemd)

### Automatic Installation

```bash
# Clone repository
git clone https://github.com/apimgr/weather.git
cd weather

# Run installer (requires root)
sudo ./scripts/linux.sh
```

### Manual Installation

1. **Build the binary:**
   ```bash
   make build
   ```

2. **Copy binary to system:**
   ```bash
   sudo cp binaries/weather-linux-amd64 /usr/local/bin/weather
   sudo chmod +x /usr/local/bin/weather
   ```

3. **Create service user:**
   ```bash
   sudo useradd --system --no-create-home --shell /bin/false weather
   ```

4. **Create directories:**
   ```bash
   sudo mkdir -p /etc/weather/{certs,databases}
   sudo mkdir -p /var/lib/weather/backups
   sudo mkdir -p /var/log/weather
   sudo mkdir -p /var/cache/weather/weather
   sudo chown -R weather:weather /etc/weather /var/lib/weather /var/log/weather /var/cache/weather
   ```

5. **Install systemd service:**
   ```bash
   sudo cp scripts/weather.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable weather
   sudo systemctl start weather
   ```

6. **Check status:**
   ```bash
   sudo systemctl status weather
   sudo journalctl -u weather -f
   ```

### systemd Commands

```bash
# Start service
sudo systemctl start weather

# Stop service
sudo systemctl stop weather

# Restart service
sudo systemctl restart weather

# Check status
sudo systemctl status weather

# Enable auto-start on boot
sudo systemctl enable weather

# Disable auto-start
sudo systemctl disable weather

# View logs
sudo journalctl -u weather -f

# View last 100 log lines
sudo journalctl -u weather -n 100
```

## macOS (launchd)

### Automatic Installation

```bash
# Clone repository
git clone https://github.com/apimgr/weather.git
cd weather

# Run installer
./scripts/macos.sh
```

### Manual Installation (User)

1. **Build the binary:**
   ```bash
   make build
   ```

2. **Copy binary to system:**
   ```bash
   sudo cp binaries/weather-macos-amd64 /usr/local/bin/weather
   sudo chmod +x /usr/local/bin/weather
   ```

3. **Create directories:**
   ```bash
   mkdir -p ~/Library/Application\ Support/weather/{certs,databases,data}
   mkdir -p ~/Library/Logs/weather
   mkdir -p ~/Library/Caches/weather/weather
   ```

4. **Install launchd agent:**
   ```bash
   cp scripts/com.apimgr.weather.plist ~/Library/LaunchAgents/
   launchctl load ~/Library/LaunchAgents/com.apimgr.weather.plist
   ```

5. **Check status:**
   ```bash
   launchctl list | grep weather
   tail -f ~/Library/Logs/weather/stdout.log
   ```

### Manual Installation (System-wide, requires root)

1. **Copy binary:**
   ```bash
   sudo cp binaries/weather-macos-amd64 /usr/local/bin/weather
   sudo chmod +x /usr/local/bin/weather
   ```

2. **Create service user:**
   ```bash
   # macOS uses dscl for user management
   sudo dscl . -create /Users/_weather
   sudo dscl . -create /Users/_weather UserShell /usr/bin/false
   sudo dscl . -create /Users/_weather RealName "Weather Service"
   sudo dscl . -create /Users/_weather UniqueID 300
   sudo dscl . -create /Users/_weather PrimaryGroupID 300
   sudo dscl . -create /Users/_weather NFSHomeDirectory /var/empty
   ```

3. **Create directories:**
   ```bash
   sudo mkdir -p /Library/Application\ Support/weather/{certs,databases,data}
   sudo mkdir -p /Library/Logs/weather
   sudo mkdir -p /Library/Caches/weather/weather
   sudo chown -R _weather:_weather /Library/Application\ Support/weather /Library/Logs/weather /Library/Caches/weather
   ```

4. **Install launchd daemon:**
   ```bash
   sudo cp scripts/com.apimgr.weather.plist /Library/LaunchDaemons/
   sudo launchctl load /Library/LaunchDaemons/com.apimgr.weather.plist
   ```

### launchd Commands

```bash
# Load (start) service
launchctl load ~/Library/LaunchAgents/com.apimgr.weather.plist

# Unload (stop) service
launchctl unload ~/Library/LaunchAgents/com.apimgr.weather.plist

# Check if running
launchctl list | grep weather

# View logs
tail -f ~/Library/Logs/weather/stdout.log
tail -f ~/Library/Logs/weather/stderr.log

# For system service (requires sudo)
sudo launchctl load /Library/LaunchDaemons/com.apimgr.weather.plist
sudo launchctl unload /Library/LaunchDaemons/com.apimgr.weather.plist
```

## Windows

### Using Automated Script

```powershell
# Run PowerShell as Administrator
cd weather
.\scripts\windows.ps1 -InstallService
```

### Manual Installation

See [WINDOWS_SERVICE.md](WINDOWS_SERVICE.md) for detailed Windows service installation options including:
- NSSM (Non-Sucking Service Manager) - Recommended
- sc.exe (Windows Service Control)
- PowerShell wrapper

## Docker

If you prefer running as a container instead of a system service:

```bash
# Using docker-compose
docker-compose up -d

# Using docker run
docker run -d \
  --name weather \
  -p 127.0.0.1:64080:80 \
  -v ./rootfs/data/weather:/data \
  -v ./rootfs/config/weather:/config \
  --restart unless-stopped \
  ghcr.io/apimgr/weather:latest
```

## Verification

After installation on any platform:

1. **Check if service is running:**
   ```bash
   # Linux
   systemctl status weather

   # macOS
   launchctl list | grep weather

   # Windows
   Get-Service Weather
   ```

2. **Check logs:**
   ```bash
   # Linux
   tail -f /var/log/weather/access.log

   # macOS
   tail -f ~/Library/Logs/weather/stdout.log

   # Windows
   Get-Content C:\ProgramData\weather\logs\access.log -Tail 50 -Wait
   ```

3. **Test the service:**
   ```bash
   curl http://localhost:3000/healthz
   ```

## Uninstallation

### Linux

```bash
sudo systemctl stop weather
sudo systemctl disable weather
sudo rm /etc/systemd/system/weather.service
sudo systemctl daemon-reload
sudo rm /usr/local/bin/weather
sudo userdel weather
sudo rm -rf /etc/weather /var/lib/weather /var/log/weather /var/cache/weather
```

### macOS

```bash
launchctl unload ~/Library/LaunchAgents/com.apimgr.weather.plist
rm ~/Library/LaunchAgents/com.apimgr.weather.plist
sudo rm /usr/local/bin/weather
rm -rf ~/Library/Application\ Support/weather ~/Library/Logs/weather ~/Library/Caches/weather
```

### Windows

```powershell
# Using NSSM
nssm stop Weather
nssm remove Weather confirm

# Using sc.exe
sc.exe stop Weather
sc.exe delete Weather

# Remove files
Remove-Item "C:\Program Files\Weather" -Recurse -Force
Remove-Item "C:\ProgramData\weather" -Recurse -Force
```

## Troubleshooting

### Service won't start

1. **Check permissions:**
   ```bash
   # Linux
   sudo ls -la /var/lib/weather /var/log/weather /etc/weather

   # macOS
   ls -la ~/Library/Application\ Support/weather
   ```

2. **Run manually to see errors:**
   ```bash
   # Linux
   sudo -u weather /usr/local/bin/weather

   # macOS/Windows
   /usr/local/bin/weather
   ```

3. **Check logs:**
   ```bash
   # Linux
   sudo journalctl -u weather -n 50

   # macOS
   tail -50 ~/Library/Logs/weather/stderr.log
   ```

### Port already in use

Edit the service file to change the port:

```bash
# Linux
sudo systemctl edit weather
# Add: Environment="PORT=8080"

# macOS - edit plist file and add to EnvironmentVariables dict:
# <key>PORT</key>
# <string>8080</string>
```

## Security Considerations

- Service runs as a dedicated non-privileged user (Linux/macOS)
- Limited file system access via systemd security features (Linux)
- Logs rotation prevents disk filling
- SSL certificates stored in protected config directory
- Database access restricted to service user only

## Support

For issues or questions:
- GitHub Issues: https://github.com/apimgr/weather/issues
- Documentation: https://github.com/apimgr/weather/blob/main/README.md

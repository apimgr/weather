# Installing Weather as a Windows Service

There are two methods to install Weather as a Windows service:

## Method 1: Using NSSM (Recommended)

NSSM (Non-Sucking Service Manager) is the easiest way to run applications as Windows services.

### Install NSSM:

```powershell
# Using Chocolatey
choco install nssm

# Or download from https://nssm.cc/download
```

### Install Weather Service:

```powershell
# Install the service
nssm install Weather "C:\Program Files\Weather\weather.exe"

# Set service description
nssm set Weather Description "Weather Service - Beautiful weather forecasts and data"

# Set startup directory
nssm set Weather AppDirectory "C:\ProgramData\weather"

# Set service to auto-start
nssm set Weather Start SERVICE_AUTO_START

# Start the service
nssm start Weather
```

### Manage the Service:

```powershell
# Start service
nssm start Weather

# Stop service
nssm stop Weather

# Restart service
nssm restart Weather

# Check status
nssm status Weather

# View service configuration
nssm dump Weather

# Remove service
nssm remove Weather confirm
```

## Method 2: Using sc.exe (Built-in)

Windows built-in Service Control utility.

### Create the Service:

```powershell
# Run as Administrator
sc.exe create Weather binPath= "C:\Program Files\Weather\weather.exe" DisplayName= "Weather Service" start= auto

# Set description
sc.exe description Weather "Weather Service - Beautiful weather forecasts and data"

# Start the service
sc.exe start Weather
```

### Manage the Service:

```powershell
# Start service
sc.exe start Weather

# Stop service
sc.exe stop Weather

# Query status
sc.exe query Weather

# Delete service
sc.exe delete Weather
```

## Method 3: Using PowerShell (Advanced)

Create a wrapper service using PowerShell.

### Create Service Script:

Create `C:\Program Files\Weather\service-wrapper.ps1`:

```powershell
$ServiceName = "Weather"
$BinaryPath = "C:\Program Files\Weather\weather.exe"
$WorkingDir = "C:\ProgramData\weather"

Set-Location $WorkingDir

# Start the application
& $BinaryPath

# Keep the service alive
while ($true) {
    Start-Sleep -Seconds 30

    # Check if process is still running
    $process = Get-Process weather -ErrorAction SilentlyContinue
    if (-not $process) {
        Write-EventLog -LogName Application -Source $ServiceName -EntryType Error -EventId 1001 -Message "Weather service stopped unexpectedly. Restarting..."
        & $BinaryPath
    }
}
```

### Install with NSSM:

```powershell
nssm install Weather powershell.exe "-ExecutionPolicy Bypass -File \"C:\Program Files\Weather\service-wrapper.ps1\""
```

## Verify Installation

After installing with any method:

```powershell
# Check service status
Get-Service Weather

# View service details
Get-Service Weather | Format-List *

# Check if service is running
Get-Process weather

# View logs (if using NSSM)
nssm set Weather AppStdout "C:\ProgramData\weather\logs\stdout.log"
nssm set Weather AppStderr "C:\ProgramData\weather\logs\stderr.log"
nssm restart Weather
```

## Troubleshooting

### Service won't start:

```powershell
# Check Windows Event Viewer
Get-EventLog -LogName Application -Source Weather -Newest 10

# Run manually to see errors
cd "C:\Program Files\Weather"
.\weather.exe
```

### Change service settings:

```powershell
# Using NSSM
nssm edit Weather

# Using sc.exe
sc.exe config Weather start= demand  # Change to manual start
sc.exe config Weather start= auto    # Change to automatic
```

### View service logs:

```powershell
# Application logs
Get-Content "C:\ProgramData\weather\logs\access.log" -Tail 50 -Wait

# Error logs
Get-Content "C:\ProgramData\weather\logs\error.log" -Tail 50 -Wait
```

## Uninstall Service

```powershell
# Stop the service first
Stop-Service Weather

# Using NSSM
nssm remove Weather confirm

# Using sc.exe
sc.exe delete Weather
```

## Directory Permissions

Ensure the service account has access to:

```
C:\ProgramData\weather\           # Data directory
C:\Users\[User]\AppData\Local\weather\  # Logs/cache (if running as user)
```

## Notes

- The service runs as `Local System` by default with NSSM
- For production, consider running as a dedicated service account
- Logs are written to `C:\ProgramData\weather\logs\` by default
- Configuration is stored in the database at `C:\ProgramData\weather\data\weather.db`

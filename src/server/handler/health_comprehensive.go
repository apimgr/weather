package handlers

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/apimgr/weather/src/database"
	"github.com/apimgr/weather/src/utils"
)

// ComprehensiveHealthCheck handles GET /healthz with full system health data
func ComprehensiveHealthCheck(db *database.DB, httpPort string, httpsPort int, sslManager interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := GetInitStatus()
		startTime := status.Started

		// Read version from release.txt
		version := readVersion()

		// Overall status determination
		overallStatus := "healthy"
		httpStatus := http.StatusOK

		// Database check
		dbStatus, dbLatency, dbErr := db.HealthCheck()
		if dbErr != nil || dbStatus != "connected" {
			overallStatus = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
		}

		// Memory stats
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		memUsedBytes := int64(m.Alloc)
		memTotalBytes := int64(m.Sys)
		memUsedPercent := int(float64(memUsedBytes) / float64(memTotalBytes) * 100)

		if memUsedPercent > 95 {
			overallStatus = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
		} else if memUsedPercent > 80 {
			if overallStatus == "healthy" {
				overallStatus = "degraded"
			}
		}

		memStatus := "ok"
		if memUsedPercent > 95 {
			memStatus = "critical"
		} else if memUsedPercent > 80 {
			memStatus = "warning"
		}

		// Disk usage
		dataDir := getDataDir()
		logDir := getLogDir()

		dataDiskUsage := getDiskUsage(dataDir)
		logDiskUsage := getDiskUsage(logDir)

		if dataDiskUsage.UsedPercent > 95 || logDiskUsage.UsedPercent > 95 {
			overallStatus = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
		} else if dataDiskUsage.UsedPercent > 80 || logDiskUsage.UsedPercent > 80 {
			if overallStatus == "healthy" {
				overallStatus = "degraded"
			}
		}

		diskStatus := "ok"
		if dataDiskUsage.UsedPercent > 95 || logDiskUsage.UsedPercent > 95 {
			diskStatus = "critical"
		} else if dataDiskUsage.UsedPercent > 80 || logDiskUsage.UsedPercent > 80 {
			diskStatus = "warning"
		}

		// Session count
		sessionCount, _ := db.GetSessionCount()

		// SSL status
		sslStatus := getSSLStatus(sslManager)

		// Scheduler status (placeholder)
		schedulerStatus := getSchedulerStatus()

		// Request stats (placeholder - will track in middleware)
		requestStats := getRequestStats()

		// Server info
		serverInfo := getServerInfo(c, httpPort, httpsPort, sslManager)

		// Feature flags
		features := getFeatureFlags(db)

		// Build response according to spec
		response := gin.H{
			"status":         overallStatus,
			"timestamp":      time.Now().Format(time.RFC3339),
			"version":        version,
			"uptime_seconds": int64(time.Since(startTime).Seconds()),
			"checks": gin.H{
				"database": gin.H{
					"status":     dbStatus,
					"type":       "sqlite",
					"latency_ms": dbLatency,
					"connection_pool": gin.H{
						// SQLite doesn't have connection pooling
						"active": 1,
						"idle":   0,
						"max":    1,
					},
				},
				"location_databases": gin.H{
					"countries": gin.H{
						"status": getStatusString(initStatus.Countries),
						"loaded": initStatus.Countries,
					},
					"cities": gin.H{
						"status": getStatusString(initStatus.Cities),
						"loaded": initStatus.Cities,
					},
					"zipcodes": gin.H{
						"status": "loaded",
						"loaded": true,
					},
					"geoip": gin.H{
						"status":    "loaded",
						"loaded":    true,
						"databases": 4,
						"types":     []string{"IPv4 City", "IPv6 City", "Country", "ASN"},
					},
				},
				"cache": gin.H{
					"status":     "inactive",
					"type":       "none",
					"hit_rate":   0.0,
					"size_bytes": 0,
					"entries":    0,
				},
				"disk": gin.H{
					"status": diskStatus,
					"data_dir": gin.H{
						"path":         dataDir,
						"used_bytes":   dataDiskUsage.UsedBytes,
						"free_bytes":   dataDiskUsage.FreeBytes,
						"total_bytes":  dataDiskUsage.TotalBytes,
						"used_percent": dataDiskUsage.UsedPercent,
					},
					"log_dir": gin.H{
						"path":         logDir,
						"used_bytes":   logDiskUsage.UsedBytes,
						"free_bytes":   logDiskUsage.FreeBytes,
						"total_bytes":  logDiskUsage.TotalBytes,
						"used_percent": logDiskUsage.UsedPercent,
					},
				},
				"memory": gin.H{
					"status":       memStatus,
					"used_bytes":   memUsedBytes,
					"total_bytes":  memTotalBytes,
					"used_percent": memUsedPercent,
					"heap_bytes":   int64(m.HeapAlloc),
					"gc_runs":      m.NumGC,
				},
				"ssl":       sslStatus,
				"scheduler": schedulerStatus,
				"sessions": gin.H{
					"active":      sessionCount,
					// Placeholder
					"total_today": sessionCount,
				},
				"requests": requestStats,
			},
			"server":   serverInfo,
			"features": features,
		}

		// TEMPLATE.md: /healthz must return HTML (line 4670)
		// /api/v1/healthz returns JSON (handled by different handler)
		c.HTML(httpStatus, "pages/healthz.tmpl", response)
	}
}

// Helper functions

func readVersion() string {
	data, err := os.ReadFile("release.txt")
	if err != nil {
		return "dev"
	}
	return string(data)
}

// Global variables to store directory paths
var (
	globalDataDir string
	globalLogDir  string
)

// SetDirectoryPaths sets the global directory paths for health checks
func SetDirectoryPaths(dataDir, logDir string) {
	globalDataDir = dataDir
	globalLogDir = logDir
}

func getDataDir() string {
	if globalDataDir != "" {
		return globalDataDir
	}
	dir := os.Getenv("DATA_DIR")
	if dir == "" {
		dir = "./data"
	}
	return dir
}

func getLogDir() string {
	if globalLogDir != "" {
		return globalLogDir
	}
	dir := os.Getenv("LOG_DIR")
	if dir == "" {
		dir = "./logs"
	}
	return dir
}

type DiskUsage struct {
	Path        string
	UsedBytes   int64
	FreeBytes   int64
	TotalBytes  int64
	UsedPercent int
}

// getDiskUsage is implemented in disk_unix.go and disk_windows.go

func getSSLStatus(sslManager interface{}) gin.H {
	// Check if SSL manager is provided and has GetCertInfo method
	if sslManager == nil {
		return gin.H{
			"enabled":        false,
			"status":         "none",
			"expires_at":     nil,
			"days_remaining": 0,
			"issuer":         "Unknown",
		}
	}

	// Use type assertion to get cert info
	type certInfoGetter interface {
		GetCertInfo() map[string]interface{}
	}

	if manager, ok := sslManager.(certInfoGetter); ok {
		info := manager.GetCertInfo()
		return gin.H(info)
	}

	// Fallback if type assertion fails
	return gin.H{
		"enabled":        false,
		"status":         "none",
		"expires_at":     nil,
		"days_remaining": 0,
		"issuer":         "Unknown",
	}
}

func getSchedulerStatus() gin.H {
	// Placeholder - will integrate with actual scheduler
	return gin.H{
		"status":        "running",
		"tasks_total":   6,
		"tasks_enabled": 6,
		"next_run":      time.Now().Add(5 * time.Minute).Format(time.RFC3339),
	}
}

func getRequestStats() gin.H {
	// Placeholder - will track in middleware
	return gin.H{
		"total_today":     0,
		"rate_per_minute": 0,
		"errors_today":    0,
		"error_rate":      0.0,
	}
}

func getServerInfo(c *gin.Context, httpPort string, httpsPort int, sslManager interface{}) gin.H {
	hostInfo := utils.GetHostInfo(c)

	httpsEnabled := false
	if sslManager != nil {
		type httpsChecker interface {
			IsHTTPSEnabled() bool
		}
		if manager, ok := sslManager.(httpsChecker); ok {
			httpsEnabled = manager.IsHTTPSEnabled()
		}
	}

	return gin.H{
		"address":       hostInfo.Hostname,
		"http_port":     httpPort,
		"https_port":    httpsPort,
		"https_enabled": httpsEnabled,
		"pid":           os.Getpid(),
		"started_at":    GetInitStatus().Started.Format(time.RFC3339),
	}
}

func getFeatureFlags(db *database.DB) gin.H {
	// Read from database settings
	regEnabled, _ := db.GetSetting("registration.enabled")
	apiEnabled, _ := db.GetSetting("api.enabled")
	maintenanceMode, _ := db.GetSetting("maintenance.mode")

	return gin.H{
		"registration_enabled": regEnabled != "false",
		"api_enabled":          apiEnabled != "false",
		"maintenance_mode":     maintenanceMode == "true",
		"graphql_enabled":      false,
		"websocket_enabled":    false,
	}
}

func getStatusString(loaded bool) string {
	if loaded {
		return "loaded"
	}
	return "loading"
}

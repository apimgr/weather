package handlers

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/apimgr/weather/src/database"
	"github.com/apimgr/weather/src/utils"
)

var (
	initStatus = &utils.InitializationStatus{
		Countries: false,
		Cities:    false,
		Weather:   false,
		Started:   time.Now(),
	}
	initMutex sync.RWMutex
)

// SetInitStatus updates initialization status
func SetInitStatus(countries, cities, weather bool) {
	initMutex.Lock()
	defer initMutex.Unlock()

	initStatus.Countries = countries
	initStatus.Cities = cities
	initStatus.Weather = weather
}

// IsInitialized checks if all services are initialized
func IsInitialized() bool {
	initMutex.RLock()
	defer initMutex.RUnlock()

	return initStatus.Countries && initStatus.Cities && initStatus.Weather
}

// GetInitStatus returns current initialization status
func GetInitStatus() *utils.InitializationStatus {
	initMutex.RLock()
	defer initMutex.RUnlock()

	return &utils.InitializationStatus{
		Countries: initStatus.Countries,
		Cities:    initStatus.Cities,
		Weather:   initStatus.Weather,
		Started:   initStatus.Started,
	}
}

// HealthCheck handles GET /healthz
func HealthCheck(c *gin.Context) {
	status := GetInitStatus()
	uptime := time.Since(status.Started)

	health := gin.H{
		"status":    "OK",
		"timestamp": utils.Now(),
		"service":   "Weather",
		"version":   "2.0.0-go",
		"uptime":    uptime.String(),
		"ready":     IsInitialized(),
		"initialization": gin.H{
			"countries": status.Countries,
			"cities":    status.Cities,
			"weather":   status.Weather,
		},
	}

	if !IsInitialized() {
		health["status"] = "Initializing"
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	c.JSON(http.StatusOK, health)
}

// ReadinessCheck handles GET /readyz (Kubernetes readiness probe)
func ReadinessCheck(c *gin.Context) {
	if IsInitialized() {
		c.JSON(http.StatusOK, gin.H{
			"ready":     true,
			"timestamp": utils.Now(),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"ready":     false,
			"timestamp": utils.Now(),
			"message":   "Services still initializing",
		})
	}
}

// LivenessCheck handles GET /livez (Kubernetes liveness probe)
func LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"alive":     true,
		"timestamp": utils.Now(),
	})
}

// DebugInfo handles GET /debug/info
func DebugInfo(c *gin.Context) {
	status := GetInitStatus()
	uptime := time.Since(status.Started)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := gin.H{
		"service": gin.H{
			"name":    "Weather",
			"version": "2.0.0-go",
			"uptime":  uptime.String(),
			"started": status.Started.Format(time.RFC3339),
		},
		"initialization": gin.H{
			"ready":     IsInitialized(),
			"countries": status.Countries,
			"cities":    status.Cities,
			"weather":   status.Weather,
		},
		"runtime": gin.H{
			"go_version":    runtime.Version(),
			"num_cpu":       runtime.NumCPU(),
			"num_goroutine": runtime.NumGoroutine(),
		},
		"memory": gin.H{
			"alloc_mb":       fmt.Sprintf("%.2f", float64(m.Alloc)/1024/1024),
			"total_alloc_mb": fmt.Sprintf("%.2f", float64(m.TotalAlloc)/1024/1024),
			"sys_mb":         fmt.Sprintf("%.2f", float64(m.Sys)/1024/1024),
			"num_gc":         m.NumGC,
		},
		"timestamp": utils.Now(),
	}

	c.JSON(http.StatusOK, info)
}

// ServeLoadingPage renders the loading/initialization page
func ServeLoadingPage(c *gin.Context) {
	status := GetInitStatus()
	uptime := time.Since(status.Started)

	// Check if it's an API request (wants JSON)
	if c.GetHeader("Accept") == "application/json" || c.Query("format") == "json" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "Initializing",
			"message": "Services are starting up. Please wait a moment.",
			"initialization": gin.H{
				"countries": status.Countries,
				"cities":    status.Cities,
				"weather":   status.Weather,
			},
			"uptime":    uptime.String(),
			"timestamp": utils.Now(),
		})
		return
	}

	// Check if it's a console client (curl/wget)
	userAgent := c.GetHeader("User-Agent")
	isCurl := contains(userAgent, "curl") || contains(userAgent, "wget") || contains(userAgent, "HTTPie")

	if isCurl {
		// Console-friendly ASCII output
		output := fmt.Sprintf(`🚀 Weather - Starting Up

Services Initialization:
  [%s] Countries Database
  [%s] Cities Database
  [%s] Weather Service

Uptime: %s

⏳ Please wait a moment and try again...

Tip: Check status with:
  curl -q -LSs %s/healthz
`,
			checkmark(status.Countries),
			checkmark(status.Cities),
			checkmark(status.Weather),
			uptime.Round(time.Second).String(),
			utils.GetHostInfo(c).FullHost,
		)

		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusServiceUnavailable, output)
		return
	}

	// Browser gets HTML loading page
	hostInfo := utils.GetHostInfo(c)

	c.HTML(http.StatusServiceUnavailable, "components/loading.tmpl", gin.H{
		"Title":    "Starting Up - Weather",
		"Status":   status,
		"Uptime":   uptime.String(),
		"HostInfo": hostInfo,
	})
}

// Helper functions

func checkmark(ready bool) string {
	if ready {
		return "✓"
	}
	return "⋯"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// APIHealthCheck handles GET /api/v1/healthz - TEMPLATE.md compliant format
func APIHealthCheck(db *database.DB, startTime time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Calculate uptime in human-readable format
		uptime := time.Since(startTime)
		uptimeStr := formatUptime(uptime)

		// Get mode (production/development)
		mode := os.Getenv("MODE")
		if mode == "" {
			mode = "production"
		}

		// Get hostname for node info
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "unknown"
		}

		// Database health check
		dbStatus, _, dbErr := db.HealthCheck()
		dbCheck := "ok"
		if dbErr != nil || dbStatus != "connected" {
			dbCheck = "error"
		}

		// Cache check (currently none/memory)
		cacheCheck := "ok"

		// Disk check (simplified)
		diskCheck := "ok"

		// Overall status
		status := "healthy"
		httpStatus := http.StatusOK
		if dbCheck == "error" {
			status = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
		}

		// Build TEMPLATE.md compliant response
		response := gin.H{
			"status":    status,
			"version":   getVersionFromEnv(),
			"mode":      mode,
			"uptime":    uptimeStr,
			"timestamp": time.Now().Format(time.RFC3339),
			"node": gin.H{
				"id":       hostname,
				"hostname": hostname,
			},
			"cluster": gin.H{
				"enabled": false,
				"status":  "disabled",
			},
			"checks": gin.H{
				"database": dbCheck,
				"cache":    cacheCheck,
				"disk":     diskCheck,
			},
		}

		c.JSON(httpStatus, response)
	}
}

// formatUptime converts duration to human-readable format (e.g., "2d 5h 30m")
func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// getVersionFromEnv gets version from environment or returns "dev"
func getVersionFromEnv() string {
	// Try to read from release.txt
	data, err := os.ReadFile("release.txt")
	if err == nil {
		return string(data)
	}
	return "dev"
}

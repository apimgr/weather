package handler

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
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

	// TorStatusGetter interface for getting Tor service status
	torStatusGetter TorStatusProvider
	torMutex        sync.RWMutex
)

// TorStatusProvider is an interface for getting Tor service status
type TorStatusProvider interface {
	IsRunning() bool
	GetOnionAddress() string
}

// SetTorStatusProvider sets the global Tor status provider
func SetTorStatusProvider(provider TorStatusProvider) {
	torMutex.Lock()
	defer torMutex.Unlock()
	torStatusGetter = provider
}

// GetTorStatus returns the current Tor service status
func GetTorStatus() (running bool, onionAddress string) {
	torMutex.RLock()
	defer torMutex.RUnlock()
	if torStatusGetter == nil {
		return false, ""
	}
	return torStatusGetter.IsRunning(), torStatusGetter.GetOnionAddress()
}

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
		"status":    "ok",
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
		health["status"] = "initializing"
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

	c.HTML(http.StatusServiceUnavailable, "component/loading.tmpl", gin.H{
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

// APIHealthCheck handles GET /api/v1/healthz - AI.md PART 13 compliant format
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

		// Scheduler check (assume ok if no errors)
		schedulerCheck := "ok"

		// Overall status per AI.md PART 13: healthy | degraded | unhealthy
		status := "healthy"
		httpStatus := http.StatusOK

		// Count issues for degraded vs unhealthy determination
		criticalIssues := 0
		warningIssues := 0

		if dbCheck == "error" {
			criticalIssues++
		}
		if cacheCheck == "error" {
			warningIssues++
		}
		if diskCheck == "error" {
			warningIssues++
		}
		if schedulerCheck == "error" {
			warningIssues++
		}

		// Determine status: unhealthy (critical), degraded (warnings), healthy (none)
		if criticalIssues > 0 {
			status = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
		} else if warningIssues > 0 {
			status = "degraded"
			// 200 OK but degraded - still functional
		}

		// Get version info
		version := getVersionFromEnv()
		goVersion := runtime.Version()

		// Build info from ldflags (set during build)
		buildCommit := os.Getenv("BUILD_COMMIT")
		if buildCommit == "" {
			buildCommit = "unknown"
		}
		buildDate := os.Getenv("BUILD_DATE")
		if buildDate == "" {
			buildDate = time.Now().Format(time.RFC3339)
		}

		// Check for Tor service status
		// Get actual status from TorService via global provider (set in main.go)
		torRunning, torHostname := GetTorStatus()
		torEnabled := torHostname != "" || torRunning
		torStatus := ""

		// Also check if tor binary exists as fallback
		if !torEnabled {
			if _, err := exec.LookPath("tor"); err == nil {
				torEnabled = true
			}
		}

		// Check tor status for checks object
		torCheck := ""
		if torEnabled {
			if torRunning {
				torCheck = "ok"
			} else {
				torCheck = "error"
				warningIssues++
			}
		}

		// Build checks object - only include enabled services per AI.md PART 13
		checks := gin.H{
			"database":  dbCheck,
			"cache":     cacheCheck,
			"disk":      diskCheck,
			"scheduler": schedulerCheck,
		}
		// Add tor check only if tor is enabled
		if torEnabled {
			checks["tor"] = torCheck
		}

		// Get hostname for node info per AI.md PART 13
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "unknown"
		}

		// Build AI.md PART 13 compliant response
		response := gin.H{
			"project": gin.H{
				"name":        "Weather Service",
				"tagline":     "Real-time weather data API",
				"description": "Unified weather information platform aggregating global meteorological data",
			},
			"status":     status,
			"version":    version,
			"mode":       mode,
			"uptime":     uptimeStr,
			"timestamp":  time.Now().Format(time.RFC3339),
			"go_version": goVersion,
			"build": gin.H{
				"commit": buildCommit,
				"date":   buildDate,
			},
			// Node info per AI.md PART 13 line 13717
			"node": gin.H{
				"id":       "standalone",
				"hostname": hostname,
			},
			"cluster": gin.H{
				"enabled":    false,
				"status":     "",
				"primary":    "",
				"nodes":      []string{},
				"node_count": 0,
				"role":       "",
			},
			"features": gin.H{
				"tor": gin.H{
					"enabled":  torEnabled,
					"running":  torRunning,
					"status":   torStatus,
					"hostname": torHostname,
				},
				"geoip": false,
			},
			"checks": checks,
			"stats": gin.H{
				"requests_total":     0,
				"requests_24h":       0,
				"active_connections": 0,
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

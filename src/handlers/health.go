package handlers

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"weather-go/src/utils"
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

	c.HTML(http.StatusServiceUnavailable, "loading.tmpl", gin.H{
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

package handler

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NOTE: Manual OpenAPI spec handlers removed per TEMPLATE.md requirements
// TEMPLATE.md Rule: "NEVER manually edit OpenAPI JSON" - specs must be auto-generated only
// All OpenAPI specs are now auto-generated from swag annotations and embedded in binary

// GetSwaggerUIAuto returns the auto-generated Swagger UI using swaggo/gin-swagger
// Serves Swagger UI at /openapi (TEMPLATE.md compliant)
func GetSwaggerUIAuto() gin.HandlerFunc {
	// Custom Dracula theme configuration
	// TEMPLATE.md: Swagger UI must match site theme (Dracula dark)
	config := ginSwagger.Config{
		// Relative URL for the JSON spec
		URL:                      "doc.json",
		DocExpansion:         "list",
		DeepLinking:          true,
		PersistAuthorization: true,
	}

	return ginSwagger.CustomWrapHandler(&config, swaggerFiles.Handler)
}

var startTime = time.Now()

// PrometheusMetrics returns Prometheus-compatible metrics
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		uptime := time.Since(startTime).Seconds()

		// Prometheus text format
		metrics := fmt.Sprintf(`# HELP weather_uptime_seconds Server uptime in seconds
# TYPE weather_uptime_seconds gauge
weather_uptime_seconds %.0f

# HELP weather_memory_alloc_bytes Current allocated memory in bytes
# TYPE weather_memory_alloc_bytes gauge
weather_memory_alloc_bytes %.0f

# HELP weather_memory_sys_bytes Total memory obtained from OS in bytes
# TYPE weather_memory_sys_bytes gauge
weather_memory_sys_bytes %.0f

# HELP weather_goroutines_total Number of goroutines
# TYPE weather_goroutines_total gauge
weather_goroutines_total %.0f

# HELP weather_gc_runs_total Total number of GC runs
# TYPE weather_gc_runs_total counter
weather_gc_runs_total %.0f
`, uptime, float64(m.Alloc), float64(m.Sys), float64(runtime.NumGoroutine()), float64(m.NumGC))

		c.Header("Content-Type", "text/plain; version=0.0.4")
		c.String(http.StatusOK, metrics)
	}
}

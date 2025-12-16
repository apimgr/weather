package handlers

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// GetOpenAPISpec returns the OpenAPI 3.0 specification
func GetOpenAPISpec(c *gin.Context) {
	// Dynamically detect server URL from request
	protocol := c.GetHeader("X-Forwarded-Proto")
	if protocol == "" {
		if c.Request.TLS != nil {
			protocol = "https"
		} else {
			protocol = "http"
		}
	}

	// Get hostname from request
	hostname := c.Request.Host

	// Build server URL
	serverURL := protocol + "://" + hostname

	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Weather API",
			"description": "Production-grade weather API with global forecasts, severe weather alerts, earthquake tracking, and moon phase information",
			"version":     "1.0.0",
			"contact": map[string]string{
				"name": "Weather API Support",
				"url":  "https://github.com/apimgr/weather",
			},
			"license": map[string]string{
				"name": "MIT",
				"url":  "https://github.com/apimgr/weather/blob/main/LICENSE.md",
			},
		},
		"servers": []map[string]string{
			{
				"url":         serverURL,
				"description": "Current server",
			},
		},
		"paths": map[string]interface{}{
			"/api/v1/weather": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get weather forecast",
					"description": "Retrieve weather forecast for a location",
					"tags":        []string{"Weather"},
					"parameters": []map[string]interface{}{
						{
							"name":        "location",
							"in":          "query",
							"description": "Location (city, coordinates, ZIP)",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string"},
							"example":     "Brooklyn, NY",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"location":  map[string]interface{}{"type": "string"},
											"latitude":  map[string]interface{}{"type": "number"},
											"longitude": map[string]interface{}{"type": "number"},
											"current": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"temperature": map[string]interface{}{"type": "number"},
													"humidity":    map[string]interface{}{"type": "number"},
													"wind_speed":  map[string]interface{}{"type": "number"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/severe-weather": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get severe weather alerts",
					"description": "Retrieve severe weather alerts for a location",
					"tags":        []string{"Severe Weather"},
					"parameters": []map[string]interface{}{
						{
							"name":        "location",
							"in":          "query",
							"description": "Location to check alerts",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string"},
						},
						{
							"name":        "distance",
							"in":          "query",
							"description": "Radius in miles",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer"},
							"example":     "50",
						},
						{
							"name":        "type",
							"in":          "query",
							"description": "Alert type filter",
							"required":    false,
							"schema": map[string]interface{}{
								"type": "string",
								"enum": []string{"hurricanes", "tornadoes", "storms", "winter", "floods"},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
						},
					},
				},
			},
			"/api/v1/moon": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get moon phase",
					"description": "Retrieve moon phase information",
					"tags":        []string{"Moon"},
					"parameters": []map[string]interface{}{
						{
							"name":        "location",
							"in":          "query",
							"description": "Location for rise/set times",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
						},
					},
				},
			},
			"/api/v1/earthquakes": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get recent earthquakes",
					"description": "Retrieve recent earthquakes near a location",
					"tags":        []string{"Earthquakes"},
					"parameters": []map[string]interface{}{
						{
							"name":        "location",
							"in":          "query",
							"description": "Center location",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string"},
						},
						{
							"name":        "radius",
							"in":          "query",
							"description": "Search radius in km",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer"},
							"example":     "500",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
						},
					},
				},
			},
			"/healthz": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Health check",
					"description": "Check service health status",
					"tags":        []string{"System"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Service is healthy",
						},
						"503": map[string]interface{}{
							"description": "Service is initializing or unhealthy",
						},
					},
				},
			},
		},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"bearerAuth": map[string]interface{}{
					"type":   "http",
					"scheme": "bearer",
				},
			},
		},
		"tags": []map[string]string{
			{"name": "Weather", "description": "Weather forecast operations"},
			{"name": "Severe Weather", "description": "Severe weather alerts"},
			{"name": "Moon", "description": "Moon phase information"},
			{"name": "Earthquakes", "description": "Earthquake data"},
			{"name": "System", "description": "System health and status"},
		},
	}

	c.JSON(http.StatusOK, spec)
}

// GetSwaggerUI returns the Swagger UI HTML page
func GetSwaggerUI(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Weather API - Swagger UI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
    <style>
        body {
            margin: 0;
            padding: 0;
            background: #282a36;
        }
        .swagger-ui .topbar { display: none; }
        /* Dracula theme for Swagger */
        .swagger-ui { background: #282a36; }
        .swagger-ui .info .title { color: #bd93f9; }
        .swagger-ui .info .description { color: #f8f8f2; }
        .swagger-ui .opblock-tag { color: #8be9fd; border-color: #44475a; background: #44475a; }
        .swagger-ui .opblock { background: #44475a; border-color: #6272a4; }
        .swagger-ui .opblock.opblock-get { border-color: #50fa7b; background: rgba(80, 250, 123, 0.1); }
        .swagger-ui .opblock.opblock-post { border-color: #8be9fd; background: rgba(139, 233, 253, 0.1); }
        .swagger-ui .opblock.opblock-put { border-color: #ffb86c; background: rgba(255, 184, 108, 0.1); }
        .swagger-ui .opblock.opblock-delete { border-color: #ff5555; background: rgba(255, 85, 85, 0.1); }
        .swagger-ui .opblock .opblock-summary-method { background: #bd93f9; color: #282a36; }
        .swagger-ui .opblock .opblock-summary-path { color: #f8f8f2; }
        .swagger-ui .opblock-description-wrapper, .swagger-ui .opblock-body { background: #282a36; color: #f8f8f2; }
        .swagger-ui table thead tr td, .swagger-ui table thead tr th { border-color: #6272a4; color: #bd93f9; }
        .swagger-ui .response-col_status { color: #50fa7b; }
        .swagger-ui .parameter__name { color: #ff79c6; }
        .swagger-ui .response-col_description { color: #f8f8f2; }
        .swagger-ui input[type=text], .swagger-ui textarea, .swagger-ui select { background: #44475a; color: #f8f8f2; border-color: #6272a4; }
        .swagger-ui .btn { background: #bd93f9; color: #282a36; }
        .swagger-ui .btn.execute { background: #50fa7b; color: #282a36; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/api/openapi.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// GetOpenAPISpecYAML returns the OpenAPI 3.0 specification in YAML format
func GetOpenAPISpecYAML(c *gin.Context) {
	// Dynamically detect server URL from request
	protocol := c.GetHeader("X-Forwarded-Proto")
	if protocol == "" {
		if c.Request.TLS != nil {
			protocol = "https"
		} else {
			protocol = "http"
		}
	}

	hostname := c.Request.Host
	serverURL := protocol + "://" + hostname

	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Weather API",
			"description": "Production-grade weather API with global forecasts, severe weather alerts, earthquake tracking, and moon phase information",
			"version":     "1.0.0",
			"contact": map[string]string{
				"name": "Weather API Support",
				"url":  "https://github.com/apimgr/weather",
			},
			"license": map[string]string{
				"name": "MIT",
				"url":  "https://github.com/apimgr/weather/blob/main/LICENSE.md",
			},
		},
		"servers": []map[string]string{
			{
				"url":         serverURL,
				"description": "Current server",
			},
		},
		"paths": map[string]interface{}{
			"/api/v1/weather": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get weather forecast",
					"description": "Retrieve weather forecast for a location",
					"tags":        []string{"Weather"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
						},
					},
				},
			},
			"/healthz": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Health check",
					"description": "Check service health status",
					"tags":        []string{"System"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Service is healthy",
						},
					},
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate YAML"})
		return
	}

	c.Header("Content-Type", "application/x-yaml")
	c.String(http.StatusOK, string(yamlData))
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

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetOpenAPISpec returns the OpenAPI 3.0 specification
func GetOpenAPISpec(c *gin.Context) {
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
				"url":         "http://localhost",
				"description": "Local development server",
			},
			{
				"url":         "https://weather.example.com",
				"description": "Production server",
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
        body { margin: 0; padding: 0; }
        .swagger-ui .topbar { display: none; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/openapi.json",
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

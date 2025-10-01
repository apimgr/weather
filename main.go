package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"weather-go/handlers"
	"weather-go/services"
	"weather-go/utils"
)

func main() {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.New()

	// Trust reverse proxy headers
	r.SetTrustedProxies([]string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})

	// Custom logging middleware (Apache2 combined format)
	r.Use(apacheLoggingMiddleware())

	// Recovery middleware
	r.Use(gin.Recovery())

	// Security headers middleware
	r.Use(securityHeadersMiddleware())

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Path normalization middleware - fix double slashes
	r.Use(func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "//") {
			normalizedPath := strings.ReplaceAll(c.Request.URL.Path, "//", "/")
			query := c.Request.URL.RawQuery
			redirectURL := normalizedPath
			if query != "" {
				redirectURL += "?" + query
			}
			c.Redirect(http.StatusMovedPermanently, redirectURL)
			c.Abort()
			return
		}
		c.Next()
	})

	// Serve static files
	r.Static("/css", "./static/css")
	r.Static("/js", "./static/js")
	r.Static("/images", "./static/images")

	// Load templates
	r.LoadHTMLGlob("templates/*")

	// Initialize services
	log.Println("🚀 Starting Console Weather Service...")
	log.Println("📍 Initializing location databases...")

	locationEnhancer := services.NewLocationEnhancer()

	// Set callback to mark initialization complete
	locationEnhancer.SetOnInitComplete(func(countries, cities bool) {
		// Mark weather service as always ready (no initialization needed)
		handlers.SetInitStatus(countries, cities, true)
		log.Printf("✅ Service ready! Countries: %v, Cities: %v\n", countries, cities)
	})

	weatherService := services.NewWeatherService(locationEnhancer)

	// Data loads automatically in the background via loadData()
	// Mark service as ready after 2 minute initialization timeout (keep as fallback)
	go func() {
		time.Sleep(2 * time.Minute)
		if !handlers.IsInitialized() {
			log.Println("⏰ Initialization timeout reached, marking service as ready (fallback)")
			handlers.SetInitStatus(true, true, true)
		}
	}()

	// Create handlers
	weatherHandler := handlers.NewWeatherHandler(weatherService, locationEnhancer)
	apiHandler := handlers.NewAPIHandler(weatherService, locationEnhancer)
	webHandler := handlers.NewWebHandler(weatherService, locationEnhancer)

	// Health check endpoints (Kubernetes standard)
	r.GET("/healthz", handlers.HealthCheck)
	r.GET("/health", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/healthz")
	})
	r.GET("/readyz", handlers.ReadinessCheck)
	r.GET("/livez", handlers.LivenessCheck)

	// Debug endpoints
	r.GET("/debug/info", handlers.DebugInfo)
	r.GET("/debug/params", func(c *gin.Context) {
		// Parse parameters and return debug info
		c.JSON(http.StatusOK, gin.H{
			"query":  c.Request.URL.Query(),
			"params": "Parameter parsing debug",
		})
	})
	r.GET("/debug/ip", func(c *gin.Context) {
		// IP detection for My Location button
		clientIP := utils.GetClientIP(c)

		// Try to get location from IP
		coords, err := weatherService.GetCoordinatesFromIP(clientIP)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"clientIP": clientIP,
				"location": gin.H{
					"value": "", // Empty means fallback to manual entry
				},
				"error": err.Error(),
			})
			return
		}

		// Enhance location
		enhanced := locationEnhancer.EnhanceLocation(coords)

		c.JSON(http.StatusOK, gin.H{
			"clientIP": clientIP,
			"location": gin.H{
				"value": enhanced.ShortName, // e.g., "Albany, NY"
			},
			"coordinates": gin.H{
				"latitude":  coords.Latitude,
				"longitude": coords.Longitude,
			},
		})
	})

	// API routes - must come before catch-all weather routes
	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/weather", apiHandler.GetWeather)
		apiV1.GET("/weather/:location", apiHandler.GetWeatherByLocation)
		apiV1.GET("/forecast", apiHandler.GetForecast)
		apiV1.GET("/forecast/:location", apiHandler.GetForecastByLocation)
		apiV1.GET("/search", apiHandler.SearchLocations)
		apiV1.GET("/ip", apiHandler.GetIP)
		apiV1.GET("/location", apiHandler.GetLocation)
		apiV1.GET("/docs", apiHandler.GetDocsJSON)

		// Root /api/v1 endpoint - return all endpoints
		apiV1.GET("", func(c *gin.Context) {
			hostInfo := utils.GetHostInfo(c)
			c.JSON(http.StatusOK, gin.H{
				"version": "v1",
				"endpoints": []string{
					hostInfo.FullHost + "/api/v1/weather",
					hostInfo.FullHost + "/api/v1/weather/:location",
					hostInfo.FullHost + "/api/v1/forecast",
					hostInfo.FullHost + "/api/v1/forecast/:location",
					hostInfo.FullHost + "/api/v1/search",
					hostInfo.FullHost + "/api/v1/ip",
					hostInfo.FullHost + "/api/v1/location",
					hostInfo.FullHost + "/api/v1/docs",
				},
				"documentation": hostInfo.FullHost + "/docs",
			})
		})
	}

	// Main /api endpoint - API version information
	r.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "Console Weather Service API",
			"version": "2.0.0",
			"api_versions": []string{
				"v1",
			},
			"current_version": "v1",
			"documentation":   "http://" + c.Request.Host + "/docs",
		})
	})

	// HTML documentation page at /docs
	r.GET("/docs", apiHandler.GetDocsHTML)

	// Examples endpoint
	r.GET("/examples", func(c *gin.Context) {
		hostInfo := utils.GetHostInfo(c)
		examples := fmt.Sprintf(`Weather API Examples

Console Interface:
  curl %s/
  curl %s/London
  curl %s/Paris?format=1
  curl %s/Tokyo?units=metric

JSON API:
  curl %s/api/v1/weather?location=London
  curl %s/api/v1/forecast?location=Paris&days=5
  curl %s/api/v1/search?q=New+York
  curl %s/api/v1/ip
`,
			hostInfo.FullHost, hostInfo.FullHost, hostInfo.FullHost, hostInfo.FullHost,
			hostInfo.FullHost, hostInfo.FullHost, hostInfo.FullHost, hostInfo.FullHost)

		c.String(http.StatusOK, examples)
	})

	// Web interface routes
	r.GET("/web", webHandler.ServeWebInterface)
	r.GET("/web/:location", webHandler.ServeWebInterface)

	// Moon interface routes
	r.GET("/moon", webHandler.ServeMoonInterface)
	r.GET("/moon/:location", webHandler.ServeMoonInterface)

	// Initialization check middleware - show loading page if not ready
	r.Use(func(c *gin.Context) {
		// Skip for health checks, API routes, and static files
		if strings.HasPrefix(c.Request.URL.Path, "/healthz") ||
			strings.HasPrefix(c.Request.URL.Path, "/api") ||
			strings.HasPrefix(c.Request.URL.Path, "/debug") ||
			strings.Contains(c.Request.URL.Path, ".") {
			c.Next()
			return
		}

		// Show loading page if not initialized
		if !handlers.IsInitialized() {
			handlers.ServeLoadingPage(c)
			c.Abort()
			return
		}

		c.Next()
	})

	// Main weather routes (catch-all, must be last)
	r.GET("/", weatherHandler.HandleRoot)
	r.GET("/:location", weatherHandler.HandleLocation)

	// Get port from environment or default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Show startup message
	protocol := "http"
	if os.Getenv("NODE_ENV") == "production" {
		protocol = "https"
	}
	hostname := os.Getenv("HOST")
	if hostname == "" {
		hostname = os.Getenv("HOSTNAME")
	}
	if hostname == "" {
		hostname = "localhost"
	}

	baseURL := fmt.Sprintf("%s://%s", protocol, hostname)
	if (protocol == "http" && port != "80") || (protocol == "https" && port != "443") {
		baseURL += ":" + port
	}

	log.Printf("🌤️  Console Weather Service starting on port %s\n", port)
	log.Printf("📡 API Documentation: %s/api/docs\n", baseURL)
	log.Printf("💡 Examples: %s/examples\n", baseURL)
	log.Printf("🏥 Health Check: %s/healthz\n", baseURL)

	// Start server
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// apacheLoggingMiddleware logs requests in Apache2 combined format
func apacheLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Get client IP
		clientIP := c.ClientIP()

		// Skip logging for localhost/private IPs
		if isLocalIP(clientIP) {
			return
		}

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Get method and path
		method := c.Request.Method
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path += "?" + c.Request.URL.RawQuery
		}

		// Get user agent
		userAgent := c.Request.UserAgent()
		if userAgent == "" {
			userAgent = "-"
		}

		// Get referer
		referer := c.Request.Referer()
		if referer == "" {
			referer = "-"
		}

		// Apache2 combined log format
		// 127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache.gif HTTP/1.0" 200 2326 "http://www.example.com/start.html" "Mozilla/4.08"
		log.Printf("%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %.3fms",
			clientIP,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			method,
			path,
			c.Request.Proto,
			statusCode,
			c.Writer.Size(),
			referer,
			userAgent,
			float64(latency.Microseconds())/1000.0,
		)
	}
}

// isLocalIP checks if an IP is localhost or private
func isLocalIP(ip string) bool {
	localIPs := []string{
		"127.0.0.1",
		"::1",
		"localhost",
		"172.17.0.1", // Docker bridge
		"172.18.0.1",
		"172.19.0.1",
	}

	for _, localIP := range localIPs {
		if ip == localIP {
			return true
		}
	}

	// Check for private IP ranges
	if len(ip) > 3 {
		if strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "192.168.") {
			return true
		}
	}

	return false
}

// securityHeadersMiddleware adds security headers
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'")

		// Other security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

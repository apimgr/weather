package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"weather-go/src/database"
	"weather-go/src/handlers"
	"weather-go/src/middleware"
	"weather-go/src/scheduler"
	"weather-go/src/services"
	"weather-go/src/utils"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

func main() {
	// CLI flags
	var (
		showStatus  = flag.Bool("status", false, "Show server status and configuration")
		showVersion = flag.Bool("version", false, "Show version information")
		healthcheck = flag.Bool("healthcheck", false, "Run healthcheck and exit (for Docker)")
		configPort  = flag.String("port", "", "Override PORT environment variable")
		dataDir     = flag.String("data", "", "Data directory (will store weather.db)")
		configDir   = flag.String("config", "", "Configuration directory")
		configAddr  = flag.String("address", "", "Override server listen address (default: 0.0.0.0)")
	)
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("Weather Service v%s\n", Version)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Go Version: %s\n", strings.TrimPrefix(os.Getenv("GOVERSION"), "go"))
		os.Exit(0)
	}

	// Handle healthcheck flag (for Docker HEALTHCHECK)
	if *healthcheck {
		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}
		resp, err := http.Get(fmt.Sprintf("http://localhost:%s/healthz", port))
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Apply CLI overrides to environment
	if *configPort != "" {
		os.Setenv("PORT", *configPort)
	}
	if *dataDir != "" {
		// Create data directory if it doesn't exist
		if err := os.MkdirAll(*dataDir, 0755); err != nil {
			log.Fatalf("Failed to create data directory %s: %v", *dataDir, err)
		}
		// Set DATABASE_PATH to <dataDir>/weather.db
		dbPath := filepath.Join(*dataDir, "weather.db")
		os.Setenv("DATABASE_PATH", dbPath)
	}
	if *configDir != "" {
		// Create config directory if it doesn't exist
		if err := os.MkdirAll(*configDir, 0755); err != nil {
			log.Fatalf("Failed to create config directory %s: %v", *configDir, err)
		}
		// Future: Load config files from this directory
	}
	if *configAddr != "" {
		os.Setenv("SERVER_ADDRESS", *configAddr)
	}

	// Disable log timestamps
	log.SetFlags(0)

	// Print startup timestamp
	startTime := time.Now()
	fmt.Printf("🕐 %s\n", startTime.Format("2006-01-02 at 15:04:05"))

	// Initialize database
	fmt.Println("💾 Initializing database...")
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./data/weather.db"
	}

	// Ensure data directory exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	fmt.Printf("✅ Database initialized: %s\n", dbPath)

	// Check if this is first run (no users)
	isFirstRun, err := db.IsFirstRun()
	if err != nil {
		log.Printf("⚠️  Warning: Could not check first run status: %v", err)
		isFirstRun = false
	}
	if isFirstRun {
		fmt.Println("🆕 First run detected - please create an admin account at /register")
	}

	// Handle status flag
	if *showStatus {
		showServerStatus(db, dbPath, isFirstRun)
		os.Exit(0)
	}

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
	r.Static("/static/css", "./static/css")
	r.Static("/static/js", "./static/js")
	r.Static("/static/images", "./static/images")

	// PWA files
	r.StaticFile("/manifest.json", "./static/manifest.json")
	r.StaticFile("/sw.js", "./static/js/sw.js")

	// Load templates
	r.LoadHTMLGlob("templates/*.html")

	// Initialize services
	fmt.Println("🚀 Starting Weather...")
	fmt.Println("📍 Initializing location databases...")

	locationEnhancer := services.NewLocationEnhancer()

	// Set callback to mark initialization complete
	locationEnhancer.SetOnInitComplete(func(countries, cities bool) {
		// Mark weather service as always ready (no initialization needed)
		handlers.SetInitStatus(countries, cities, true)
		fmt.Printf("✅ Service ready! Countries: %v, Cities: %v\n", countries, cities)
		// Print ready timestamp
		fmt.Printf("🕐 %s\n", time.Now().Format("2006-01-02 at 15:04:05"))
	})

	weatherService := services.NewWeatherService(locationEnhancer)

	// Data loads automatically in the background via loadData()
	// Mark service as ready after 2 minute initialization timeout (keep as fallback)
	go func() {
		time.Sleep(2 * time.Minute)
		if !handlers.IsInitialized() {
			fmt.Println("⏰ Initialization timeout reached, marking service as ready (fallback)")
			fmt.Printf("🕐 %s\n", time.Now().Format("2006-01-02 at 15:04:05"))
			handlers.SetInitStatus(true, true, true)
		}
	}()

	// Initialize scheduler for periodic tasks
	taskScheduler := scheduler.NewScheduler(db.DB)

	// Register cleanup tasks - run every hour
	taskScheduler.AddTask("cleanup-sessions", 1*time.Hour, func() error {
		return scheduler.CleanupOldSessions(db.DB)
	})

	taskScheduler.AddTask("cleanup-rate-limits", 1*time.Hour, func() error {
		return scheduler.CleanupRateLimitCounters(db.DB)
	})

	taskScheduler.AddTask("cleanup-audit-logs", 24*time.Hour, func() error {
		return scheduler.CleanupOldAuditLogs(db.DB)
	})

	// Register weather alert checks - run every 15 minutes
	taskScheduler.AddTask("check-weather-alerts", 15*time.Minute, func() error {
		return scheduler.CheckWeatherAlerts(db.DB)
	})

	// Register backup task - run every 6 hours
	taskScheduler.AddTask("system-backup", 6*time.Hour, func() error {
		return scheduler.CreateSystemBackup(db.DB)
	})

	// Register weather cache refresh - run every 30 minutes
	taskScheduler.AddTask("refresh-weather-cache", 30*time.Minute, func() error {
		return scheduler.RefreshWeatherCache(db.DB)
	})

	// Start the scheduler
	taskScheduler.Start()

	// Create services
	earthquakeService := services.NewEarthquakeService()
	hurricaneService := services.NewHurricaneService()

	// Create handlers
	weatherHandler := handlers.NewWeatherHandler(weatherService, locationEnhancer)
	apiHandler := handlers.NewAPIHandler(weatherService, locationEnhancer)
	webHandler := handlers.NewWebHandler(weatherService, locationEnhancer)
	earthquakeHandler := handlers.NewEarthquakeHandler(earthquakeService, weatherService, locationEnhancer)
	hurricaneHandler := handlers.NewHurricaneHandler(hurricaneService)

	// Create auth handlers
	authHandler := &handlers.AuthHandler{DB: db.DB}
	dashboardHandler := &handlers.DashboardHandler{DB: db.DB}
	adminHandler := &handlers.AdminHandler{DB: db.DB}
	locationHandler := &handlers.LocationHandler{DB: db.DB}
	notificationHandler := &handlers.NotificationHandler{DB: db.DB}

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

	// Authentication routes (public)
	r.GET("/login", authHandler.ShowLoginPage)
	r.POST("/login", authHandler.HandleLogin)
	r.GET("/register", authHandler.ShowRegisterPage)
	r.POST("/register", authHandler.HandleRegister)
	r.GET("/logout", authHandler.HandleLogout)

	// Protected routes (require authentication)
	r.GET("/dashboard", middleware.RequireAuth(db.DB), dashboardHandler.ShowDashboard)
	r.GET("/admin", middleware.RequireAuth(db.DB), middleware.RequireAdmin(), dashboardHandler.ShowAdminPanel)
	r.GET("/notifications", middleware.RequireAuth(db.DB), notificationHandler.ShowNotificationsPage)

	// Location management pages
	r.GET("/locations/new", middleware.RequireAuth(db.DB), locationHandler.ShowAddLocationPage)
	r.GET("/locations/:id/edit", middleware.RequireAuth(db.DB), locationHandler.ShowEditLocationPage)

	// User info API (requires auth)
	r.GET("/api/user", middleware.RequireAuth(db.DB), authHandler.GetCurrentUser)

	// Location API routes (require auth)
	locationAPI := r.Group("/api/locations")
	locationAPI.Use(middleware.RequireAuth(db.DB))
	{
		locationAPI.GET("", locationHandler.ListLocations)
		locationAPI.GET("/:id", locationHandler.GetLocation)
		locationAPI.POST("", locationHandler.CreateLocation)
		locationAPI.PUT("/:id", locationHandler.UpdateLocation)
		locationAPI.DELETE("/:id", locationHandler.DeleteLocation)
		locationAPI.PUT("/:id/alerts", locationHandler.ToggleAlerts)
	}

	// Notification API routes (require auth)
	notificationAPI := r.Group("/api/notifications")
	notificationAPI.Use(middleware.RequireAuth(db.DB))
	{
		notificationAPI.GET("", notificationHandler.ListNotifications)
		notificationAPI.GET("/unread", notificationHandler.GetUnreadCount)
		notificationAPI.PUT("/:id/read", notificationHandler.MarkAsRead)
		notificationAPI.PUT("/read-all", notificationHandler.MarkAllAsRead)
		notificationAPI.DELETE("/:id", notificationHandler.DeleteNotification)
	}

	// Admin API routes (require admin role)
	adminAPI := r.Group("/api/admin")
	adminAPI.Use(middleware.RequireAuth(db.DB))
	adminAPI.Use(middleware.RequireAdmin())
	{
		// User management
		adminAPI.GET("/users", adminHandler.ListUsers)
		adminAPI.POST("/users", adminHandler.CreateUser)
		adminAPI.PUT("/users/:id", adminHandler.UpdateUser)
		adminAPI.DELETE("/users/:id", adminHandler.DeleteUser)
		adminAPI.PUT("/users/:id/password", adminHandler.UpdateUserPassword)

		// Settings management
		adminAPI.GET("/settings", adminHandler.ListSettings)
		adminAPI.GET("/settings/:key", adminHandler.GetSetting)
		adminAPI.PUT("/settings/:key", adminHandler.UpdateSetting)

		// API token management
		adminAPI.GET("/tokens", adminHandler.ListTokens)
		adminAPI.POST("/tokens", adminHandler.GenerateToken)
		adminAPI.DELETE("/tokens/:id", adminHandler.RevokeToken)

		// Audit logs
		adminAPI.GET("/logs", adminHandler.ListAuditLogs)
		adminAPI.DELETE("/logs", adminHandler.ClearAuditLogs)

		// System stats
		adminAPI.GET("/stats", adminHandler.GetSystemStats)
	}

	// API routes - must come before catch-all weather routes
	// Apply optional auth (to check for API tokens) and rate limiting
	apiV1 := r.Group("/api/v1")
	apiV1.Use(middleware.OptionalAuth(db.DB))
	apiV1.Use(middleware.RateLimitMiddleware(db.DB))
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
			"service": "Weather API",
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

	// Earthquake routes
	r.GET("/earthquake", earthquakeHandler.HandleEarthquakeRequest)
	r.GET("/earthquake/*location", earthquakeHandler.HandleEarthquakeRequest)
	r.GET("/api/earthquakes", earthquakeHandler.HandleEarthquakeAPI)

	// Hurricane routes
	r.GET("/hurricane", hurricaneHandler.HandleHurricaneRequest)
	r.GET("/hurricanes", hurricaneHandler.HandleHurricaneRequest)
	r.GET("/api/hurricanes", hurricaneHandler.HandleHurricaneAPI)

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

	fmt.Printf("🌤️  Weather starting on port %s\n", port)
	fmt.Printf("📡 API Documentation: %s/api/docs\n", baseURL)
	fmt.Printf("💡 Examples: %s/examples\n", baseURL)
	fmt.Printf("🏥 Health Check: %s/healthz\n", baseURL)

	// Create HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Stop scheduler
	taskScheduler.Stop()

	// Shutdown HTTP server with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️  Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
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
		c.Header("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://unpkg.com; script-src 'self' 'unsafe-inline' https://unpkg.com; img-src 'self' data: https: http:; font-src 'self' data:; connect-src 'self' https://unpkg.com https://*.tile.openstreetmap.org")

		// Other security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// showServerStatus displays comprehensive server status information
func showServerStatus(db *database.DB, dbPath string, isFirstRun bool) {
	// Get configuration values
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "release"
	}

	address := os.Getenv("SERVER_ADDRESS")
	if address == "" {
		address = "0.0.0.0"
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	hasSessionSecret := sessionSecret != ""

	// Get database statistics
	var userCount, locationCount, tokenCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM locations").Scan(&locationCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM api_tokens WHERE expires_at > datetime('now')").Scan(&tokenCount)

	// Display status
	fmt.Println("\n╔══════════════════════════════════════════════════════╗")
	fmt.Println("║          🌤️  Weather Service - Status              ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")

	fmt.Println("\n📊 Server Configuration:")
	fmt.Printf("   Version:        %s\n", Version)
	fmt.Printf("   Build Date:     %s\n", BuildDate)
	fmt.Printf("   Git Commit:     %s\n", GitCommit)
	fmt.Printf("   Listen Address: %s:%s\n", address, port)
	fmt.Printf("   Gin Mode:       %s\n", ginMode)

	fmt.Println("\n💾 Database:")
	fmt.Printf("   Path:           %s\n", dbPath)
	fmt.Printf("   Users:          %d\n", userCount)
	fmt.Printf("   Locations:      %d\n", locationCount)
	fmt.Printf("   Active Tokens:  %d\n", tokenCount)
	fmt.Printf("   First Run:      %v\n", isFirstRun)

	fmt.Println("\n🔐 Security:")
	if hasSessionSecret {
		fmt.Println("   Session Secret: ✅ Configured")
	} else {
		fmt.Println("   Session Secret: ⚠️  Using default (not recommended for production)")
	}

	fmt.Println("\n🌐 Endpoints:")
	fmt.Printf("   Web Interface:  http://%s:%s/\n", address, port)
	fmt.Printf("   API Docs:       http://%s:%s/api/docs\n", address, port)
	fmt.Printf("   Health Check:   http://%s:%s/healthz\n", address, port)
	fmt.Printf("   Admin Panel:    http://%s:%s/admin\n", address, port)

	fmt.Println("\n📡 Features:")
	fmt.Println("   ✅ Weather forecasts (Open-Meteo)")
	fmt.Println("   ✅ Moon phases")
	fmt.Println("   ✅ Earthquakes (USGS)")
	fmt.Println("   ✅ Hurricanes (NOAA)")
	fmt.Println("   ✅ Authentication & Sessions")
	fmt.Println("   ✅ Saved Locations")
	fmt.Println("   ✅ Weather Alerts")
	fmt.Println("   ✅ API Tokens")
	fmt.Println("   ✅ PWA Support")
	fmt.Println("   ✅ Rate Limiting")

	fmt.Println("\n💡 CLI Commands:")
	fmt.Println("   --status        Show this status information")
	fmt.Println("   --version       Show version information")
	fmt.Println("   --healthcheck   Run health check (for Docker)")
	fmt.Println("   --port PORT     Override PORT environment variable")
	fmt.Println("   --data DIR      Data directory (will store weather.db)")
	fmt.Println("   --config DIR    Configuration directory")
	fmt.Println("   --address ADDR  Override server listen address")

	fmt.Println("\n" + strings.Repeat("─", 56))
	fmt.Println()
}

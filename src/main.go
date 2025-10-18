package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"weather-go/src/database"
	"weather-go/src/handlers"
	"weather-go/src/middleware"
	"weather-go/src/models"
	"weather-go/src/scheduler"
	"weather-go/src/services"
	"weather-go/src/utils"
)

//go:embed all:templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

// getDefaultListenAddress auto-detects IPv6 support and returns dual-stack (::) or IPv4-only (0.0.0.0)
func getDefaultListenAddress() string {
	// Try to listen on dual-stack IPv6
	listener, err := net.Listen("tcp", "[::]:0")
	if err == nil {
		listener.Close()
		return "::" // IPv6 dual-stack supported (includes IPv4)
	}

	// Fallback to IPv4 only
	return "0.0.0.0"
}

func main() {
	// Check for DEBUG environment variable
	debugEnv := strings.ToLower(os.Getenv("DEBUG"))
	debugMode := debugEnv != "" && debugEnv != "0" && debugEnv != "false" && debugEnv != "no"

	if debugMode {
		log.Println("⚠️  DEBUG MODE ENABLED")
		log.Println("⚠️  This mode should NEVER be used in production!")
	}

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

	// Get OS-appropriate directory paths
	dirPaths, err := utils.GetDirectoryPaths()
	if err != nil {
		log.Fatalf("Failed to determine directory paths: %v", err)
	}

	// Apply environment variable overrides
	if envDataDir := os.Getenv("DATA_DIR"); envDataDir != "" && *dataDir == "" {
		*dataDir = envDataDir
	}
	if envConfigDir := os.Getenv("CONFIG_DIR"); envConfigDir != "" && *configDir == "" {
		*configDir = envConfigDir
	}
	if envLogDir := os.Getenv("LOG_DIR"); envLogDir != "" {
		dirPaths.Log = envLogDir
	}

	// Apply CLI overrides (CLI takes precedence over env vars)
	if *configPort != "" {
		os.Setenv("PORT", *configPort)
	}
	if *dataDir != "" {
		// CLI override for data directory
		if info, err := os.Stat(*dataDir); err == nil {
			if !info.IsDir() {
				if err := os.Remove(*dataDir); err != nil {
					log.Fatalf("Failed to remove file at %s: %v", *dataDir, err)
				}
			}
		}
		if err := os.MkdirAll(*dataDir, 0755); err != nil {
			log.Fatalf("Failed to create data directory %s: %v", *dataDir, err)
		}
		dirPaths.Data = *dataDir
	}
	if *configDir != "" {
		// CLI override for config directory
		if info, err := os.Stat(*configDir); err == nil {
			if !info.IsDir() {
				if err := os.Remove(*configDir); err != nil {
					log.Fatalf("Failed to remove file at %s: %v", *configDir, err)
				}
			}
		}
		if err := os.MkdirAll(*configDir, 0755); err != nil {
			log.Fatalf("Failed to create config directory %s: %v", *configDir, err)
		}
		dirPaths.Config = *configDir
	}
	if *configAddr != "" {
		os.Setenv("SERVER_LISTEN", *configAddr)
	}

	// Create all required directories
	if err := utils.CreateDirectories(dirPaths); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// Initialize logger
	appLogger, err := utils.NewLogger(dirPaths.Log)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Print startup timestamp
	startTime := time.Now()
	appLogger.Printf("🕐 %s", startTime.Format("2006-01-02 at 15:04:05"))

	// Initialize database
	// Priority: 1. Connection string, 2. Individual params (DB_TYPE, DB_HOST...), 3. SQLite path
	dbConnString := os.Getenv("DATABASE_URL")
	if dbConnString == "" {
		dbConnString = os.Getenv("DB_CONNECTION_STRING")
	}

	var dbPath string
	var db *database.DB

	if dbConnString != "" {
		// Use connection string (postgres://user:pass@host/db, mysql://user:pass@host/db, etc.)
		db, err = database.InitDBFromConnectionString(dbConnString)
	} else if dbType := os.Getenv("DB_TYPE"); dbType != "" && dbType != "sqlite" {
		// Use individual database parameters (for PostgreSQL, MySQL, MSSQL)
		config := &database.DatabaseConfig{
			Type:     dbType,
			Host:     os.Getenv("DB_HOST"),
			Database: os.Getenv("DB_NAME"),
			Username: os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			SSLMode:  os.Getenv("DB_SSLMODE"), // PostgreSQL only
		}

		// Parse port
		if portStr := os.Getenv("DB_PORT"); portStr != "" {
			var port int
			fmt.Sscanf(portStr, "%d", &port)
			config.Port = port
		} else {
			// Default ports
			switch dbType {
			case "postgres", "postgresql":
				config.Port = 5432
			case "mysql", "mariadb":
				config.Port = 3306
			case "mssql", "sqlserver":
				config.Port = 1433
			}
		}

		// Set defaults if not provided
		if config.Host == "" {
			config.Host = "localhost"
		}
		if config.Database == "" {
			config.Database = "weather"
		}

		db, err = database.InitDBWithConfig(config)
		dbPath = fmt.Sprintf("%s://%s:%d/%s", dbType, config.Host, config.Port, config.Database)
	} else {
		// Use SQLite file path (default)
		dbPath = os.Getenv("DATABASE_PATH")
		if dbPath == "" {
			dbPath = utils.GetDatabasePath(dirPaths)
		}
		db, err = database.InitDB(dbPath)
	}
	if err != nil {
		appLogger.Fatal("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Check if setup is complete
	var setupComplete bool
	var setupValue string
	err = db.DB.QueryRow("SELECT value FROM settings WHERE key = 'setup.completed'").Scan(&setupValue)
	setupComplete = (err == nil && setupValue == "true")

	if setupComplete {
		appLogger.Printf("✅ Database initialized: %s", dbPath)
	} else {
		appLogger.Printf("✅ Database initialized: %s (setup mode)", dbPath)
	}

	// Initialize default settings with proper backup path
	settingsModel := &models.SettingsModel{DB: db.DB}
	backupPath := utils.GetBackupPath(dirPaths)
	if err := settingsModel.InitializeDefaults(backupPath); err != nil {
		appLogger.Error("⚠️  Warning: Could not initialize default settings: %v", err)
	}

	// Initialize cache manager (Valkey/Redis support, optional)
	cacheManager := services.NewCacheManager()
	if cacheManager.IsEnabled() {
		appLogger.Printf("✅ Cache enabled (Redis/Valkey)")
	}

	// Auto-detect SMTP server at 172.17.0.1 (Docker bridge) and configure defaults
	smtpService := services.NewSMTPService(db.DB)
	if err := smtpService.LoadConfig(); err == nil {
		// Check if SMTP is not already configured
		smtpHost := settingsModel.GetString("smtp.host", "")
		if smtpHost == "" {
			// Try auto-detect
			if detected, _ := smtpService.AutoDetect(); detected {
				// SMTP detected, enable it
				settingsModel.SetBool("smtp.enabled", true)
				appLogger.Printf("✉️  SMTP server auto-detected and enabled")
			}
		}

		// Set default from_address if not set
		fromAddr := settingsModel.GetString("smtp.from_address", "")
		if fromAddr == "" {
			hostname, _ := os.Hostname()
			if hostname == "" {
				hostname = "localhost"
			}
			defaultFromAddr := fmt.Sprintf("no-reply@%s", hostname)
			settingsModel.SetString("smtp.from_address", defaultFromAddr)
		}

		// Set default from_name to server.title if not set
		fromName := settingsModel.GetString("smtp.from_name", "")
		if fromName == "" {
			serverTitle := settingsModel.GetString("server.title", "Weather Service")
			settingsModel.SetString("smtp.from_name", serverTitle)
		}
	}

	// Check if this is first run (no users)
	isFirstRun, err := db.IsFirstRun()
	if err != nil {
		appLogger.Error("⚠️  Warning: Could not check first run status: %v", err)
		isFirstRun = false
	}
	if isFirstRun {
		appLogger.Printf("🆕 First run detected - please create an admin account at /register")
	}

	// Handle status flag
	if *showStatus {
		showServerStatus(db, dbPath, isFirstRun)
		os.Exit(0)
	}

	// Set Gin mode based on ENV variable (development, production, test)
	envMode := os.Getenv("ENV")
	if envMode == "" {
		envMode = os.Getenv("ENVIRONMENT") // Alternative
	}

	switch envMode {
	case "development", "dev":
		gin.SetMode(gin.DebugMode)
	case "test", "testing":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.New()

	// Trust reverse proxy headers
	r.SetTrustedProxies([]string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})

	// Access logging middleware (writes to log files)
	r.Use(middleware.AccessLogger(appLogger))

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

	// Global rate limiting middleware (100 req/s)
	r.Use(middleware.GlobalRateLimitMiddleware())

	// Server context middleware - injects server title/tagline/description
	r.Use(middleware.InjectServerContext(db.DB, Version))

	// Check for first user setup - redirects to /user/setup if no users exist
	r.Use(middleware.CheckFirstUserSetup(db.DB))

	// Restrict admin users to only access /admin routes - all other routes treat them as anonymous
	r.Use(middleware.RestrictAdminToAdminRoutes())

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

	// Serve embedded static files
	staticSubFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Failed to get static subdirectory: %v", err)
	}
	r.StaticFS("/static", http.FS(staticSubFS))

	// Load embedded templates with custom functions
	// Create sub-filesystem starting at "templates/" so template names don't include "templates/" prefix
	templatesSubFS, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		log.Fatalf("Failed to get templates subdirectory: %v", err)
	}

	// Walk the filesystem and collect all .html files
	var templatePaths []string
	fs.WalkDir(templatesSubFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".html") {
			templatePaths = append(templatePaths, path)
		}
		return nil
	})

	// Debug: Print loaded templates
	if gin.Mode() == gin.DebugMode {
		fmt.Printf("📝 Loading %d templates:\n", len(templatePaths))
		for _, path := range templatePaths {
			fmt.Printf("   - %s\n", path)
		}
	}

	// Parse all templates
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}).ParseFS(templatesSubFS, templatePaths...))

	// Debug: Print registered template names
	if gin.Mode() == gin.DebugMode {
		fmt.Println("📋 Registered template names:")
		for _, t := range tmpl.Templates() {
			fmt.Printf("   - %s\n", t.Name())
		}
	}

	r.SetHTMLTemplate(tmpl)

	// Live reload templates in debug mode (loads from filesystem if available)
	if gin.Mode() == gin.DebugMode {
		if _, err := os.Stat("templates"); err == nil {
			r.Use(func(c *gin.Context) {
				// Try to reload from filesystem in debug mode
				t := template.New("").Funcs(template.FuncMap{
					"upper": strings.ToUpper,
					"lower": strings.ToLower,
				})
				// Load all templates including subdirectories
				// Note: This loads from filesystem, so paths are relative to templates/
				patterns := []string{
					"templates/*.html",
					"templates/*/*.html",
					"templates/*/*/*.html",
				}
				for _, pattern := range patterns {
					t, _ = t.ParseGlob(pattern)
				}
				// Need to rename templates to remove "templates/" prefix for consistency
				// This is a bit hacky but necessary for live reload
				r.SetHTMLTemplate(t)
				c.Next()
			})
			fmt.Println("🔄 Live reload enabled for templates (using filesystem)")
		} else {
			fmt.Println("📦 Using embedded templates (no filesystem templates found)")
		}
	} else {
		fmt.Println("📦 Using embedded templates and static files")
	}

	// Initialize location enhancer
	locationEnhancer := services.NewLocationEnhancer(db.DB)

	// Set callback to mark initialization complete
	locationEnhancer.SetOnInitComplete(func(countries, cities bool) {
		// Mark weather service as always ready (no initialization needed)
		handlers.SetInitStatus(countries, cities, true)
		fmt.Printf("✅ Countries: %v, Cities: %v, zipcodes: true, airportcodes: true\n", countries, cities)
	})

	// Initialize GeoIP service (downloads database on first run, updates weekly)
	geoipService := services.NewGeoIPService(dirPaths.Config)

	weatherService := services.NewWeatherService(locationEnhancer, geoipService)

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

	// Initialize notification system services (silent)
	channelManager := services.NewChannelManager(db.DB)
	templateEngine := services.NewTemplateEngine(db.DB)
	deliverySystem := services.NewDeliverySystem(db.DB, channelManager, templateEngine)

	// Load delivery system settings from database
	_ = deliverySystem.LoadSettings()

	// Initialize default templates
	_ = templateEngine.InitializeDefaultTemplates()

	// Initialize channels in database
	_ = channelManager.InitializeChannels()

	// Create weather notification service
	weatherNotifications := services.NewWeatherNotificationService(db.DB, weatherService, deliverySystem, templateEngine)

	// Initialize notification metrics service
	notificationMetrics := services.NewNotificationMetrics(db.DB)

	// Initialize scheduler for periodic tasks
	taskScheduler := scheduler.NewScheduler(db.DB)

	// Register log rotation task - run daily at midnight
	taskScheduler.AddTask("rotate-logs", 24*time.Hour, func() error {
		return appLogger.RotateLogs()
	})

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
		return weatherNotifications.CheckWeatherAlerts()
	})

	// Register daily forecast - run once per day at 7 AM
	taskScheduler.AddTask("daily-forecast", 24*time.Hour, func() error {
		return weatherNotifications.SendDailyForecast()
	})

	// Register notification queue processing - run every 2 minutes
	taskScheduler.AddTask("process-notification-queue", 2*time.Minute, func() error {
		return deliverySystem.ProcessQueue()
	})

	// Register cleanup of old delivered notifications - run daily
	taskScheduler.AddTask("cleanup-notifications", 24*time.Hour, func() error {
		return deliverySystem.CleanupOld(30) // Keep 30 days
	})

	// Register backup task - run every 6 hours
	taskScheduler.AddTask("system-backup", 6*time.Hour, func() error {
		return scheduler.CreateSystemBackup(db.DB)
	})

	// Register weather cache refresh - run every 30 minutes
	taskScheduler.AddTask("refresh-weather-cache", 30*time.Minute, func() error {
		return scheduler.RefreshWeatherCache(db.DB)
	})

	// Register GeoIP database update - run weekly
	taskScheduler.AddTask("update-geoip-database", 7*24*time.Hour, func() error {
		fmt.Println("🌍 Weekly GeoIP database update starting...")
		if err := geoipService.UpdateDatabase(); err != nil {
			fmt.Printf("⚠️ GeoIP update failed: %v\n", err)
			return err
		}
		return nil
	})

	// Start the scheduler
	taskScheduler.Start()

	// Create services
	earthquakeService := services.NewEarthquakeService()
	hurricaneService := services.NewHurricaneService()
	severeWeatherService := services.NewSevereWeatherService()

	// Create handlers
	weatherHandler := handlers.NewWeatherHandler(weatherService, locationEnhancer)
	apiHandler := handlers.NewAPIHandler(weatherService, locationEnhancer)
	webHandler := handlers.NewWebHandler(weatherService, locationEnhancer)
	earthquakeHandler := handlers.NewEarthquakeHandler(earthquakeService, weatherService, locationEnhancer)
	hurricaneHandler := handlers.NewHurricaneHandler(hurricaneService)
	severeWeatherHandler := handlers.NewSevereWeatherHandler(severeWeatherService, locationEnhancer, weatherService)

	// Create auth handlers
	authHandler := &handlers.AuthHandler{DB: db.DB}
	setupHandler := &handlers.SetupHandler{DB: db.DB}
	dashboardHandler := &handlers.DashboardHandler{DB: db.DB}
	adminHandler := &handlers.AdminHandler{DB: db.DB}
	locationHandler := &handlers.LocationHandler{
		DB:               db.DB,
		WeatherService:   weatherService,
		LocationEnhancer: locationEnhancer,
	}
	notificationHandler := &handlers.NotificationHandler{DB: db.DB}

	// Create notification system handlers
	channelHandler := handlers.NewNotificationChannelHandler(db.DB)
	preferencesHandler := handlers.NewNotificationPreferencesHandler(db.DB)
	templateHandler := handlers.NewNotificationTemplateHandler(db.DB)
	metricsHandler := handlers.NewNotificationMetricsHandler(notificationMetrics)

	// Get port configuration using comprehensive port manager
	// Priority: 1) Database saved ports, 2) PORT env variable, 3) Random port (64000-64999)
	portManager := utils.NewPortManager(db.DB)
	httpPortInt, httpsPortInt, err := portManager.GetServerPorts()
	if err != nil {
		log.Fatalf("Failed to configure server ports: %v", err)
	}

	port := fmt.Sprintf("%d", httpPortInt)

	// Get listen address - auto-detect reverse proxy and IPv6 support
	listenAddress := os.Getenv("SERVER_LISTEN")
	if listenAddress == "" {
		listenAddress = os.Getenv("SERVER_ADDRESS") // Backward compatibility
	}
	mode := ""
	if listenAddress == "" {
		// Check for reverse proxy indicator
		reverseProxy := os.Getenv("REVERSE_PROXY") == "true"

		if reverseProxy {
			listenAddress = "127.0.0.1"
			mode = " in reverse proxy mode"
		} else {
			// Auto-detect IPv6 support and use dual-stack if available
			listenAddress = getDefaultListenAddress()
			if listenAddress == "::" {
				mode = " (dual-stack: IPv4 + IPv6)"
			} else {
				mode = " (IPv4 only)"
			}
		}
	}

	// Print startup messages
	appLogger.Printf("🚀 Starting Weather%s on %s:%s", mode, listenAddress, port)
	appLogger.Info("Data directory: %s", dirPaths.Data)
	appLogger.Info("Config directory: %s", dirPaths.Config)
	appLogger.Info("Log directory: %s", dirPaths.Log)

	// Initialize SSL manager
	sslCertsDir := utils.GetCertsPath(dirPaths)
	sslManager := utils.NewSSLManager(db.DB, sslCertsDir)
	httpsPort := httpsPortInt

	// Check for SSL configuration
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	// Try to check for existing Let's Encrypt certs and enable HTTPS if configured
	if httpsPort > 0 {
		found, err := sslManager.CheckExistingCerts(hostname)
		if err != nil {
			appLogger.Error("⚠️  SSL check failed: %v", err)
		} else if found {
			appLogger.Printf("🔒 Found Let's Encrypt certificate for %s", hostname)
			appLogger.Printf("🔌 HTTPS enabled on port: %d", httpsPort)
		} else {
			appLogger.Printf("ℹ️  HTTPS port configured (%d) but no certificates found", httpsPort)
		}
	}
	// Note: Self-signed cert generation is optional and disabled by default
	// Can be enabled via CLI flag or environment variable if needed

	// Set directory paths for handlers
	handlers.SetDirectoryPaths(dirPaths.Data, dirPaths.Log)

	// Health check endpoints (Kubernetes standard)
	r.GET("/healthz", handlers.ComprehensiveHealthCheck(db, port, httpsPort, sslManager))
	r.GET("/health", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/healthz")
	})
	r.GET("/readyz", handlers.ReadinessCheck)
	r.GET("/livez", handlers.LivenessCheck)
	r.GET("/healthz/setup", setupHandler.GetSetupStatus)

	// Debug endpoints (only enabled when DEBUG environment variable is set)
	if debugMode {
		debugHandlers := handlers.NewDebugHandlers(db.DB, r)
		debugHandlers.RegisterDebugRoutes(r)

		log.Println("🔧 Debug endpoints enabled:")
		log.Println("   GET  /debug/routes  - List all routes")
		log.Println("   GET  /debug/config  - Show configuration")
		log.Println("   GET  /debug/memory  - Memory statistics")
		log.Println("   GET  /debug/db      - Database statistics")
		log.Println("   POST /debug/reload  - Reload configuration")
		log.Println("   POST /debug/gc      - Trigger garbage collection")
	}

	// IP detection endpoint (always available for My Location feature)
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

	// First-run setup routes (public)
	// First user setup routes (blocked after setup complete)
	setupRoutes := r.Group("/user/setup")
	setupRoutes.Use(middleware.BlockSetupAfterComplete(db.DB))
	{
		setupRoutes.GET("", setupHandler.ShowWelcome)
		setupRoutes.GET("/register", setupHandler.ShowUserRegister)
		setupRoutes.POST("/register", setupHandler.CreateUser)
		setupRoutes.GET("/admin", setupHandler.ShowAdminSetup)
		setupRoutes.POST("/admin", setupHandler.CreateAdmin)
		setupRoutes.GET("/complete", setupHandler.CompleteSetup)
	}

	// Setup wizard routes
	setupWizard := r.Group("/setup")
	setupWizard.Use(middleware.RequireAuth(db.DB))
	{
		// Admin account creation (step 1 - accessible only when no admin exists)
		adminSetup := setupWizard.Group("/admin")
		adminSetup.Use(middleware.BlockSetupAfterAdminExists(db.DB))
		{
			adminSetup.GET("/welcome", setupHandler.ShowAdminSetup)
			adminSetup.POST("/create", setupHandler.CreateAdmin)
		}

		// Server configuration wizard (step 2 - admin only, after admin account created)
		serverSetup := setupWizard.Group("/server")
		serverSetup.Use(middleware.RequireAdmin())
		serverSetup.Use(middleware.BlockSetupAfterComplete(db.DB))
		{
			serverSetup.GET("/welcome", setupHandler.ShowServerSetupWelcome)
			serverSetup.GET("/settings", setupHandler.ShowServerSetupSettings)
			serverSetup.POST("/settings", setupHandler.SaveServerSettings)
		}

		// Setup completion page (admin only)
		setupWizard.GET("/complete", middleware.RequireAdmin(), setupHandler.ShowServerSetupComplete)
	}

	// Authentication routes (public)
	r.GET("/login", authHandler.ShowLoginPage)
	r.POST("/login", authHandler.HandleLogin)
	r.GET("/register", authHandler.ShowRegisterPage)
	r.POST("/register", authHandler.HandleRegister)
	r.GET("/logout", authHandler.HandleLogout)

	// User routes (require authentication)
	userRoutes := r.Group("/user")
	userRoutes.Use(middleware.RequireAuth(db.DB))
	userRoutes.Use(middleware.BlockAdminFromUserRoutes())
	{
		userRoutes.GET("", dashboardHandler.ShowDashboard)          // /user -> user dashboard
		userRoutes.GET("/dashboard", dashboardHandler.ShowDashboard) // /user/dashboard -> user dashboard
	}

	// Admin routes (require admin role + stricter rate limiting)
	adminRoutes := r.Group("/admin")
	adminRoutes.Use(middleware.RequireAuth(db.DB))
	adminRoutes.Use(middleware.RequireAdmin())
	adminRoutes.Use(middleware.AdminRateLimitMiddleware())
	adminRoutes.Use(middleware.AuditLogger(db.DB)) // Log all admin actions
	{
		adminRoutes.GET("", dashboardHandler.ShowAdminPanel) // /admin -> admin dashboard

		// Admin management pages
		adminRoutes.GET("/users", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/users.html", utils.TemplateData(c, gin.H{
				"title":       "User Management - Admin",
				"page":        "users",
				"breadcrumb": "Users",
			}))
		})

		adminRoutes.GET("/settings", adminHandler.ShowSettingsPage)

		adminRoutes.GET("/tokens", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/tokens.html", utils.TemplateData(c, gin.H{
				"title":       "API Tokens - Admin",
				"page":        "tokens",
				"breadcrumb": "API Tokens",
			}))
		})

		adminRoutes.GET("/logs", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/logs.html", utils.TemplateData(c, gin.H{
				"title":       "Audit Logs - Admin",
				"page":        "logs",
				"breadcrumb": "Audit Logs",
			}))
		})

		adminRoutes.GET("/tasks", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/tasks.html", utils.TemplateData(c, gin.H{
				"title":       "Scheduled Tasks - Admin",
				"page":        "tasks",
				"breadcrumb": "Scheduled Tasks",
			}))
		})

		adminRoutes.GET("/backup", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/backup.html", utils.TemplateData(c, gin.H{
				"title":       "Backup Management - Admin",
				"page":        "backup",
				"breadcrumb": "Backup",
			}))
		})

		adminRoutes.GET("/channels", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin_channels.html", utils.TemplateData(c, gin.H{
				"title":       "Notification Channels - Admin",
				"page":        "channels",
				"breadcrumb": "Channels",
			}))
		})

		adminRoutes.GET("/templates", func(c *gin.Context) {
			c.HTML(http.StatusOK, "template_editor.html", utils.TemplateData(c, gin.H{
				"title":       "Template Editor - Admin",
				"page":        "templates",
				"breadcrumb": "Templates",
			}))
		})
	}
	r.GET("/notifications", middleware.RequireAuth(db.DB), notificationHandler.ShowNotificationsPage)

	// User profile page
	r.GET("/profile", middleware.RequireAuth(db.DB), func(c *gin.Context) {
		c.HTML(http.StatusOK, "user/profile.html", utils.TemplateData(c, gin.H{
			"title": "Profile",
			"page":  "profile",
		}))
	})
	r.GET("/user/profile", middleware.RequireAuth(db.DB), middleware.BlockAdminFromUserRoutes(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "user/profile.html", utils.TemplateData(c, gin.H{
			"title": "Profile",
			"page":  "profile",
		}))
	})

	// User notification preferences page
	r.GET("/preferences", middleware.RequireAuth(db.DB), func(c *gin.Context) {
		c.HTML(http.StatusOK, "user_preferences.html", utils.TemplateData(c, gin.H{
			"title": "Preferences",
			"page":  "preferences",
		}))
	})
	r.GET("/user/preferences", middleware.RequireAuth(db.DB), middleware.BlockAdminFromUserRoutes(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "user_preferences.html", utils.TemplateData(c, gin.H{
			"title": "Preferences",
			"page":  "preferences",
		}))
	})

	// Removed - moved to adminRoutes group above

	// Location management pages
	r.GET("/locations/new", middleware.RequireAuth(db.DB), locationHandler.ShowAddLocationPage)
	r.GET("/locations/:id/edit", middleware.RequireAuth(db.DB), locationHandler.ShowEditLocationPage)

	// API v1 routes - all API endpoints under /api/v1
	apiV1 := r.Group("/api/v1")

	// Weather API routes (optional auth + API rate limiting)
	weatherAPI := apiV1.Group("")
	weatherAPI.Use(middleware.OptionalAuth(db.DB))
	weatherAPI.Use(middleware.APIRateLimitMiddleware())
	{
		weatherAPI.GET("/weather", apiHandler.GetWeather)
		weatherAPI.GET("/weather/:location", apiHandler.GetWeatherByLocation)
		weatherAPI.GET("/forecast", apiHandler.GetForecast)
		weatherAPI.GET("/forecast/:location", apiHandler.GetForecastByLocation)
		weatherAPI.GET("/search", apiHandler.SearchLocations)
		weatherAPI.GET("/ip", apiHandler.GetIP)
		weatherAPI.GET("/location", apiHandler.GetLocation)
		weatherAPI.GET("/docs", apiHandler.GetDocsJSON)
		weatherAPI.GET("/earthquakes", earthquakeHandler.HandleEarthquakeAPI)
		weatherAPI.GET("/hurricanes", hurricaneHandler.HandleHurricaneAPI) // Backwards compat
		weatherAPI.GET("/severe-weather", severeWeatherHandler.HandleSevereWeatherAPI)

		// Root /api/v1 endpoint - return all endpoints
		weatherAPI.GET("", func(c *gin.Context) {
			hostInfo := utils.GetHostInfo(c)
			c.JSON(http.StatusOK, gin.H{
				"version": "v1",
				"endpoints": []string{
					hostInfo.FullHost + "/api/v1/user",
					hostInfo.FullHost + "/api/v1/locations",
					hostInfo.FullHost + "/api/v1/notifications",
					hostInfo.FullHost + "/api/v1/admin",
					hostInfo.FullHost + "/api/v1/weather",
					hostInfo.FullHost + "/api/v1/weather/:location",
					hostInfo.FullHost + "/api/v1/forecast",
					hostInfo.FullHost + "/api/v1/forecast/:location",
					hostInfo.FullHost + "/api/v1/search",
					hostInfo.FullHost + "/api/v1/ip",
					hostInfo.FullHost + "/api/v1/location",
					hostInfo.FullHost + "/api/v1/docs",
					hostInfo.FullHost + "/api/v1/blocklist",
					hostInfo.FullHost + "/api/v1/earthquakes",
					hostInfo.FullHost + "/api/v1/hurricanes",
					hostInfo.FullHost + "/api/v1/severe-weather",
				},
				"documentation": hostInfo.FullHost + "/docs",
			})
		})
	}

	// Public blocklist endpoint (no auth required)
	apiV1.GET("/blocklist", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"blocklist": utils.UsernameBlocklist,
			"count":     utils.GetBlocklistSize(),
			"public":    utils.IsBlocklistPublic(),
			"note":      "These usernames are reserved and cannot be used for registration. The blocklist does not apply to the first user (admin setup).",
		})
	})

	// User info API (requires auth)
	apiV1.GET("/user", middleware.RequireAuth(db.DB), middleware.BlockAdminFromUserRoutes(), authHandler.GetCurrentUser)
	apiV1.PUT("/user/profile", middleware.RequireAuth(db.DB), middleware.BlockAdminFromUserRoutes(), authHandler.UpdateProfile)

	// Location API routes (require auth)
	// Public location endpoints (no auth required)
	apiV1.GET("/locations/search", locationHandler.SearchLocations)
	apiV1.GET("/locations/lookup/zip/:code", locationHandler.LookupZipCode)
	apiV1.GET("/locations/lookup/coords", locationHandler.LookupCoordinates)

	// Protected location endpoints (require auth)
	locationAPI := apiV1.Group("/locations")
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
	notificationAPI := apiV1.Group("/notifications")
	notificationAPI.Use(middleware.RequireAuth(db.DB))
	{
		notificationAPI.GET("", notificationHandler.ListNotifications)
		notificationAPI.GET("/unread", notificationHandler.GetUnreadCount)
		notificationAPI.PUT("/:id/read", notificationHandler.MarkAsRead)
		notificationAPI.PUT("/read-all", notificationHandler.MarkAllAsRead)
		notificationAPI.DELETE("/:id", notificationHandler.DeleteNotification)
	}

	// Admin API routes (require admin role + stricter rate limiting)
	adminAPI := apiV1.Group("/admin")
	adminAPI.Use(middleware.RequireAuth(db.DB))
	adminAPI.Use(middleware.RequireAdmin())
	adminAPI.Use(middleware.AdminRateLimitMiddleware())
	adminAPI.Use(middleware.AuditLogger(db.DB)) // Log all admin API actions
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

		// Scheduled tasks
		adminAPI.GET("/tasks", adminHandler.GetScheduledTasks)

		// Notification channel management (admin only)
		adminAPI.GET("/channels", channelHandler.ListChannels)
		adminAPI.GET("/channels/definitions", channelHandler.GetChannelDefinitions)
		adminAPI.GET("/channels/queue/stats", channelHandler.GetQueueStats)
		adminAPI.GET("/channels/history", channelHandler.GetNotificationHistory)
		adminAPI.POST("/channels/initialize", channelHandler.InitializeChannels)
		adminAPI.GET("/channels/:type", channelHandler.GetChannel)
		adminAPI.PUT("/channels/:type", channelHandler.UpdateChannel)
		adminAPI.POST("/channels/:type/enable", channelHandler.EnableChannel)
		adminAPI.POST("/channels/:type/disable", channelHandler.DisableChannel)
		adminAPI.POST("/channels/:type/test", channelHandler.TestChannel)
		adminAPI.GET("/channels/:type/stats", channelHandler.GetChannelStats)

		// SMTP provider management
		adminAPI.GET("/smtp/providers", channelHandler.ListSMTPProviders)
		adminAPI.POST("/smtp/autodetect", channelHandler.AutoDetectSMTP)

		// Template management (admin only)
		adminAPI.GET("/templates", templateHandler.ListTemplates)
		adminAPI.GET("/templates/variables", templateHandler.GetTemplateVariables)
		adminAPI.POST("/templates/preview", templateHandler.PreviewTemplate)
		adminAPI.POST("/templates/initialize", templateHandler.InitializeDefaults)
		adminAPI.GET("/templates/:id", templateHandler.GetTemplate)
		adminAPI.POST("/templates", templateHandler.CreateTemplate)
		adminAPI.PUT("/templates/:id", templateHandler.UpdateTemplate)
		adminAPI.DELETE("/templates/:id", templateHandler.DeleteTemplate)
		adminAPI.POST("/templates/:id/clone", templateHandler.CloneTemplate)

		// Notification metrics management (admin only)
		adminAPI.GET("/metrics/notifications/summary", metricsHandler.GetSummary)
		adminAPI.GET("/metrics/notifications/channels/:type", metricsHandler.GetChannelMetrics)
		adminAPI.GET("/metrics/notifications/errors", metricsHandler.GetRecentErrors)
		adminAPI.GET("/metrics/notifications/health", metricsHandler.GetHealthStatus)
	}

	// User notification preferences API (authenticated users)
	userPrefAPI := r.Group("/api/user")
	userPrefAPI.Use(middleware.RequireAuth(db.DB))
	{
		// Channel preferences
		userPrefAPI.GET("/preferences", preferencesHandler.GetUserPreferences)
		userPrefAPI.PUT("/preferences/:id", preferencesHandler.UpdatePreference)
		userPrefAPI.POST("/preferences", preferencesHandler.CreatePreference)
		userPrefAPI.DELETE("/preferences/:id", preferencesHandler.DeletePreference)

		// Subscriptions
		userPrefAPI.GET("/subscriptions", preferencesHandler.GetSubscriptions)
		userPrefAPI.PUT("/subscriptions/:id", preferencesHandler.UpdateSubscription)
		userPrefAPI.POST("/subscriptions", preferencesHandler.CreateSubscription)
	}

	// API routes are now consolidated under /api/v1 above

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
	r.GET("/earthquake/:location", earthquakeHandler.HandleEarthquakeRequest)

	// Hurricane routes (keep for backwards compatibility, redirect to severe-weather)
	r.GET("/hurricane", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/severe-weather")
	})
	r.GET("/hurricanes", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/severe-weather")
	})

	// Severe Weather routes (new comprehensive severe weather page)
	r.GET("/severe-weather", severeWeatherHandler.HandleSevereWeatherRequest)
	r.GET("/severe-weather/:location", severeWeatherHandler.HandleSevereWeatherRequest)

	// Type-filtered severe weather routes
	r.GET("/severe/:type", severeWeatherHandler.HandleSevereWeatherByType)
	r.GET("/severe/:type/:location", severeWeatherHandler.HandleSevereWeatherByType)

	// Backwards compatibility redirects for old API routes
	r.GET("/api/earthquakes", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/v1/earthquakes?"+c.Request.URL.RawQuery)
	})
	r.GET("/api/hurricanes", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/v1/hurricanes?"+c.Request.URL.RawQuery)
	})

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

	// Main weather routes
	r.GET("/", weatherHandler.HandleRoot)                      // Uses IP/cookie lookup
	r.GET("/weather/:location", weatherHandler.HandleLocation) // Explicit location
	r.GET("/:location", weatherHandler.HandleLocation)         // Backwards compatibility catch-all

	// Build final URL for documentation
	finalHostname := os.Getenv("DOMAIN")
	if finalHostname == "" {
		finalHostname = os.Getenv("HOSTNAME")
	}
	if finalHostname == "" {
		finalHostname = "localhost"
	}

	protocol := "http"
	if os.Getenv("TLS_ENABLED") == "true" || httpsPortInt > 0 {
		protocol = "https"
	}

	finalURL := fmt.Sprintf("%s://%s", protocol, finalHostname)
	if (protocol == "http" && port != "80") || (protocol == "https" && port != "443") {
		finalURL += ":" + port
	}

	// Print final startup messages
	fmt.Printf("📡 For documentation see: %s/docs\n", finalURL)
	fmt.Printf("🕐 Ready: %s: %s\n", time.Now().Format("2006-01-02 at 15:04:05"), finalURL)

	// Create HTTP server with graceful shutdown
	// Format address properly for IPv6
	serverAddr := listenAddress + ":" + port
	if listenAddress == "::" {
		serverAddr = "[" + listenAddress + "]:" + port
	}

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGTERM, // Graceful shutdown (systemctl stop)
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGHUP,  // Reload config
		syscall.SIGUSR1, // Reopen log files
		syscall.SIGUSR2, // Toggle debug mode
	)

	// Handle signals
	for sig := range sigChan {
		switch sig {
		case syscall.SIGTERM, syscall.SIGINT:
			log.Println("🛑 Received shutdown signal, shutting down gracefully...")

			// Stop scheduler
			taskScheduler.Stop()

			// Close cache connection
			if err := cacheManager.Close(); err != nil {
				log.Printf("⚠️  Cache shutdown error: %v", err)
			}

			// Shutdown HTTP server with 5 second timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("⚠️  Server forced to shutdown: %v", err)
			}

			log.Println("✅ Server exited gracefully")
			return

		case syscall.SIGHUP:
			log.Println("🔄 Received SIGHUP, reloading configuration...")
			// Reload settings from database
			settingsModel := &models.SettingsModel{DB: db.DB}
			if err := settingsModel.InitializeDefaults(utils.GetBackupPath(dirPaths)); err != nil {
				log.Printf("⚠️  Failed to reload settings: %v", err)
			} else {
				log.Println("✅ Configuration reloaded")
			}

		case syscall.SIGUSR1:
			log.Println("📝 Received SIGUSR1, reopening log files...")
			// Rotate logs
			if err := appLogger.RotateLogs(); err != nil {
				log.Printf("⚠️  Failed to rotate logs: %v", err)
			} else {
				log.Println("✅ Log files reopened")
			}

		case syscall.SIGUSR2:
			log.Println("🔧 Received SIGUSR2, toggling debug mode...")
			// Toggle Gin mode between debug and release
			if gin.Mode() == gin.DebugMode {
				gin.SetMode(gin.ReleaseMode)
				log.Println("✅ Debug mode: OFF (release mode)")
			} else {
				gin.SetMode(gin.DebugMode)
				log.Println("✅ Debug mode: ON (debug mode)")
			}
		}
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

// isLocalIP checks if an IP is localhost or private (supports IPv4 and IPv6)
func isLocalIP(ip string) bool {
	// Parse IP address properly
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		// If we can't parse it, assume local for safety
		return true
	}

	// Check if loopback (127.0.0.1 or ::1)
	if parsedIP.IsLoopback() {
		return true
	}

	// Check if private (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, fc00::/7, fe80::/10)
	if parsedIP.IsPrivate() {
		return true
	}

	// Check for link-local IPv6 (fe80::/10)
	if parsedIP.IsLinkLocalUnicast() {
		return true
	}

	// Check for unique local IPv6 (fc00::/7)
	if len(parsedIP) == 16 && (parsedIP[0]&0xfe) == 0xfc {
		return true
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

	envMode := os.Getenv("ENV")
	if envMode == "" {
		envMode = os.Getenv("ENVIRONMENT")
	}
	if envMode == "" {
		envMode = "production"
	}

	address := os.Getenv("SERVER_LISTEN")
	if address == "" {
		address = os.Getenv("SERVER_ADDRESS") // Backward compatibility
	}
	addressMode := ""
	if address == "" {
		// Check for reverse proxy indicators
		reverseProxy := os.Getenv("REVERSE_PROXY") == "true"

		if reverseProxy {
			address = "127.0.0.1"
			addressMode = " (reverse proxy mode)"
		} else {
			address = "::"
			addressMode = " (all interfaces)"
		}
	}

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
	fmt.Printf("   Listen Address: %s:%s%s\n", address, port, addressMode)
	fmt.Printf("   Environment:    %s\n", envMode)

	fmt.Println("\n💾 Database:")
	fmt.Printf("   Path:           %s\n", dbPath)
	fmt.Printf("   Users:          %d\n", userCount)
	fmt.Printf("   Locations:      %d\n", locationCount)
	fmt.Printf("   Active Tokens:  %d\n", tokenCount)
	fmt.Printf("   First Run:      %v\n", isFirstRun)

	fmt.Println("\n🔐 Security:")
	fmt.Println("   Session Secret: ✅ Configured")

	fmt.Println("\n🌐 Endpoints:")
	fmt.Printf("   Web Interface:  http://%s:%s/\n", address, port)
	fmt.Printf("   API Docs:       http://%s:%s/docs\n", address, port)
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

	fmt.Println("\n🌐 Network Configuration:")
	fmt.Println("   Default:        :: (all interfaces, IPv4 + IPv6)")
	fmt.Println("   Reverse Proxy:  127.0.0.1 (set REVERSE_PROXY=true)")
	fmt.Println("   Custom:         Set SERVER_LISTEN environment variable")

	fmt.Println("\n" + strings.Repeat("─", 56))
	fmt.Println()
}

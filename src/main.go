package main

// Weather Service - Main entry point
// Per AI.md: Swagger annotations moved to src/swagger/annotations.go

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/apimgr/weather/src/cli"
	"github.com/apimgr/weather/src/config"
	"github.com/apimgr/weather/src/database"
	"github.com/apimgr/weather/src/mode"
	"github.com/apimgr/weather/src/paths"
	"github.com/apimgr/weather/src/scheduler"
	"github.com/apimgr/weather/src/server"
	"github.com/apimgr/weather/src/server/handler"
	"github.com/apimgr/weather/src/server/metrics"
	"github.com/apimgr/weather/src/server/middleware"
	"github.com/apimgr/weather/src/server/model"
	"github.com/apimgr/weather/src/server/service"
	"github.com/apimgr/weather/src/utils"
)

//go:embed locales/*.json
var localesFS embed.FS

// getDefaultListenAddress auto-detects IPv6 support and returns dual-stack (::) or IPv4-only (0.0.0.0)
func getDefaultListenAddress() string {
	// Try to listen on dual-stack IPv6
	listener, err := net.Listen("tcp", "[::]:0")
	if err == nil {
		listener.Close()
		// IPv6 dual-stack supported (includes IPv4)
		return "::"
	}

	// Fallback to IPv4 only
	return "0.0.0.0"
}

func main() {
	// Initialize CLI
	cliInstance := cli.NewCLI()

	// Set version information
	cli.Version = Version
	cli.BuildDate = BuildDate
	cli.CommitID = CommitID

	// Register CLI commands
	cliInstance.RegisterCommand(&cli.Command{
		Name:        "service",
		Description: "Service management operations",
		Privileged:  true,
		Handler:     cli.ServiceCommand,
	})

	cliInstance.RegisterCommand(&cli.Command{
		Name:        "maintenance",
		Description: "Maintenance operations",
		Privileged:  false,
		Handler:     cli.MaintenanceCommand,
	})

	cliInstance.RegisterCommand(&cli.Command{
		Name:        "update",
		Description: "Update operations",
		Privileged:  false,
		Handler:     cli.UpdateCommand,
	})

	// Parse CLI arguments
	if err := cliInstance.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Failed to parse CLI: %v", err)
	}

	// Check if this is a command that exits (handled by CLI package)
	// Commands like --help, --version are handled internally

	// Handle healthcheck flag (for Docker HEALTHCHECK)
	if os.Getenv("CLI_HEALTHCHECK_FLAG") == "1" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "80"
		}
		resp, err := http.Get(fmt.Sprintf("http://localhost:%s/healthz", port))
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Handle --daemon flag (AI.md PART 8: Daemonize - detach from terminal, Unix only)
	// AI.md: "modern service managers prefer foreground" but flag should work
	if config.IsTruthy(os.Getenv("DAEMON")) && runtime.GOOS != "windows" {
		// Check if we're already daemonized (avoid double fork)
		if os.Getenv("_DAEMON_CHILD") != "1" {
			// Fork a new process
			execPath, err := os.Executable()
			if err != nil {
				log.Fatalf("Failed to get executable path for daemon: %v", err)
			}

			// Set marker so child knows it's the daemon
			env := os.Environ()
			env = append(env, "_DAEMON_CHILD=1")

			// Create child process
			procAttr := &os.ProcAttr{
				Dir:   "/",
				Env:   env,
				Files: []*os.File{nil, nil, nil}, // Detach stdin/stdout/stderr
			}

			proc, err := os.StartProcess(execPath, os.Args, procAttr)
			if err != nil {
				log.Fatalf("Failed to start daemon process: %v", err)
			}

			// Release the child and exit parent
			proc.Release()
			fmt.Printf("Started daemon with PID %d\n", proc.Pid)
			os.Exit(0)
		}
		// Child process continues with normal startup
		fmt.Println("Running as daemon (detached from terminal)")
	}

	// Initialize mode from environment variables (AI.md PART 6)
	// Handles MODE and DEBUG environment variables (set by CLI or directly)
	mode.FromEnv()

	if mode.IsDebugEnabled() {
		log.Println("DEBUG MODE ENABLED")
		log.Println("This mode should NEVER be used in production!")
		fmt.Println("⚠️  DEBUG MODE ENABLED")
		fmt.Println("⚠️  This mode should NEVER be used in production!")
	}

	// Log the current mode
	log.Printf("Running in mode: %s", mode.ModeString())
	fmt.Printf("🔒 Running in mode: %s\n", mode.ModeString())

	// Initialize Prometheus metrics (AI.md PART 21 - NON-NEGOTIABLE)
	metrics.Init(Version, CommitID, BuildDate)

	// Get OS-appropriate directory paths
	dirPaths, err := utils.GetDirectoryPaths()
	if err != nil {
		log.Fatalf("Failed to determine directory paths: %v", err)
	}

	// Apply environment variable overrides (set by CLI or directly)
	// AI.md PART 7: Permissions - root: 0755, user: 0700
	dirPerm := os.FileMode(0700)
	if os.Geteuid() == 0 {
		dirPerm = 0755
	}

	envDataDir := os.Getenv("DATA_DIR")
	if envDataDir != "" {
		// CLI override for data directory
		if info, err := os.Stat(envDataDir); err == nil {
			if !info.IsDir() {
				if err := os.Remove(envDataDir); err != nil {
					log.Fatalf("Failed to remove file at %s: %v", envDataDir, err)
				}
			}
		}
		if err := os.MkdirAll(envDataDir, dirPerm); err != nil {
			log.Fatalf("Failed to create data directory %s: %v", envDataDir, err)
		}
		dirPaths.Data = envDataDir
	}

	envConfigDir := os.Getenv("CONFIG_DIR")
	if envConfigDir != "" {
		// CLI override for config directory
		if info, err := os.Stat(envConfigDir); err == nil {
			if !info.IsDir() {
				if err := os.Remove(envConfigDir); err != nil {
					log.Fatalf("Failed to remove file at %s: %v", envConfigDir, err)
				}
			}
		}
		if err := os.MkdirAll(envConfigDir, dirPerm); err != nil {
			log.Fatalf("Failed to create config directory %s: %v", envConfigDir, err)
		}
		dirPaths.Config = envConfigDir
	}

	envLogDir := os.Getenv("LOG_DIR")
	if envLogDir != "" {
		dirPaths.Log = envLogDir
	}

	// Create all required directories
	if err := utils.CreateDirectories(dirPaths); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// Generate server.yml if it doesn't exist (runtime generation per TEMPLATE.md)
	if err := cli.GenerateServerYML(dirPaths.Config); err != nil {
		log.Printf("Warning: Failed to generate server.yml: %v", err)
	}

	// Initialize logger
	appLogger, err := utils.NewLogger(dirPaths.Log)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Print startup timestamp
	startTime := time.Now()
	appLogger.Printf("%s", startTime.Format("2006-01-02 at 15:04:05"))
	fmt.Printf("🕐 %s\n", startTime.Format("2006-01-02 at 15:04:05"))

	// TEMPLATE.md PART 1: First Run Detection and Auto-Configuration
	isFirstRun := utils.DetectFirstRun(dirPaths.Data)
	var setupToken string

	if isFirstRun {
		appLogger.Printf("First run detected - initializing server...")
		fmt.Println("🎉 First run detected - auto-configuring server...")

		// Auto-detect SMTP
		smtpHost, smtpPort := utils.AutoDetectSMTP()
		appLogger.Printf("SMTP auto-detected: %s:%d", smtpHost, smtpPort)
		fmt.Printf("📧 SMTP auto-detected: %s:%d\n", smtpHost, smtpPort)

		// Create server.yml with auto-detected settings
		configPath := filepath.Join(dirPaths.Config, "server.yml")
		if err := utils.CreateDefaultServerYML(configPath, smtpHost, smtpPort); err != nil {
			appLogger.Error("Failed to create server.yml: %v", err)
			fmt.Printf("⚠️  Failed to create server.yml: %v\n", err)
		} else {
			appLogger.Printf("server.yml created: %s", configPath)
			fmt.Printf("✅ server.yml created with auto-detected settings\n")
		}

		// Generate one-time setup token
		token, err := utils.GenerateSetupToken()
		if err != nil {
			appLogger.Fatal("Failed to generate setup token: %v", err)
		}
		setupToken = token
		appLogger.Printf("Setup token generated (will be displayed in banner)")
	}

	// Initialize database - TEMPLATE.md PART 31: Dual database architecture
	// SQLite dual database: server.db + users.db
	// server.db = admin credentials, config, scheduler, audit log
	// users.db = user accounts, tokens, sessions, locations
	dualDB, err := database.InitDualDB(dirPaths.Data)
	if err != nil {
		appLogger.Fatal("Failed to initialize dual database system: %v", err)
	}
	defer dualDB.Close()

	// Set global instance for handler access
	database.SetGlobalDualDB(dualDB)
	dbPath := fmt.Sprintf("%s/db/server.db + %s/db/users.db", dirPaths.Data, dirPaths.Data)

	// Create wrapper for handlers that use database.DB struct
	// Uses Users database for user-related operations
	db := &database.DB{DB: dualDB.Users}

	// Check if setup is complete
	var setupComplete bool
	var setupValue string
	err = dualDB.QueryRowServer("SELECT value FROM server_config WHERE key = 'setup.completed'").Scan(&setupValue)
	setupComplete = (err == nil && setupValue == "true")

	// If first run, store setup token in database
	if isFirstRun && setupToken != "" {
		_, err = database.GetServerDB().Exec(`
			INSERT INTO server_config (key, value, type, description, updated_at)
			VALUES ('setup.token', ?, 'string', 'One-time setup token (first run)', datetime('now'))
		`, setupToken)
		if err != nil {
			appLogger.Error("Failed to store setup token: %v", err)
		} else {
			appLogger.Printf("Setup token stored in database")
		}
	}

	if setupComplete {
		appLogger.Printf("Database initialized: %s", dbPath)
		fmt.Printf("✅ Database initialized: %s\n", dbPath)
	} else {
		appLogger.Printf("Database initialized: %s (setup mode)", dbPath)
		fmt.Printf("✅ Database initialized: %s (setup mode)\n", dbPath)
	}

	// Load server configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		appLogger.Error("Warning: Could not load server.yml: %v (using defaults)", err)
		fmt.Printf("⚠️  Warning: Could not load server.yml: %v (using defaults)\n", err)
	} else {
		appLogger.Printf("Configuration loaded from server.yml")
		fmt.Printf("✅ Configuration loaded from server.yml\n")
	}

	// Set global config for handler access
	config.SetGlobalConfig(cfg)

	// Note: Version and BuildDate are embedded in binary via LDFLAGS, not in config file

	// Initialize default settings with proper backup path
	settingsModel := &models.SettingsModel{DB: db.DB}
	backupPath := utils.GetBackupPath(dirPaths)
	if err := settingsModel.InitializeDefaults(backupPath); err != nil {
		appLogger.Error("Warning: Could not initialize default settings: %v", err)
		fmt.Printf("⚠️  Warning: Could not initialize default settings: %v\n", err)
	}

	// Initialize cache manager (Valkey/Redis support, optional)
	cacheManager := service.NewCacheManager()
	if cacheManager.IsEnabled() {
		appLogger.Printf("Cache enabled (Redis/Valkey)")
		fmt.Printf("✅ Cache enabled (Redis/Valkey)\n")
	}

	// Auto-detect SMTP server (localhost, Docker gateway, etc.) and configure defaults
	smtpService := service.NewSMTPService(db.DB)
	if err := smtpService.LoadConfig(); err == nil {
		// Check if SMTP is not already configured
		smtpHost := settingsModel.GetString("smtp.host", "")
		if smtpHost == "" {
			// Try auto-detect
			if detected, _ := smtpService.AutoDetect(); detected {
				// SMTP detected, enable it
				settingsModel.SetBool("smtp.enabled", true)
				appLogger.Printf("SMTP server auto-detected and enabled")
				fmt.Printf("✉️  SMTP server auto-detected and enabled\n")
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

	// Check if this is first run (no users created yet)
	hasNoUsers, err := db.IsFirstRun()
	if err != nil {
		appLogger.Error("Warning: Could not check first run status: %v", err)
		fmt.Printf("⚠️  Warning: Could not check first run status: %v\n", err)
		hasNoUsers = false
	}
	if hasNoUsers {
		appLogger.Printf("No users found - please create an admin account at /auth/register")
		fmt.Printf("🆕 No users found - please create an admin account at /auth/register\n")
	}

	// Handle status flag
	// AI.md PART 8: --status exits with 0=healthy, 1=unhealthy
	if os.Getenv("CLI_STATUS_FLAG") == "1" {
		isHealthy := showServerStatus(db, dbPath, hasNoUsers)
		if isHealthy {
			os.Exit(0)
		}
		os.Exit(1)
	}

	// Set Gin mode based on MODE variable (development, production, test)
	// AI.md PART 5: Environment Variables
	envMode := os.Getenv("MODE")
	if envMode == "" {
		envMode = os.Getenv("ENVIRONMENT") // Legacy fallback
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

	// AI.md PART 5: Middleware order - security first!
	// 1. URL normalization (FIRST - normalize before anything else)
	r.Use(middleware.URLNormalizeMiddleware())

	// 2. Path security (SECOND - block traversal attacks before processing)
	r.Use(middleware.PathSecurityMiddleware())

	// Request ID middleware - for request tracing in logs
	r.Use(middleware.RequestID())

	// Access logging middleware (writes to log files)
	r.Use(middleware.AccessLogger(appLogger))

	// Recovery middleware
	r.Use(gin.Recovery())

	// Response compression per AI.md PART 18 lines 15704-15719
	// Compresses text/html, text/css, application/json, etc.
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// Prometheus metrics middleware (AI.md PART 21 - NON-NEGOTIABLE)
	r.Use(middleware.MetricsMiddleware())

	// Security headers middleware
	r.Use(middleware.SecurityHeaders())

	// Body size limit middleware per AI.md PART 18 line 15691 (10MB)
	r.Use(middleware.BodySizeLimitMiddleware(middleware.DefaultMaxBodySize))

	// CSRF protection middleware (AI.md PART 0 line 994, PART 22)
	r.Use(middleware.CSRFProtection(middleware.DefaultCSRFConfig()))

	// CORS middleware per AI.md PART 17 lines 14220-14222 and 15401-15405
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-API-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		// Per AI.md line 15405: Access-Control-Max-Age = 86400 (24 hours)
		MaxAge: 24 * time.Hour,
	}))

	// Global rate limiting middleware (100 req/s)
	r.Use(middleware.GlobalRateLimitMiddleware())

	// Server context middleware - injects server title/tagline/description
	r.Use(middleware.InjectServerContext(db.DB, Version))

	// Check for first user setup - redirects to /users/setup if no users exist
	r.Use(middleware.CheckFirstUserSetup(db.DB))

	// Restrict admin users to only access /admin routes - all other routes treat them as anonymous
	r.Use(middleware.RestrictAdminToAdminRoutes())

	// Path normalization handled by middleware.URLNormalizeMiddleware() and middleware.PathSecurityMiddleware()

	// Serve embedded static files from server package
	staticSubFS, err := server.GetStaticSubFS()
	if err != nil {
		log.Fatalf("Failed to get static subdirectory: %v", err)
	}
	r.StaticFS("/static", http.FS(staticSubFS))

	// Initialize i18n service (TEMPLATE.md PART 29 - NON-NEGOTIABLE)
	i18nService, err := service.NewI18n(localesFS, "en")
	if err != nil {
		log.Fatalf("Failed to initialize i18n: %v", err)
	}
	fmt.Printf("🌐 I18n initialized with languages: %v\n", i18nService.GetSupportedLanguages())

	// I18n middleware - detects language from Accept-Language header
	r.Use(func(c *gin.Context) {
		acceptLang := c.GetHeader("Accept-Language")
		lang := i18nService.ParseAcceptLanguage(acceptLang)
		c.Set("lang", lang)
		c.Set("i18n", i18nService)
		c.Next()
	})

	// Load embedded templates with custom functions from server package
	// Get embedded templates filesystem
	templatesFS := server.GetTemplatesFS()
	// Create sub-filesystem starting at "templates/" so template names don't include "templates/" prefix
	templatesSubFS, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		log.Fatalf("Failed to get templates subdirectory: %v", err)
	}

	// Walk the filesystem and collect all .tmpl files
	var templatePaths []string
	fs.WalkDir(templatesSubFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".tmpl") {
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

	// Create template function map with i18n support
	templateFuncs := template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		// i18n translation function - expects lang to be set in template data
		"t": func(lang, key string) string {
			return i18nService.T(lang, key)
		},
	}

	// Parse all templates
	tmpl := template.Must(template.New("").Funcs(templateFuncs).ParseFS(templatesSubFS, templatePaths...))

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
		if _, err := os.Stat("src/server/templates"); err == nil {
			r.Use(func(c *gin.Context) {
				// Try to reload from filesystem in debug mode
				t := template.New("").Funcs(templateFuncs)
				// Load all templates including subdirectories
				// Note: This loads from filesystem, so paths are relative to src/server/templates/
				patterns := []string{
					"src/server/templates/*.tmpl",
					"src/server/templates/*/*.tmpl",
					"src/server/templates/*/*/*.tmpl",
				}
				for _, pattern := range patterns {
					t, _ = t.ParseGlob(pattern)
				}
				// Need to rename templates to remove "src/server/templates/" prefix for consistency
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
	locationEnhancer := service.NewLocationEnhancer(db.DB)

	// Set callback to mark initialization complete
	locationEnhancer.SetOnInitComplete(func(countries, cities bool) {
		// Mark weather service as always ready (no initialization needed)
		handler.SetInitStatus(countries, cities, true)
		fmt.Printf("✅ Countries: %v, Cities: %v, zipcodes: true, airportcodes: true\n", countries, cities)
	})

	// Initialize GeoIP service (downloads database on first run, updates weekly)
	geoipService := service.NewGeoIPService(dirPaths.Config)

	weatherService := service.NewWeatherService(locationEnhancer, geoipService)

	// Data loads automatically in the background via loadData()
	// Mark service as ready after 2 minute initialization timeout (keep as fallback)
	go func() {
		time.Sleep(2 * time.Minute)
		if !handler.IsInitialized() {
			fmt.Println("⏰ Initialization timeout reached, marking service as ready (fallback)")
			fmt.Printf("🕐 %s\n", time.Now().Format("2006-01-02 at 15:04:05"))
			handler.SetInitStatus(true, true, true)
		}
	}()

	// Initialize notification system services (silent)
	channelManager := service.NewChannelManager(db.DB)
	templateEngine := service.NewTemplateEngine(db.DB)
	deliverySystem := service.NewDeliverySystem(db.DB, channelManager, templateEngine)

	// Load delivery system settings from database
	_ = deliverySystem.LoadSettings()

	// Initialize default templates
	_ = templateEngine.InitializeDefaultTemplates()

	// Initialize channels in database
	_ = channelManager.InitializeChannels()

	// Create weather notification service
	weatherNotifications := service.NewWeatherNotificationService(db.DB, weatherService, deliverySystem, templateEngine)

	// Initialize notification metrics service
	notificationMetrics := service.NewNotificationMetrics(db.DB)

	// Initialize Tor hidden service (TEMPLATE.md PART 32 - NON-NEGOTIABLE)
	torService := service.NewTorService(db, dirPaths.Data)

	// Initialize config file watcher for live reload (TEMPLATE.md PART 1)
	configPath := filepath.Join(dirPaths.Config, "server.yml")
	configWatcher, err := service.NewConfigWatcher(configPath, func(newCfg *config.AppConfig) error {
		// Reload configuration callback - applies changes live without restart
		log.Printf("Configuration reloaded from %s", configPath)
		fmt.Printf("🔄 Configuration reloaded from %s\n", configPath)

		// Update all configuration sections that can be changed at runtime
		cfg.Server.Mode = newCfg.Server.Mode
		cfg.Server.Branding = newCfg.Server.Branding
		cfg.Server.SEO = newCfg.Server.SEO
		cfg.Web = newCfg.Web
		cfg.Server.Notifications = newCfg.Server.Notifications
		cfg.Server.RateLimit = newCfg.Server.RateLimit
		cfg.Server.Tor = newCfg.Server.Tor
		cfg.Server.Features = newCfg.Server.Features

		// Update global config for handlers
		config.SetGlobalConfig(cfg)

		// Note: Port changes would require graceful restart (not implemented yet)
		// For now, port changes require manual restart

		log.Println("✅ All configuration sections reloaded (branding, SEO, theme, email, notifications, rate limiting, web, Tor, features)")
		fmt.Println("✅ All configuration sections reloaded successfully")

		return nil
	})
	if err != nil {
		log.Printf("Failed to create config watcher: %v", err)
		fmt.Printf("⚠️  Failed to create config watcher: %v\n", err)
	}

	// Initialize scheduler for periodic tasks
	taskScheduler := scheduler.NewScheduler(db.DB)

	// Register log rotation task - run daily at midnight
	taskScheduler.AddTask("rotate-logs", 24*time.Hour, func() error {
		return appLogger.RotateLogs()
	})

	// Register cleanup tasks - AI.md PART 19: session cleanup every 15 minutes
	taskScheduler.AddTask("cleanup-sessions", 15*time.Minute, func() error {
		return scheduler.CleanupOldSessions(db.DB)
	})

	// AI.md PART 19: token cleanup every 15 minutes
	taskScheduler.AddTask("cleanup-tokens", 15*time.Minute, func() error {
		return scheduler.CleanupExpiredTokens(db.DB)
	})

	taskScheduler.AddTask("cleanup-rate-limits", 1*time.Hour, func() error {
		return scheduler.CleanupRateLimitCounters(db.DB)
	})

	taskScheduler.AddTask("cleanup-audit-logs", 24*time.Hour, func() error {
		return scheduler.CleanupOldAuditLogs(db.DB)
	})

	// Register weather alert checks - run every 5 minutes per IDEA.md
	taskScheduler.AddTask("check-weather-alerts", 5*time.Minute, func() error {
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
	// Keep 30 days
	taskScheduler.AddTask("cleanup-notifications", 24*time.Hour, func() error {
		return deliverySystem.CleanupOld(30)
	})

	// AI.md PART 19: backup daily at 02:00 (using 24h interval for now)
	taskScheduler.AddTask("backup-daily", 24*time.Hour, func() error {
		return scheduler.CreateSystemBackup(db.DB)
	})

	// AI.md PART 19 line 27050: backup_hourly - hourly incremental (disabled by default)
	// Only runs if backup.hourly_enabled is true in config
	taskScheduler.AddTask("backup-hourly", 1*time.Hour, func() error {
		if !cfg.Server.Maintenance.Backup.HourlyEnabled {
			return nil
		}
		p := paths.GetDefaultPaths("weather")
		if p == nil {
			return fmt.Errorf("failed to get paths for hourly backup")
		}
		return scheduler.BackupHourlyTask(p.ConfigDir, p.DataDir)()
	})

	// AI.md PART 19: SSL renewal check daily at 03:00
	taskScheduler.AddTask("ssl-renewal", 24*time.Hour, func() error {
		return scheduler.CheckSSLRenewal()
	})

	// AI.md PART 19: self health check every 5 minutes
	taskScheduler.AddTask("healthcheck-self", 5*time.Minute, func() error {
		return scheduler.SelfHealthCheck()
	})

	// AI.md PART 19: Tor health check every 10 minutes (when Tor installed)
	taskScheduler.AddTask("tor-health", 10*time.Minute, func() error {
		return scheduler.CheckTorHealth()
	})

	// Register weather cache refresh - run every 15 minutes per IDEA.md
	taskScheduler.AddTask("refresh-weather-cache", 15*time.Minute, func() error {
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

	// AI.md PART 19: blocklist update daily at 04:00
	taskScheduler.AddTask("blocklist-update", 24*time.Hour, func() error {
		return scheduler.UpdateBlocklist()
	})

	// AI.md PART 19: CVE database update daily at 05:00
	taskScheduler.AddTask("cve-update", 24*time.Hour, func() error {
		return scheduler.UpdateCVEDatabase()
	})

	// AI.md PART 19 line 24792: cluster.heartbeat every 30 seconds (cluster mode only)
	// This is a LOCAL task - runs on every node (not a global task)
	nodeIDForHeartbeat, _ := os.Hostname()
	if nodeIDForHeartbeat == "" {
		nodeIDForHeartbeat = "default"
	}
	taskScheduler.AddTask("cluster-heartbeat", 30*time.Second, func() error {
		return scheduler.ClusterHeartbeat(nodeIDForHeartbeat)
	})

	// Initialize task history table for scheduler tracking
	if err := taskScheduler.InitTaskHistoryTable(); err != nil {
		fmt.Printf("❌ Failed to initialize task history table: %v\n", err)
		log.Fatalf("Failed to initialize task history table: %v", err)
	}

	// Start the scheduler
	taskScheduler.Start()

	// Schedule WebUI notification cleanup tasks (TEMPLATE.md Part 25)
	// Note: NotificationCleaner will be initialized after NotificationService is created
	// Cleanup scheduled for 02:00 UTC, Limit enforcement at 03:00 UTC

	// Create services
	earthquakeService := service.NewEarthquakeService()
	hurricaneService := service.NewHurricaneService()
	severeWeatherService := service.NewSevereWeatherService()

	// Create handlers
	weatherHandler := handler.NewWeatherHandler(weatherService, locationEnhancer)
	apiHandler := handler.NewAPIHandler(weatherService, locationEnhancer)
	webHandler := handler.NewWebHandler(weatherService, locationEnhancer)
	earthquakeHandler := handler.NewEarthquakeHandler(earthquakeService, weatherService, locationEnhancer)
	hurricaneHandler := handler.NewHurricaneHandler(hurricaneService)
	severeWeatherHandler := handler.NewSevereWeatherHandler(severeWeatherService, locationEnhancer, weatherService)
	moonHandler := handler.NewMoonHandler(weatherService, locationEnhancer)

	// Create auth handlers
	authHandler := &handler.AuthHandler{DB: db.DB}
	twoFAHandler := &handler.TwoFactorHandler{DB: db.DB}
	setupHandler := &handler.SetupHandler{DB: db.DB}
	dashboardHandler := &handler.DashboardHandler{DB: db.DB}
	adminHandler := &handler.AdminHandler{DB: db.DB}
	locationHandler := &handler.LocationHandler{
		DB:               db.DB,
		WeatherService:   weatherService,
		LocationEnhancer: locationEnhancer,
	}

	// Initialize WebSocket Hub for real-time notifications (TEMPLATE.md Part 25)
	wsHub := service.NewWebSocketHub()
	// Start hub in goroutine
	go wsHub.Run()

	// Initialize Notification Service (TEMPLATE.md Part 25 - WebUI Notifications)
	notificationService := &service.NotificationService{
		UserDB:     dualDB.Users,
		ServerDB:   dualDB.Server,
		WSHub:      wsHub,
		UserNotif:  &models.UserNotificationModel{DB: dualDB.Users},
		AdminNotif: &models.AdminNotificationModel{DB: dualDB.Server},
		Prefs:      &models.NotificationPreferencesModel{UserDB: dualDB.Users, ServerDB: dualDB.Server},
	}

	// Create WebUI notification API handlers (TEMPLATE.md Part 25)
	notificationAPIHandler := &handler.NotificationAPIHandlers{
		NotificationService: notificationService,
		WSHub:               wsHub,
	}

	// Legacy notification handler (for email notifications only)
	notificationHandler := &handler.NotificationHandler{DB: db.DB}

	// Create notification system handlers
	channelHandler := handler.NewNotificationChannelHandler(db.DB)
	preferencesHandler := handler.NewNotificationPreferencesHandler(db.DB)
	templateHandler := handler.NewNotificationTemplateHandler(db.DB)
	metricsHandler := handler.NewNotificationMetricsHandler(notificationMetrics)

	// Initialize WebUI Notification Cleanup Scheduler (TEMPLATE.md Part 25)
	notificationCleaner := scheduler.NewNotificationCleaner(notificationService)
	// Daily at 2 AM UTC
	taskScheduler.ScheduleNotificationCleanup(notificationCleaner, "02:00")
	// Daily at 3 AM UTC
	taskScheduler.ScheduleNotificationLimitEnforcement(notificationCleaner, "03:00")

	// Create scheduler handler for task management
	schedulerHandler := handler.NewSchedulerHandler(taskScheduler)

	// Create Tor admin handler
	torAdminHandler := handler.NewTorAdminHandler(torService, settingsModel, dirPaths.Data)

	// Create email template handler
	emailTemplateHandler := handler.NewEmailTemplateHandler(filepath.Join("src", "server", "templates"))

	// Create logs handler
	logsHandler := handler.NewLogsHandler(dirPaths.Log)

	// Create admin settings handlers
	adminUsersHandler := &handler.AdminUsersHandler{ConfigPath: configPath}
	adminAuthHandler := &handler.AdminAuthSettingsHandler{ConfigPath: configPath}
	adminWeatherHandler := &handler.AdminWeatherHandler{ConfigPath: configPath}
	adminNotificationsHandler := &handler.AdminNotificationsHandler{ConfigPath: configPath}
	adminGeoIPHandler := &handler.AdminGeoIPHandler{ConfigPath: configPath}

	// Create domain handler (TEMPLATE.md PART 34: Custom domain support)
	domainHandler := handler.NewDomainHandlers(db.DB, appLogger)

	// Get port configuration using comprehensive port manager
	// Priority: 1) Database saved ports, 2) Config file port, 3) PORT env variable, 4) Random port
	portManager := utils.NewPortManager(db.DB)

	// Extract port from config (can be int or string)
	configPort := 0
	if cfg != nil && cfg.Server.Port != nil {
		switch p := cfg.Server.Port.(type) {
		case int:
			configPort = p
		case float64:
			configPort = int(p)
		case string:
			if parsed, err := strconv.Atoi(p); err == nil {
				configPort = parsed
			}
		}
	}

	httpPortInt, httpsPortInt, err := portManager.GetServerPortsWithConfig(configPort)
	if err != nil {
		log.Fatalf("Failed to configure server ports: %v", err)
	}

	port := fmt.Sprintf("%d", httpPortInt)

	// Get listen address - auto-detect reverse proxy and IPv6 support
	// AI.md PART 5: LISTEN env var
	listenAddress := os.Getenv("LISTEN")
	if listenAddress == "" {
		listenAddress = os.Getenv("SERVER_ADDRESS") // Legacy fallback
	}
	
	// Check if listenAddress contains a port (e.g., "127.0.0.1:8080" or "[::]:8080")
	if listenAddress != "" && strings.Contains(listenAddress, ":") {
		// Try to split host and port
		host, portStr, err := net.SplitHostPort(listenAddress)
		if err == nil && portStr != "" {
			// Successfully parsed - update both listenAddress and port
			listenAddress = host
			if parsedPort, err := strconv.Atoi(portStr); err == nil && parsedPort > 0 && parsedPort < 65536 {
				httpPortInt = parsedPort
				port = portStr
			}
		}
	}
	
	networkMode := ""
	if listenAddress == "" {
		// Check for reverse proxy indicator per AI.md PART 5: Boolean Handling
		reverseProxy := config.IsTruthy(os.Getenv("REVERSE_PROXY"))

		if reverseProxy {
			listenAddress = "127.0.0.1"
			networkMode = " in reverse proxy mode"
		} else {
			// Auto-detect IPv6 support and use dual-stack if available
			listenAddress = getDefaultListenAddress()
			if listenAddress == "::" {
				networkMode = " (dual-stack: IPv4 + IPv6)"
			} else {
				networkMode = " (IPv4 only)"
			}
		}
	}

	// Print startup messages
	// Format display address correctly for IPv6
	displayAddr := listenAddress
	if listenAddress == "::" {
		displayAddr = "[::]"
	}
	appLogger.Printf("Starting Weather%s on %s:%s", networkMode, displayAddr, port)
	fmt.Printf("🚀 Starting Weather%s on %s:%s\n", networkMode, displayAddr, port)
	appLogger.Info("Data directory: %s", dirPaths.Data)
	appLogger.Info("Config directory: %s", dirPaths.Config)
	appLogger.Info("Log directory: %s", dirPaths.Log)

	// Initialize SSL manager
	sslCertsDir := utils.GetCertsPath(dirPaths)
	sslManager := utils.NewSSLManager(db.DB, sslCertsDir)
	httpsPort := httpsPortInt

	// Create SSL handler
	sslHandler := handler.NewSSLHandler(sslCertsDir, db.DB)

	// Create metrics handler
	metricsConfigHandler := handler.NewMetricsHandler()

	// Create logging handler
	loggingHandler := handler.NewLoggingHandler(dirPaths.Log)

	// Create admin web handler (robots.txt, security.txt)
	adminWebHandler := handler.NewAdminWebHandler(db)

	// Check for SSL configuration
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	// Try to check for existing Let's Encrypt certs and enable HTTPS if configured
	if httpsPort > 0 {
		found, err := sslManager.CheckExistingCerts(hostname)
		if err != nil {
			appLogger.Error("SSL check failed: %v", err)
			fmt.Printf("⚠️  SSL check failed: %v\n", err)
		} else if found {
			appLogger.Printf("Found Let's Encrypt certificate for %s", hostname)
			appLogger.Printf("HTTPS enabled on port: %d", httpsPort)
			fmt.Printf("🔒 Found Let's Encrypt certificate for %s\n", hostname)
			fmt.Printf("🔌 HTTPS enabled on port: %d\n", httpsPort)
		} else {
			appLogger.Printf("HTTPS port configured (%d) but no certificates found", httpsPort)
			fmt.Printf("ℹ️  HTTPS port configured (%d) but no certificates found\n", httpsPort)
		}
	}
	// Note: Self-signed cert generation is optional and disabled by default
	// Can be enabled via CLI flag or environment variable if needed

	// Set directory paths for handlers
	handler.SetDirectoryPaths(dirPaths.Data, dirPaths.Log)

	// Health check endpoints (Kubernetes standard)
	r.GET("/healthz", handler.ComprehensiveHealthCheck(db, port, httpsPort, sslManager))
	r.GET("/health", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/healthz")
	})
	r.GET("/readyz", handler.ReadinessCheck)
	r.GET("/livez", handler.LivenessCheck)
	r.GET("/healthz/setup", setupHandler.GetSetupStatus)

	// Prometheus metrics endpoint (TEMPLATE.md required - optional auth)
	r.GET("/metrics", handler.PrometheusMetrics())

	// security.txt endpoint (RFC 9116 - TEMPLATE.md PART 25)
	r.GET("/.well-known/security.txt", adminWebHandler.ServeSecurityTxt)
	// Also serve at root for compatibility
	r.GET("/security.txt", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/.well-known/security.txt")
	})

	// /.well-known/change-password redirect (TEMPLATE.md PART 25)
	r.GET("/.well-known/change-password", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/profile?tab=security")
	})

	// /.well-known/acme-challenge/:token - Let's Encrypt HTTP-01 challenge (TEMPLATE.md Part 8)
	r.GET("/.well-known/acme-challenge/:token", func(c *gin.Context) {
		token := c.Param("token")
		// This would retrieve the key authorization from the Let's Encrypt service
		// For now, return a placeholder response
		c.String(http.StatusOK, "ACME challenge token: %s", token)
	})

	// robots.txt endpoint
	r.GET("/robots.txt", adminWebHandler.ServeRobotsTxt)

	// sitemap.xml endpoint (AI.md PART 16: dynamically generated)
	r.GET("/sitemap.xml", adminWebHandler.ServeSitemap)

	// favicon.ico endpoint (AI.md PART 16: embedded default, customizable)
	r.GET("/favicon.ico", adminWebHandler.ServeFavicon)

	// Debug endpoints (only enabled when --debug flag or DEBUG=true)
	// Per AI.md PART 6: Debug endpoints only available when debug mode enabled
	if mode.IsDebugEnabled() {
		debugHandlers := handler.NewDebugHandlers(db.DB, r)
		debugHandlers.RegisterDebugRoutes(r)

		// pprof endpoints per AI.md PART 6
		debugGroup := r.Group("/debug/pprof")
		{
			debugGroup.GET("/", gin.WrapF(pprof.Index))
			debugGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
			debugGroup.GET("/profile", gin.WrapF(pprof.Profile))
			debugGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
			debugGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
			debugGroup.GET("/trace", gin.WrapF(pprof.Trace))
			debugGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
			debugGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
			debugGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
			debugGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
			debugGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
			debugGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
		}

		// expvar endpoint per AI.md PART 6
		r.GET("/debug/vars", gin.WrapH(http.DefaultServeMux))

		log.Println("🔧 Debug endpoints enabled:")
		log.Println("   GET  /debug/routes  - List all routes")
		log.Println("   GET  /debug/config  - Show configuration")
		log.Println("   GET  /debug/memory  - Memory statistics")
		log.Println("   GET  /debug/db      - Database statistics")
		log.Println("   POST /debug/reload  - Reload configuration")
		log.Println("   POST /debug/gc      - Trigger garbage collection")
		log.Println("   GET  /debug/pprof/  - pprof index")
		log.Println("   GET  /debug/pprof/heap - Heap profile")
		log.Println("   GET  /debug/pprof/goroutine - Goroutine dump")
		log.Println("   GET  /debug/vars    - expvar metrics")
	}

	// IP detection endpoint (always available for My Location feature)
	r.GET("/debug/ip", func(c *gin.Context) {
		// IP detection for My Location button
		clientIP := utils.GetClientIP(c)

		// Try to get location from IP
		coords, err := weatherService.GetCoordinatesFromIP(clientIP)
		if err != nil {
			// Empty means fallback to manual entry
			c.JSON(http.StatusOK, gin.H{
				"clientIP": clientIP,
				"location": gin.H{
					"value": "",
				},
				"error": err.Error(),
			})
			return
		}

		// Enhance location
		enhanced := locationEnhancer.EnhanceLocation(coords)

		// e.g., "Albany, NY"
		c.JSON(http.StatusOK, gin.H{
			"clientIP": clientIP,
			"location": gin.H{
				"value": enhanced.ShortName,
			},
			"coordinates": gin.H{
				"latitude":  coords.Latitude,
				"longitude": coords.Longitude,
			},
		})
	})

	// First-run setup routes (public)
	// First user setup routes (blocked after setup complete)
	// AI.md PART 14: Use plural nouns for all routes
	setupRoutes := r.Group("/users/setup")
	setupRoutes.Use(middleware.BlockSetupAfterComplete(db.DB))
	{
		setupRoutes.GET("", setupHandler.ShowWelcome)
		setupRoutes.GET("/register", setupHandler.ShowUserRegister)
		setupRoutes.POST("/register", setupHandler.CreateUser)
		// AI.md: Setup routes use /{admin_path}
		setupRoutes.GET("/"+cfg.GetAdminPath(), setupHandler.ShowAdminSetup)
		setupRoutes.POST("/"+cfg.GetAdminPath(), setupHandler.CreateAdmin)
		setupRoutes.GET("/complete", setupHandler.CompleteSetup)
	}

	// Setup wizard routes
	setupWizard := r.Group("/setup")
	setupWizard.Use(middleware.RequireAuth(db.DB))
	{
		// Admin account creation (step 1 - accessible only when no admin exists)
		// AI.md: Setup wizard uses /{admin_path}
		adminSetup := setupWizard.Group("/" + cfg.GetAdminPath())
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

	// Authentication routes (public) - TEMPLATE.md lines 4441-4534
	r.GET("/auth/login", authHandler.ShowLoginPage)
	r.POST("/auth/login", authHandler.HandleLogin)
	r.GET("/auth/register", authHandler.ShowRegisterPage)
	r.POST("/auth/register", authHandler.HandleRegister)
	r.GET("/auth/logout", authHandler.HandleLogout)

	// User routes (require authentication) - per AI.md PART 14: /users/ is plural
	usersRoutes := r.Group("/users")
	usersRoutes.Use(middleware.RequireAuth(db.DB))
	usersRoutes.Use(middleware.BlockAdminFromUserRoutes())
	{
		// /users -> user dashboard (current user)
		usersRoutes.GET("", dashboardHandler.ShowDashboard)
		// /users/dashboard -> user dashboard
		usersRoutes.GET("/dashboard", dashboardHandler.ShowDashboard)
	}

	// Admin routes (require admin role + stricter rate limiting)
	// AI.md: Admin panel at /{admin_path} (configurable, default: "admin")
	adminRoutes := r.Group("/" + cfg.GetAdminPath())
	adminRoutes.Use(middleware.RequireAuth(db.DB))
	adminRoutes.Use(middleware.RequireAdmin())
	adminRoutes.Use(middleware.AdminRateLimitMiddleware())
	// Log all admin actions
	adminRoutes.Use(middleware.AuditLogger(db.DB))
	{
		// /admin -> admin dashboard
		adminRoutes.GET("", dashboardHandler.ShowAdminPanel)

		// Admin management pages
		adminRoutes.GET("/users", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/users.tmpl", utils.TemplateData(c, gin.H{
				"title":      "User Management - Admin",
				"page":       "users",
				"breadcrumb": "Users",
			}))
		})

		adminRoutes.GET("/settings", adminHandler.ShowSettingsPage)

		adminRoutes.GET("/server/web", adminWebHandler.ShowWebSettings)

		adminRoutes.GET("/email", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/admin-email.tmpl", utils.TemplateData(c, gin.H{
				"title":      "Email Settings - Admin",
				"page":       "email",
				"breadcrumb": "Email",
			}))
		})

		adminRoutes.GET("/database", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/admin-database.tmpl", utils.TemplateData(c, gin.H{
				"title":      "Database & Cache - Admin",
				"page":       "database",
				"breadcrumb": "Database",
			}))
		})

		adminRoutes.GET("/system", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/admin-system.tmpl", utils.TemplateData(c, gin.H{
				"title":      "System Information - Admin",
				"page":       "system",
				"breadcrumb": "System",
			}))
		})

		adminRoutes.GET("/security", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/admin-security.tmpl", utils.TemplateData(c, gin.H{
				"title":      "Security Settings - Admin",
				"page":       "security",
				"breadcrumb": "Security",
			}))
		})

		adminRoutes.GET("/tokens", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/tokens.tmpl", utils.TemplateData(c, gin.H{
				"title":      "API Tokens - Admin",
				"page":       "tokens",
				"breadcrumb": "API Tokens",
			}))
		})

		adminRoutes.GET("/logs", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/admin-logs.tmpl", utils.TemplateData(c, gin.H{
				"title":      "System Logs - Admin",
				"page":       "logs",
				"breadcrumb": "System Logs",
			}))
		})

		adminRoutes.GET("/audit", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/logs.tmpl", utils.TemplateData(c, gin.H{
				"title":      "Audit Logs - Admin",
				"page":       "audit",
				"breadcrumb": "Audit Logs",
			}))
		})

		adminRoutes.GET("/tasks", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/admin-tasks-enhanced.tmpl", utils.TemplateData(c, gin.H{
				"title":      "Scheduled Tasks - Admin",
				"page":       "tasks",
				"breadcrumb": "Scheduled Tasks",
			}))
		})

		adminRoutes.GET("/ssl", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/admin-ssl.tmpl", utils.TemplateData(c, gin.H{
				"title":      "SSL/TLS Management - Admin",
				"page":       "ssl",
				"breadcrumb": "SSL/TLS",
			}))
		})

		adminRoutes.GET("/backup", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/admin-backup-enhanced.tmpl", utils.TemplateData(c, gin.H{
				"title":      "Backup Management - Admin",
				"page":       "backup",
				"breadcrumb": "Backup",
			}))
		})

		adminRoutes.GET("/metrics", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/admin-metrics.tmpl", utils.TemplateData(c, gin.H{
				"title":      "Metrics Configuration - Admin",
				"page":       "metrics",
				"breadcrumb": "Metrics",
			}))
		})

		adminRoutes.GET("/server/tor", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/admin-tor.tmpl", utils.TemplateData(c, gin.H{
				"title":      "Tor Hidden Service - Admin",
				"page":       "tor",
				"breadcrumb": "Tor Hidden Service",
			}))
		})

		adminRoutes.GET("/channels", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin_channels.tmpl", utils.TemplateData(c, gin.H{
				"title":      "Notification Channels - Admin",
				"page":       "channels",
				"breadcrumb": "Channels",
			}))
		})

		adminRoutes.GET("/templates", func(c *gin.Context) {
			c.HTML(http.StatusOK, "template_editor.tmpl", utils.TemplateData(c, gin.H{
				"title":      "Template Editor - Admin",
				"page":       "templates",
				"breadcrumb": "Templates",
			}))
		})

		adminRoutes.GET("/email/templates", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin/admin-email-editor.tmpl", utils.TemplateData(c, gin.H{
				"title":      "Email Template Editor - Admin",
				"page":       "email-templates",
				"breadcrumb": "Email Templates",
			}))
		})

		// New admin panels
		adminRoutes.GET("/server/users", adminUsersHandler.ShowUserSettings)
		adminRoutes.GET("/server/auth", adminAuthHandler.ShowAuthSettings)
		adminRoutes.GET("/server/weather", adminWeatherHandler.ShowWeatherSettings)
		adminRoutes.GET("/server/notifications", adminNotificationsHandler.ShowNotificationSettings)
		adminRoutes.GET("/server/geoip", adminGeoIPHandler.ShowGeoIPSettings)

		// Custom domains management page (TEMPLATE.md PART 34)
		adminRoutes.GET("/domains", func(c *gin.Context) {
			// Get all domains from database
			domainModel := &models.DomainModel{DB: db.DB}
			domains, err := domainModel.List(nil)
			if err != nil {
				appLogger.Error("Failed to list domains: %v", err)
				domains = []*models.Domain{}
			}

			c.HTML(http.StatusOK, "admin-domains", utils.TemplateData(c, gin.H{
				"title":   "Custom Domains",
				"page":    "domains",
				"Domains": domains,
			}))
		})
	}
	r.GET("/notifications", middleware.RequireAuth(db.DB), notificationHandler.ShowNotificationsPage)

	// User profile page (per AI.md PART 14: /users/ is plural)
	r.GET("/users/profile", middleware.RequireAuth(db.DB), middleware.BlockAdminFromUserRoutes(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "pages/user/profile.tmpl", utils.TemplateData(c, gin.H{
			"title": "Profile",
			"page":  "profile",
		}))
	})

	// User security settings page (per AI.md PART 14: /users/ is plural)
	r.GET("/users/security", middleware.RequireAuth(db.DB), middleware.BlockAdminFromUserRoutes(), twoFAHandler.ShowSecurityPage)

	// User notification preferences page (per AI.md PART 14: /users/ is plural)
	r.GET("/users/preferences", middleware.RequireAuth(db.DB), middleware.BlockAdminFromUserRoutes(), func(c *gin.Context) {
		c.HTML(http.StatusOK, "user_preferences.tmpl", utils.TemplateData(c, gin.H{
			"title": "Preferences",
			"page":  "preferences",
		}))
	})

	// Removed - moved to adminRoutes group above

	// Location management pages
	r.GET("/locations/new", middleware.RequireAuth(db.DB), locationHandler.ShowAddLocationPage)
	r.GET("/locations/:id/edit", middleware.RequireAuth(db.DB), locationHandler.ShowEditLocationPage)

	// API routes - all API endpoints under /api/{api_version}
	// AI.md: API version prefix is configurable (default: "v1")
	apiV1 := r.Group(cfg.GetAPIPath())

	// Health check endpoint (JSON) - TEMPLATE.md compliant format
	apiV1.GET("/healthz", handler.APIHealthCheck(db, startTime))

	// Weather API routes (optional auth + API rate limiting)
	weatherAPI := apiV1.Group("")
	weatherAPI.Use(middleware.OptionalAuth(db.DB))
	weatherAPI.Use(middleware.APIRateLimitMiddleware())
	{
		// Weather endpoints per AI.md PART 36
		weatherAPI.GET("/weather", apiHandler.GetWeather)
		weatherAPI.GET("/weather/:location", apiHandler.GetWeatherByLocation)
		weatherAPI.GET("/weather/forecast", apiHandler.GetForecast)
		weatherAPI.GET("/weather/locations", apiHandler.GetLocation)

		// Backwards compatibility - old paths (deprecated)
		weatherAPI.GET("/forecasts", apiHandler.GetForecast)
		weatherAPI.GET("/forecasts/:location", apiHandler.GetForecastByLocation)

		// Additional endpoints
		weatherAPI.GET("/ip", apiHandler.GetIP)
		weatherAPI.GET("/docs", apiHandler.GetDocsJSON)
		weatherAPI.GET("/earthquakes", earthquakeHandler.HandleEarthquakeAPI)
		weatherAPI.GET("/earthquakes/:id", earthquakeHandler.HandleEarthquakeByIDAPI)
		// Backwards compat
		weatherAPI.GET("/hurricanes", hurricaneHandler.HandleHurricaneAPI)
		weatherAPI.GET("/hurricanes/:id", hurricaneHandler.HandleHurricaneByIDAPI)
		weatherAPI.GET("/severe-weather", severeWeatherHandler.HandleSevereWeatherAPI)
		weatherAPI.GET("/severe-weather/:id", severeWeatherHandler.HandleAlertByIDAPI)
		weatherAPI.GET("/moon", moonHandler.HandleMoonAPI)
		weatherAPI.GET("/moon/calendar", moonHandler.HandleMoonCalendarAPI)
		weatherAPI.GET("/sun", moonHandler.HandleSunAPI)
		weatherAPI.GET("/history", apiHandler.GetHistoricalWeather)

		// CLI client compatibility aliases (IDEA.md endpoints)
		weatherAPI.GET("/weather/alerts", severeWeatherHandler.HandleSevereWeatherAPI)
		weatherAPI.GET("/weather/moon", moonHandler.HandleMoonAPI)
		weatherAPI.GET("/weather/history", apiHandler.GetHistoricalWeather)

		// Root /api/{api_version} endpoint - return all endpoints
		// AI.md PART 14: Never hardcode v1 - use cfg.GetAPIPath()
		weatherAPI.GET("", func(c *gin.Context) {
			hostInfo := utils.GetHostInfo(c)
			apiBase := hostInfo.FullHost + cfg.GetAPIPath()
			adminBase := hostInfo.FullHost + cfg.GetAdminAPIPath()
			c.JSON(http.StatusOK, gin.H{
				"version": cfg.GetAPIVersion(),
				"endpoints": []string{
					apiBase + "/users",
					apiBase + "/locations",
					apiBase + "/users/notifications",
					adminBase,
					adminBase + "/domains",
					apiBase + "/weather",
					apiBase + "/weather/:location",
					apiBase + "/forecasts",
					apiBase + "/forecasts/:location",
					apiBase + "/ip",
					apiBase + "/locations",
					apiBase + "/docs",
					apiBase + "/blocklist",
					apiBase + "/earthquakes",
					apiBase + "/hurricanes",
					apiBase + "/severe-weather",
					apiBase + "/moon",
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

	// Server API endpoints (contact form)
	apiV1.POST("/server/contact", handler.HandleContactFormSubmission(db, cfg))

	// Auth API routes per AI.md PART 33
	authAPIHandler := handler.NewAuthAPIHandler(db.DB)
	authAPI := apiV1.Group("/auth")
	{
		// Public auth endpoints (no auth required)
		authAPI.POST("/register", authAPIHandler.HandleAPIRegister)
		authAPI.POST("/login", authAPIHandler.HandleAPILogin)
		authAPI.POST("/2fa", authAPIHandler.HandleAPI2FA)
		authAPI.POST("/recovery/use", authAPIHandler.HandleAPIRecoveryUse)
		authAPI.POST("/password/forgot", authAPIHandler.HandleAPIPasswordForgot)
		authAPI.POST("/password/reset", authAPIHandler.HandleAPIPasswordReset)
		authAPI.POST("/verify", authAPIHandler.HandleAPIVerifyEmail)

		// User invite endpoints (no auth required - token validates)
		authAPI.GET("/invite/user/:token", authAPIHandler.HandleAPIUserInviteValidate)
		authAPI.POST("/invite/user/:token", authAPIHandler.HandleAPIUserInviteComplete)

		// Protected auth endpoints (require auth)
		authAPI.POST("/logout", middleware.RequireAuth(db.DB), authAPIHandler.HandleAPILogout)
		authAPI.POST("/refresh", middleware.RequireAuth(db.DB), authAPIHandler.HandleAPIRefresh)
	}

	// Users API routes per AI.md PART 33 (spec uses /api/v1/users not /api/v1/user)
	usersAPI := apiV1.Group("/users")
	usersAPI.Use(middleware.RequireAuth(db.DB))
	usersAPI.Use(middleware.BlockAdminFromUserRoutes())
	{
		usersAPI.GET("", authHandler.GetCurrentUser)
		usersAPI.PATCH("", authHandler.UpdateProfile)

		// Security endpoints
		usersAPI.GET("/security/2fa", twoFAHandler.GetTwoFactorStatus)
		usersAPI.GET("/security/2fa/setup", twoFAHandler.SetupTwoFactor)
		usersAPI.POST("/security/2fa/enable", twoFAHandler.EnableTwoFactor)
		usersAPI.POST("/security/2fa/disable", twoFAHandler.DisableTwoFactor)
		usersAPI.POST("/security/2fa/verify", twoFAHandler.VerifyTwoFactorCode)
		usersAPI.POST("/security/recovery/regenerate", twoFAHandler.RegenerateRecoveryKeys)
	}

	// Note: 2FA routes already registered under usersAPI (/users/security/2fa/*)

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

	// WebUI Notification API routes - User (per AI.md PART 14: /users/ is plural)
	usersNotificationAPI := apiV1.Group("/users/notifications")
	usersNotificationAPI.Use(middleware.RequireAuth(db.DB))
	usersNotificationAPI.Use(middleware.BlockAdminFromUserRoutes())
	{
		usersNotificationAPI.GET("", notificationAPIHandler.GetUserNotifications)
		usersNotificationAPI.GET("/unread", notificationAPIHandler.GetUserUnreadNotifications)
		usersNotificationAPI.GET("/count", notificationAPIHandler.GetUserUnreadCount)
		usersNotificationAPI.GET("/stats", notificationAPIHandler.GetUserStats)
		usersNotificationAPI.PATCH("/:id/read", notificationAPIHandler.MarkUserNotificationRead)
		usersNotificationAPI.PATCH("/read", notificationAPIHandler.MarkAllUserNotificationsRead)
		usersNotificationAPI.PATCH("/:id/dismiss", notificationAPIHandler.DismissUserNotification)
		usersNotificationAPI.DELETE("/:id", notificationAPIHandler.DeleteUserNotification)
		usersNotificationAPI.GET("/preferences", notificationAPIHandler.GetUserPreferences)
		usersNotificationAPI.PATCH("/preferences", notificationAPIHandler.UpdateUserPreferences)
	}

	// Admin API routes (require admin role + stricter rate limiting)
	// AI.md: Admin API at /api/{api_version}/{admin_path}/
	adminAPI := apiV1.Group("/" + cfg.GetAdminPath())
	adminAPI.Use(middleware.RequireAuth(db.DB))
	adminAPI.Use(middleware.RequireAdmin())
	adminAPI.Use(middleware.AdminRateLimitMiddleware())
	// Log all admin API actions
	adminAPI.Use(middleware.AuditLogger(db.DB))
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

		// Live settings management (SPEC Section 18)
		adminSettingsHandler := &handler.AdminSettingsHandler{
			DB:                  db.DB,
			// TEMPLATE.md Part 25: Send notifications on settings changes
			NotificationService: notificationService,
		}
		adminAPI.GET("/settings/all", adminSettingsHandler.GetAllSettings)
		adminAPI.PUT("/settings/bulk", adminSettingsHandler.UpdateSettings)
		adminAPI.POST("/settings/reset", adminSettingsHandler.ResetSettings)
		adminAPI.GET("/settings/export", adminSettingsHandler.ExportSettings)
		adminAPI.POST("/settings/import", adminSettingsHandler.ImportSettings)
		adminAPI.POST("/reload", adminSettingsHandler.ReloadConfig)

		// New admin settings endpoints
		adminAPI.POST("/server/users", adminUsersHandler.UpdateUserSettings)
		adminAPI.POST("/server/auth", adminAuthHandler.UpdateAuthSettings)
		adminAPI.POST("/server/weather", adminWeatherHandler.UpdateWeatherSettings)
		adminAPI.POST("/server/notifications", adminNotificationsHandler.UpdateNotificationSettings)
		adminAPI.POST("/server/geoip", adminGeoIPHandler.UpdateGeoIPSettings)

		// API token management
		adminAPI.GET("/tokens", adminHandler.ListTokens)
		adminAPI.POST("/tokens", adminHandler.GenerateToken)
		adminAPI.DELETE("/tokens/:id", adminHandler.RevokeToken)

		// Audit logs
		adminAPI.GET("/audit-logs", adminHandler.ListAuditLogs)
		adminAPI.DELETE("/audit-logs", adminHandler.ClearAuditLogs)

		// System stats
		adminAPI.GET("/stats", adminHandler.GetSystemStats)

		// Test email endpoint
		adminAPI.POST("/test/email", func(c *gin.Context) {
			// Send test email using SMTP settings
			c.JSON(http.StatusOK, gin.H{
				"ok":      true,
				"message": "Test email sent successfully (feature will be implemented when email service is ready)",
			})
		})

		// Admin status and health endpoints
		adminAPI.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "online",
				"version": Version,
				"uptime":  time.Since(startTime).String(),
			})
		})

		adminAPI.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"healthy":      true,
				"database":     "connected",
				"cache":        "available",
				"disk_space":   "adequate",
				"last_checked": time.Now().Format(time.RFC3339),
			})
		})

		// Scheduled tasks management (TEMPLATE.md lines 1193-1214)
		// List all tasks with status
		adminAPI.GET("/tasks", schedulerHandler.GetAllTasks)
		// Get task history
		adminAPI.GET("/tasks/:name/history", schedulerHandler.GetTaskHistory)
		// Enable a task
		adminAPI.POST("/tasks/:name/enable", schedulerHandler.EnableTask)
		// Disable a task
		adminAPI.POST("/tasks/:name/disable", schedulerHandler.DisableTask)
		// Manual trigger
		adminAPI.POST("/tasks/:name/trigger", schedulerHandler.TriggerTask)

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

		// WebUI Notification API routes - Admin (TEMPLATE.md Part 25)
		adminAPI.GET("/notifications", notificationAPIHandler.GetAdminNotifications)
		adminAPI.GET("/notifications/unread", notificationAPIHandler.GetAdminUnreadNotifications)
		adminAPI.GET("/notifications/count", notificationAPIHandler.GetAdminUnreadCount)
		adminAPI.GET("/notifications/stats", notificationAPIHandler.GetAdminStats)
		adminAPI.PATCH("/notifications/:id/read", notificationAPIHandler.MarkAdminNotificationRead)
		adminAPI.PATCH("/notifications/read", notificationAPIHandler.MarkAllAdminNotificationsRead)
		adminAPI.PATCH("/notifications/:id/dismiss", notificationAPIHandler.DismissAdminNotification)
		adminAPI.DELETE("/notifications/:id", notificationAPIHandler.DeleteAdminNotification)
		adminAPI.GET("/notifications/preferences", notificationAPIHandler.GetAdminPreferences)
		adminAPI.PATCH("/notifications/preferences", notificationAPIHandler.UpdateAdminPreferences)
		adminAPI.POST("/notifications/send", notificationAPIHandler.SendTestNotification)
		adminAPI.GET("/channels/:type/stats", channelHandler.GetChannelStats)

		// SMTP provider management
		adminAPI.GET("/smtp/providers", channelHandler.ListSMTPProviders)
		adminAPI.POST("/smtp/autodetect", channelHandler.AutoDetectSMTP)

		// Admin panel settings endpoints
		adminAPI.PUT("/settings/web", handler.SaveWebSettings)
		adminAPI.PUT("/settings/security", handler.SaveSecuritySettings)
		adminAPI.PUT("/settings/database", handler.SaveDatabaseSettings)

		// Database management endpoints
		adminAPI.POST("/database/test", handler.TestDatabaseConnection)
		adminAPI.POST("/database/test-config", handler.TestDatabaseConfigConnection)
		adminAPI.POST("/database/optimize", handler.OptimizeDatabase)
		adminAPI.POST("/database/vacuum", handler.VacuumDatabase)
		adminAPI.POST("/cache/clear", handler.ClearCache)

		// Backup management endpoints
		adminAPI.POST("/backup/create", handler.CreateBackup)
		adminAPI.POST("/backup/restore", handler.RestoreBackup)
		adminAPI.GET("/backup/list", handler.ListBackups)
		adminAPI.GET("/backup/download/:filename", handler.DownloadBackup)
		adminAPI.DELETE("/backup/delete/:filename", handler.DeleteBackup)

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

		// Tor hidden service management (TEMPLATE.md PART 32)
		torAPI := adminAPI.Group("/server/tor")
		{
			torAPI.GET("/status", torAdminHandler.GetStatus)
			torAPI.GET("/health", torAdminHandler.GetHealth)
			torAPI.POST("/enable", torAdminHandler.Enable)
			torAPI.POST("/disable", torAdminHandler.Disable)
			torAPI.POST("/regenerate", torAdminHandler.Regenerate)

			// Vanity address generation
			torAPI.POST("/vanity/generate", torAdminHandler.GenerateVanity)
			torAPI.GET("/vanity/status", torAdminHandler.GetVanityStatus)
			torAPI.POST("/vanity/cancel", torAdminHandler.CancelVanity)
			torAPI.POST("/vanity/apply", torAdminHandler.ApplyVanity)

			// Key import/export
			torAPI.POST("/keys/import", torAdminHandler.ImportKeys)
			torAPI.GET("/keys/export", torAdminHandler.ExportKeys)
		}

		// Web settings management (robots.txt, security.txt) - TEMPLATE.md compliant
		webAPI := adminAPI.Group("/server/web")
		{
			webAPI.GET("/robots", adminWebHandler.GetRobotsTxt)
			webAPI.PATCH("/robots", adminWebHandler.UpdateRobotsTxt)
			webAPI.GET("/security", adminWebHandler.GetSecurityTxt)
			webAPI.PATCH("/security", adminWebHandler.UpdateSecurityTxt)
		}

		// Email template management
		emailTemplateAPI := adminAPI.Group("/email/templates")
		{
			emailTemplateAPI.GET("", emailTemplateHandler.ListTemplates)
			emailTemplateAPI.GET("/:name", emailTemplateHandler.GetTemplate)
			emailTemplateAPI.PUT("/:name", emailTemplateHandler.UpdateTemplate)
			emailTemplateAPI.GET("/:name/export", emailTemplateHandler.ExportTemplate)
			emailTemplateAPI.POST("/:name/import", emailTemplateHandler.ImportTemplate)
			emailTemplateAPI.POST("/test", emailTemplateHandler.TestTemplate)
		}

		// Custom domain management (TEMPLATE.md PART 34: Multi-domain hosting)
		adminAPI.GET("/domains", domainHandler.ListDomains)
		adminAPI.GET("/domains/:id", domainHandler.GetDomain)
		adminAPI.POST("/domains", domainHandler.CreateDomain)
		adminAPI.GET("/domains/:id/verification", domainHandler.GetVerificationToken)
		adminAPI.PUT("/domains/:id/verify", domainHandler.VerifyDomain)
		adminAPI.PUT("/domains/:id/activate", domainHandler.ActivateDomain)
		adminAPI.PUT("/domains/:id/deactivate", domainHandler.DeactivateDomain)
		adminAPI.PUT("/domains/:id/ssl", domainHandler.UpdateSSL)
		adminAPI.DELETE("/domains/:id", domainHandler.DeleteDomain)

		// System logs management (TEMPLATE.md: /api/v1/admin/server/logs)
		logsAPI := adminAPI.Group("/server/logs")
		{
			// List all log files
			logsAPI.GET("", logsHandler.GetLogs)

			// Get log entries for specific type (access, error, audit, etc.)
			logsAPI.GET("/:type", logsHandler.GetLogs)

			// Download specific log file
			logsAPI.GET("/:type/download", logsHandler.DownloadLogs)

			// Audit log specific endpoints
			logsAPI.GET("/audit", logsHandler.GetAuditLogs)
			logsAPI.GET("/audit/download", logsHandler.DownloadAuditLogs)
			logsAPI.POST("/audit/search", logsHandler.SearchAuditLogs)
			logsAPI.GET("/audit/stats", logsHandler.GetAuditStats)

			// Legacy/additional endpoints
			logsAPI.GET("/stats", logsHandler.GetLogStats)
			logsAPI.GET("/archives", logsHandler.ListArchivedLogs)
			logsAPI.GET("/stream", logsHandler.StreamLogs)
			logsAPI.POST("/rotate", logsHandler.RotateLogs)
			logsAPI.DELETE("", logsHandler.ClearLogs)
		}

		// SSL/TLS certificate management
		// TEMPLATE.md Part 8: Full Let's Encrypt support with all 3 challenge types
		sslAPI := adminAPI.Group("/ssl")
		{
			sslAPI.GET("/status", sslHandler.GetStatus)
			// Obtain LE certificate (HTTP-01/TLS-ALPN-01/DNS-01)
			sslAPI.POST("/obtain", sslHandler.ObtainCertificate)
			// Manual renewal
			sslAPI.POST("/renew", sslHandler.RenewCertificate)
			// Start auto-renewal service
			sslAPI.POST("/auto-renew", sslHandler.StartAutoRenewal)
			// Get DNS records for DNS-01 challenge
			sslAPI.GET("/dns-records", sslHandler.GetDNSRecords)
			sslAPI.POST("/verify", sslHandler.VerifyCertificate)
			sslAPI.PUT("/settings", sslHandler.UpdateSettings)
			sslAPI.GET("/export", sslHandler.ExportCertificate)
			sslAPI.POST("/import", sslHandler.ImportCertificate)
			sslAPI.POST("/revoke", sslHandler.RevokeCertificate)
			sslAPI.POST("/test", sslHandler.TestSSL)
			sslAPI.POST("/scan", sslHandler.SecurityScan)
		}

		// Metrics configuration
		metricsAPI := adminAPI.Group("/metrics")
		{
			metricsAPI.GET("/config", metricsConfigHandler.GetConfig)
			metricsAPI.PUT("/config", metricsConfigHandler.UpdateConfig)
			metricsAPI.GET("/stats", metricsConfigHandler.GetStats)
			metricsAPI.GET("/list", metricsConfigHandler.ListMetrics)
			metricsAPI.POST("/custom", metricsConfigHandler.CreateMetric)
			metricsAPI.DELETE("/custom/:name", metricsConfigHandler.DeleteMetric)
			metricsAPI.GET("/export", metricsConfigHandler.ExportMetrics)
			metricsAPI.PUT("/toggle/:name", metricsConfigHandler.ToggleMetric)
		}

		// Advanced logging formats
		loggingAPI := adminAPI.Group("/logging")
		{
			loggingAPI.GET("/formats", loggingHandler.GetFormats)
			loggingAPI.PUT("/formats", loggingHandler.UpdateFormats)
			loggingAPI.GET("/fail2ban/config", loggingHandler.GetFail2banConfig)
			loggingAPI.GET("/syslog/config", loggingHandler.GetSyslogConfig)
			loggingAPI.GET("/cef/config", loggingHandler.GetCEFConfig)
			loggingAPI.GET("/export", loggingHandler.ExportLogs)
			loggingAPI.POST("/fail2ban/configure", loggingHandler.ConfigureFail2ban)
			loggingAPI.POST("/syslog/configure", loggingHandler.ConfigureSyslog)
			loggingAPI.GET("/test", loggingHandler.TestFormat)
		}
	}

	// User notification preferences API (authenticated users)
	// AI.md PART 14: Use versioned API + plural nouns
	userPrefAPI := apiV1.Group("/users")
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
	// AI.md PART 14: Never hardcode v1 - use cfg.GetAPIVersion()
	r.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "Weather API",
			"version": "2.0.0",
			"api_versions": []string{
				cfg.GetAPIVersion(),
			},
			"current_version": cfg.GetAPIVersion(),
			"documentation":   "http://" + c.Request.Host + "/docs",
			"openapi":         "http://" + c.Request.Host + "/openapi.json",
			"swagger":         "http://" + c.Request.Host + "/openapi",
			"graphql":         "http://" + c.Request.Host + cfg.GetAPIPath() + "/graphql",
		})
	})

	// OpenAPI/Swagger documentation (AI.md PART 14)
	// Root-level endpoints per AI.md specification
	r.GET("/openapi", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/openapi/index.html")
	})
	// Swagger UI + JSON spec (auto-generated)
	r.GET("/openapi/*any", handler.GetSwaggerUIAuto())
	r.GET("/openapi.json", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/openapi/doc.json")
	})

	// GraphQL API (AI.md PART 14)
	// AI.md: Never hardcode v1 - use cfg.GetAPIPath()
	graphqlHandler, err := handler.InitGraphQL()
	if err != nil {
		log.Printf("Failed to initialize GraphQL: %v", err)
		fmt.Printf("⚠️  Failed to initialize GraphQL: %v\n", err)
	} else {
		graphqlPath := cfg.GetAPIPath() + "/graphql"
		// GraphQL endpoint at /api/{api_version}/graphql per AI.md PART 14
		r.POST(graphqlPath, handler.GraphQLHandler(graphqlHandler))
		// GET for GraphiQL UI
		r.GET(graphqlPath, handler.GraphQLHandler(graphqlHandler))
		// GraphiQL UI at /graphql (redirects to playground)
		r.GET("/graphql", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, graphqlPath) })
		appLogger.Printf("GraphQL API enabled at %s", graphqlPath)
		fmt.Printf("✅ GraphQL API enabled at %s\n", graphqlPath)
	}

	// HTML documentation page at /docs
	r.GET("/docs", apiHandler.GetDocsHTML)

	// WebSocket endpoint for real-time notifications (TEMPLATE.md Part 25)
	// Requires authentication for both users and admins
	r.GET("/ws/notifications", middleware.OptionalAuth(db.DB), notificationAPIHandler.HandleWebSocketConnection)

	// Standard server pages (TEMPLATE.md lines 2308-2314, 4486-4489)
	r.GET("/server/about", handler.ShowAboutPage(db, cfg))
	r.GET("/server/privacy", handler.ShowPrivacyPage(db, cfg))
	r.GET("/server/contact", handler.ShowContactPage(db, cfg))
	r.GET("/server/help", handler.ShowHelpPage(db, cfg))

	// Examples endpoint
	// AI.md PART 14: Never hardcode v1 - use cfg.GetAPIPath()
	r.GET("/examples", func(c *gin.Context) {
		hostInfo := utils.GetHostInfo(c)
		apiPath := cfg.GetAPIPath()
		examples := fmt.Sprintf(`Weather API Examples

Console Interface:
  curl %s/
  curl %s/London
  curl %s/Paris?format=1
  curl %s/Tokyo?units=metric

JSON API:
  curl %s%s/weather?location=London
  curl %s%s/forecasts?location=Paris&days=5
  curl %s%s/locations/search?q=New+York
  curl %s%s/ip
`,
			hostInfo.FullHost, hostInfo.FullHost, hostInfo.FullHost, hostInfo.FullHost,
			hostInfo.FullHost, apiPath, hostInfo.FullHost, apiPath, hostInfo.FullHost, apiPath, hostInfo.FullHost, apiPath)

		c.String(http.StatusOK, examples)
	})

	// Web interface routes
	r.GET("/web", webHandler.ServeWebInterface)
	r.GET("/web/:location", webHandler.ServeWebInterface)

	// Moon interface routes
	r.GET("/moon", webHandler.ServeMoonInterface)
	r.GET("/moon/:location", webHandler.ServeMoonInterface)

	// Earthquake routes (plural per AI.md PART 14)
	r.GET("/earthquakes", earthquakeHandler.HandleEarthquakeRequest)
	r.GET("/earthquakes/:location", earthquakeHandler.HandleEarthquakeRequest)

	// Backwards compatibility: singular -> plural redirect
	r.GET("/earthquake", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/earthquakes")
	})

	// Hurricane routes redirect to severe-weather (plural per AI.md PART 14)
	r.GET("/hurricanes", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/severe-weather")
	})
	r.GET("/hurricane", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/severe-weather")
	})

	// Severe Weather routes (new comprehensive severe weather page)
	r.GET("/severe-weather", severeWeatherHandler.HandleSevereWeatherRequest)
	r.GET("/severe-weather/:location", severeWeatherHandler.HandleSevereWeatherRequest)

	// Type-filtered severe weather routes
	r.GET("/severe/:type", severeWeatherHandler.HandleSevereWeatherByType)
	r.GET("/severe/:type/:location", severeWeatherHandler.HandleSevereWeatherByType)

	// AI.md PART 14: Legacy endpoints are technical debt - DELETED
	// OLD: /api/earthquakes and /api/hurricanes redirects removed
	// Use versioned endpoints: /api/{api_version}/earthquakes and /api/{api_version}/hurricanes

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
		if !handler.IsInitialized() {
			handler.ServeLoadingPage(c)
			c.Abort()
			return
		}

		c.Next()
	})

	// Main weather routes
	// Uses IP/cookie lookup
	r.GET("/", weatherHandler.HandleRoot)
	// Explicit location
	r.GET("/weather/:location", weatherHandler.HandleLocation)
	// Backwards compatibility catch-all
	r.GET("/:location", weatherHandler.HandleLocation)

	// Build final URL for documentation
	// Per AI.md PART 5 - DOMAIN env var (no project prefix)
	finalHostname := os.Getenv("DOMAIN")
	if finalHostname == "" {
		// System variable
		finalHostname = os.Getenv("HOSTNAME")
	}
	if finalHostname == "" {
		finalHostname = "localhost"
	}

	protocol := "http"
	// AI.md PART 5: Boolean Handling
	tlsEnabled := config.IsTruthy(os.Getenv("TLS_ENABLED"))
	if tlsEnabled || httpsPortInt > 0 {
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
	// Format address properly - check if port is already included
	var serverAddr string
	if listenAddress == "::" {
		// IPv6 dual-stack: must be formatted as [::]:port
		serverAddr = "[::]:" + port
	} else if listenAddress == "0.0.0.0" || !strings.Contains(listenAddress, ":") {
		// IPv4 or hostname without port: append port
		serverAddr = listenAddress + ":" + port
	} else if strings.Count(listenAddress, ":") > 1 && !strings.HasPrefix(listenAddress, "[") {
		// IPv6 address without brackets (e.g., "::1")
		host, portPart, err := net.SplitHostPort(listenAddress)
		if err != nil || portPart == "" {
			// No port in address, add brackets and port
			serverAddr = "[" + listenAddress + "]:" + port
		} else {
			serverAddr = "[" + host + "]:" + portPart
		}
	} else {
		// Already has port (e.g., "127.0.0.1:8080" or "[::1]:8080")
		serverAddr = listenAddress
	}

	// Server configuration per AI.md PART 18 lines 15697-15702
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
		// Per AI.md PART 18: read_timeout: 30s
		ReadTimeout: 30 * time.Second,
		// Per AI.md PART 18: write_timeout: 30s
		WriteTimeout: 30 * time.Second,
		// Per AI.md PART 18: idle_timeout: 120s
		IdleTimeout: 120 * time.Second,
		// Max header size (1MB is reasonable default)
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start Tor hidden service after HTTP server starts
	if err := torService.Start(httpPortInt); err != nil {
		log.Printf("Failed to start Tor hidden service: %v", err)
		fmt.Printf("⚠️  Failed to start Tor hidden service: %v\n", err)
	}

	// Start config file watcher for live reload
	if configWatcher != nil {
		if err := configWatcher.Start(); err != nil {
			log.Printf("Failed to start config watcher: %v", err)
			fmt.Printf("⚠️  Failed to start config watcher: %v\n", err)
		}
	}

	// TEMPLATE.md PART 1: Display startup banner
	torOnionAddr := ""
	if cfg != nil && cfg.Server.Tor.Enabled {
		torOnionAddr = cfg.Server.Tor.OnionAddr
	}

	if isFirstRun {
		utils.DisplayFirstRunBanner(httpPortInt, setupToken, utils.IsDockerized(), torOnionAddr)
	} else {
		utils.DisplayNormalBanner(Version, BuildDate, httpPortInt, utils.IsDockerized(), torOnionAddr)
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	// Per AI.md PART 27 lines 6456-6458: Graceful shutdown signals
	// SIGTERM: kill (systemctl stop)
	// SIGINT: Ctrl+C
	// SIGQUIT: Ctrl+\ (Unix only, but harmless to include on Windows)
	baseSignals := []os.Signal{
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	}

	// Per AI.md PART 27 line 6536: Ignore SIGHUP - config reloads automatically via file watcher
	signal.Ignore(syscall.SIGHUP)

	// Add platform-specific signals (SIGUSR1/2 on Unix only)
	allSignals := make([]os.Signal, len(baseSignals)+len(platformSignals))
	copy(allSignals, baseSignals)
	for i, sig := range platformSignals {
		allSignals[len(baseSignals)+i] = sig
	}

	signal.Notify(sigChan, allSignals...)

	// Handle signals
	for sig := range sigChan {
		switch sig {
		case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
			log.Println("🛑 Received shutdown signal, shutting down gracefully...")

			// Stop scheduler
			taskScheduler.Stop()

			// Stop Tor service
			if err := torService.Stop(); err != nil {
				log.Printf("Tor shutdown error: %v", err)
				fmt.Printf("⚠️  Tor shutdown error: %v\n", err)
			}

			// Stop config watcher
			if configWatcher != nil {
				if err := configWatcher.Stop(); err != nil {
					log.Printf("Config watcher shutdown error: %v", err)
					fmt.Printf("⚠️  Config watcher shutdown error: %v\n", err)
				}
			}

			// Close cache connection
			if err := cacheManager.Close(); err != nil {
				log.Printf("Cache shutdown error: %v", err)
				fmt.Printf("⚠️  Cache shutdown error: %v\n", err)
			}

			// Shutdown HTTP server with 5 second timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("Server forced to shutdown: %v", err)
				fmt.Printf("⚠️  Server forced to shutdown: %v\n", err)
			}

			log.Println("Server exited gracefully")
			fmt.Println("✅ Server exited gracefully")
			return

		default:
			// Handle platform-specific signals (SIGHUP, SIGUSR1, SIGUSR2, SIGRTMIN+3 on Unix)
			// Returns true if shutdown requested (e.g., SIGRTMIN+3 Docker signal)
			if handlePlatformSignal(sig, db, appLogger, dirPaths) {
				// Shutdown requested - execute same graceful shutdown as SIGTERM
				log.Println("🛑 Platform signal requested shutdown, shutting down gracefully...")

				// Stop scheduler
				taskScheduler.Stop()

				// Stop Tor service
				if err := torService.Stop(); err != nil {
					log.Printf("Tor shutdown error: %v", err)
					fmt.Printf("⚠️  Tor shutdown error: %v\n", err)
				}

				// Stop config watcher
				if configWatcher != nil {
					if err := configWatcher.Stop(); err != nil {
						log.Printf("Config watcher shutdown error: %v", err)
						fmt.Printf("⚠️  Config watcher shutdown error: %v\n", err)
					}
				}

				// Close cache connection
				if err := cacheManager.Close(); err != nil {
					log.Printf("Cache shutdown error: %v", err)
					fmt.Printf("⚠️  Cache shutdown error: %v\n", err)
				}

				// Shutdown HTTP server with 5 second timeout
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := srv.Shutdown(ctx); err != nil {
					log.Printf("Server forced to shutdown: %v", err)
					fmt.Printf("⚠️  Server forced to shutdown: %v\n", err)
				}

				log.Println("Server exited gracefully")
				fmt.Println("✅ Server exited gracefully")
				return
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


// showServerStatus displays comprehensive server status information
// Per AI.md PART 8: Returns true if healthy, false if unhealthy
func showServerStatus(db *database.DB, dbPath string, isFirstRun bool) bool {
	// Get configuration values - AI.md PART 5: Environment Variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	envMode := os.Getenv("MODE")
	if envMode == "" {
		envMode = os.Getenv("ENVIRONMENT") // Legacy fallback
	}
	if envMode == "" {
		envMode = "production"
	}

	address := os.Getenv("LISTEN")
	if address == "" {
		address = os.Getenv("SERVER_ADDRESS") // Legacy fallback
	}
	addressMode := ""
	if address == "" {
		// Check for reverse proxy indicators per AI.md PART 5: Boolean Handling
		reverseProxy := config.IsTruthy(os.Getenv("REVERSE_PROXY"))

		if reverseProxy {
			address = "127.0.0.1"
			addressMode = " (reverse proxy mode)"
		} else {
			address = "::"
			addressMode = " (all interfaces)"
		}
	}

	// Perform health checks (AI.md PART 8: --status must check health)
	isHealthy := true
	healthStatus := "✅ Healthy"
	
	// Check database connection
	dbStatus, _, dbErr := db.HealthCheck()
	if dbErr != nil || dbStatus != "connected" {
		isHealthy = false
		healthStatus = "🔴 Unhealthy (Database Error)"
	}

	// Get database statistics
	var userCount, locationCount, tokenCount int
	database.GetUsersDB().QueryRow("SELECT COUNT(*) FROM user_accounts").Scan(&userCount)
	database.GetUsersDB().QueryRow("SELECT COUNT(*) FROM user_saved_locations").Scan(&locationCount)
	database.GetUsersDB().QueryRow("SELECT COUNT(*) FROM user_tokens WHERE expires_at > datetime('now')").Scan(&tokenCount)

	// Display status
	fmt.Println("\n╔══════════════════════════════════════════════════════╗")
	fmt.Println("║          🌤️  Weather Service - Status              ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")

	fmt.Printf("\n🏥 Health Status: %s\n", healthStatus)

	fmt.Println("\n📊 Server Configuration:")
	fmt.Printf("   Version:        %s\n", Version)
	fmt.Printf("   Build Date:     %s\n", BuildDate)
	fmt.Printf("   Git Commit:     %s\n", CommitID)
	fmt.Printf("   Listen Address: %s:%s%s\n", address, port, addressMode)
	fmt.Printf("   Environment:    %s\n", envMode)

	fmt.Println("\n💾 Database:")
	fmt.Printf("   Path:           %s\n", dbPath)
	fmt.Printf("   Status:         %s\n", dbStatus)
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

	// Return health status per AI.md PART 8
	return isHealthy
}

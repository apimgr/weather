package config

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration from server.yml per AI.md PART 4
type Config struct {
	Server ServerConfig `yaml:"server"`
	Web    WebConfig    `yaml:"web"`
}

// ServerConfig represents server-specific configuration per AI.md PART 4
type ServerConfig struct {
	// Port: random 64xxx on first run, then persisted
	Port     interface{}        `yaml:"port"` // int or string (for dual port "8090,8443")
	FQDN     string             `yaml:"fqdn"`
	Address  string             `yaml:"address"` // Default: [::]
	Mode     string             `yaml:"mode"`    // production or development
	Branding BrandingConfig     `yaml:"branding"`
	SEO      SEOConfig          `yaml:"seo"`
	User     string             `yaml:"user"`
	Group    string             `yaml:"group"`
	PIDFile  interface{}        `yaml:"pidfile"` // bool or string path
	Daemonize bool              `yaml:"daemonize"`
	Admin    AdminConfig        `yaml:"admin"`
	SSL      SSLConfig          `yaml:"ssl"`
	Scheduler SchedulerConfig   `yaml:"scheduler"`
	RateLimit RateLimitConfig   `yaml:"rate_limit"`
	Database DatabaseConfig     `yaml:"database"`
	Maintenance MaintenanceConfig `yaml:"maintenance"`
	Notifications NotificationConfig `yaml:"notifications"`
	Tor      TorConfig          `yaml:"tor"`
	Features FeatureConfig      `yaml:"features"`
}

// AdminConfig represents admin panel configuration per AI.md PART 4
type AdminConfig struct {
	Email string `yaml:"email"`
	// Note: username, password, and token stored in database, not config file
}

// SSLConfig represents SSL/TLS configuration per AI.md PART 4
type SSLConfig struct {
	Enabled    bool              `yaml:"enabled"`
	Cert       string            `yaml:"cert"`        // Manual cert path (optional)
	Key        string            `yaml:"key"`         // Manual key path (optional)
	MinVersion string            `yaml:"min_version"` // TLS1.2, TLS1.3
	LetsEncrypt LetsEncryptConfig `yaml:"letsencrypt"`
}

// LetsEncryptConfig represents Let's Encrypt configuration per AI.md PART 4
type LetsEncryptConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Email     string `yaml:"email"`
	Challenge string `yaml:"challenge"` // http-01, tls-alpn-01, dns-01
	Staging   bool   `yaml:"staging"`   // Use staging server for testing
}

// SchedulerConfig represents scheduler configuration per AI.md PART 4
type SchedulerConfig struct {
	Enabled bool                   `yaml:"enabled"`
	Tasks   map[string]SchedulerTask `yaml:"tasks"`
}

// SchedulerTask represents a scheduled task per AI.md PART 4
type SchedulerTask struct {
	Enabled      bool   `yaml:"enabled"`
	Schedule     string `yaml:"schedule"`      // Cron format or @hourly/@daily
	RetryOnFail  bool   `yaml:"retry_on_fail"`
	RetryDelay   string `yaml:"retry_delay"`   // e.g., "1h"
	MaxAge       string `yaml:"max_age"`       // e.g., "30d" (for log_rotation)
	MaxSize      string `yaml:"max_size"`      // e.g., "100MB" (for log_rotation)
	Retention    int    `yaml:"retention"`     // e.g., 4 (for backup)
	RenewBefore  string `yaml:"renew_before"`  // e.g., "7d" (for ssl_renewal)
}

// DatabaseConfig represents database configuration per AI.md PART 4
type DatabaseConfig struct {
	Driver   string `yaml:"driver"`   // file, sqlite, postgres, mysql, mariadb, mssql, mongodb
	Host     string `yaml:"host"`     // For remote databases
	Port     int    `yaml:"port"`     // For remote databases
	Name     string `yaml:"name"`     // Database name
	Username string `yaml:"username"` // For remote databases
	Password string `yaml:"password"` // For remote databases
	SSLMode  string `yaml:"sslmode"`  // For PostgreSQL
}

// MaintenanceConfig represents maintenance mode configuration per AI.md PART 4
type MaintenanceConfig struct {
	SelfHealing SelfHealingConfig `yaml:"self_healing"`
	Cleanup     CleanupConfig     `yaml:"cleanup"`
	Notify      NotifyConfig      `yaml:"notify"`
	Backup      BackupConfig      `yaml:"backup"`
}

// SelfHealingConfig represents self-healing settings per AI.md PART 4
type SelfHealingConfig struct {
	Enabled       bool `yaml:"enabled"`
	RetryInterval int  `yaml:"retry_interval"` // seconds between retry attempts
	MaxAttempts   int  `yaml:"max_attempts"`   // 0 = unlimited
}

// CleanupConfig represents auto-cleanup thresholds per AI.md PART 4
type CleanupConfig struct {
	DiskThreshold     int `yaml:"disk_threshold"`      // Start cleanup when disk > X% full
	LogRetentionDays  int `yaml:"log_retention_days"`  // Delete logs older than X days
	BackupKeepCount   int `yaml:"backup_keep_count"`   // Keep last X backups
}

// NotifyConfig represents maintenance notification settings per AI.md PART 4
type NotifyConfig struct {
	OnEnter bool `yaml:"on_enter"` // Notify when entering maintenance mode
	OnExit  bool `yaml:"on_exit"`  // Notify when exiting maintenance mode
}

// BackupConfig represents backup encryption settings per AI.md PART 24
type BackupConfig struct {
	Encryption BackupEncryptionConfig `yaml:"encryption"`
}

// BackupEncryptionConfig represents backup encryption settings per AI.md PART 24
type BackupEncryptionConfig struct {
	Enabled bool   `yaml:"enabled"` // true if password was set during setup
	Hint    string `yaml:"hint"`    // Optional password hint (e.g., "First pet's name + year")
	// Password is NEVER stored - derived on-demand
}

// RateLimitConfig represents rate limiting configuration per AI.md PART 4
type RateLimitConfig struct {
	Enabled  bool `yaml:"enabled"`
	Requests int  `yaml:"requests"` // Requests per window
	Window   int  `yaml:"window"`   // Window in seconds
}

// BrandingConfig represents branding configuration per AI.md PART 4
type BrandingConfig struct {
	Title       string `yaml:"title"`
	Tagline     string `yaml:"tagline"`
	Description string `yaml:"description"`
}

// SEOConfig represents SEO configuration per AI.md PART 4
type SEOConfig struct {
	Keywords []string `yaml:"keywords"` // Array of keywords
}

// NotificationConfig represents notification settings per AI.md PART 4
type NotificationConfig struct {
	Enabled        bool `yaml:"enabled"`
	EmailEnabled   bool `yaml:"email_enabled"`
	WebhookEnabled bool `yaml:"webhook_enabled"`
}

// WebConfig represents web-specific configuration per AI.md PART 4
// WebConfig represents web interface configuration
type WebConfig struct {
	UI          UIConfig `yaml:"ui"`
	CORS        string   `yaml:"cors"`        // CORS setting, e.g., "*"
	RobotsTxt   string   `yaml:"robots_txt"`  // Custom robots.txt content
	SecurityTxt string   `yaml:"security_txt"` // Custom security.txt content
}

// UIConfig represents UI configuration per AI.md PART 4
type UIConfig struct {
	Theme string `yaml:"theme"` // dark, light
}

// TorConfig represents Tor hidden service configuration per AI.md PART 4
type TorConfig struct {
	Enabled   bool   `yaml:"enabled"`
	OnionAddr string `yaml:"onion_addr"`
}

// FeatureConfig represents feature toggles per AI.md PART 4
type FeatureConfig struct {
	Earthquakes   bool `yaml:"earthquakes"`
	Hurricanes    bool `yaml:"hurricanes"`
	MoonPhases    bool `yaml:"moon_phases"`
	SevereWeather bool `yaml:"severe_weather"`
	AuditLog      bool `yaml:"audit_log"`
}

// randomPort returns a random port in the 64000-64999 range per AI.md PART 4
func randomPort() int {
	rand.Seed(time.Now().UnixNano())
	return 64000 + rand.Intn(1000)
}

// getDefaultFQDN returns the default FQDN (hostname) per AI.md PART 4
func getDefaultFQDN() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}

// LoadConfig loads configuration from server.yml per AI.md PART 4
func LoadConfig() (*Config, error) {
	// Get hostname for defaults
	hostname := getDefaultFQDN()
	adminEmail := fmt.Sprintf("admin@%s", hostname)

	// Default config with sane defaults per AI.md PART 4
	cfg := &Config{
		Server: ServerConfig{
			Port:      randomPort(), // Random 64xxx on first run
			FQDN:      hostname,
			Address:   "[::]", // All interfaces IPv4/IPv6
			Mode:      "production",
			User:      "{auto}",
			Group:     "{auto}",
			PIDFile:   true,
			Daemonize: false,
			Branding: BrandingConfig{
				Title:       "weather",
				Tagline:     "",
				Description: "",
			},
			SEO: SEOConfig{
				Keywords: []string{},
			},
			Admin: AdminConfig{
				Email: adminEmail,
			},
			SSL: SSLConfig{
				Enabled:    false,
				Cert:       "",
				Key:        "",
				MinVersion: "TLS1.2",
				LetsEncrypt: LetsEncryptConfig{
					Enabled:   false,
					Email:     adminEmail,
					Challenge: "http-01",
					Staging:   false,
				},
			},
			Scheduler: SchedulerConfig{
				Enabled: true,
				Tasks: map[string]SchedulerTask{
					"geoip_update": {
						Enabled:     true,
						Schedule:    "0 3 * * 0", // Weekly Sunday 3am
						RetryOnFail: true,
						RetryDelay:  "1h",
					},
					"blocklist_update": {
						Enabled:     true,
						Schedule:    "0 4 * * *", // Daily 4am
						RetryOnFail: true,
						RetryDelay:  "1h",
					},
					"cve_update": {
						Enabled:     true,
						Schedule:    "0 5 * * *", // Daily 5am
						RetryOnFail: true,
						RetryDelay:  "1h",
					},
					"log_rotation": {
						Enabled:  true,
						Schedule: "0 0 * * *", // Daily midnight
						MaxAge:   "30d",
						MaxSize:  "100MB",
					},
					"session_cleanup": {
						Enabled:  true,
						Schedule: "@hourly",
					},
					"backup": {
						Enabled:   true,
						Schedule:  "0 2 * * *", // Daily 2am
						Retention: 4,
					},
					"ssl_renewal": {
						Enabled:     true,
						Schedule:    "0 3 * * *", // Daily 3am
						RenewBefore: "7d",
					},
					"health_check": {
						Enabled:  true,
						Schedule: "*/5 * * * *", // Every 5 minutes
					},
					"tor_health": {
						Enabled:  true,
						Schedule: "*/10 * * * *", // Every 10 minutes
					},
				},
			},
			RateLimit: RateLimitConfig{
				Enabled:  true,
				Requests: 120,
				Window:   60,
			},
			Database: DatabaseConfig{
				Driver: "file",
			},
			Maintenance: MaintenanceConfig{
				SelfHealing: SelfHealingConfig{
					Enabled:       true,
					RetryInterval: 30,
					MaxAttempts:   0, // Unlimited
				},
				Cleanup: CleanupConfig{
					DiskThreshold:    90,
					LogRetentionDays: 7,
					BackupKeepCount:  5,
				},
				Notify: NotifyConfig{
					OnEnter: true,
					OnExit:  true,
				},
				Backup: BackupConfig{
					Encryption: BackupEncryptionConfig{
						Enabled: false, // Set to true during setup wizard if password provided
						Hint:    "",    // Optional password hint
					},
				},
			},
			Notifications: NotificationConfig{
				Enabled:      true,
				EmailEnabled: true,
			},
			Tor: TorConfig{
				Enabled: false,
			},
			Features: FeatureConfig{
				Earthquakes:   true,
				Hurricanes:    true,
				MoonPhases:    true,
				SevereWeather: true,
				AuditLog:      true,
			},
		},
		Web: WebConfig{
			UI: UIConfig{
				Theme: "dark",
			},
			CORS: "*",
		},
	}

	// Try to load from server.yml
	configPath := findConfigFile()
	if configPath == "" {
		// No config file found - create it on first run per AI.md PART 4
		configPath = getConfigPath()
		if err := createDefaultConfig(cfg, configPath); err != nil {
			// Log error but continue with defaults
			fmt.Fprintf(os.Stderr, "Warning: Could not create config file: %v\n", err)
		}
		return cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, err
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// getConfigPath returns the config file path based on user privileges per AI.md PART 4
func getConfigPath() string {
	// Check if running as root
	if os.Geteuid() == 0 {
		// Root user: /etc/apimgr/weather/server.yml
		return "/etc/apimgr/weather/server.yml"
	}

	// Regular user: ~/.config/apimgr/weather/server.yml
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home not found
		return "server.yml"
	}
	return filepath.Join(home, ".config", "apimgr", "weather", "server.yml")
}

// findConfigFile searches for server.yml in common locations per AI.md PART 4
func findConfigFile() string {
	// Priority 1: Environment variable CONFIG_DIR
	if configDir := os.Getenv("CONFIG_DIR"); configDir != "" {
		path := filepath.Join(configDir, "server.yml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Priority 2: Standard location based on user
	standardPath := getConfigPath()
	if _, err := os.Stat(standardPath); err == nil {
		return standardPath
	}

	// Priority 3: Check for server.yaml (migrate to server.yml per AI.md PART 4)
	yamlPath := filepath.Join(filepath.Dir(standardPath), "server.yaml")
	if _, err := os.Stat(yamlPath); err == nil {
		// Auto-migrate from .yaml to .yml
		if err := os.Rename(yamlPath, standardPath); err == nil {
			return standardPath
		}
		return yamlPath
	}

	return ""
}

// createDefaultConfig creates a default server.yml file per AI.md PART 4
func createDefaultConfig(cfg *Config, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comment
	header := `# =============================================================================
# Weather Service Configuration (AI.md PART 4)
# =============================================================================
# This file was auto-generated on first run with sane defaults.
# Edit as needed and restart the service to apply changes.
# =============================================================================

`
	fullData := append([]byte(header), data...)

	// Write to file
	if err := os.WriteFile(path, fullData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Created default configuration: %s\n", path)
	return nil
}

// Global config instance for handler access
var globalConfig *Config

// SetGlobalConfig sets the global config instance
func SetGlobalConfig(cfg *Config) {
	globalConfig = cfg
}

// GetGlobalConfig returns the global config instance
func GetGlobalConfig() *Config {
	return globalConfig
}

// SaveConfig saves the current configuration to server.yml per AI.md PART 4
func SaveConfig(cfg *Config) error {
	configPath := findConfigFile()
	if configPath == "" {
		configPath = getConfigPath()
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// UpdateWebRobotsTxt updates the robots.txt content in server.yml
func UpdateWebRobotsTxt(content string) error {
	cfg := GetGlobalConfig()
	if cfg == nil {
		return fmt.Errorf("global config not initialized")
	}

	cfg.Web.RobotsTxt = content
	return SaveConfig(cfg)
}

// UpdateWebSecurityTxt updates the security.txt content in server.yml
func UpdateWebSecurityTxt(content string) error {
	cfg := GetGlobalConfig()
	if cfg == nil {
		return fmt.Errorf("global config not initialized")
	}

	cfg.Web.SecurityTxt = content
	return SaveConfig(cfg)
}

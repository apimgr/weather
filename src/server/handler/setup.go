package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/apimgr/weather/src/config"
	"github.com/apimgr/weather/src/database"
	"github.com/apimgr/weather/src/paths"
	"github.com/apimgr/weather/src/utils"

	"github.com/gin-gonic/gin"
)

type SetupHandler struct {
	DB *sql.DB
}

// ShowSetupTokenEntry shows the setup token entry form
// AI.md: First step of setup - user must enter setup token displayed in console
func (h *SetupHandler) ShowSetupTokenEntry(c *gin.Context) {
	c.HTML(http.StatusOK, "page/setup_token.tmpl", gin.H{
		"Title": "Server Setup - Enter Setup Token",
	})
}

// VerifySetupToken validates the setup token and allows access to admin creation
// AI.md: Setup token stored as SHA-256 hash in {config_dir}/setup_token.txt
func (h *SetupHandler) VerifySetupToken(c *gin.Context) {
	var input struct {
		SetupToken string `json:"setup_token" form:"setup_token" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setup token is required"})
		return
	}

	// Validate setup token against stored hash
	configDir := paths.GetConfigDir()
	valid, err := utils.ValidateSetupToken(configDir, input.SetupToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setup token not found or already used"})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid setup token"})
		return
	}

	// Store validated token in session for the admin creation step
	// Use a secure session cookie to track that token was validated
	c.SetCookie("setup_token_verified", "true", 3600, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"redirect": c.Request.URL.Path[:strings.LastIndex(c.Request.URL.Path, "/")] + "/admin",
	})
}

// ShowAdminSetup shows the admin creation form (step 4)
func (h *SetupHandler) ShowAdminSetup(c *gin.Context) {
	c.HTML(http.StatusOK, "page/setup_admin.tmpl", gin.H{
		"Title": "Create Administrator Account - Weather Setup",
	})
}

// CreateAdmin creates the Primary Admin account
// AI.md: Setup creates Primary Admin, requires setup token verification
func (h *SetupHandler) CreateAdmin(c *gin.Context) {
	// Verify setup token was validated
	verified, err := c.Cookie("setup_token_verified")
	if err != nil || verified != "true" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Setup token not verified"})
		return
	}

	var input struct {
		Username        string `json:"username" form:"username"`
		Email           string `json:"email" form:"email"`
		UseRandom       bool   `json:"use_random" form:"use_random"`
		Password        string `json:"password" form:"password"`
		ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
	}

	// Accept both JSON and form data
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate email
	email := strings.TrimSpace(input.Email)
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// Set default username if empty
	username := strings.TrimSpace(input.Username)
	if username == "" {
		username = "administrator"
	}

	// Normalize username
	username = strings.ToLower(username)

	var password string
	var generatedPassword string

	if input.UseRandom {
		// Generate random password (32 characters, alphanumeric + special)
		var err error
		generatedPassword, err = generateRandomPassword(32)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate secure password"})
			return
		}
		password = generatedPassword
	} else {
		// Use custom password - must be confirmed
		if input.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
			return
		}
		// Passwords cannot start or end with whitespace
		if input.Password != strings.TrimSpace(input.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password cannot start or end with whitespace"})
			return
		}
		if len(input.Password) < 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 12 characters"})
			return
		}
		if input.Password != input.ConfirmPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
			return
		}
		password = input.Password
	}

	// Check if admin username already exists in server_admin_credentials
	var count int
	err := database.GetServerDB().QueryRow("SELECT COUNT(*) FROM server_admin_credentials WHERE username = ?", username).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Hash password using Argon2id (TEMPLATE.md Part 0 requirement)
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create administrator account in server_admin_credentials (NOT user_accounts)
	result, err := database.GetServerDB().Exec(`
		INSERT INTO server_admin_credentials (username, email, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, username, email, hashedPassword)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create administrator"})
		return
	}

	// Get the newly created admin ID
	adminID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve admin ID"})
		return
	}

	// Create admin session (auto-login) in server_admin_sessions
	sessionID, err := generateSessionID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session ID"})
		return
	}
	// 7 days
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err = database.GetServerDB().Exec(`
		INSERT INTO server_admin_sessions (id, admin_id, ip_address, user_agent, created_at, expires_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
	`, sessionID, adminID, c.ClientIP(), c.Request.UserAgent(), expiresAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	isHTTPS := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"

	// Set admin_session cookie (separate from weather_session)
	c.SetCookie(
		"admin_session",
		sessionID,
		int(7*24*time.Hour.Seconds()),
		"/",
		"",
		isHTTPS,
		true,
	)

	// Delete setup token file after successful admin creation
	// AI.md: File deleted after successful setup completion
	configDir := paths.GetConfigDir()
	if err := utils.DeleteSetupToken(configDir); err != nil {
		// Log but don't fail - admin was created successfully
		fmt.Printf("Warning: failed to delete setup token file: %v\n", err)
	}

	// Clear the setup_token_verified cookie
	c.SetCookie("setup_token_verified", "", -1, "/", "", false, true)

	response := gin.H{"ok": true}
	if generatedPassword != "" {
		// Include generated password in response (shown only once)
		response["generated_password"] = generatedPassword
		response["username"] = username
	}

	// Redirect to setup complete page (same path prefix as current request)
	basePath := c.Request.URL.Path[:strings.LastIndex(c.Request.URL.Path, "/")]
	response["redirect"] = basePath + "/complete"

	c.JSON(http.StatusOK, response)
}

// generateRandomPassword generates a cryptographically secure random password
func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#-_$"
	password := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range password {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate secure random password: %w", err)
		}
		password[i] = charset[n.Int64()]
	}
	return string(password), nil
}

// generateSessionID generates a cryptographically secure random session ID
func generateSessionID() (string, error) {
	// Generate 48 random bytes and encode as base64 (results in 64 chars)
	bytes := make([]byte, 48)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure session ID: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CompleteSetup shows the setup completion page
// AI.md: Setup is complete when Primary Admin is created
func (h *SetupHandler) CompleteSetup(c *gin.Context) {
	// Mark setup as complete in database
	_, err := database.GetServerDB().Exec(`
		INSERT INTO server_config (key, value, type, description, updated_at)
		VALUES ('setup.completed', 'true', 'bool', 'Server setup completed', datetime('now'))
		ON CONFLICT(key) DO UPDATE SET value = 'true', updated_at = datetime('now')
	`)
	if err != nil {
		fmt.Printf("Warning: failed to mark setup complete: %v\n", err)
	}

	// Get the admin path from config (derive from current URL)
	// Current URL is /{admin_path}/server/setup/complete
	path := c.Request.URL.Path
	parts := strings.Split(path, "/")
	adminPath := "/admin" // default
	if len(parts) > 1 && parts[1] != "" {
		adminPath = "/" + parts[1]
	}

	c.HTML(http.StatusOK, "page/setup_complete.tmpl", gin.H{
		"Title":     "Setup Complete - Weather",
		"AdminPath": adminPath,
	})
}

// =============================================================================
// Server Setup Wizard (Admin Only)
// =============================================================================

// ShowServerSetupWelcome shows the server setup welcome page
func (h *SetupHandler) ShowServerSetupWelcome(c *gin.Context) {
	c.HTML(http.StatusOK, "page/server_setup_welcome.tmpl", gin.H{
		"Title": "Server Setup - Weather",
	})
}

// ShowServerSetupSettings shows the server settings configuration page
func (h *SetupHandler) ShowServerSetupSettings(c *gin.Context) {
	c.HTML(http.StatusOK, "page/server_setup_settings.tmpl", gin.H{
		"Title": "Server Settings - Weather",
	})
}

// SaveServerSettings saves server configuration settings
func (h *SetupHandler) SaveServerSettings(c *gin.Context) {
	var input struct {
		ServerName          string `json:"serverName"`
		ServerDescription   string `json:"serverDescription"`
		TemperatureUnit     string `json:"temperatureUnit"`
		WindSpeedUnit       string `json:"windSpeedUnit"`
		PrecipitationUnit   string `json:"precipitationUnit"`
		RateLimitAnon       int    `json:"rateLimitAnon"`
		RateLimitAuth       int    `json:"rateLimitAuth"`
		EnableRegistration  bool   `json:"enableRegistration"`
		EnableAlerts        bool   `json:"enableAlerts"`
		EnableNotifications bool   `json:"enableNotifications"`
		EnableAuditLog      bool   `json:"enableAuditLog"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save all settings to database
	settings := map[string]string{
		"server.name":              input.ServerName,
		"server.description":       input.ServerDescription,
		"units.temperature":        input.TemperatureUnit,
		"units.wind_speed":         input.WindSpeedUnit,
		"units.precipitation":      input.PrecipitationUnit,
		"rate_limit.anonymous":     fmt.Sprintf("%d", input.RateLimitAnon),
		"rate_limit.authenticated": fmt.Sprintf("%d", input.RateLimitAuth),
		"features.registration":    boolToString(input.EnableRegistration),
		"features.alerts":          boolToString(input.EnableAlerts),
		"features.notifications":   boolToString(input.EnableNotifications),
		"features.audit_log":       boolToString(input.EnableAuditLog),
	}

	for key, value := range settings {
		_, err := database.GetServerDB().Exec(`
			INSERT INTO server_config (key, value, updated_at)
			VALUES (?, ?, datetime('now'))
			ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = datetime('now')
		`, key, value, value)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save setting: " + key})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":  true,
		"redirect": "/setup/complete",
	})
}

// ShowServerSetupComplete shows the completion page and marks setup as done
func (h *SetupHandler) ShowServerSetupComplete(c *gin.Context) {
	// Mark server setup as complete
	_, err := database.GetServerDB().Exec(`
		INSERT INTO server_config (key, value, updated_at)
		VALUES ('setup.completed', 'true', datetime('now'))
		ON CONFLICT(key) DO UPDATE SET value = 'true', updated_at = datetime('now')
	`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark setup as complete"})
		return
	}

	c.HTML(http.StatusOK, "page/server_setup_complete.tmpl", gin.H{
		"Title": "Setup Complete - Weather",
	})
}

// GetSetupStatus returns the current setup status as a healthz endpoint
func (h *SetupHandler) GetSetupStatus(c *gin.Context) {
	// Check if any users exist
	var userCount int
	err := database.GetUsersDB().QueryRow("SELECT COUNT(*) FROM user_accounts").Scan(&userCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to check setup status",
		})
		return
	}

	// No users = setup not started
	if userCount == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":      "not_started",
			"step":        0,
			"total_steps": 3,
			"message":     "Setup not started",
			"next_action": "Create first user account",
			"next_route":  "/auth/register",
			"is_complete": false,
		})
		return
	}

	// Check if admin exists in server_admin_credentials
	var adminCount int
	err = database.GetServerDB().QueryRow("SELECT COUNT(*) FROM server_admin_credentials").Scan(&adminCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to check admin status",
		})
		return
	}

	// First user created but no admin = step 1 complete
	if adminCount == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":      "user_created",
			"step":        1,
			"total_steps": 3,
			"message":     "First user created, admin account needed",
			"next_action": "Create admin account",
			"next_route":  "/setup/admin/welcome",
			"is_complete": false,
		})
		return
	}

	// Check if setup is completed
	var setupComplete string
	err = database.GetServerDB().QueryRow("SELECT value FROM server_config WHERE key = 'setup.completed'").Scan(&setupComplete)

	// If setup.completed doesn't exist or is not "true", server settings needed
	if err != nil || setupComplete != "true" {
		c.JSON(http.StatusOK, gin.H{
			"status":      "admin_created",
			"step":        2,
			"total_steps": 3,
			"message":     "Admin account created, server configuration needed",
			"next_action": "Configure server settings",
			"next_route":  "/setup/server/welcome",
			"is_complete": false,
		})
		return
	}

	// Setup is complete
	// Get admin path from config (AI.md: use configurable admin_path)
	cfg, _ := config.LoadConfig()
	adminPath := "/admin"
	if cfg != nil {
		adminPath = "/" + cfg.GetAdminPath()
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "completed",
		"step":        3,
		"total_steps": 3,
		"message":     "Setup completed successfully",
		"next_action": "Access admin dashboard",
		"next_route":  adminPath,
		"is_complete": true,
	})
}

// Helper function to convert bool to string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

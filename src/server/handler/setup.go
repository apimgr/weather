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

	"github.com/apimgr/weather/src/database"
	"github.com/apimgr/weather/src/utils"

	"github.com/gin-gonic/gin"
)

type SetupHandler struct {
	DB *sql.DB
}

// ShowWelcome shows the welcome screen (step 1)
func (h *SetupHandler) ShowWelcome(c *gin.Context) {
	c.HTML(http.StatusOK, "page/setup_welcome.tmpl", gin.H{
		"Title": "Welcome - Weather Setup",
	})
}

// ShowUserRegister shows the user registration form (step 2)
func (h *SetupHandler) ShowUserRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "page/setup_user.tmpl", gin.H{
		"Title": "Create Your Account - Weather Setup",
	})
}

// CreateUser creates the first user account (step 3)
func (h *SetupHandler) CreateUser(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Trim whitespace from non-password fields
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(input.Email)

	// Passwords cannot start or end with whitespace
	if input.Password != strings.TrimSpace(input.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password cannot start or end with whitespace"})
		return
	}

	// Normalize username
	username := strings.ToLower(input.Username)

	// Check if user already exists
	var count int
	err := database.GetUsersDB().QueryRow("SELECT COUNT(*) FROM user_accounts WHERE email = ? OR username = ?", input.Email, username).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Email or username already exists"})
		return
	}

	// Hash password using Argon2id (TEMPLATE.md Part 0 requirement)
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create first user with 'user' role (NOT admin)
	_, err = database.GetUsersDB().Exec(`
		INSERT INTO user_accounts (username, email, password_hash, role, created_at, updated_at)
		VALUES (?, ?, ?, 'user', datetime('now'), datetime('now'))
	`, username, input.Email, hashedPassword)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ShowAdminSetup shows the admin creation form (step 4)
func (h *SetupHandler) ShowAdminSetup(c *gin.Context) {
	c.HTML(http.StatusOK, "page/setup_admin.tmpl", gin.H{
		"Title": "Create Administrator Account - Weather Setup",
	})
}

// CreateAdmin creates the administrator account (step 5)
func (h *SetupHandler) CreateAdmin(c *gin.Context) {
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

	response := gin.H{"ok": true}
	if generatedPassword != "" {
		// Include generated password in response (shown only once)
		response["generated_password"] = generatedPassword
		response["username"] = username
	}

	// Add redirect to server setup
	response["redirect"] = "/setup/server/welcome"

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

// CompleteSetup performs the final redirect based on current user context
func (h *SetupHandler) CompleteSetup(c *gin.Context) {
	// Check if admin session exists
	adminSessionID, err := c.Cookie("admin_session")
	if err == nil && adminSessionID != "" {
		// Check if admin session is valid
		var adminID int
		err := database.GetServerDB().QueryRow(`
			SELECT admin_id FROM server_admin_sessions
			WHERE session_id = ? AND expires_at > CURRENT_TIMESTAMP
		`, adminSessionID).Scan(&adminID)
		if err == nil {
			c.Redirect(http.StatusFound, "/setup/server/welcome")
			return
		}
	}

	// Not admin - redirect to dashboard or login
	_, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	c.Redirect(http.StatusFound, "/users/dashboard")
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
	c.JSON(http.StatusOK, gin.H{
		"status":      "completed",
		"step":        3,
		"total_steps": 3,
		"message":     "Setup completed successfully",
		"next_action": "Access admin dashboard",
		"next_route":  "/admin",
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

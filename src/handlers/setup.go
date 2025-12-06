package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type SetupHandler struct {
	DB *sql.DB
}

// ShowWelcome shows the welcome screen (step 1)
func (h *SetupHandler) ShowWelcome(c *gin.Context) {
	c.HTML(http.StatusOK, "setup_welcome.tmpl", gin.H{
		"Title": "Welcome - Weather Setup",
	})
}

// ShowUserRegister shows the user registration form (step 2)
func (h *SetupHandler) ShowUserRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "setup_user.tmpl", gin.H{
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

	// Normalize username
	username := input.Username
	// Username can be anything for first user, but normalize it
	username = strings.ToLower(strings.TrimSpace(username))

	// Check if user already exists
	var count int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? OR username = ?", input.Email, username).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Email or username already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create first user with 'user' role (NOT admin)
	_, err = h.DB.Exec(`
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES (?, ?, ?, 'user', datetime('now'), datetime('now'))
	`, username, input.Email, string(hashedPassword))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ShowAdminSetup shows the admin creation form (step 4)
func (h *SetupHandler) ShowAdminSetup(c *gin.Context) {
	c.HTML(http.StatusOK, "setup_admin.tmpl", gin.H{
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

	// Check if admin username already exists
	var count int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create administrator account with custom username
	result, err := h.DB.Exec(`
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES (?, ?, ?, 'admin', datetime('now'), datetime('now'))
	`, username, email, string(hashedPassword))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create administrator"})
		return
	}

	// Get the newly created admin user ID
	adminID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve admin ID"})
		return
	}

	// Create session for the admin user (auto-login)
	sessionID, err := generateSessionID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session ID"})
		return
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days

	_, err = h.DB.Exec(`
		INSERT INTO sessions (id, user_id, expires_at, created_at)
		VALUES (?, ?, ?, datetime('now'))
	`, sessionID, adminID, expiresAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set session cookie
	c.SetCookie(
		"weather_session",
		sessionID,
		int(7*24*time.Hour.Seconds()),
		"/",
		"",
		false, // secure (set to true in production with HTTPS)
		true,  // httpOnly
	)

	response := gin.H{"success": true}
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
	// Check if user is logged in
	userID, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get user role
	var role string
	err := h.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Redirect based on role - admin goes to server setup, user to dashboard
	if role == "admin" {
		c.Redirect(http.StatusFound, "/setup/server/welcome")
	} else {
		c.Redirect(http.StatusFound, "/user/dashboard")
	}
}

// =============================================================================
// Server Setup Wizard (Admin Only)
// =============================================================================

// ShowServerSetupWelcome shows the server setup welcome page
func (h *SetupHandler) ShowServerSetupWelcome(c *gin.Context) {
	c.HTML(http.StatusOK, "server_setup_welcome.tmpl", gin.H{
		"Title": "Server Setup - Weather",
	})
}

// ShowServerSetupSettings shows the server settings configuration page
func (h *SetupHandler) ShowServerSetupSettings(c *gin.Context) {
	c.HTML(http.StatusOK, "server_setup_settings.tmpl", gin.H{
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
		"server.name":               input.ServerName,
		"server.description":        input.ServerDescription,
		"units.temperature":         input.TemperatureUnit,
		"units.wind_speed":          input.WindSpeedUnit,
		"units.precipitation":       input.PrecipitationUnit,
		"rate_limit.anonymous":      fmt.Sprintf("%d", input.RateLimitAnon),
		"rate_limit.authenticated":  fmt.Sprintf("%d", input.RateLimitAuth),
		"features.registration":     boolToString(input.EnableRegistration),
		"features.alerts":           boolToString(input.EnableAlerts),
		"features.notifications":    boolToString(input.EnableNotifications),
		"features.audit_log":        boolToString(input.EnableAuditLog),
	}

	for key, value := range settings {
		_, err := h.DB.Exec(`
			INSERT INTO settings (key, value, updated_at)
			VALUES (?, ?, datetime('now'))
			ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = datetime('now')
		`, key, value, value)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save setting: " + key})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"redirect": "/setup/complete",
	})
}

// ShowServerSetupComplete shows the completion page and marks setup as done
func (h *SetupHandler) ShowServerSetupComplete(c *gin.Context) {
	// Mark server setup as complete
	_, err := h.DB.Exec(`
		INSERT INTO settings (key, value, updated_at)
		VALUES ('setup.completed', 'true', datetime('now'))
		ON CONFLICT(key) DO UPDATE SET value = 'true', updated_at = datetime('now')
	`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark setup as complete"})
		return
	}

	c.HTML(http.StatusOK, "server_setup_complete.tmpl", gin.H{
		"Title": "Setup Complete - Weather",
	})
}

// GetSetupStatus returns the current setup status as a healthz endpoint
func (h *SetupHandler) GetSetupStatus(c *gin.Context) {
	// Check if any users exist
	var userCount int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
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
			"status":       "not_started",
			"step":         0,
			"total_steps":  3,
			"message":      "Setup not started",
			"next_action":  "Create first user account",
			"next_route":   "/register",
			"is_complete":  false,
		})
		return
	}

	// Check if admin exists
	var adminCount int
	err = h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)
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
			"status":       "user_created",
			"step":         1,
			"total_steps":  3,
			"message":      "First user created, admin account needed",
			"next_action":  "Create admin account",
			"next_route":   "/setup/admin/welcome",
			"is_complete":  false,
		})
		return
	}

	// Check if setup is completed
	var setupComplete string
	err = h.DB.QueryRow("SELECT value FROM settings WHERE key = 'setup.completed'").Scan(&setupComplete)

	// If setup.completed doesn't exist or is not "true", server settings needed
	if err != nil || setupComplete != "true" {
		c.JSON(http.StatusOK, gin.H{
			"status":       "admin_created",
			"step":         2,
			"total_steps":  3,
			"message":      "Admin account created, server configuration needed",
			"next_action":  "Configure server settings",
			"next_route":   "/setup/server/welcome",
			"is_complete":  false,
		})
		return
	}

	// Setup is complete
	c.JSON(http.StatusOK, gin.H{
		"status":       "completed",
		"step":         3,
		"total_steps":  3,
		"message":      "Setup completed successfully",
		"next_action":  "Access admin dashboard",
		"next_route":   "/admin",
		"is_complete":  true,
	})
}

// Helper function to convert bool to string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

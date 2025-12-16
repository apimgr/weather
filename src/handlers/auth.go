package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"strings"

	"weather-go/src/middleware"
	"weather-go/src/models"
	"weather-go/src/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	DB *sql.DB
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // Can be username, email, or phone
	Password   string `json:"password" binding:"required"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Username        string `json:"username" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// ShowLoginPage renders the login page
func (h *AuthHandler) ShowLoginPage(c *gin.Context) {
	// Check if already authenticated
	if middleware.IsAuthenticated(c) {
		c.Redirect(http.StatusFound, "/user/dashboard")
		return
	}

	c.HTML(http.StatusOK, "pages/login.tmpl", utils.TemplateData(c, gin.H{
		"title": "Login",
	}))
}

// ShowRegisterPage renders the registration page
func (h *AuthHandler) ShowRegisterPage(c *gin.Context) {
	// Check if already authenticated
	if middleware.IsAuthenticated(c) {
		c.Redirect(http.StatusFound, "/user/dashboard")
		return
	}

	// Check if this is first user (setup mode)
	userModel := &models.UserModel{DB: h.DB}
	count, err := userModel.Count()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "pages/error.tmpl", utils.TemplateData(c, gin.H{
			"error": "Database error",
		}))
		return
	}

	isSetup := count == 0

	c.HTML(http.StatusOK, "pages/register.tmpl", utils.TemplateData(c, gin.H{
		"title":   "Register",
		"isSetup": isSetup,
	}))
}

// HandleLogin processes login requests
func (h *AuthHandler) HandleLogin(c *gin.Context) {
	var req LoginRequest

	// Support both JSON and form data
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
	} else {
		// Accept both "identifier" and legacy "email" field names
		req.Identifier = c.PostForm("identifier")
		if req.Identifier == "" {
			req.Identifier = c.PostForm("email") // Backward compatibility
		}
		req.Password = c.PostForm("password")
	}

	// Validate credentials - try username, email, or phone
	userModel := &models.UserModel{DB: h.DB}
	user, err := userModel.GetByIdentifier(req.Identifier)
	if err != nil {
		respondWithError(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !userModel.CheckPassword(user, req.Password) {
		respondWithError(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Get session timeout from settings
	sessionTimeout, err := h.getSessionTimeout()
	if err != nil {
		sessionTimeout = 2592000 // Default 30 days
	}

	// Create session
	sessionModel := &models.SessionModel{DB: h.DB}
	session, err := sessionModel.Create(user.ID, sessionTimeout)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Set session cookie with Secure flag in production mode
	// Per TEMPLATE.md security requirements and OWASP guidelines
	isProduction := os.Getenv("MODE") == "production"
	c.SetCookie(
		middleware.SessionCookieName,
		session.ID,
		sessionTimeout,
		"/",
		"",
		isProduction, // Secure flag: true in production (requires HTTPS)
		true,         // HttpOnly: always true to prevent XSS
	)

	// Respond based on request type
	if strings.Contains(contentType, "application/json") {
		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
			},
		})
	} else {
		c.Redirect(http.StatusFound, "/user/dashboard")
	}
}

// HandleRegister processes registration requests
func (h *AuthHandler) HandleRegister(c *gin.Context) {
	var req RegisterRequest

	// Support both JSON and form data
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
	} else {
		req.Username = c.PostForm("username")
		req.Email = c.PostForm("email")
		req.Password = c.PostForm("password")
		req.ConfirmPassword = c.PostForm("confirm_password")
	}

	// Validate passwords match
	if req.Password != req.ConfirmPassword {
		respondWithError(c, http.StatusBadRequest, "Passwords do not match")
		return
	}

	userModel := &models.UserModel{DB: h.DB}

	// Check if first user - they will be regular user, then create admin separately
	count, err := userModel.Count()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	isFirstUser := count == 0

	// Validate username (all users, including first user)
	if err := utils.ValidateUsername(req.Username); err != nil {
		respondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Normalize username
	username := utils.NormalizeUsername(req.Username)

	// All users created via /register are regular users
	// Admin accounts are created through /user/setup/admin wizard (first run only)
	role := "user"

	// Create user
	user, err := userModel.Create(username, req.Email, req.Password, role)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			respondWithError(c, http.StatusBadRequest, "Unable to create account. Please try a different email address.")
			return
		}
		respondWithError(c, http.StatusInternalServerError, "Failed to create account. Please try again later.")
		return
	}

	// Get session timeout from settings
	sessionTimeout, err := h.getSessionTimeout()
	if err != nil {
		sessionTimeout = 2592000 // Default 30 days
	}

	// Auto-login after registration
	sessionModel := &models.SessionModel{DB: h.DB}
	session, err := sessionModel.Create(user.ID, sessionTimeout)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "User created but failed to login")
		return
	}

	// Set session cookie with Secure flag in production mode
	// Per TEMPLATE.md security requirements and OWASP guidelines
	isProduction := os.Getenv("MODE") == "production"
	c.SetCookie(
		middleware.SessionCookieName,
		session.ID,
		sessionTimeout,
		"/",
		"",
		isProduction, // Secure flag: true in production (requires HTTPS)
		true,         // HttpOnly: always true to prevent XSS
	)

	// Respond based on request type
	if strings.Contains(contentType, "application/json") {
		response := gin.H{
			"message": "Registration successful",
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
			},
		}

		// If first user, indicate they need to create admin account
		if isFirstUser {
			response["next_step"] = "create_admin"
			response["redirect"] = "/setup/admin/welcome"
		}

		c.JSON(http.StatusCreated, response)
	} else {
		// If first user, redirect to admin setup wizard
		if isFirstUser {
			c.Redirect(http.StatusFound, "/setup/admin/welcome")
		} else {
			c.Redirect(http.StatusFound, "/user/dashboard")
		}
	}
}

// HandleLogout processes logout requests
func (h *AuthHandler) HandleLogout(c *gin.Context) {
	// Get session from context
	session, exists := middleware.GetCurrentSession(c)
	if exists {
		sessionModel := &models.SessionModel{DB: h.DB}
		sessionModel.Delete(session.ID)
	}

	// Clear session cookie
	c.SetCookie(
		middleware.SessionCookieName,
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	// Respond based on request type
	acceptHeader := c.GetHeader("Accept")
	if strings.Contains(acceptHeader, "application/json") {
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	} else {
		c.Redirect(http.StatusFound, "/")
	}
}

// GetCurrentUser returns current user info
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"phone":    user.Phone,
		"role":     user.Role,
	})
}

// UpdateProfile updates user profile (display name and phone only)
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Phone       string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userModel := &models.UserModel{DB: h.DB}
	if err := userModel.UpdateProfile(user.ID, req.DisplayName, req.Phone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// Helper functions

func (h *AuthHandler) getSessionTimeout() (int, error) {
	var timeoutStr string
	err := h.DB.QueryRow("SELECT value FROM settings WHERE key = ?", "auth.session_timeout").Scan(&timeoutStr)
	if err != nil {
		return 0, err
	}

	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return 0, err
	}

	return timeout, nil
}

func respondWithError(c *gin.Context, statusCode int, message string) {
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		c.JSON(statusCode, gin.H{"error": message})
	} else {
		// For form submissions, render inline error on same page
		// Get the current path to determine which template to render
		path := c.Request.URL.Path

		if strings.Contains(path, "login") {
			c.HTML(statusCode, "pages/login.tmpl", utils.TemplateData(c, gin.H{
				"title": "Login",
				"error": message,
			}))
		} else if strings.Contains(path, "register") {
			c.HTML(statusCode, "pages/register.tmpl", utils.TemplateData(c, gin.H{
				"title": "Register",
				"error": message,
			}))
		} else {
			// Fallback to error page for other cases
			c.HTML(statusCode, "pages/error.tmpl", gin.H{
				"error": message,
			})
		}
	}
}

package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"weather-go/src/middleware"
	"weather-go/src/models"
	"weather-go/src/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// SetupHandler handles first-run setup flow
type SetupHandler struct {
	DB *sql.DB
}

// SetupUserRegisterRequest represents user registration during setup
type SetupUserRegisterRequest struct {
	Username string `json:"username" form:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required,min=8"`
}

// SetupAdminRegisterRequest represents admin registration during setup
type SetupAdminRegisterRequest struct {
	Password        string `json:"password" form:"password" binding:"required,min=12"`
	PasswordConfirm string `json:"password_confirm" form:"password_confirm" binding:"required"`
}

// IsSetupComplete checks if setup is complete (users exist)
func (h *SetupHandler) IsSetupComplete() (bool, error) {
	userModel := &models.UserModel{DB: h.DB}
	count, err := userModel.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ShowWelcome shows the initial setup welcome screen
func (h *SetupHandler) ShowWelcome(c *gin.Context) {
	// Check if setup already complete
	complete, err := h.IsSetupComplete()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", utils.TemplateData(c, gin.H{
			"error": "Database error",
		}))
		return
	}

	if complete {
		// Setup already done, redirect to homepage
		c.Redirect(http.StatusFound, "/")
		return
	}

	c.HTML(http.StatusOK, "welcome.html", utils.TemplateData(c, gin.H{
		"title": "Welcome - Setup Required",
	}))
}

// ShowUserRegister shows the user registration form
func (h *SetupHandler) ShowUserRegister(c *gin.Context) {
	// Check if setup already complete
	complete, err := h.IsSetupComplete()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", utils.TemplateData(c, gin.H{
			"error": "Database error",
		}))
		return
	}

	if complete {
		c.Redirect(http.StatusFound, "/")
		return
	}

	c.HTML(http.StatusOK, "register-user.html", utils.TemplateData(c, gin.H{
		"title": "Create Your Account",
	}))
}

// HandleUserRegister creates the first regular user account
func (h *SetupHandler) HandleUserRegister(c *gin.Context) {
	// Check if setup already complete
	complete, err := h.IsSetupComplete()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if complete {
		respondWithError(c, http.StatusForbidden, "Setup already complete")
		return
	}

	var req SetupUserRegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	userModel := &models.UserModel{DB: h.DB}

	// Create first user as REGULAR USER (not admin)
	user, err := userModel.Create(req.Email, req.Password, "user")
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			respondWithError(c, http.StatusConflict, "Email already registered")
			return
		}
		respondWithError(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Auto-login as the new user
	sessionModel := &models.SessionModel{DB: h.DB}
	session, err := sessionModel.Create(user.ID, 2592000) // 30 days
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "User created but failed to login")
		return
	}

	// Set session cookie
	c.SetCookie(
		middleware.SessionCookieName,
		session.ID,
		2592000,
		"/",
		"",
		false,
		true,
	)

	// Redirect to admin account creation
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Account created successfully",
		"redirect": "/user/setup/admin",
	})
}

// ShowAdminRegister shows the administrator account creation form
func (h *SetupHandler) ShowAdminRegister(c *gin.Context) {
	// Must be logged in as a user to create admin
	if !middleware.IsAuthenticated(c) {
		c.Redirect(http.StatusFound, "/user/setup/register")
		return
	}

	// Check if admin already exists
	var adminCount int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", utils.TemplateData(c, gin.H{
			"error": "Database error",
		}))
		return
	}

	if adminCount > 0 {
		// Admin already created, go to complete
		c.Redirect(http.StatusFound, "/user/setup/complete")
		return
	}

	c.HTML(http.StatusOK, "register-admin.html", utils.TemplateData(c, gin.H{
		"title": "Create Administrator Account",
	}))
}

// HandleAdminRegister creates the administrator account
func (h *SetupHandler) HandleAdminRegister(c *gin.Context) {
	// Must be logged in to create admin
	if !middleware.IsAuthenticated(c) {
		respondWithError(c, http.StatusUnauthorized, "Must be logged in")
		return
	}

	// Check if admin already exists
	var adminCount int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if adminCount > 0 {
		respondWithError(c, http.StatusForbidden, "Administrator already exists")
		return
	}

	var req SetupAdminRegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Validate passwords match
	if req.Password != req.PasswordConfirm {
		respondWithError(c, http.StatusBadRequest, "Passwords do not match")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create administrator account with fixed username "administrator"
	_, err = h.DB.Exec(`
		INSERT INTO users (email, password_hash, role, created_at, updated_at)
		VALUES (?, ?, 'admin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, "administrator@localhost", string(hashedPassword))

	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to create administrator")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Administrator account created successfully",
		"redirect": "/user/setup/complete",
	})
}

// HandleComplete completes the setup and redirects appropriately
func (h *SetupHandler) HandleComplete(c *gin.Context) {
	// Check current user
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		// Not logged in, redirect to login
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Contextual redirect based on role
	if user.Role == "admin" {
		c.Redirect(http.StatusFound, "/admin")
	} else {
		c.Redirect(http.StatusFound, "/dashboard")
	}
}

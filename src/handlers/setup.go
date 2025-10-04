package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type SetupHandler struct {
	DB *sql.DB
}

// ShowWelcome shows the welcome screen (step 1)
func (h *SetupHandler) ShowWelcome(c *gin.Context) {
	c.HTML(http.StatusOK, "setup_welcome.html", gin.H{
		"Title": "Welcome - Weather Setup",
	})
}

// ShowUserRegister shows the user registration form (step 2)
func (h *SetupHandler) ShowUserRegister(c *gin.Context) {
	c.HTML(http.StatusOK, "setup_user.html", gin.H{
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

	// Check if user already exists
	var count int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", input.Email).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user with 'user' role (first user is owner, not admin)
	_, err = h.DB.Exec(`
		INSERT INTO users (email, password_hash, display_name, role, is_active)
		VALUES (?, ?, ?, 'user', 1)
	`, input.Email, string(hashedPassword), input.Username)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ShowAdminSetup shows the admin creation form (step 4)
func (h *SetupHandler) ShowAdminSetup(c *gin.Context) {
	c.HTML(http.StatusOK, "setup_admin.html", gin.H{
		"Title": "Create Administrator Account - Weather Setup",
	})
}

// CreateAdmin creates the administrator account (step 5)
func (h *SetupHandler) CreateAdmin(c *gin.Context) {
	var input struct {
		Password        string `json:"password" binding:"required,min=12"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate passwords match
	if input.Password != input.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	// Check if admin already exists
	var count int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = 'administrator@system.local'").Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Administrator already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create administrator account
	_, err = h.DB.Exec(`
		INSERT INTO users (email, password_hash, display_name, role, is_active)
		VALUES ('administrator@system.local', ?, 'Administrator', 'admin', 1)
	`, string(hashedPassword))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create administrator"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// CompleteSetup performs the final redirect based on current user context
func (h *SetupHandler) CompleteSetup(c *gin.Context) {
	// Check if user is logged in
	userID, exists := c.Get("user_id")
	if !exists {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	// Get user role
	var role string
	err := h.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
	if err != nil {
		c.Redirect(http.StatusFound, "/auth/login")
		return
	}

	// Redirect based on role
	if role == "admin" {
		c.Redirect(http.StatusFound, "/admin")
	} else {
		c.Redirect(http.StatusFound, "/dashboard")
	}
}

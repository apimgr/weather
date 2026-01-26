// Package handler provides auth API handlers per AI.md PART 33
package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/apimgr/weather/src/config"
	models "github.com/apimgr/weather/src/server/model"
	"github.com/apimgr/weather/src/server/service"
	"github.com/apimgr/weather/src/utils"

	"github.com/gin-gonic/gin"
)

// AuthAPIHandler handles auth API endpoints per AI.md PART 33
type AuthAPIHandler struct {
	DB *sql.DB
}

// NewAuthAPIHandler creates a new auth API handler
func NewAuthAPIHandler(db *sql.DB) *AuthAPIHandler {
	return &AuthAPIHandler{DB: db}
}

// APILoginRequest represents login API request per AI.md PART 33
type APILoginRequest struct {
	// Can be username, email, user_id, or phone
	Identifier    string `json:"identifier" binding:"required"`
	Password      string `json:"password" binding:"required"`
	TwoFactorCode string `json:"two_factor_code,omitempty"`
	RecoveryKey   string `json:"recovery_key,omitempty"`
}

// APIRegisterRequest represents registration API request per AI.md PART 33
type APIRegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// HandleAPILogin handles POST /api/v1/auth/login per AI.md PART 33
func (h *AuthAPIHandler) HandleAPILogin(c *gin.Context) {
	var req APILoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Invalid request format",
		})
		return
	}

	// Trim whitespace
	req.Identifier = strings.TrimSpace(req.Identifier)

	// Validate password doesn't have leading/trailing whitespace
	if req.Password != strings.TrimSpace(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Password cannot start or end with whitespace",
		})
		return
	}

	// Find user by identifier (username, email, or user_id)
	userModel := &models.UserModel{DB: h.DB}
	user, err := userModel.GetByIdentifier(req.Identifier)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"ok":    false,
			"error": "Invalid credentials",
		})
		return
	}

	// Check password
	if !userModel.CheckPassword(user, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"ok":    false,
			"error": "Invalid credentials",
		})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"ok":    false,
			"error": "Account is disabled",
		})
		return
	}

	// Check if user is banned
	if user.IsBanned {
		c.JSON(http.StatusForbidden, gin.H{
			"ok":    false,
			"error": "Account is suspended",
		})
		return
	}

	// Check if 2FA is required
	if user.TwoFactorEnabled {
		// If no 2FA code provided, return challenge
		if req.TwoFactorCode == "" && req.RecoveryKey == "" {
			c.JSON(http.StatusOK, gin.H{
				"ok": true,
				"data": gin.H{
					"requires_2fa": true,
					"user_id":      user.ID,
				},
			})
			return
		}

		// Verify 2FA code or recovery key
		var verified bool
		if req.RecoveryKey != "" {
			recoveryKeyModel := &models.RecoveryKeyModel{DB: h.DB}
			verified, err = recoveryKeyModel.VerifyAndUseRecoveryKey(int(user.ID), req.RecoveryKey)
		} else {
			verified, err = utils.VerifyTOTP(user.TwoFactorSecret, req.TwoFactorCode)
		}

		if err != nil || !verified {
			c.JSON(http.StatusUnauthorized, gin.H{
				"ok":    false,
				"error": "Invalid two-factor code",
			})
			return
		}
	}

	// Create session
	sessionModel := &models.SessionModel{DB: h.DB}
	session, err := sessionModel.Create(user.ID, 2592000) // 30 days
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Failed to create session",
		})
		return
	}

	// Update last login
	userModel.UpdateLastLogin(user.ID, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
		"data": gin.H{
			"token": session.ID,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    utils.MaskEmail(user.Email),
				"role":     user.Role,
			},
			"expires_at": session.ExpiresAt,
		},
	})
}

// HandleAPIRegister handles POST /api/v1/auth/register per AI.md PART 33
func (h *AuthAPIHandler) HandleAPIRegister(c *gin.Context) {
	// Check if registration is enabled
	if !config.IsMultiUserEnabled() {
		c.JSON(http.StatusNotFound, gin.H{
			"ok":    false,
			"error": "Registration is not available",
		})
		return
	}

	// Check registration mode
	if config.IsRegistrationDisabled() || config.IsRegistrationPrivate() {
		c.JSON(http.StatusNotFound, gin.H{
			"ok":    false,
			"error": "Public registration is not available",
		})
		return
	}

	var req APIRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Invalid request format",
		})
		return
	}

	// Trim whitespace
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	// Validate username per AI.md PART 33
	if err := utils.ValidateUsername(req.Username); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	// Validate email per AI.md PART 33
	if err := utils.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	// Check password strength
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Password must be at least 8 characters",
		})
		return
	}

	// Normalize username
	username := utils.NormalizeUsername(req.Username)

	// Create user
	userModel := &models.UserModel{DB: h.DB}
	user, err := userModel.Create(username, req.Email, req.Password, "user")
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{
				"ok":    false,
				"error": "Username or email already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Failed to create account",
		})
		return
	}

	// Create session (auto-login after registration)
	sessionModel := &models.SessionModel{DB: h.DB}
	session, err := sessionModel.Create(user.ID, 2592000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Account created but failed to login",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"ok": true,
		"data": gin.H{
			"token": session.ID,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    utils.MaskEmail(user.Email),
				"role":     user.Role,
			},
		},
	})
}

// HandleAPILogout handles POST /api/v1/auth/logout per AI.md PART 33
func (h *AuthAPIHandler) HandleAPILogout(c *gin.Context) {
	// Get session from context
	sessionID, exists := c.Get("session_id")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"ok":      true,
			"message": "Already logged out",
		})
		return
	}

	// Delete session
	sessionModel := &models.SessionModel{DB: h.DB}
	sessionModel.Delete(sessionID.(string))

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "Logged out successfully",
	})
}

// HandleAPI2FA handles POST /api/v1/auth/2fa per AI.md PART 33
func (h *AuthAPIHandler) HandleAPI2FA(c *gin.Context) {
	var req struct {
		UserID        int64  `json:"user_id" binding:"required"`
		TwoFactorCode string `json:"two_factor_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Invalid request format",
		})
		return
	}

	// Get user
	userModel := &models.UserModel{DB: h.DB}
	user, err := userModel.GetByID(req.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"ok":    false,
			"error": "Invalid user",
		})
		return
	}

	// Verify 2FA code
	verified, err := utils.VerifyTOTP(user.TwoFactorSecret, req.TwoFactorCode)
	if err != nil || !verified {
		c.JSON(http.StatusUnauthorized, gin.H{
			"ok":    false,
			"error": "Invalid two-factor code",
		})
		return
	}

	// Create full session
	sessionModel := &models.SessionModel{DB: h.DB}
	fullSession, err := sessionModel.Create(user.ID, 2592000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Failed to create session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
		"data": gin.H{
			"token": fullSession.ID,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    utils.MaskEmail(user.Email),
				"role":     user.Role,
			},
			"expires_at": fullSession.ExpiresAt,
		},
	})
}

// HandleAPIRecoveryUse handles POST /api/v1/auth/recovery/use per AI.md PART 33
func (h *AuthAPIHandler) HandleAPIRecoveryUse(c *gin.Context) {
	var req struct {
		UserID      int64  `json:"user_id" binding:"required"`
		RecoveryKey string `json:"recovery_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Invalid request format",
		})
		return
	}

	// Get user
	userModel := &models.UserModel{DB: h.DB}
	user, err := userModel.GetByID(req.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"ok":    false,
			"error": "Invalid user",
		})
		return
	}

	// Verify and use recovery key
	recoveryKeyModel := &models.RecoveryKeyModel{DB: h.DB}
	verified, err := recoveryKeyModel.VerifyAndUseRecoveryKey(int(user.ID), req.RecoveryKey)
	if err != nil || !verified {
		c.JSON(http.StatusUnauthorized, gin.H{
			"ok":    false,
			"error": "Invalid recovery key",
		})
		return
	}

	// Create full session
	sessionModel := &models.SessionModel{DB: h.DB}
	fullSession, err := sessionModel.Create(user.ID, 2592000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Failed to create session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
		"data": gin.H{
			"token": fullSession.ID,
		},
	})
}

// HandleAPIRefresh handles POST /api/v1/auth/refresh per AI.md PART 33
func (h *AuthAPIHandler) HandleAPIRefresh(c *gin.Context) {
	// Get current session
	sessionID, exists := c.Get("session_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"ok":    false,
			"error": "Not authenticated",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"ok":    false,
			"error": "Not authenticated",
		})
		return
	}

	// Delete old session
	sessionModel := &models.SessionModel{DB: h.DB}
	sessionModel.Delete(sessionID.(string))

	// Create new session
	newSession, err := sessionModel.Create(userID.(int64), 2592000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Failed to refresh session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
		"data": gin.H{
			"token":      newSession.ID,
			"expires_at": newSession.ExpiresAt,
		},
	})
}

// HandleAPIVerifyEmail handles POST /api/v1/auth/verify per AI.md PART 33
func (h *AuthAPIHandler) HandleAPIVerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Invalid request format",
		})
		return
	}

	// Find email verification token
	var verification struct {
		ID        int64
		UserID    int64
		Email     string
		ExpiresAt time.Time
	}

	err := h.DB.QueryRow(`
		SELECT id, user_id, email, expires_at
		FROM user_email_verifications
		WHERE token = ? AND expires_at > ?
	`, req.Token, time.Now()).Scan(&verification.ID, &verification.UserID, &verification.Email, &verification.ExpiresAt)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Invalid or expired verification token",
		})
		return
	}

	// Update user email_verified status
	_, err = h.DB.Exec(`
		UPDATE users
		SET email_verified = 1, updated_at = ?
		WHERE id = ?
	`, time.Now(), verification.UserID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Failed to verify email",
		})
		return
	}

	// Delete used verification token
	_, _ = h.DB.Exec(`DELETE FROM user_email_verifications WHERE id = ?`, verification.ID)

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "Email verified successfully",
	})
}

// HandleAPIPasswordForgot handles POST /api/v1/auth/password/forgot per AI.md PART 33
func (h *AuthAPIHandler) HandleAPIPasswordForgot(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Invalid email format",
		})
		return
	}

	// Always return success to prevent email enumeration
	// Per AI.md security requirements
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "If an account exists with that email, a password reset link will be sent",
	})

	// Async: find user and send reset email
	go func() {
		// Find user by email
		var user struct {
			ID    int64
			Email string
		}

		err := h.DB.QueryRow(`
			SELECT id, email FROM users WHERE email = ? AND is_active = 1
		`, req.Email).Scan(&user.ID, &user.Email)

		if err != nil {
			// No user found - silently return (prevent enumeration)
			return
		}

		// Generate reset token (secure random)
		token, err := models.GenerateSecureToken(32)
		if err != nil {
			return
		}

		// Store reset token (expires in 1 hour per AI.md PART 33)
		_, err = h.DB.Exec(`
			INSERT INTO user_password_resets (user_id, token, ip_address, created_at, expires_at)
			VALUES (?, ?, ?, ?, ?)
		`, user.ID, token, c.ClientIP(), time.Now(), time.Now().Add(1*time.Hour))

		if err != nil {
			return
		}

		// Get host info for reset link
		hostInfo := utils.GetHostInfo(c)
		resetURL := fmt.Sprintf("%s/auth/password/reset?token=%s", hostInfo.FullHost, token)

		// Send reset email
		smtpService := service.NewSMTPService(h.DB)
		if err := smtpService.LoadConfig(); err == nil {
			subject := "Password Reset Request"
			body := fmt.Sprintf(`
<html>
<body>
	<h2>Password Reset Request</h2>
	<p>You requested a password reset for your account.</p>
	<p>Click the link below to reset your password (expires in 1 hour):</p>
	<p><a href="%s">%s</a></p>
	<p>If you did not request this reset, please ignore this email.</p>
	<p><em>Sent at %s</em></p>
</body>
</html>
			`, resetURL, resetURL, time.Now().Format(time.RFC1123))

			_ = smtpService.SendEmail(user.Email, subject, body)
		}
	}()
}

// HandleAPIPasswordReset handles POST /api/v1/auth/password/reset per AI.md PART 33
func (h *AuthAPIHandler) HandleAPIPasswordReset(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Invalid request format",
		})
		return
	}

	// Find valid reset token
	var reset struct {
		ID        int64
		UserID    int64
		ExpiresAt time.Time
	}

	err := h.DB.QueryRow(`
		SELECT id, user_id, expires_at
		FROM user_password_resets
		WHERE token = ? AND expires_at > ?
	`, req.Token, time.Now()).Scan(&reset.ID, &reset.UserID, &reset.ExpiresAt)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Invalid or expired reset token",
		})
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Failed to process password",
		})
		return
	}

	// Update user password
	_, err = h.DB.Exec(`
		UPDATE users
		SET password_hash = ?, updated_at = ?
		WHERE id = ?
	`, hashedPassword, time.Now(), reset.UserID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Failed to reset password",
		})
		return
	}

	// Delete used reset token
	_, _ = h.DB.Exec(`DELETE FROM user_password_resets WHERE id = ?`, reset.ID)

	// Invalidate all user sessions (per AI.md PART 33)
	_, _ = h.DB.Exec(`DELETE FROM user_sessions WHERE user_id = ?`, reset.UserID)

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "Password reset successfully. Please log in with your new password.",
	})
}

// HandleAPIUserInviteValidate handles GET /api/v1/auth/invite/user/{token} per AI.md PART 33
func (h *AuthAPIHandler) HandleAPIUserInviteValidate(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Token required",
		})
		return
	}

	// Validate invite token
	inviteModel := &models.UserInviteModel{DB: h.DB}
	invite, err := inviteModel.GetByToken(token)
	if err != nil || invite == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"ok":    false,
			"error": "Invalid or expired invite",
		})
		return
	}

	// Check if expired
	if time.Now().After(invite.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{
			"ok":    false,
			"error": "Invite has expired",
		})
		return
	}

	// Check if already used
	if invite.UsedAt != nil {
		c.JSON(http.StatusGone, gin.H{
			"ok":    false,
			"error": "Invite has already been used",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
		"data": gin.H{
			"username":   invite.Username,
			"expires_at": invite.ExpiresAt,
		},
	})
}

// HandleAPIUserInviteComplete handles POST /api/v1/auth/invite/user/{token} per AI.md PART 33
func (h *AuthAPIHandler) HandleAPIUserInviteComplete(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Token required",
		})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=8"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Invalid request format",
		})
		return
	}

	// Validate invite token
	inviteModel := &models.UserInviteModel{DB: h.DB}
	invite, err := inviteModel.GetByToken(token)
	if err != nil || invite == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"ok":    false,
			"error": "Invalid or expired invite",
		})
		return
	}

	// Check if expired
	if time.Now().After(invite.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{
			"ok":    false,
			"error": "Invite has expired",
		})
		return
	}

	// Check if already used
	if invite.UsedAt != nil {
		c.JSON(http.StatusGone, gin.H{
			"ok":    false,
			"error": "Invite has already been used",
		})
		return
	}

	// Validate email
	if err := utils.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	// Create user with pre-set username from invite
	userModel := &models.UserModel{DB: h.DB}
	user, err := userModel.Create(invite.Username, req.Email, req.Password, "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Failed to create account",
		})
		return
	}

	// Mark invite as used
	inviteModel.MarkUsed(token)

	// Create session (auto-login)
	sessionModel := &models.SessionModel{DB: h.DB}
	session, err := sessionModel.Create(user.ID, 2592000)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"ok":      true,
			"message": "Account created. Please log in.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"ok": true,
		"data": gin.H{
			"token": session.ID,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    utils.MaskEmail(user.Email),
				"role":     user.Role,
			},
		},
	})
}

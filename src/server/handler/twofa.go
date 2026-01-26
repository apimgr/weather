package handler

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/apimgr/weather/src/server/middleware"
	"github.com/apimgr/weather/src/server/model"
	"github.com/apimgr/weather/src/utils"

	"github.com/gin-gonic/gin"
)

type TwoFactorHandler struct {
	DB *sql.DB
}

// GetTwoFactorStatus returns the 2FA status for the authenticated user (API endpoint)
func (h *TwoFactorHandler) GetTwoFactorStatus(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "Not authenticated"})
		return
	}

	var recoveryKeysCount int
	if user.TwoFactorEnabled {
		recoveryKeyModel := &models.RecoveryKeyModel{DB: h.DB}
		recoveryKeysCount, _ = recoveryKeyModel.GetUnusedKeysCount(int(user.ID))
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":                  true,
		"enabled":             user.TwoFactorEnabled,
		"recovery_keys_count": recoveryKeysCount,
	})
}

// ShowSecurityPage renders the security settings page
func (h *TwoFactorHandler) ShowSecurityPage(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get recovery keys count if 2FA is enabled
	var recoveryKeysCount int
	if user.TwoFactorEnabled {
		recoveryKeyModel := &models.RecoveryKeyModel{DB: h.DB}
		recoveryKeysCount, _ = recoveryKeyModel.GetUnusedKeysCount(int(user.ID))
	}

	c.HTML(http.StatusOK, "pages/user/security.tmpl", utils.TemplateData(c, gin.H{
		"title":             "Security Settings",
		"user":              user,
		"recoveryKeysCount": recoveryKeysCount,
	}))
}

// SetupTwoFactor generates a TOTP secret and QR code for setup
func (h *TwoFactorHandler) SetupTwoFactor(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// If 2FA already enabled, return error
	if user.TwoFactorEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Two-factor authentication is already enabled"})
		return
	}

	// Generate TOTP secret and QR code
	secret, qrCodeDataURL, err := utils.GenerateTOTPSecret(user.Email, "Weather Service")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate 2FA secret"})
		return
	}

	// Generate manual entry URL
	otpauthURL := utils.GenerateOTPAuthURL(user.Email, secret, "Weather Service")

	c.JSON(http.StatusOK, gin.H{
		"secret":      secret,
		"qr_code":     qrCodeDataURL,
		"manual_url":  otpauthURL,
		"account":     user.Email,
		"issuer":      "Weather Service",
	})
}

// EnableTwoFactor verifies the TOTP code and enables 2FA for the user
func (h *TwoFactorHandler) EnableTwoFactor(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// If 2FA already enabled, return error
	if user.TwoFactorEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Two-factor authentication is already enabled"})
		return
	}

	var req struct {
		Secret string `json:"secret" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Verify the TOTP code
	valid, err := utils.VerifyTOTP(req.Secret, req.Code)
	if err != nil || !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	// Enable 2FA in database
	userModel := &models.UserModel{DB: h.DB}
	if err := userModel.EnableTwoFactor(user.ID, req.Secret); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable two-factor authentication"})
		return
	}

	// Generate recovery keys (TEMPLATE.md Part 31: 10 one-time recovery keys)
	recoveryKeyModel := &models.RecoveryKeyModel{DB: h.DB}
	recoveryKeys, err := recoveryKeyModel.GenerateRecoveryKeys(int(user.ID))
	if err != nil {
		// Rollback 2FA enablement if recovery key generation fails
		userModel.DisableTwoFactor(user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate recovery keys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Two-factor authentication enabled successfully",
		"recovery_keys": recoveryKeys,
	})
}

// DisableTwoFactor disables 2FA for the user
func (h *TwoFactorHandler) DisableTwoFactor(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// If 2FA not enabled, return error
	if !user.TwoFactorEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Two-factor authentication is not enabled"})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Verify password
	userModel := &models.UserModel{DB: h.DB}
	if !userModel.CheckPassword(user, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// Disable 2FA
	if err := userModel.DisableTwoFactor(user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable two-factor authentication"})
		return
	}

	// Delete all recovery keys
	recoveryKeyModel := &models.RecoveryKeyModel{DB: h.DB}
	recoveryKeyModel.DeleteAllForUser(int(user.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Two-factor authentication disabled successfully",
	})
}

// VerifyTwoFactorCode verifies a TOTP code for an authenticated user
// This is used during sensitive operations, not during login
func (h *TwoFactorHandler) VerifyTwoFactorCode(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	if !user.TwoFactorEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Two-factor authentication is not enabled"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Verify the TOTP code
	valid, err := utils.VerifyTOTP(user.TwoFactorSecret, req.Code)
	if err != nil || !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Code verified successfully",
	})
}

// RegenerateRecoveryKeys generates new recovery keys for a user
func (h *TwoFactorHandler) RegenerateRecoveryKeys(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	if !user.TwoFactorEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Two-factor authentication is not enabled"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Verify TOTP code first
	valid, err := utils.VerifyTOTP(user.TwoFactorSecret, req.Code)
	if err != nil || !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	// Generate new recovery keys
	recoveryKeyModel := &models.RecoveryKeyModel{DB: h.DB}
	recoveryKeys, err := recoveryKeyModel.GenerateRecoveryKeys(int(user.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate recovery keys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Recovery keys regenerated successfully",
		"recovery_keys": recoveryKeys,
	})
}

// respondWith2FAError sends appropriate error response based on content type
func respondWith2FAError(c *gin.Context, statusCode int, message string) {
	contentType := c.GetHeader("Content-Type")
	acceptHeader := c.GetHeader("Accept")

	if strings.Contains(contentType, "application/json") || strings.Contains(acceptHeader, "application/json") {
		c.JSON(statusCode, gin.H{"error": message})
	} else {
		c.HTML(statusCode, "pages/error.tmpl", gin.H{
			"error": message,
		})
	}
}

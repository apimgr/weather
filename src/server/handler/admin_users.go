package handlers

import (
	"net/http"

	"github.com/apimgr/weather/src/utils"
	"github.com/gin-gonic/gin"
)

// AdminUsersHandler handles user management settings
type AdminUsersHandler struct {
	ConfigPath string
}

// ShowUserSettings displays user management settings page
func (h *AdminUsersHandler) ShowUserSettings(c *gin.Context) {
	c.HTML(http.StatusOK, "admin-users.tmpl", gin.H{
		"title": "User Management Settings",
	})
}

// UpdateUserSettings updates user management settings in server.yml
func (h *AdminUsersHandler) UpdateUserSettings(c *gin.Context) {
	var req struct {
		Enabled                      bool   `json:"enabled"`
		RegistrationEnabled          bool   `json:"registration_enabled"`
		RegistrationRequireApproval  bool   `json:"registration_require_approval"`
		RegistrationRequireEmailVerification bool `json:"registration_require_email_verification"`
		RegistrationInviteOnly       bool   `json:"registration_invite_only"`
		UsernameMinLength            int    `json:"username_min_length"`
		UsernameMaxLength            int    `json:"username_max_length"`
		UsernameAllowedChars         string `json:"username_allowed_chars"`
		UsernameBlocklistEnabled     bool   `json:"username_blocklist_enabled"`
		PasswordMinLength            int    `json:"password_min_length"`
		PasswordRequireUppercase     bool   `json:"password_require_uppercase"`
		PasswordRequireLowercase     bool   `json:"password_require_lowercase"`
		PasswordRequireNumbers       bool   `json:"password_require_numbers"`
		PasswordRequireSymbols       bool   `json:"password_require_symbols"`
		SessionTimeout               int    `json:"session_timeout"`
		SessionExtendOnActivity      bool   `json:"session_extend_on_activity"`
		SessionMaxSessionsPerUser    int    `json:"session_max_sessions_per_user"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update server.yml
	updates := map[string]interface{}{
		"server.users.enabled":                                  req.Enabled,
		"server.users.registration.enabled":                     req.RegistrationEnabled,
		"server.users.registration.require_approval":            req.RegistrationRequireApproval,
		"server.users.registration.require_email_verification":  req.RegistrationRequireEmailVerification,
		"server.users.registration.invite_only":                 req.RegistrationInviteOnly,
		"server.users.username.min_length":                      req.UsernameMinLength,
		"server.users.username.max_length":                      req.UsernameMaxLength,
		"server.users.username.allowed_chars":                   req.UsernameAllowedChars,
		"server.users.username.blocklist_enabled":               req.UsernameBlocklistEnabled,
		"server.users.password.min_length":                      req.PasswordMinLength,
		"server.users.password.require_uppercase":               req.PasswordRequireUppercase,
		"server.users.password.require_lowercase":               req.PasswordRequireLowercase,
		"server.users.password.require_numbers":                 req.PasswordRequireNumbers,
		"server.users.password.require_symbols":                 req.PasswordRequireSymbols,
		"server.users.session.timeout":                          req.SessionTimeout,
		"server.users.session.extend_on_activity":               req.SessionExtendOnActivity,
		"server.users.session.max_sessions_per_user":            req.SessionMaxSessionsPerUser,
	}

	if err := utils.UpdateYAMLConfig(h.ConfigPath, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

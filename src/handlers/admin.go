package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"weather-go/src/middleware"
	"weather-go/src/models"
	"weather-go/src/utils"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	DB *sql.DB
}

// User Management APIs

func (h *AdminHandler) ListUsers(c *gin.Context) {
	userModel := &models.UserModel{DB: h.DB}
	users, err := userModel.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Role     string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate username
	if err := utils.ValidateUsername(req.Username); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize username
	username := utils.NormalizeUsername(req.Username)

	userModel := &models.UserModel{DB: h.DB}
	user, err := userModel.Create(username, req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Prevent modifying your own account
	currentUser, _ := middleware.GetCurrentUser(c)
	if currentUser.ID == id {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify your own account credentials. Contact another administrator if you need to change your username, email, or role."})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Role     string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate username
	if err := utils.ValidateUsername(req.Username); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize username
	username := utils.NormalizeUsername(req.Username)

	userModel := &models.UserModel{DB: h.DB}
	if err := userModel.Update(id, username, req.Email, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Prevent deleting yourself
	currentUser, _ := middleware.GetCurrentUser(c)
	if currentUser.ID == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	userModel := &models.UserModel{DB: h.DB}
	if err := userModel.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *AdminHandler) UpdateUserPassword(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userModel := &models.UserModel{DB: h.DB}
	if err := userModel.UpdatePassword(id, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// Settings Management APIs

func (h *AdminHandler) ListSettings(c *gin.Context) {
	rows, err := h.DB.Query("SELECT key, value, type, COALESCE(description, ''), updated_at FROM settings ORDER BY key")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}
	defer rows.Close()

	settings := make([]map[string]interface{}, 0)
	for rows.Next() {
		var key, value, settingType, description string
		var updatedAt time.Time
		if err := rows.Scan(&key, &value, &settingType, &description, &updatedAt); err != nil {
			continue
		}
		settings = append(settings, map[string]interface{}{
			"key":         key,
			"value":       value,
			"type":        settingType,
			"description": description,
			"updated_at":  updatedAt,
		})
	}

	c.JSON(http.StatusOK, settings)
}

func (h *AdminHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	var value, settingType, description string
	var updatedAt time.Time

	err := h.DB.QueryRow("SELECT value, type, COALESCE(description, ''), updated_at FROM settings WHERE key = ?", key).
		Scan(&value, &settingType, &description, &updatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":         key,
		"value":       value,
		"type":        settingType,
		"description": description,
		"updated_at":  updatedAt,
	})
}

func (h *AdminHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")

	var req struct {
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.DB.Exec("UPDATE settings SET value = ?, updated_at = ? WHERE key = ?",
		req.Value, time.Now(), key)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setting updated successfully"})
}

// API Token Management APIs

func (h *AdminHandler) ListTokens(c *gin.Context) {
	userID := c.Query("user_id")

	var rows *sql.Rows
	var err error

	if userID != "" {
		uid, _ := strconv.Atoi(userID)
		tokenModel := &models.TokenModel{DB: h.DB}
		tokens, err := tokenModel.GetByUserID(uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tokens"})
			return
		}
		c.JSON(http.StatusOK, tokens)
		return
	}

	// Get all tokens for admin view
	rows, err = h.DB.Query(`
		SELECT t.id, t.user_id, u.email, t.name, t.created_at, t.last_used_at
		FROM api_tokens t
		JOIN users u ON t.user_id = u.id
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tokens"})
		return
	}
	defer rows.Close()

	var tokens []map[string]interface{}
	for rows.Next() {
		var id, userID int
		var email, name string
		var createdAt time.Time
		var lastUsedAt sql.NullTime

		if err := rows.Scan(&id, &userID, &email, &name, &createdAt, &lastUsedAt); err != nil {
			continue
		}

		token := map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"user_email": email,
			"name":       name,
			"created_at": createdAt,
		}

		if lastUsedAt.Valid {
			token["last_used_at"] = lastUsedAt.Time
		}

		tokens = append(tokens, token)
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *AdminHandler) GenerateToken(c *gin.Context) {
	var req struct {
		UserID int    `json:"user_id" binding:"required"`
		Name   string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenModel := &models.TokenModel{DB: h.DB}
	token, err := tokenModel.Create(req.UserID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Token generated successfully. Save it now - it won't be shown again!",
		"token":   token,
	})
}

func (h *AdminHandler) RevokeToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	tokenModel := &models.TokenModel{DB: h.DB}
	if err := tokenModel.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token revoked successfully"})
}

// Audit Log APIs

func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	rows, err := h.DB.Query(`
		SELECT a.id, a.user_id, u.email, a.action, a.resource, a.details,
		       a.ip_address, a.user_agent, a.created_at
		FROM audit_log a
		LEFT JOIN users u ON a.user_id = u.id
		ORDER BY a.created_at DESC
		LIMIT ?
	`, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var id int
		var userID sql.NullInt64
		var email, action, resource, details, ipAddress, userAgent sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &userID, &email, &action, &resource, &details, &ipAddress, &userAgent, &createdAt); err != nil {
			continue
		}

		log := map[string]interface{}{
			"id":         id,
			"action":     action.String,
			"resource":   resource.String,
			"details":    details.String,
			"ip_address": ipAddress.String,
			"user_agent": userAgent.String,
			"created_at": createdAt,
		}

		if userID.Valid {
			log["user_id"] = userID.Int64
			log["user_email"] = email.String
		}

		logs = append(logs, log)
	}

	c.JSON(http.StatusOK, logs)
}

func (h *AdminHandler) ClearAuditLogs(c *gin.Context) {
	days := 30 // Default retention
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	result, err := h.DB.Exec("DELETE FROM audit_log WHERE created_at < ?", cutoff)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear logs"})
		return
	}

	affected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{
		"message": "Audit logs cleared successfully",
		"deleted": affected,
	})
}

// System Stats APIs

func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	userModel := &models.UserModel{DB: h.DB}

	totalUsers, _ := userModel.Count()
	adminCount, _ := userModel.CountByRole("admin")

	var totalLocations, totalTokens, totalSessions, totalNotifications int
	h.DB.QueryRow("SELECT COUNT(*) FROM saved_locations").Scan(&totalLocations)
	h.DB.QueryRow("SELECT COUNT(*) FROM api_tokens").Scan(&totalTokens)
	h.DB.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&totalSessions)
	h.DB.QueryRow("SELECT COUNT(*) FROM notifications").Scan(&totalNotifications)

	c.JSON(http.StatusOK, gin.H{
		"users": gin.H{
			"total": totalUsers,
			"admin": adminCount,
			"user":  totalUsers - adminCount,
		},
		"locations":     totalLocations,
		"tokens":        totalTokens,
		"sessions":      totalSessions,
		"notifications": totalNotifications,
	})
}

// GetScheduledTasks returns status of all scheduled tasks
func (h *AdminHandler) GetScheduledTasks(c *gin.Context) {
	// Get scheduled tasks from database
	rows, err := h.DB.Query(`
		SELECT task_name, last_run, next_run, status
		FROM scheduled_tasks
		ORDER BY task_name
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch scheduled tasks"})
		return
	}
	defer rows.Close()

	var tasks []map[string]interface{}
	for rows.Next() {
		var taskName, status string
		var lastRun, nextRun sql.NullTime

		if err := rows.Scan(&taskName, &lastRun, &nextRun, &status); err != nil {
			continue
		}

		task := map[string]interface{}{
			"name":    taskName,
			"status":  status,
			"running": status == "running",
		}

		if lastRun.Valid {
			task["last_run"] = lastRun.Time
		}
		if nextRun.Valid {
			task["next_run"] = nextRun.Time
		}

		// Determine interval from task name (this is approximate)
		interval := "Unknown"
		switch taskName {
		case "cleanup-sessions", "cleanup-rate-limits":
			interval = "Every 1 hour"
		case "cleanup-audit-logs":
			interval = "Every 24 hours"
		case "check-weather-alerts":
			interval = "Every 15 minutes"
		case "daily-forecast":
			interval = "Every 24 hours"
		case "process-notification-queue":
			interval = "Every 2 minutes"
		case "cleanup-notifications":
			interval = "Every 24 hours"
		case "system-backup":
			interval = "Every 6 hours"
		case "refresh-weather-cache":
			interval = "Every 30 minutes"
		}
		task["interval"] = interval

		tasks = append(tasks, task)
	}

	// If no tasks in database, return empty array
	if len(tasks) == 0 {
		tasks = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, tasks)
}

// ShowSettingsPage renders the admin settings page
func (h *AdminHandler) ShowSettingsPage(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Load all settings
	settingsModel := &models.SettingsModel{DB: h.DB}
	settings := make(map[string]interface{})

	// Server settings
	settings["server_title"] = settingsModel.GetString("server.title", "Weather Service")
	settings["server_tagline"] = settingsModel.GetString("server.tagline", "Your personal weather dashboard")
	settings["server_description"] = settingsModel.GetString("server.description", "A comprehensive platform for weather forecasts, moon phases, earthquakes, and hurricane tracking.")
	settings["server_http_port"] = settingsModel.GetInt("server.http_port", 3000)
	settings["server_https_port"] = settingsModel.GetInt("server.https_port", 0)
	settings["server_timezone"] = settingsModel.GetString("server.timezone", "America/New_York")
	settings["server_date_format"] = settingsModel.GetString("server.date_format", "US")
	settings["server_time_format"] = settingsModel.GetString("server.time_format", "12-hour")

	// Security settings
	settings["security_session_timeout"] = settingsModel.GetInt("security.session_timeout", 2592000)
	settings["security_max_login_attempts"] = settingsModel.GetInt("security.max_login_attempts", 5)
	settings["security_lockout_duration"] = settingsModel.GetInt("security.lockout_duration", 30)
	settings["security_password_min_length"] = settingsModel.GetInt("security.password_min_length", 8)

	// Features
	settings["features_registration_enabled"] = settingsModel.GetBool("features.registration_enabled", true)
	settings["features_api_enabled"] = settingsModel.GetBool("features.api_enabled", true)
	settings["features_weather_alerts"] = settingsModel.GetBool("features.weather_alerts", true)

	// Backup settings
	settings["backup_enabled"] = settingsModel.GetBool("backup.enabled", true)
	settings["backup_interval"] = settingsModel.GetInt("backup.interval", 6)
	settings["backup_retention"] = settingsModel.GetInt("backup.retention", 30)
	settings["backup_location"] = settingsModel.GetString("backup.location", "/data/backups")

	// Logging settings
	settings["logging_level"] = settingsModel.GetString("logging.level", "info")
	settings["logging_format"] = settingsModel.GetString("logging.format", "apache")
	settings["logging_access_log"] = settingsModel.GetBool("logging.access_log", true)
	settings["logging_error_log"] = settingsModel.GetBool("logging.error_log", true)
	settings["logging_audit_log"] = settingsModel.GetBool("logging.audit_log", true)
	settings["logging_rotation_days"] = settingsModel.GetInt("logging.rotation_days", 30)

	c.HTML(http.StatusOK, "admin-settings.html", gin.H{
		"title":    "Server Settings",
		"user":     user,
		"settings": settings,
	})
}

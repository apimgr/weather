package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"weather-go/src/middleware"
	"weather-go/src/models"

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
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Role     string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userModel := &models.UserModel{DB: h.DB}
	user, err := userModel.Create(req.Email, req.Password, req.Role)
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

	var req struct {
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userModel := &models.UserModel{DB: h.DB}
	if err := userModel.Update(id, req.Email, req.Role); err != nil {
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

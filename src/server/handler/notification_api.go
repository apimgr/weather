package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/apimgr/weather/src/config"
	"github.com/apimgr/weather/src/mode"
	"github.com/apimgr/weather/src/server/model"
	"github.com/apimgr/weather/src/server/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Implement proper origin checking per AI.md PART 17
		// Development mode: allow all origins
		if !mode.IsProduction() {
			return true
		}

		// Production mode: check against configured CORS origins
		origin := r.Header.Get("Origin")
		if origin == "" {
			// No origin header - allow (same-origin)
			return true
		}

		// Load config to check allowed origins
		cfg, err := config.LoadConfig()
		if err != nil || cfg == nil {
			// Config load failed - reject for safety
			return false
		}

		// Check if CORS is set to allow all
		if cfg.Web.CORS == "*" {
			return true
		}

		// Check if origin is in allowed CORS list
		if cfg.Web.CORS != "" {
			allowedOrigins := strings.Split(cfg.Web.CORS, ",")
			for _, allowed := range allowedOrigins {
				if strings.TrimSpace(allowed) == origin {
					return true
				}
			}
		}

		// Origin not allowed
		return false
	},
}

// NotificationAPIHandlers handles all notification API endpoints
type NotificationAPIHandlers struct {
	NotificationService *service.NotificationService
	WSHub               *service.WebSocketHub
}

// NewNotificationAPIHandlers creates a new notification API handlers instance
func NewNotificationAPIHandlers(notificationService *service.NotificationService, wsHub *service.WebSocketHub) *NotificationAPIHandlers {
	return &NotificationAPIHandlers{
		NotificationService: notificationService,
		WSHub:               wsHub,
	}
}

// ========== USER NOTIFICATION ENDPOINTS ==========

// GetUserNotifications returns all notifications for the authenticated user
// GET /api/v1/user/notifications
func (h *NotificationAPIHandlers) GetUserNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit < 1 || limit > 100 {
		limit = 50
	}

	notifications, err := h.NotificationService.GetUserNotifications(userID.(int), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"limit":         limit,
		"offset":        offset,
		"count":         len(notifications),
	})
}

// GetUserUnreadNotifications returns unread notifications for the authenticated user
// GET /api/v1/user/notifications/unread
func (h *NotificationAPIHandlers) GetUserUnreadNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notifications, err := h.NotificationService.GetUserUnreadNotifications(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve unread notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// GetUserUnreadCount returns the count of unread notifications
// GET /api/v1/user/notifications/count
func (h *NotificationAPIHandlers) GetUserUnreadCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := h.NotificationService.GetUserUnreadCount(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

// GetUserNotificationStats returns notification statistics
// GET /api/v1/user/notifications/stats
func (h *NotificationAPIHandlers) GetUserNotificationStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	stats, err := h.NotificationService.GetUserStatistics(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// MarkUserNotificationRead marks a notification as read
// PATCH /api/v1/user/notifications/:id/read
func (h *NotificationAPIHandlers) MarkUserNotificationRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification ID required"})
		return
	}

	err := h.NotificationService.MarkUserNotificationAsRead(notificationID, userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification marked as read",
		"id":      notificationID,
	})
}

// MarkAllUserNotificationsRead marks all notifications as read
// PATCH /api/v1/user/notifications/read
func (h *NotificationAPIHandlers) MarkAllUserNotificationsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := h.NotificationService.MarkAllUserNotificationsAsRead(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "all notifications marked as read",
	})
}

// DismissUserNotification dismisses a notification
// PATCH /api/v1/user/notifications/:id/dismiss
func (h *NotificationAPIHandlers) DismissUserNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification ID required"})
		return
	}

	err := h.NotificationService.DismissUserNotification(notificationID, userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification dismissed",
		"id":      notificationID,
	})
}

// DeleteUserNotification deletes a notification
// DELETE /api/v1/user/notifications/:id
func (h *NotificationAPIHandlers) DeleteUserNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification ID required"})
		return
	}

	err := h.NotificationService.DeleteUserNotification(notificationID, userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification deleted",
		"id":      notificationID,
	})
}

// GetUserNotificationPreferences returns notification preferences
// GET /api/v1/user/notifications/preferences
func (h *NotificationAPIHandlers) GetUserNotificationPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	prefs, err := h.NotificationService.GetUserPreferences(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get preferences"})
		return
	}

	c.JSON(http.StatusOK, prefs)
}

// UpdateUserNotificationPreferences updates notification preferences
// PATCH /api/v1/user/notifications/preferences
func (h *NotificationAPIHandlers) UpdateUserNotificationPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var prefs models.NotificationPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate toast durations (1-60 seconds)
	if prefs.ToastDurationSuccess < 1 || prefs.ToastDurationSuccess > 60 ||
		prefs.ToastDurationInfo < 1 || prefs.ToastDurationInfo > 60 ||
		prefs.ToastDurationWarning < 1 || prefs.ToastDurationWarning > 60 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "toast durations must be between 1 and 60 seconds"})
		return
	}

	err := h.NotificationService.UpdateUserPreferences(userID.(int), &prefs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update preferences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "preferences updated successfully",
	})
}

// ========== ADMIN NOTIFICATION ENDPOINTS ==========

// GetAdminNotifications returns all notifications for the authenticated admin
// GET /api/v1/admin/notifications
func (h *NotificationAPIHandlers) GetAdminNotifications(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit < 1 || limit > 100 {
		limit = 50
	}

	notifications, err := h.NotificationService.GetAdminNotifications(adminID.(int), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"limit":         limit,
		"offset":        offset,
		"count":         len(notifications),
	})
}

// GetAdminUnreadNotifications returns unread notifications for the authenticated admin
// GET /api/v1/admin/notifications/unread
func (h *NotificationAPIHandlers) GetAdminUnreadNotifications(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notifications, err := h.NotificationService.GetAdminUnreadNotifications(adminID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve unread notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// GetAdminUnreadCount returns the count of unread notifications
// GET /api/v1/admin/notifications/count
func (h *NotificationAPIHandlers) GetAdminUnreadCount(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := h.NotificationService.GetAdminUnreadCount(adminID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

// GetAdminNotificationStats returns notification statistics
// GET /api/v1/admin/notifications/stats
func (h *NotificationAPIHandlers) GetAdminNotificationStats(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	stats, err := h.NotificationService.GetAdminStatistics(adminID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// MarkAdminNotificationRead marks a notification as read
// PATCH /api/v1/admin/notifications/:id/read
func (h *NotificationAPIHandlers) MarkAdminNotificationRead(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification ID required"})
		return
	}

	err := h.NotificationService.MarkAdminNotificationAsRead(notificationID, adminID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification marked as read",
		"id":      notificationID,
	})
}

// MarkAllAdminNotificationsRead marks all notifications as read
// PATCH /api/v1/admin/notifications/read
func (h *NotificationAPIHandlers) MarkAllAdminNotificationsRead(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := h.NotificationService.MarkAllAdminNotificationsAsRead(adminID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "all notifications marked as read",
	})
}

// DismissAdminNotification dismisses a notification
// PATCH /api/v1/admin/notifications/:id/dismiss
func (h *NotificationAPIHandlers) DismissAdminNotification(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification ID required"})
		return
	}

	err := h.NotificationService.DismissAdminNotification(notificationID, adminID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification dismissed",
		"id":      notificationID,
	})
}

// DeleteAdminNotification deletes a notification
// DELETE /api/v1/admin/notifications/:id
func (h *NotificationAPIHandlers) DeleteAdminNotification(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification ID required"})
		return
	}

	err := h.NotificationService.DeleteAdminNotification(notificationID, adminID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notification deleted",
		"id":      notificationID,
	})
}

// GetAdminNotificationPreferences returns notification preferences
// GET /api/v1/admin/notifications/preferences
func (h *NotificationAPIHandlers) GetAdminNotificationPreferences(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	prefs, err := h.NotificationService.GetAdminPreferences(adminID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get preferences"})
		return
	}

	c.JSON(http.StatusOK, prefs)
}

// UpdateAdminNotificationPreferences updates notification preferences
// PATCH /api/v1/admin/notifications/preferences
func (h *NotificationAPIHandlers) UpdateAdminNotificationPreferences(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var prefs models.NotificationPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate toast durations (1-60 seconds)
	if prefs.ToastDurationSuccess < 1 || prefs.ToastDurationSuccess > 60 ||
		prefs.ToastDurationInfo < 1 || prefs.ToastDurationInfo > 60 ||
		prefs.ToastDurationWarning < 1 || prefs.ToastDurationWarning > 60 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "toast durations must be between 1 and 60 seconds"})
		return
	}

	err := h.NotificationService.UpdateAdminPreferences(adminID.(int), &prefs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update preferences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "preferences updated successfully",
	})
}

// SendTestNotification sends a test notification to the authenticated admin
// POST /api/v1/admin/notifications/send
func (h *NotificationAPIHandlers) SendTestNotification(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Type    models.NotificationType    `json:"type" binding:"required"`
		Display models.NotificationDisplay `json:"display" binding:"required"`
		Title   string                     `json:"title" binding:"required"`
		Message string                     `json:"message" binding:"required"`
		Action  *models.NotificationAction `json:"action,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate notification type
	validTypes := map[models.NotificationType]bool{
		models.NotificationTypeSuccess:  true,
		models.NotificationTypeInfo:     true,
		models.NotificationTypeWarning:  true,
		models.NotificationTypeError:    true,
		models.NotificationTypeSecurity: true,
	}
	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification type"})
		return
	}

	// Validate display type
	validDisplays := map[models.NotificationDisplay]bool{
		models.NotificationDisplayToast:  true,
		models.NotificationDisplayBanner: true,
		models.NotificationDisplayCenter: true,
	}
	if !validDisplays[req.Display] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid display type"})
		return
	}

	// Send test notification
	notification, err := h.NotificationService.SendAdminNotification(
		adminID.(int),
		req.Type,
		req.Display,
		req.Title,
		req.Message,
		req.Action,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test notification"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "test notification sent successfully",
		"notification": notification,
	})
}

// ========== WEBSOCKET ENDPOINT ==========

// HandleWebSocketConnection handles WebSocket connections for real-time notifications
// GET /ws/notifications
func (h *NotificationAPIHandlers) HandleWebSocketConnection(c *gin.Context) {
	// Check if user or admin is authenticated
	userID, userExists := c.Get("user_id")
	adminID, adminExists := c.Get("admin_id")

	if !userExists && !adminExists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upgrade connection"})
		return
	}

	// Create WebSocket client
	var client *service.WebSocketClient
	if userExists {
		userIDInt := userID.(int)
		client = &service.WebSocketClient{
			ID:       fmt.Sprintf("user-%d", userIDInt),
			Conn:     conn,
			Hub:      h.WSHub,
			Send:     make(chan []byte, 256),
			UserID:   &userIDInt,
			LastPing: time.Now(),
		}
	} else {
		adminIDInt := adminID.(int)
		client = &service.WebSocketClient{
			ID:       fmt.Sprintf("admin-%d", adminIDInt),
			Conn:     conn,
			Hub:      h.WSHub,
			Send:     make(chan []byte, 256),
			AdminID:  &adminIDInt,
			LastPing: time.Now(),
		}
	}

	// Register client with hub
	h.WSHub.RegisterClient(client)

	// Start read and write pumps
	go client.WritePump()
	go client.ReadPump()
}

// ========== METHOD ALIASES FOR BACKWARDS COMPATIBILITY ==========

// GetUserStats is an alias for GetUserNotificationStats
func (h *NotificationAPIHandlers) GetUserStats(c *gin.Context) {
	h.GetUserNotificationStats(c)
}

// GetUserPreferences is an alias for GetUserNotificationPreferences
func (h *NotificationAPIHandlers) GetUserPreferences(c *gin.Context) {
	h.GetUserNotificationPreferences(c)
}

// UpdateUserPreferences is an alias for UpdateUserNotificationPreferences
func (h *NotificationAPIHandlers) UpdateUserPreferences(c *gin.Context) {
	h.UpdateUserNotificationPreferences(c)
}

// GetAdminStats is an alias for GetAdminNotificationStats
func (h *NotificationAPIHandlers) GetAdminStats(c *gin.Context) {
	h.GetAdminNotificationStats(c)
}

// GetAdminPreferences is an alias for GetAdminNotificationPreferences
func (h *NotificationAPIHandlers) GetAdminPreferences(c *gin.Context) {
	h.GetAdminNotificationPreferences(c)
}

// UpdateAdminPreferences is an alias for UpdateAdminNotificationPreferences
func (h *NotificationAPIHandlers) UpdateAdminPreferences(c *gin.Context) {
	h.UpdateAdminNotificationPreferences(c)
}

// RegisterNotificationAPIRoutes registers all notification API routes
func RegisterNotificationAPIRoutes(router *gin.Engine, handlers *NotificationAPIHandlers, userAuth, adminAuth gin.HandlerFunc) {
	// WebSocket endpoint (must be registered before /api/v1 to avoid conflicts)
	router.GET("/ws/notifications", handlers.HandleWebSocketConnection)

	// User notification endpoints
	userAPI := router.Group("/api/v1/user")
	if userAuth != nil {
		userAPI.Use(userAuth)
	}
	{
		notifications := userAPI.Group("/notifications")
		{
			notifications.GET("", handlers.GetUserNotifications)
			notifications.GET("/unread", handlers.GetUserUnreadNotifications)
			notifications.GET("/count", handlers.GetUserUnreadCount)
			notifications.GET("/stats", handlers.GetUserNotificationStats)

			notifications.PATCH("/:id/read", handlers.MarkUserNotificationRead)
			notifications.PATCH("/read", handlers.MarkAllUserNotificationsRead)

			notifications.PATCH("/:id/dismiss", handlers.DismissUserNotification)
			notifications.DELETE("/:id", handlers.DeleteUserNotification)

			notifications.GET("/preferences", handlers.GetUserNotificationPreferences)
			notifications.PATCH("/preferences", handlers.UpdateUserNotificationPreferences)
		}
	}

	// Admin notification endpoints
	adminAPI := router.Group("/api/v1/admin")
	if adminAuth != nil {
		adminAPI.Use(adminAuth)
	}
	{
		notifications := adminAPI.Group("/notifications")
		{
			notifications.GET("", handlers.GetAdminNotifications)
			notifications.GET("/unread", handlers.GetAdminUnreadNotifications)
			notifications.GET("/count", handlers.GetAdminUnreadCount)
			notifications.GET("/stats", handlers.GetAdminNotificationStats)

			notifications.PATCH("/:id/read", handlers.MarkAdminNotificationRead)
			notifications.PATCH("/read", handlers.MarkAllAdminNotificationsRead)

			notifications.PATCH("/:id/dismiss", handlers.DismissAdminNotification)
			notifications.DELETE("/:id", handlers.DeleteAdminNotification)

			notifications.GET("/preferences", handlers.GetAdminNotificationPreferences)
			notifications.PATCH("/preferences", handlers.UpdateAdminNotificationPreferences)

			notifications.POST("/send", handlers.SendTestNotification)
		}
	}
}

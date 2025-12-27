package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apimgr/weather/src/server/handler"
	"github.com/apimgr/weather/src/server/model"
	"github.com/apimgr/weather/src/server/service"
	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func setupNotificationAPITest(t *testing.T) (*gin.Engine, *sql.DB, *sql.DB, *services.NotificationService) {
	gin.SetMode(gin.TestMode)

	// Create test databases
	userDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open user database: %v", err)
	}

	serverDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open server database: %v", err)
	}

	// Create tables
	createNotificationTables(t, userDB, serverDB)

	// Create WebSocket hub and notification service
	wsHub := services.NewWebSocketHub()
	go wsHub.Run()

	notificationService := &services.NotificationService{
		UserDB:     userDB,
		ServerDB:   serverDB,
		WSHub:      wsHub,
		UserNotif:  &models.UserNotificationModel{DB: userDB},
		AdminNotif: &models.AdminNotificationModel{DB: serverDB},
		Prefs:      &models.NotificationPreferencesModel{UserDB: userDB, ServerDB: serverDB},
	}

	// Create handlers
	notificationAPIHandler := &handlers.NotificationAPIHandlers{
		NotificationService: notificationService,
		WSHub:               wsHub,
	}

	// Create router
	r := gin.New()

	// User notification routes
	user := r.Group("/api/v1/user/notifications")
	user.Use(mockAuthMiddleware(1, false)) // Mock user ID = 1
	{
		user.GET("", notificationAPIHandler.GetUserNotifications)
		user.GET("/unread", notificationAPIHandler.GetUserUnreadNotifications)
		user.GET("/count", notificationAPIHandler.GetUserUnreadCount)
		user.GET("/stats", notificationAPIHandler.GetUserStats)
		user.PATCH("/:id/read", notificationAPIHandler.MarkUserNotificationRead)
		user.PATCH("/read", notificationAPIHandler.MarkAllUserNotificationsRead)
		user.PATCH("/:id/dismiss", notificationAPIHandler.DismissUserNotification)
		user.DELETE("/:id", notificationAPIHandler.DeleteUserNotification)
		user.GET("/preferences", notificationAPIHandler.GetUserPreferences)
		user.PATCH("/preferences", notificationAPIHandler.UpdateUserPreferences)
	}

	// Admin notification routes
	admin := r.Group("/api/v1/admin/notifications")
	admin.Use(mockAuthMiddleware(1, true)) // Mock admin ID = 1
	{
		admin.GET("", notificationAPIHandler.GetAdminNotifications)
		admin.GET("/unread", notificationAPIHandler.GetAdminUnreadNotifications)
		admin.GET("/count", notificationAPIHandler.GetAdminUnreadCount)
		admin.GET("/stats", notificationAPIHandler.GetAdminStats)
		admin.PATCH("/:id/read", notificationAPIHandler.MarkAdminNotificationRead)
		admin.PATCH("/read", notificationAPIHandler.MarkAllAdminNotificationsRead)
		admin.PATCH("/:id/dismiss", notificationAPIHandler.DismissAdminNotification)
		admin.DELETE("/:id", notificationAPIHandler.DeleteAdminNotification)
		admin.GET("/preferences", notificationAPIHandler.GetAdminPreferences)
		admin.PATCH("/preferences", notificationAPIHandler.UpdateAdminPreferences)
		admin.POST("/send", notificationAPIHandler.SendTestNotification)
	}

	return r, userDB, serverDB, notificationService
}

func createNotificationTables(t *testing.T, userDB, serverDB *sql.DB) {
	// Create user_notifications table
	_, err := userDB.Exec(`
		CREATE TABLE user_notifications (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('success', 'info', 'warning', 'error', 'security')),
			display TEXT NOT NULL CHECK(display IN ('toast', 'banner', 'center')) DEFAULT 'toast',
			title TEXT NOT NULL,
			message TEXT NOT NULL,
			action_json TEXT,
			read BOOLEAN DEFAULT 0,
			dismissed BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create user_notifications table: %v", err)
	}

	// Create server_admin_notifications table
	_, err = serverDB.Exec(`
		CREATE TABLE server_admin_notifications (
			id TEXT PRIMARY KEY,
			admin_id INTEGER NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('success', 'info', 'warning', 'error', 'security')),
			display TEXT NOT NULL CHECK(display IN ('toast', 'banner', 'center')) DEFAULT 'toast',
			title TEXT NOT NULL,
			message TEXT NOT NULL,
			action_json TEXT,
			read BOOLEAN DEFAULT 0,
			dismissed BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create server_admin_notifications table: %v", err)
	}

	// Create user_notification_preferences table
	_, err = userDB.Exec(`
		CREATE TABLE user_notification_preferences (
			user_id INTEGER PRIMARY KEY,
			enable_toast BOOLEAN DEFAULT 1,
			enable_banner BOOLEAN DEFAULT 1,
			enable_center BOOLEAN DEFAULT 1,
			enable_sound BOOLEAN DEFAULT 0,
			toast_duration_success INTEGER DEFAULT 5,
			toast_duration_info INTEGER DEFAULT 5,
			toast_duration_warning INTEGER DEFAULT 10,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create user_notification_preferences table: %v", err)
	}

	// Create server_admin_notification_preferences table
	_, err = serverDB.Exec(`
		CREATE TABLE server_admin_notification_preferences (
			admin_id INTEGER PRIMARY KEY,
			enable_toast BOOLEAN DEFAULT 1,
			enable_banner BOOLEAN DEFAULT 1,
			enable_center BOOLEAN DEFAULT 1,
			enable_sound BOOLEAN DEFAULT 0,
			toast_duration_success INTEGER DEFAULT 5,
			toast_duration_info INTEGER DEFAULT 5,
			toast_duration_warning INTEGER DEFAULT 10,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create server_admin_notification_preferences table: %v", err)
	}
}

func mockAuthMiddleware(id int, isAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isAdmin {
			c.Set("admin_id", id)
		} else {
			c.Set("user_id", id)
		}
		c.Next()
	}
}

func TestUserNotificationAPI_GetNotifications(t *testing.T) {
	r, userDB, serverDB, service := setupNotificationAPITest(t)
	defer userDB.Close()
	defer serverDB.Close()

	// Create test notifications
	_, _ = service.SendUserSuccess(1, "Test 1", "Message 1", nil)
	_, _ = service.SendUserInfo(1, "Test 2", "Message 2", nil)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/user/notifications", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	var response []models.Notification
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Response length = %v, want 2", len(response))
	}
}

func TestUserNotificationAPI_GetUnreadCount(t *testing.T) {
	r, userDB, serverDB, service := setupNotificationAPITest(t)
	defer userDB.Close()
	defer serverDB.Close()

	// Create test notifications
	notif1, _ := service.SendUserSuccess(1, "Test 1", "Message 1", nil)
	_, _ = service.SendUserInfo(1, "Test 2", "Message 2", nil)

	// Mark one as read
	_ = service.MarkUserNotificationRead(notif1.ID, 1)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/user/notifications/count", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	var response map[string]int
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["unread_count"] != 1 {
		t.Errorf("Unread count = %v, want 1", response["unread_count"])
	}
}

func TestUserNotificationAPI_MarkAsRead(t *testing.T) {
	r, userDB, serverDB, service := setupNotificationAPITest(t)
	defer userDB.Close()
	defer serverDB.Close()

	// Create test notification
	notif, _ := service.SendUserSuccess(1, "Test", "Message", nil)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/user/notifications/"+notif.ID+"/read", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify it's marked as read
	retrieved, _ := service.UserNotif.GetByID(notif.ID, 1)
	if !retrieved.Read {
		t.Error("Notification should be marked as read")
	}
}

func TestUserNotificationAPI_MarkAllAsRead(t *testing.T) {
	r, userDB, serverDB, service := setupNotificationAPITest(t)
	defer userDB.Close()
	defer serverDB.Close()

	// Create test notifications
	_, _ = service.SendUserSuccess(1, "Test 1", "Message 1", nil)
	_, _ = service.SendUserInfo(1, "Test 2", "Message 2", nil)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/user/notifications/read", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify all are marked as read
	count, _ := service.GetUserUnreadCount(1)
	if count != 0 {
		t.Errorf("Unread count = %v, want 0", count)
	}
}

func TestUserNotificationAPI_DismissNotification(t *testing.T) {
	r, userDB, serverDB, service := setupNotificationAPITest(t)
	defer userDB.Close()
	defer serverDB.Close()

	// Create test notification
	notif, _ := service.SendUserSuccess(1, "Test", "Message", nil)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/user/notifications/"+notif.ID+"/dismiss", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify it's dismissed
	retrieved, _ := service.UserNotif.GetByID(notif.ID, 1)
	if !retrieved.Dismissed {
		t.Error("Notification should be dismissed")
	}
}

func TestUserNotificationAPI_DeleteNotification(t *testing.T) {
	r, userDB, serverDB, service := setupNotificationAPITest(t)
	defer userDB.Close()
	defer serverDB.Close()

	// Create test notification
	notif, _ := service.SendUserSuccess(1, "Test", "Message", nil)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/user/notifications/"+notif.ID, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify it's deleted
	_, err := service.UserNotif.GetByID(notif.ID, 1)
	if err == nil {
		t.Error("Notification should be deleted")
	}
}

func TestUserNotificationAPI_GetPreferences(t *testing.T) {
	r, userDB, serverDB, _ := setupNotificationAPITest(t)
	defer userDB.Close()
	defer serverDB.Close()

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/user/notifications/preferences", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	var prefs models.NotificationPreferences
	err := json.Unmarshal(w.Body.Bytes(), &prefs)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check defaults
	if !prefs.EnableToast {
		t.Error("Default EnableToast should be true")
	}
	if prefs.ToastDurationSuccess != 5 {
		t.Errorf("Default ToastDurationSuccess = %v, want 5", prefs.ToastDurationSuccess)
	}
}

func TestUserNotificationAPI_UpdatePreferences(t *testing.T) {
	r, userDB, serverDB, service := setupNotificationAPITest(t)
	defer userDB.Close()
	defer serverDB.Close()

	// Prepare request body
	updateData := models.NotificationPreferences{
		EnableToast:           false,
		EnableBanner:          true,
		EnableCenter:          true,
		EnableSound:           true,
		ToastDurationSuccess:  3,
		ToastDurationInfo:     7,
		ToastDurationWarning:  15,
	}

	body, _ := json.Marshal(updateData)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/user/notifications/preferences", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify preferences were updated
	prefs, err := service.GetUserPreferences(1)
	if err != nil {
		t.Fatalf("Failed to get preferences: %v", err)
	}

	if prefs.EnableToast {
		t.Error("EnableToast should be false after update")
	}
	if !prefs.EnableSound {
		t.Error("EnableSound should be true after update")
	}
	if prefs.ToastDurationSuccess != 3 {
		t.Errorf("ToastDurationSuccess = %v, want 3", prefs.ToastDurationSuccess)
	}
}

func TestUserNotificationAPI_GetStatistics(t *testing.T) {
	r, userDB, serverDB, service := setupNotificationAPITest(t)
	defer userDB.Close()
	defer serverDB.Close()

	// Create test notifications
	_, _ = service.SendUserSuccess(1, "Success", "Message", nil)
	_, _ = service.SendUserInfo(1, "Info", "Message", nil)
	notif3, _ := service.SendUserWarning(1, "Warning", "Message", nil, models.NotificationDisplayToast)

	// Mark one as read
	_ = service.MarkUserNotificationRead(notif3.ID, 1)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/user/notifications/stats", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	var stats models.NotificationStatistics
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if stats.Total != 3 {
		t.Errorf("Total = %v, want 3", stats.Total)
	}
	if stats.Unread != 2 {
		t.Errorf("Unread = %v, want 2", stats.Unread)
	}
	if stats.Read != 1 {
		t.Errorf("Read = %v, want 1", stats.Read)
	}
}

func TestAdminNotificationAPI_SendTestNotification(t *testing.T) {
	r, userDB, serverDB, service := setupNotificationAPITest(t)
	defer userDB.Close()
	defer serverDB.Close()

	// Prepare request body
	testData := map[string]interface{}{
		"type":    "success",
		"display": "toast",
		"title":   "Test Notification",
		"message": "This is a test notification",
	}

	body, _ := json.Marshal(testData)

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/admin/notifications/send", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusCreated)
		t.Logf("Response body: %s", w.Body.String())
	}

	// Verify notification was created
	count, _ := service.GetAdminUnreadCount(1)
	if count != 1 {
		t.Errorf("Admin unread count = %v, want 1", count)
	}
}

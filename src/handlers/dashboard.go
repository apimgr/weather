package handlers

import (
	"database/sql"
	"net/http"

	"weather-go/src/middleware"
	"weather-go/src/models"
	"weather-go/src/utils"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	DB *sql.DB
}

// ShowDashboard renders the user dashboard
func (h *DashboardHandler) ShowDashboard(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get user's saved locations
	locationModel := &models.LocationModel{DB: h.DB}
	locations, err := locationModel.GetByUserID(user.ID)
	if err != nil {
		locations = []*models.SavedLocation{} // Empty array on error
	}

	// Get unread notification count
	notificationModel := &models.NotificationModel{DB: h.DB}
	unreadCount, err := notificationModel.GetUnreadCount(user.ID)
	if err != nil {
		unreadCount = 0
	}

	c.HTML(http.StatusOK, "dashboard.html", utils.TemplateData(c, gin.H{
		"title":         "Dashboard - Weather Service",
		"user":          user,
		"locations":     locations,
		"unreadCount":   unreadCount,
		"locationCount": len(locations),
		"page":          "dashboard",
	}))
}

// ShowAdminPanel renders the admin panel
func (h *DashboardHandler) ShowAdminPanel(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok || user.Role != "admin" {
		c.Redirect(http.StatusFound, "/user/dashboard")
		return
	}

	// Get system statistics
	userModel := &models.UserModel{DB: h.DB}

	totalUsers, _ := userModel.Count()
	adminCount, _ := userModel.CountByRole("admin")

	// Count total locations across all users
	var totalLocations int
	h.DB.QueryRow("SELECT COUNT(*) FROM saved_locations").Scan(&totalLocations)

	c.HTML(http.StatusOK, "admin.html", utils.TemplateData(c, gin.H{
		"title":          "Admin Panel - Weather Service",
		"user":           user,
		"totalUsers":     totalUsers,
		"adminCount":     adminCount,
		"totalLocations": totalLocations,
		"page":           "admin",
	}))
}

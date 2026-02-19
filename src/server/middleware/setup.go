package middleware

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/apimgr/weather/src/config"
	"github.com/apimgr/weather/src/database"
	"github.com/apimgr/weather/src/paths"
	"github.com/apimgr/weather/src/utils"

	"github.com/gin-gonic/gin"
)

// AdminSetupRequired checks if admin setup is complete when accessing admin routes
// AI.md: Server is FULLY FUNCTIONAL without setup - only admin panel requires setup
// AI.md: Setup flow is at /{admin_path}/server/setup, requires setup token
func AdminSetupRequired(db *sql.DB, cfg *config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		adminPath := "/" + cfg.GetAdminPath()

		// Only apply to admin routes
		if !strings.HasPrefix(path, adminPath) {
			c.Next()
			return
		}

		// Skip check for setup routes within admin
		setupPath := adminPath + "/server/setup"
		if strings.HasPrefix(path, setupPath) {
			c.Next()
			return
		}

		// Check if any admin exists (setup is complete when admin exists)
		var count int
		err := database.GetServerDB().QueryRow("SELECT COUNT(*) FROM server_admin_credentials").Scan(&count)
		if err != nil {
			c.Next()
			return
		}

		// If no admin exists, redirect to setup wizard
		if count == 0 {
			c.Redirect(http.StatusFound, setupPath)
			c.Abort()
			return
		}

		c.Next()
	}
}

// BlockSetupAfterComplete blocks access to setup routes after server setup is complete
// AI.md: Setup token file deleted after successful setup completion
func BlockSetupAfterComplete(db *sql.DB, cfg *config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if setup token file still exists
		configDir := paths.GetConfigDir()
		if !utils.SetupTokenExists(configDir) {
			// Setup complete (token file deleted), redirect to admin dashboard
			adminPath := "/" + cfg.GetAdminPath()
			c.Redirect(http.StatusFound, adminPath+"/dashboard")
			c.Abort()
			return
		}

		c.Next()
	}
}

// BlockSetupAfterAdminExists blocks access to admin setup if admin account already exists
func BlockSetupAfterAdminExists(db *sql.DB, cfg *config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if admin exists in server_admin_credentials
		var count int
		err := database.GetServerDB().QueryRow("SELECT COUNT(*) FROM server_admin_credentials").Scan(&count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			c.Abort()
			return
		}

		// If admin exists, redirect to admin dashboard
		if count > 0 {
			adminPath := "/" + cfg.GetAdminPath()
			c.Redirect(http.StatusFound, adminPath+"/dashboard")
			c.Abort()
			return
		}

		c.Next()
	}
}

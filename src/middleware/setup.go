package middleware

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CheckFirstUserSetup checks if any users exist and redirects to setup if needed
// Only applies to web/HTML requests, not API requests
func CheckFirstUserSetup(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip check for setup routes, static files, and API
		if strings.HasPrefix(path, "/user/setup") ||
			strings.HasPrefix(path, "/static/") ||
			strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/healthz") {
			c.Next()
			return
		}

		// Only apply to HTML requests (check Accept header)
		accept := c.GetHeader("Accept")
		if !strings.Contains(accept, "text/html") && accept != "" {
			c.Next()
			return
		}

		// Check if admin user exists (setup is complete when any admin exists)
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
		if err != nil {
			c.Next()
			return
		}

		// If no admin exists, redirect to setup
		if count == 0 {
			c.Redirect(http.StatusFound, "/user/setup")
			c.Abort()
			return
		}

		c.Next()
	}
}

// BlockSetupAfterComplete blocks access to setup routes after server setup is complete
func BlockSetupAfterComplete(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if server setup is marked as complete
		var setupComplete string
		err := db.QueryRow("SELECT value FROM settings WHERE key = 'setup.completed'").Scan(&setupComplete)

		// If setting exists and is true, setup is complete
		if err == nil && setupComplete == "true" {
			c.Redirect(http.StatusFound, "/admin")
			c.Abort()
			return
		}

		c.Next()
	}
}

// BlockSetupAfterAdminExists blocks access to admin setup if admin account already exists
func BlockSetupAfterAdminExists(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if admin user exists
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			c.Abort()
			return
		}

		// If admin exists, redirect to appropriate page
		if count > 0 {
			// Check if current user is admin
			user, ok := GetCurrentUser(c)
			if ok && user.Role == "admin" {
				c.Redirect(http.StatusFound, "/admin")
			} else {
				c.Redirect(http.StatusFound, "/user/dashboard")
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

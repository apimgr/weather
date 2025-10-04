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

		// Check if admin user exists (setup is complete when 'administrator' exists)
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'administrator'").Scan(&count)
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

// BlockSetupAfterComplete blocks access to setup routes after setup is complete
func BlockSetupAfterComplete(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if admin user exists (username='administrator')
		// Setup is only complete when the dedicated admin account exists
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'administrator'").Scan(&count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			c.Abort()
			return
		}

		// If admin exists, setup is complete - redirect to login
		if count > 0 {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}

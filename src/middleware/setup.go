package middleware

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CheckFirstUserSetup checks if any users exist and redirects to setup if needed
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

		// Check if any users exist
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		if err != nil {
			c.Next()
			return
		}

		// If no users exist, redirect to setup
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
		// Check if any users exist
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			c.Abort()
			return
		}

		// If users exist, setup is complete - return 404
		if count > 0 {
			c.HTML(http.StatusNotFound, "error.html", gin.H{
				"Title":   "Not Found",
				"Message": "Setup already completed",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

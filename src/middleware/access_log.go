package middleware

import (
	"time"
	"weather-go/src/utils"

	"github.com/gin-gonic/gin"
)

// AccessLogger creates middleware for logging HTTP requests
func AccessLogger(logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Extract request details
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		protocol := c.Request.Proto
		statusCode := c.Writer.Status()
		bodySize := int64(c.Writer.Size())
		referer := c.Request.Referer()
		userAgent := c.Request.UserAgent()

		// Get username from context (if authenticated)
		username := ""
		if user, exists := c.Get("user"); exists {
			if userMap, ok := user.(map[string]interface{}); ok {
				if uname, ok := userMap["username"].(string); ok {
					username = uname
				}
			}
		}

		// Log access
		logger.Access(clientIP, username, method, path, protocol, statusCode, bodySize, referer, userAgent)

		// Also log slow requests as warnings
		if duration > 1*time.Second {
			logger.Error("Slow request: %s %s took %v", method, path, duration)
		}
	}
}

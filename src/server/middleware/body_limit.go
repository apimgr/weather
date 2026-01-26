// Package middleware - Request body size limiter per AI.md PART 18
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodySizeLimit constants per AI.md PART 18 line 15691
const (
	// DefaultMaxBodySize is 10MB per AI.md PART 18
	DefaultMaxBodySize = 10 << 20
)

// BodySizeLimitMiddleware limits request body size per AI.md PART 18 line 15691
// Default: 10MB (max_body_size: 10MB)
func BodySizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	if maxSize <= 0 {
		maxSize = DefaultMaxBodySize
	}

	return func(c *gin.Context) {
		// Skip body size check for GET, HEAD, OPTIONS requests
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Check Content-Length header if present
		if c.Request.ContentLength > maxSize {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":  "Request body too large",
				"code":   "BODY_TOO_LARGE",
				"status": http.StatusRequestEntityTooLarge,
			})
			return
		}

		// Wrap body with size limiter
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()
	}
}

// Package middleware - Security headers per AI.md PART 22
// AI.md Reference: Lines 17695-17718
package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security headers to all responses
// Per AI.md lines 17697-17706
func SecurityHeaders(sslEnabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Per AI.md line 17697: "All responses MUST include:"
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Per AI.md lines 17708-17712: "When SSL is enabled, also include:"
		if sslEnabled {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

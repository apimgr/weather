package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security headers to all responses per TEMPLATE.md requirements
// These headers protect against common web vulnerabilities:
// - XSS attacks
// - Clickjacking
// - MIME type sniffing
// - Information disclosure
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		// Browsers should not try to detect content type, trust the Content-Type header
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking attacks
		// Deny embedding this site in frames/iframes from other origins
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// XSS Protection (legacy browsers)
		// Modern browsers use CSP instead, but this helps older browsers
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content Security Policy
		// Restrictive policy that only allows resources from same origin
		// Allows inline styles for Dracula theme and inline scripts for admin panels
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'self'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
		c.Header("Content-Security-Policy", csp)

		// Referrer Policy
		// Only send referrer for same-origin requests
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy (formerly Feature-Policy)
		// Disable potentially dangerous browser features
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

		// Strict Transport Security (HSTS)
		// Force HTTPS for 1 year, include subdomains
		// Only set if connection is HTTPS
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Remove server identification header for security
		// Don't leak server software version
		c.Header("Server", "")

		// Cross-Origin policies
		// Prevent other sites from embedding resources
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// Continue processing request
		c.Next()
	}
}

// SecurityHeadersAPI adds security headers optimized for API endpoints
// Less restrictive CSP since API endpoints don't serve HTML
func SecurityHeadersAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Same basic security headers
		c.Header("X-Content-Type-Options", "nosniff")
		// API responses should never be framed
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		// Simpler CSP for API (no inline scripts/styles needed for JSON responses)
		c.Header("Content-Security-Policy", "default-src 'none'")

		// API-specific headers
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Server", "")

		// HSTS for HTTPS API requests
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// Package middleware - Security middleware per AI.md PART 22
// AI.md Reference: Lines 17693-19697 (Security & Logging)
package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CSRFConfig holds CSRF protection configuration
// Per AI.md lines 14787-14799
type CSRFConfig struct {
	Enabled     bool
	TokenLength int
	CookieName  string
	HeaderName  string
	Secure      string
}

// DefaultCSRFConfig returns default CSRF configuration
// Per AI.md lines 14791-14798
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		Enabled:     true,
		TokenLength: 32,
		CookieName:  "csrf_token",
		HeaderName:  "X-CSRF-Token",
		Secure:      "auto",
	}
}

// CSRFProtection provides CSRF protection middleware
// Per AI.md line 14785: "ALL forms MUST have CSRF protection"
func CSRFProtection(cfg CSRFConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if disabled
		if !cfg.Enabled {
			c.Next()
			return
		}

		// GET requests don't need validation, but should get token
		if c.Request.Method == "GET" {
			// Generate token if not present
			token, err := c.Cookie(cfg.CookieName)
			if err != nil || token == "" {
				token = generateCSRFToken(cfg.TokenLength)
				setCSRFCookie(c, cfg, token)
			}

			// Make token available to templates
			c.Set("csrf_token", token)
			c.Next()
			return
		}

		// Per AI.md line 14804: "All non-GET requests validate CSRF token"
		cookieToken, err := c.Cookie(cfg.CookieName)
		if err != nil || cookieToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "CSRF token missing",
			})
			return
		}

		// Check token from form or header
		// Per AI.md line 14805: "Token stored in cookie and must match form/header value"
		formToken := c.PostForm("csrf_token")
		headerToken := c.GetHeader(cfg.HeaderName)

		requestToken := formToken
		if requestToken == "" {
			requestToken = headerToken
		}

		if requestToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "CSRF token not provided",
			})
			return
		}

		if requestToken != cookieToken {
			// Per AI.md line 18164: Log CSRF failures to audit
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "CSRF token validation failed",
			})
			return
		}

		c.Next()
	}
}

// generateCSRFToken generates a random CSRF token
func generateCSRFToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// setCSRFCookie sets the CSRF token cookie
func setCSRFCookie(c *gin.Context, cfg CSRFConfig, token string) {
	secure := false
	if cfg.Secure == "auto" {
		// Auto-detect based on scheme
		secure = c.Request.TLS != nil
	} else if cfg.Secure == "true" {
		secure = true
	}

	c.SetCookie(
		cfg.CookieName,
		token,
		3600,
		"/",
		"",
		secure,
		true,
	)
}

// RegenerateCSRFToken regenerates CSRF token (call on login)
// Per AI.md line 14806: "Tokens regenerated on login"
func RegenerateCSRFToken(c *gin.Context, cfg CSRFConfig) {
	token := generateCSRFToken(cfg.TokenLength)
	setCSRFCookie(c, cfg, token)
	c.Set("csrf_token", token)
}

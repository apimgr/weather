package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DefaultAnonymousLimit = 120  // requests per hour
	DefaultWindowSeconds  = 3600 // 1 hour
)

// RateLimitMiddleware implements rate limiting for anonymous users
func RateLimitMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for authenticated users with API tokens
		authMethod, exists := c.Get("auth_method")
		if exists && authMethod == "api_token" {
			c.Next()
			return
		}

		// Get identifier (IP address for anonymous users)
		identifier := c.ClientIP()
		endpoint := c.Request.Method + " " + c.Request.URL.Path

		// Get rate limit settings from database
		limit, window := getRateLimitSettings(db)

		// Check rate limit
		allowed, remaining, resetTime, err := checkRateLimit(db, identifier, endpoint, limit, window)
		if err != nil {
			// Log error but don't block request
			fmt.Printf("Rate limit error: %v\n", err)
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("You have exceeded the rate limit of %d requests per hour", limit),
				"retry_after": resetTime.Sub(time.Now()).Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getRateLimitSettings retrieves rate limit configuration from database
func getRateLimitSettings(db *sql.DB) (int, int) {
	var limitStr, windowStr string

	err := db.QueryRow("SELECT value FROM settings WHERE key = ?", "rate_limit.anonymous").Scan(&limitStr)
	if err != nil {
		return DefaultAnonymousLimit, DefaultWindowSeconds
	}

	err = db.QueryRow("SELECT value FROM settings WHERE key = ?", "rate_limit.window").Scan(&windowStr)
	if err != nil {
		windowStr = strconv.Itoa(DefaultWindowSeconds)
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = DefaultAnonymousLimit
	}

	window, err := strconv.Atoi(windowStr)
	if err != nil {
		window = DefaultWindowSeconds
	}

	return limit, window
}

// checkRateLimit checks if request is allowed and updates counters
func checkRateLimit(db *sql.DB, identifier, endpoint string, limit, windowSeconds int) (bool, int, time.Time, error) {
	now := time.Now()
	windowStart := now.Truncate(time.Duration(windowSeconds) * time.Second)
	resetTime := windowStart.Add(time.Duration(windowSeconds) * time.Second)

	// Get current count for this window
	var count int
	err := db.QueryRow(`
		SELECT count FROM rate_limits
		WHERE identifier = ? AND endpoint = ? AND window_start = ?
	`, identifier, endpoint, windowStart).Scan(&count)

	if err == sql.ErrNoRows {
		// First request in this window, insert new record
		_, err = db.Exec(`
			INSERT INTO rate_limits (identifier, endpoint, count, window_start)
			VALUES (?, ?, 1, ?)
		`, identifier, endpoint, windowStart)
		if err != nil {
			return false, 0, resetTime, err
		}
		return true, limit - 1, resetTime, nil
	} else if err != nil {
		return false, 0, resetTime, err
	}

	// Check if limit exceeded
	if count >= limit {
		return false, 0, resetTime, nil
	}

	// Increment counter
	_, err = db.Exec(`
		UPDATE rate_limits
		SET count = count + 1
		WHERE identifier = ? AND endpoint = ? AND window_start = ?
	`, identifier, endpoint, windowStart)
	if err != nil {
		return false, 0, resetTime, err
	}

	remaining := limit - count - 1
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, resetTime, nil
}

// CleanupOldRateLimits removes rate limit entries older than 2 hours
func CleanupOldRateLimits(db *sql.DB) error {
	cutoff := time.Now().Add(-2 * time.Hour)
	_, err := db.Exec("DELETE FROM rate_limits WHERE window_start < ?", cutoff)
	return err
}

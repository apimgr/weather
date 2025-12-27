package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/httprate"
)

const (
	// Global rate limit (all endpoints)
	// 100 requests per second
	GlobalRPS   = 100
	// Burst allowance
	GlobalBurst = 200

	// API rate limit (stricter for API)
	// 100 requests per 15 minutes
	APIRequestsPerWindow = 100
	APIWindowDuration    = 15 * time.Minute

	// Admin rate limit (most restrictive)
	// 30 requests per 15 minutes
	AdminRequestsPerWindow = 30
	AdminWindowDuration    = 15 * time.Minute
)

var (
	// Global rate limiter (applied to all routes)
	globalLimiter *httprate.RateLimiter

	// API rate limiter (applied to /api/* routes)
	apiLimiter *httprate.RateLimiter

	// Admin rate limiter (applied to /admin/* routes)
	adminLimiter *httprate.RateLimiter
)

func init() {
	// Initialize rate limiters
	globalLimiter = httprate.NewRateLimiter(
		GlobalRPS,
		time.Second,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	)

	apiLimiter = httprate.NewRateLimiter(
		APIRequestsPerWindow,
		APIWindowDuration,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	)

	adminLimiter = httprate.NewRateLimiter(
		AdminRequestsPerWindow,
		AdminWindowDuration,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	)
}

// GlobalRateLimitMiddleware applies global rate limiting (100 req/s)
func GlobalRateLimitMiddleware() gin.HandlerFunc {
	return wrapRateLimiter(globalLimiter, GlobalRPS, time.Second)
}

// APIRateLimitMiddleware applies API rate limiting (100 req/15min)
func APIRateLimitMiddleware() gin.HandlerFunc {
	return wrapRateLimiter(apiLimiter, APIRequestsPerWindow, APIWindowDuration)
}

// AdminRateLimitMiddleware applies admin rate limiting (30 req/15min)
func AdminRateLimitMiddleware() gin.HandlerFunc {
	return wrapRateLimiter(adminLimiter, AdminRequestsPerWindow, AdminWindowDuration)
}

// wrapRateLimiter wraps httprate.RateLimiter for Gin
func wrapRateLimiter(limiter *httprate.RateLimiter, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a wrapper to capture rate limit status
		rateLimitExceeded := false

		handler := limiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if rate limit headers were set by httprate
			if w.Header().Get("X-RateLimit-Remaining") == "0" {
				rateLimitExceeded = true
			}
			c.Next()
		}))

		// Create response writer wrapper
		writer := &rateLimitResponseWriter{
			ResponseWriter: c.Writer,
			ginContext:     c,
		}
		c.Writer = writer

		// Call httprate handler
		handler.ServeHTTP(writer, c.Request)

		// If rate limited, abort with 429
		if rateLimitExceeded {
			retryAfter := int(window.Seconds())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}

		// Set rate limit headers for successful requests
		c.Header("X-RateLimit-Limit", writer.Header().Get("X-RateLimit-Limit"))
		c.Header("X-RateLimit-Remaining", writer.Header().Get("X-RateLimit-Remaining"))
		c.Header("X-RateLimit-Reset", writer.Header().Get("X-RateLimit-Reset"))
	}
}

// rateLimitResponseWriter wraps gin.ResponseWriter to work with httprate
type rateLimitResponseWriter struct {
	gin.ResponseWriter
	ginContext *gin.Context
	statusCode int
}

func (w *rateLimitResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	if statusCode == http.StatusTooManyRequests {
		// Don't write header yet, let Gin handle it
		return
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *rateLimitResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode == http.StatusTooManyRequests {
		// Don't write body for rate limited requests
		// Gin will handle the JSON response
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}

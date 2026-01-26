package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/httprate"
)

// Rate limit constants per AI.md PART 1: Security-First Design
const (
	// Login attempts: 5 per 15 minutes
	LoginRequestsPerWindow = 5
	LoginWindowDuration    = 15 * time.Minute

	// Password reset: 3 per 1 hour
	PasswordResetRequestsPerWindow = 3
	PasswordResetWindowDuration    = 1 * time.Hour

	// API (authenticated): 100 per 1 minute per AI.md PART 1
	APIAuthRequestsPerWindow = 100
	APIAuthWindowDuration    = 1 * time.Minute

	// API (unauthenticated): 20 per 1 minute per AI.md PART 1
	APIUnauthRequestsPerWindow = 20
	APIUnauthWindowDuration    = 1 * time.Minute

	// Registration: 5 per 1 hour
	RegistrationRequestsPerWindow = 5
	RegistrationWindowDuration    = 1 * time.Hour

	// File upload: 10 per 1 hour
	FileUploadRequestsPerWindow = 10
	FileUploadWindowDuration    = 1 * time.Hour

	// Admin: 30 per 15 minutes
	AdminRequestsPerWindow = 30
	AdminWindowDuration    = 15 * time.Minute

	// Global rate limit (DDoS protection)
	GlobalRPS   = 100
	GlobalBurst = 200
)

var (
	// Rate limiters initialized in init()
	globalLimiter        *httprate.RateLimiter
	loginLimiter         *httprate.RateLimiter
	passwordResetLimiter *httprate.RateLimiter
	apiAuthLimiter       *httprate.RateLimiter
	apiUnauthLimiter     *httprate.RateLimiter
	registrationLimiter  *httprate.RateLimiter
	fileUploadLimiter    *httprate.RateLimiter
	adminLimiter         *httprate.RateLimiter
)

func init() {
	// Initialize all rate limiters per AI.md PART 1 specifications
	globalLimiter = httprate.NewRateLimiter(
		GlobalRPS,
		time.Second,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	)

	loginLimiter = httprate.NewRateLimiter(
		LoginRequestsPerWindow,
		LoginWindowDuration,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	)

	passwordResetLimiter = httprate.NewRateLimiter(
		PasswordResetRequestsPerWindow,
		PasswordResetWindowDuration,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	)

	apiAuthLimiter = httprate.NewRateLimiter(
		APIAuthRequestsPerWindow,
		APIAuthWindowDuration,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	)

	apiUnauthLimiter = httprate.NewRateLimiter(
		APIUnauthRequestsPerWindow,
		APIUnauthWindowDuration,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	)

	registrationLimiter = httprate.NewRateLimiter(
		RegistrationRequestsPerWindow,
		RegistrationWindowDuration,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	)

	fileUploadLimiter = httprate.NewRateLimiter(
		FileUploadRequestsPerWindow,
		FileUploadWindowDuration,
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

// LoginRateLimitMiddleware applies login rate limiting (5 req/15min)
func LoginRateLimitMiddleware() gin.HandlerFunc {
	return wrapRateLimiter(loginLimiter, LoginRequestsPerWindow, LoginWindowDuration)
}

// PasswordResetRateLimitMiddleware applies password reset rate limiting (3 req/1hr)
func PasswordResetRateLimitMiddleware() gin.HandlerFunc {
	return wrapRateLimiter(passwordResetLimiter, PasswordResetRequestsPerWindow, PasswordResetWindowDuration)
}

// APIAuthRateLimitMiddleware applies authenticated API rate limiting (100 req/1min)
func APIAuthRateLimitMiddleware() gin.HandlerFunc {
	return wrapRateLimiter(apiAuthLimiter, APIAuthRequestsPerWindow, APIAuthWindowDuration)
}

// APIUnauthRateLimitMiddleware applies unauthenticated API rate limiting (20 req/1min)
func APIUnauthRateLimitMiddleware() gin.HandlerFunc {
	return wrapRateLimiter(apiUnauthLimiter, APIUnauthRequestsPerWindow, APIUnauthWindowDuration)
}

// APIRateLimitMiddleware applies API rate limiting based on authentication status
func APIRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		_, exists := c.Get("user_id")
		if exists {
			// Authenticated: 100 req/min
			wrapRateLimiter(apiAuthLimiter, APIAuthRequestsPerWindow, APIAuthWindowDuration)(c)
		} else {
			// Unauthenticated: 20 req/min
			wrapRateLimiter(apiUnauthLimiter, APIUnauthRequestsPerWindow, APIUnauthWindowDuration)(c)
		}
	}
}

// RegistrationRateLimitMiddleware applies registration rate limiting (5 req/1hr)
func RegistrationRateLimitMiddleware() gin.HandlerFunc {
	return wrapRateLimiter(registrationLimiter, RegistrationRequestsPerWindow, RegistrationWindowDuration)
}

// FileUploadRateLimitMiddleware applies file upload rate limiting (10 req/1hr)
func FileUploadRateLimitMiddleware() gin.HandlerFunc {
	return wrapRateLimiter(fileUploadLimiter, FileUploadRequestsPerWindow, FileUploadWindowDuration)
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

package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIError represents a standardized error response per AI.md PART 19
type APIError struct {
	// Human-readable error message
	Error   string                 `json:"error"`
	// Machine-readable error code
	Code    string                 `json:"code"`
	// HTTP status code
	Status  int                    `json:"status"`
	// Optional additional context
	Details map[string]interface{} `json:"details,omitempty"`
}

// APISuccess represents a standardized success response per AI.md PART 19
type APISuccess struct {
	// Always true for successful actions
	Success bool                   `json:"success"`
	// Optional human-readable message
	Message string                 `json:"message,omitempty"`
	// ID of created/affected resource
	ID      string                 `json:"id,omitempty"`
	// Optional additional response data
	Data    map[string]interface{} `json:"data,omitempty"`
}

// Common error codes per AI.md PART 19
const (
	// Client errors (4xx)
	ErrInvalidInput     = "INVALID_INPUT"
	ErrNotFound         = "NOT_FOUND"
	ErrUnauthorized     = "UNAUTHORIZED"
	ErrForbidden        = "FORBIDDEN"
	ErrConflict         = "CONFLICT"
	ErrValidationFailed = "VALIDATION_FAILED"
	ErrRateLimited      = "RATE_LIMITED"
	ErrBadRequest       = "BAD_REQUEST"

	// Server errors (5xx)
	ErrInternal         = "INTERNAL_ERROR"
	ErrServiceUnavail   = "SERVICE_UNAVAILABLE"
	ErrDatabaseError    = "DATABASE_ERROR"
	ErrExternalService  = "EXTERNAL_SERVICE_ERROR"
)

// RespondError sends a standardized error response per AI.md PART 19
func RespondError(c *gin.Context, status int, code string, message string, details ...map[string]interface{}) {
	var detailsMap map[string]interface{}
	if len(details) > 0 {
		detailsMap = details[0]
	}

	errResponse := APIError{
		Error:   message,
		Code:    code,
		Status:  status,
		Details: detailsMap,
	}

	// AI.md PART 19: Support .txt extension for text responses
	if shouldRespondText(c) {
		c.String(status, "%s: %s\n", code, message)
		return
	}

	c.JSON(status, errResponse)
}

// RespondSuccess sends a standardized success response per AI.md PART 19
func RespondSuccess(c *gin.Context, message string, data ...map[string]interface{}) {
	response := APISuccess{
		Success: true,
		Message: message,
	}

	// Merge optional data
	if len(data) > 0 {
		for key, value := range data[0] {
			if key == "id" {
				if id, ok := value.(string); ok {
					response.ID = id
				}
			} else {
				if response.Data == nil {
					response.Data = make(map[string]interface{})
				}
				response.Data[key] = value
			}
		}
	}

	// AI.md PART 19: Support .txt extension for text responses
	if shouldRespondText(c) {
		c.String(http.StatusOK, "%s\n", message)
		return
	}

	c.JSON(http.StatusOK, response)
}

// RespondCreated sends a standardized response for resource creation
func RespondCreated(c *gin.Context, message string, id string, data ...map[string]interface{}) {
	response := APISuccess{
		Success: true,
		Message: message,
		ID:      id,
	}

	// Merge optional data
	if len(data) > 0 {
		response.Data = data[0]
	}

	// AI.md PART 19: Support .txt extension for text responses
	if shouldRespondText(c) {
		c.String(http.StatusCreated, "Created: %s (ID: %s)\n", message, id)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// RespondData sends a direct data response (no wrapper) per AI.md PART 19
func RespondData(c *gin.Context, data interface{}) {
	// AI.md PART 19: Support .txt extension for text responses
	if shouldRespondText(c) {
		// Convert data to string representation
		c.String(http.StatusOK, "%v\n", data)
		return
	}

	c.JSON(http.StatusOK, data)
}

// RespondPaginated sends a paginated response per AI.md PART 19
func RespondPaginated(c *gin.Context, data interface{}, page, limit, total int) {
	pages := total / limit
	if total%limit != 0 {
		pages++
	}

	response := gin.H{
		"data": data,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": pages,
		},
	}

	// AI.md PART 19: Support .txt extension for text responses
	if shouldRespondText(c) {
		c.String(http.StatusOK, "Page %d/%d (Total: %d items)\n", page, pages, total)
		return
	}

	c.JSON(http.StatusOK, response)
}

// shouldRespondText checks if the request wants text response per AI.md PART 19
// Supports both .txt extension and Accept: text/plain header
func shouldRespondText(c *gin.Context) bool {
	// Check URL extension: /api/v1/weather.txt
	path := c.Request.URL.Path
	ext := filepath.Ext(path)
	if ext == ".txt" {
		return true
	}

	// Check Accept header
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "text/plain") {
		return true
	}

	return false
}

// Helper functions for common error scenarios

// BadRequest returns a 400 Bad Request error
func BadRequest(c *gin.Context, message string, details ...map[string]interface{}) {
	RespondError(c, http.StatusBadRequest, ErrBadRequest, message, details...)
}

// InvalidInput returns a 400 Invalid Input error
func InvalidInput(c *gin.Context, message string, details ...map[string]interface{}) {
	RespondError(c, http.StatusBadRequest, ErrInvalidInput, message, details...)
}

// NotFound returns a 404 Not Found error
func NotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, ErrNotFound, message)
}

// Unauthorized returns a 401 Unauthorized error
func Unauthorized(c *gin.Context, message string) {
	RespondError(c, http.StatusUnauthorized, ErrUnauthorized, message)
}

// Forbidden returns a 403 Forbidden error
func Forbidden(c *gin.Context, message string) {
	RespondError(c, http.StatusForbidden, ErrForbidden, message)
}

// Conflict returns a 409 Conflict error
func Conflict(c *gin.Context, message string, details ...map[string]interface{}) {
	RespondError(c, http.StatusConflict, ErrConflict, message, details...)
}

// InternalError returns a 500 Internal Server Error
func InternalError(c *gin.Context, message string) {
	RespondError(c, http.StatusInternalServerError, ErrInternal, message)
}

// ValidationFailed returns a 422 Validation Failed error
func ValidationFailed(c *gin.Context, message string, details map[string]interface{}) {
	RespondError(c, http.StatusUnprocessableEntity, ErrValidationFailed, message, details)
}

// RateLimited returns a 429 Rate Limited error
func RateLimited(c *gin.Context, message string) {
	RespondError(c, http.StatusTooManyRequests, ErrRateLimited, message)
}

package utils

import (
	"github.com/gin-gonic/gin"
)

// ServerContext holds server information for templates
type ServerContext struct {
	Title       string
	Tagline     string
	Description string
	Version     string
}

// TemplateData enriches template data with server context and user info
func TemplateData(c *gin.Context, data gin.H) gin.H {
	// Get server context from Gin context (set by middleware)
	serverCtxInterface, exists := c.Get("server")

	var serverCtx interface{}
	if !exists {
		// Fallback to defaults
		serverCtx = map[string]string{
			"Title":       "Weather Service",
			"Tagline":     "Your personal weather dashboard",
			"Description": "Weather information service",
			"Version":     "unknown",
		}
	} else {
		serverCtx = serverCtxInterface
	}

	// Get user context from Gin context (set by auth middleware)
	userCtxInterface, userExists := c.Get("user")
	var userCtx interface{}
	if !userExists {
		// Fallback to empty user (guest)
		userCtx = map[string]string{
			"Email": "",
			"Role":  "guest",
		}
	} else {
		userCtx = userCtxInterface
	}

	// Create enriched data
	enriched := gin.H{
		"server": serverCtx,
		"user":   userCtx,
	}

	// Merge user-provided data
	for k, v := range data {
		enriched[k] = v
	}

	return enriched
}

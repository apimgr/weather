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

// TemplateData enriches template data with server context
func TemplateData(c *gin.Context, data gin.H) gin.H {
	// Get server context from Gin context (set by middleware)
	serverCtxInterface, exists := c.Get("server")

	var serverCtx interface{}
	if !exists {
		// Fallback to defaults
		serverCtx = map[string]string{
			"title":       "Weather Service",
			"tagline":     "Your personal weather dashboard",
			"description": "Weather information service",
			"version":     "unknown",
		}
	} else {
		serverCtx = serverCtxInterface
	}

	// Create enriched data
	enriched := gin.H{
		"server": serverCtx,
	}

	// Merge user-provided data
	for k, v := range data {
		enriched[k] = v
	}

	return enriched
}

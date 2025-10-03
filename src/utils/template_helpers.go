package utils

import (
	"weather-go/src/middleware"

	"github.com/gin-gonic/gin"
)

// TemplateData enriches template data with server context
func TemplateData(c *gin.Context, data gin.H) gin.H {
	// Get server context from middleware
	serverCtx, exists := middleware.GetServerContext(c)
	if !exists {
		// Fallback to defaults
		serverCtx = middleware.ServerContext{
			Title:       "Weather Service",
			Tagline:     "Your personal weather dashboard",
			Description: "Weather information service",
		}
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

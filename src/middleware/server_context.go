package middleware

import (
	"database/sql"

	"weather-go/src/models"

	"github.com/gin-gonic/gin"
)

// ServerContext holds server-wide configuration
type ServerContext struct {
	Title       string
	Tagline     string
	Description string
	Version     string
}

// InjectServerContext adds server configuration to all requests
func InjectServerContext(db *sql.DB, version string) gin.HandlerFunc {
	settingsModel := &models.SettingsModel{DB: db}

	return func(c *gin.Context) {
		// Get server settings
		title := settingsModel.GetString("server.title", "Weather Service")
		tagline := settingsModel.GetString("server.tagline", "Your personal weather dashboard")
		description := settingsModel.GetString("server.description", "A comprehensive platform for weather forecasts, moon phases, earthquakes, and hurricane tracking.")

		// Create server context
		serverCtx := ServerContext{
			Title:       title,
			Tagline:     tagline,
			Description: description,
			Version:     version,
		}

		// Add to gin context for handlers to use
		c.Set("server", serverCtx)

		c.Next()
	}
}

// GetServerContext retrieves server context from gin context
func GetServerContext(c *gin.Context) (ServerContext, bool) {
	serverCtx, exists := c.Get("server")
	if !exists {
		return ServerContext{
			Title:       "Weather Service",
			Tagline:     "Your personal weather dashboard",
			Description: "Weather information service",
			Version:     "unknown",
		}, false
	}

	ctx, ok := serverCtx.(ServerContext)
	return ctx, ok
}

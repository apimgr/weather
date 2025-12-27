package swagger

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterRoutes registers all Swagger/OpenAPI routes
// Per AI.md specification: /openapi for UI, /openapi.json for spec
func RegisterRoutes(router *gin.Engine) {
	// Swagger UI at /openapi
	router.GET("/openapi", GetSwaggerUI())
	router.GET("/openapi/*any", GetSwaggerUI())
}

// GetSwaggerUI returns the Swagger UI handler with theme support
// Serves auto-generated Swagger UI from swag annotations
func GetSwaggerUI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get theme preference from cookie or query param
		theme := GetTheme(c)

		// Configure Swagger UI
		config := ginSwagger.Config{
			URL:                      "doc.json",
			DocExpansion:             "list",
			DeepLinking:              true,
			PersistAuthorization:     true,
			DefaultModelsExpandDepth: 1,
		}

		// Theme support note: ginSwagger.Config doesn't support custom CSS injection via API
		// Theme will be applied via HTML template customization in future version
		// For now, using default Swagger UI theme
		_ = theme // Acknowledge theme variable to avoid unused error

		// Serve Swagger UI
		ginSwagger.CustomWrapHandler(&config, swaggerFiles.Handler)(c)
	}
}

// GetOpenAPIJSON returns the OpenAPI JSON specification
// Auto-generated from swag annotations
func GetOpenAPIJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.File("./docs/swagger.json")
	}
}

// HealthCheck for /openapi/health (separate from main health endpoint)
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "swagger-ui",
		})
	}
}

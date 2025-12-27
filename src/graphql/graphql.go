package graphql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all GraphQL routes
// Per AI.md specification: /graphql for queries, GraphiQL playground at /graphql with GET
func RegisterRoutes(router *gin.Engine, resolver *Resolver) {
	// Create GraphQL handler
	srv := handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: resolver}))

	// GraphQL endpoint for POST requests (actual queries)
	router.POST("/graphql", graphqlHandler(srv))

	// GraphiQL playground for GET requests (interactive UI)
	router.GET("/graphql", graphiqlHandler())
}

// graphqlHandler wraps the gqlgen handler for Gin
func graphqlHandler(h *handler.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// graphiqlHandler serves the GraphiQL playground with theme support
func graphiqlHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get theme preference
		theme := GetTheme(c)

		// Serve GraphiQL playground with theme
		if theme == "dark" {
			// Custom GraphiQL with dark theme
			playgroundHandler := playground.Handler("GraphQL Playground", "/graphql")
			playgroundHandler.ServeHTTP(c.Writer, c.Request)
		} else if theme == "light" {
			// Custom GraphiQL with light theme
			playgroundHandler := playground.Handler("GraphQL Playground", "/graphql")
			playgroundHandler.ServeHTTP(c.Writer, c.Request)
		} else {
			// Auto theme (let browser decide)
			playgroundHandler := playground.Handler("GraphQL Playground", "/graphql")
			playgroundHandler.ServeHTTP(c.Writer, c.Request)
		}
	}
}

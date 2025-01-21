package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Middleware to extract and validate API key
func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		// Simulated API key validation logic
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			return
		}

		// You can add logic here to validate the API key from a database or cache
		if apiKey != "test_api_key" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid API key"})
			return
		}

		// Store the API key in the context for later use
		c.Set("apiKey", apiKey)
		c.Next()
	}
}

package middleware

import (
	"WayPointPro/pkg/traffic"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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
		var cache = traffic.NewCache()
		// Check Redis cache first
		cacheKey := "api_key:" + apiKey
		val, err := cache.RedisClient.Get(cache.CTX, cacheKey).Result()

		if err == nil {
			// Cache hit: API key exists
			if val == "valid" {
				c.Set("apiKey", apiKey)
				c.Next()
				return
			}
			// Cache hit: Invalid API key
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid API key"})
			return
		}

		// If not in Redis, query the database
		var count int
		err = cache.DB.QueryRow(cache.CTX, "SELECT COUNT(*) FROM api_access_token WHERE access_token = $1", apiKey).Scan(&count)

		if err != nil || count == 0 {
			// Store invalid API key result in Redis to prevent repeated DB lookups
			cache.RedisClient.Set(cache.CTX, cacheKey, "invalid", 1*time.Minute)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid API key"})
			return
		}

		// Store valid API key result in Redis with expiration
		cache.RedisClient.Set(cache.CTX, cacheKey, "valid", 1*time.Minute)

		// Store the API key in the context for later use
		c.Set("apiKey", apiKey)
		c.Next()
	}
}

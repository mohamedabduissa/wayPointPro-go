package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	limiterpkg "github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"net/http"
	"time"
)

func RateLimit() gin.HandlerFunc {
	// Create an in-memory store
	store := memory.NewStore()
	rate := limiterpkg.Rate{
		Period: 1 * time.Minute,
		Limit:  120000,
	}

	// Create a rate limiter instance
	limiter := limiterpkg.New(store, rate)

	return func(c *gin.Context) {
		// Get the client's IP address
		API_KEY, exist := c.Get("apiKey")
		if !exist {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not check rate limit"})
			c.Abort()
			return
		}
		// Check the rate limit for the client
		context, err := limiter.Get(c, API_KEY.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not check rate limit"})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))

		// If the client is over the limit, block the request
		if context.Reached {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		// Continue to the next handler
		c.Next()
	}

}

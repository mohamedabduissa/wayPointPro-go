package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

// ExecutionTimeMiddleware logs the time taken to process each request
func ExecutionTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() // Start timer
		c.Next()
		duration := time.Since(start) // Calculate duration

		log.Printf("[%s] %s %s took %v", c.Request.Method, c.Request.URL.Path, c.Request.RemoteAddr, duration)
	}
}

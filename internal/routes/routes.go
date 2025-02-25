package routes

import (
	"WayPointPro/internal/middleware"
	"github.com/gin-gonic/gin"
)

// InitRoutes initializes and combines all route groups
func InitRoutes() *gin.Engine {
	router := gin.New()

	//router := mux.NewRouter()
	// Add essential middlewares
	router.Use(gin.Recovery())                       // Protect from panics
	router.Use(middleware.ExecutionTimeMiddleware()) // Log request execution time
	router.Use(middleware.APIKeyMiddleware())        // API Key validation
	router.Use(middleware.RateLimit())               // Rate limiting

	// Add route groups
	registerAPIRoutes(router)

	return router
}

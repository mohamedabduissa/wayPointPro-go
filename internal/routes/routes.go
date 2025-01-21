package routes

import (
	"WayPointPro/internal/middleware"
	"github.com/gin-gonic/gin"
)

// InitRoutes initializes and combines all route groups
func InitRoutes() *gin.Engine {
	router := gin.Default()

	//router := mux.NewRouter()

	// Apply global middleware
	router.Use(middleware.ExecutionTimeMiddleware())
	router.Use(middleware.APIKeyMiddleware())
	router.Use(middleware.RateLimit())

	// Add route groups
	registerAPIRoutes(router)

	return router
}

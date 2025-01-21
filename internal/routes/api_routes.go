package routes

import (
	"WayPointPro/internal/handlers/map_service"
	"github.com/gin-gonic/gin"
)

func registerAPIRoutes(router *gin.Engine) {
	//apiRouter := router.PathPrefix("/api").Subrouter()

	// Create an API route group
	apiRouter := router.Group("/api")
	{
		apiRouter.POST("/route", map_service.GetRouteHandler)                       // POST /api/route
		apiRouter.GET("/gecode", map_service.GetGeCodingHandler)                    // GET /api/gecode
		apiRouter.GET("/create_access_token", map_service.CreateAccessTokenHandler) // GET /api/gecode
		apiRouter.GET("/list_access_token", map_service.ListAccessTokenHandler)     // GET /api/gecode
		apiRouter.GET("/delete_access_token", map_service.DeleteAccessTokenHandler) // GET /api/gecode
	}

	// Example API routes
	//apiRouter.HandleFunc("/route", map_service.GetRouteHandler).Methods("POST")
	//apiRouter.HandleFunc("/gecode", map_service.GetGeCodingHandler).Methods("GET")
}

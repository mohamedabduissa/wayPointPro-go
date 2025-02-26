package map_service

import (
	"WayPointPro/internal/app/services"
	"WayPointPro/internal/models"
	"WayPointPro/pkg/traffic"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

// GetRouteHandler handles requests for route information
func GetRouteHandler(c *gin.Context) {

	trafficService := traffic.NewService()
	aggregator := services.NewRouteAggregatorService(trafficService)
	// Ensure the request method is POST
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"message": "Method not allowed"})
		return
	}

	// Parse JSON body
	var requestBody struct {
		Coordinates  string `json:"coordinates"`
		Legs         string `json:"legs"`
		Traffic      string `json:"traffic"`
		Alternatives string `json:"alternatives"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON body"})
		return
	}

	// Validate coordinates
	if requestBody.Coordinates == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing 'coordinates' parameter"})
		return
	}

	// Log received data
	//log.Printf("Coordinates: %s, Legs: %s", requestBody.Coordinates, requestBody.Legs)

	//Default 'legs' if not provided
	legs := "false"
	if requestBody.Legs != "" {
		legs = requestBody.Legs
	}

	// Options for the route
	options := map[string]string{
		"overview":     "full",
		"geometries":   "geojson",
		"steps":        legs,
		"traffic":      requestBody.Traffic,
		"alternatives": requestBody.Alternatives,
	}

	// Generate a unique cached_key
	cachedKey := trafficService.Cache.GenerateRouteCacheKey(requestBody.Coordinates)
	log.Printf("cachedKey: %s", cachedKey)
	// Check Redis cache
	cachedData, err := trafficService.Cache.GetFromRedis(cachedKey)
	if err == nil {
		log.Printf("Retreived from cache redis")
		var cachedRoute models.TransformedRoute
		err := json.Unmarshal(cachedData, &cachedRoute)
		if err != nil {
			return
		}
		c.JSON(http.StatusOK, cachedRoute)
		return
	}

	route, err := aggregator.GetAggregatedRoute(requestBody.Coordinates, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch route"})
		return
	}

	startTime := time.Now()
	response := models.TransformRoute(route)
	trafficService.Cache.CacheRouteResponse(cachedKey, response)
	duration := time.Since(startTime)
	log.Printf("response modeling API execution time: %v", duration)
	c.JSON(http.StatusOK, response)

}

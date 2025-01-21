package map_service

import (
	"WayPointPro/internal/app/services"
	"WayPointPro/internal/models"
	"WayPointPro/pkg/osrm"
	"WayPointPro/pkg/traffic"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetRouteHandler handles requests for route information
func GetRouteHandler(c *gin.Context) {

	osrmService := osrm.NewOSRMService()
	trafficService := traffic.NewService()
	aggregator := services.NewRouteAggregatorService(osrmService, trafficService)

	// Ensure the request method is POST
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"message": "Method not allowed"})
		return
	}

	// Parse JSON body
	var requestBody struct {
		Coordinates string `json:"coordinates"`
		Legs        string `json:"legs"`
		Traffic     string `json:"traffic"`
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
		"overview":   "full",
		"geometries": "geojson",
		"steps":      legs,
		"traffic":    requestBody.Traffic,
	}

	route, err := aggregator.GetAggregatedRoute(requestBody.Coordinates, options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch route"})
		return
	}

	response := models.TransformRoute(route)
	c.JSON(http.StatusOK, response)
}

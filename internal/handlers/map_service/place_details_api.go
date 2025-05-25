package map_service

import (
	"WayPointPro/internal/app/services"
	"WayPointPro/internal/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// GetPlaceDetailsHandler handles requests for route information
func GetPlaceDetailsHandler(c *gin.Context) {
	gecoderService := services.NewGecodeService()

	// Delete all rows from the table
	//_, _ = gecoderService.Cache.DB.Exec(gecoderService.Cache.CTX, `DELETE FROM geocoding_results`)

	// Parse query parameters
	query := c.Query("place_id")
	lang := c.Query("lang")
	sessionToken := c.Query("session_token")

	// Validate inputs
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request: Provide either place id please"})
		return
	}

	// Generate a unique cached_key
	var keyQuery string
	keyQuery = query
	cachedKey := gecoderService.Cache.GeneratePlaceIDCacheKey(keyQuery, lang)
	log.Printf("cachedKey: %s", cachedKey)

	// Check Redis cache
	cachedData, err := gecoderService.Cache.GetFromRedis(cachedKey)
	if err == nil {
		log.Printf("Retreived from cache redis")
		var cachedGeocoder []models.GeocodingResult
		err := json.Unmarshal(cachedData, &cachedGeocoder)
		if err != nil {
			return
		}
		responsePlaceDetails(cachedGeocoder, c)
		return
	}

	//Check the database for the cached_key
	cachedGeocoder, err := gecoderService.Cache.GetGecodeData(cachedKey)
	if err == nil && len(cachedGeocoder) > 0 {
		// Cache the database results in Redis
		gecoderService.Cache.CacheGecodeResponse(cachedKey, cachedGeocoder)
		log.Printf("Retreived from cache db")
		responsePlaceDetails(cachedGeocoder, c)
		return
	}
	if err != nil {
		log.Printf("Error fetching gecoder from cache DB")
	}

	geocoding, err := gecoderService.FetchGooglePlaceDetails(query, lang, sessionToken)
	if err != nil {
		return
	}
	// Cache and store the results
	if geocoding == nil {
		geocoding = []models.GeocodingResult{}
	} else {
		gecoderService.Cache.CacheGecodeResponse(cachedKey, geocoding)
		gecoderService.Cache.SaveGecodeData(cachedKey, geocoding)
	}

	responsePlaceDetails(geocoding, c)
}

func responsePlaceDetails(results []models.GeocodingResult, c *gin.Context) {
	// Create a new slice to hold modified results
	modifiedResults := make([]map[string]interface{}, len(results))

	// Iterate over the results and remove the "platform" key
	for i, result := range results {
		// Convert result to a map (assuming JSON serialization compatibility)
		var resultMap map[string]interface{}
		data, _ := json.Marshal(result)      // Serialize the object
		_ = json.Unmarshal(data, &resultMap) // Deserialize into a map

		// Remove the "platform" key if it exists
		delete(resultMap, "platform")
		delete(resultMap, "cached_key")
		delete(resultMap, "bbox_bottom_right_lat")
		delete(resultMap, "bbox_bottom_right_lon")
		delete(resultMap, "bbox_top_left_lat")
		delete(resultMap, "bbox_top_left_lon")
		delete(resultMap, "country")
		delete(resultMap, "address")
		delete(resultMap, "country_code")
		delete(resultMap, "name")
		delete(resultMap, "place_id")

		// Add the modified result to the new slice
		modifiedResults[i] = resultMap
	}

	c.JSON(http.StatusOK, modifiedResults)
}

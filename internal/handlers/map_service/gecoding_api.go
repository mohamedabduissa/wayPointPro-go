package map_service

import (
	"WayPointPro/internal/app/services"
	"WayPointPro/internal/models"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// GetGeCodingHandler handles requests for route information
func GetGeCodingHandler(c *gin.Context) {
	gecoderService := services.NewGecodeService()

	// Delete all rows from the table
	//_, _ = gecoderService.Cache.DB.Exec(gecoderService.Cache.CTX, `DELETE FROM geocoding_results`)

	// Parse query parameters
	query := c.Query("query")
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	country := c.Query("country")
	lang := c.Query("lang")
	limitStr := c.Query("limit")
	var radius int = 0
	var categorySet int = 0
	// Default limit to 10 if not provided
	limit := 5
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if c.Query("radius") != "" {
		if l, err := strconv.Atoi(c.Query("radius")); err == nil {
			radius = l
		}
	}

	if c.Query("categorySet") != "" {
		if l, err := strconv.Atoi(c.Query("categorySet")); err == nil {
			categorySet = l
		}
	}

	// Validate inputs
	if query == "" && (latStr == "" || lngStr == "") {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request: Provide either 'query' or 'lat' and 'lng'"})
		return
	}
	// Parse lat/lng if provided
	var lat, lng float64
	var err error
	if latStr != "" && lngStr != "" {
		lat, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid latitude value"})
			return
		}
		lng, err = strconv.ParseFloat(lngStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid longitude value"})
			return
		}
	}

	// Generate a unique cached_key
	cachedKey := gecoderService.Cache.GenerateGecodeCacheKey(query, lat, lng, country, lang, limit)
	log.Printf("cachedKey: %s", cachedKey)

	log.Printf("customParamters", c.Query("custom"))
	log.Printf("reset", c.Query("reset"))
	if c.Query("reset") != "" {
		err := gecoderService.Cache.RedisClient.FlushDB(gecoderService.Cache.CTX).Err()

		cachedKeys := []string{
			gecoderService.Cache.GenerateGecodeCacheKey("airport", lat, lng, "SA", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("airport", lat, lng, "KW", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("airport", lat, lng, "EG", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("airport", lat, lng, "KW", "ar", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("airport", lat, lng, "SA", "ar", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("airport", lat, lng, "EG", "ar", 10),

			gecoderService.Cache.GenerateGecodeCacheKey("mall", lat, lng, "SA", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("mall", lat, lng, "KW", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("mall", lat, lng, "EG", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("mall", lat, lng, "KW", "ar", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("mall", lat, lng, "SA", "ar", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("mall", lat, lng, "EG", "ar", 10),

			gecoderService.Cache.GenerateGecodeCacheKey("tourist attraction", lat, lng, "SA", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("tourist attraction", lat, lng, "KW", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("tourist attraction", lat, lng, "EG", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("tourist attraction", lat, lng, "KW", "ar", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("tourist attraction", lat, lng, "SA", "ar", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("tourist attraction", lat, lng, "EG", "ar", 10),
		}

		// Build placeholder string: $1, $2, ..., $n
		placeholders := make([]string, len(cachedKeys))
		args := make([]interface{}, len(cachedKeys))

		for i, key := range cachedKeys {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = key
		}

		query := fmt.Sprintf(
			`DELETE FROM geocoding_results WHERE cached_key IN (%s);`,
			strings.Join(placeholders, ", "),
		)

		gecoderService.Cache.DB.Exec(gecoderService.Cache.CTX, query, args...)

	}

	// Check Redis cache
	cachedData, err := gecoderService.Cache.GetFromRedis(cachedKey)
	if err == nil {
		log.Printf("Retreived from cache redis")
		var cachedGeocoder []models.GeocodingResult
		err := json.Unmarshal(cachedData, &cachedGeocoder)
		if err != nil {
			return
		}
		response(cachedGeocoder, c)
		return
	}

	//Check the database for the cached_key
	cachedGeocoder, err := gecoderService.Cache.GetGecodeData(cachedKey)
	if err == nil && len(cachedGeocoder) > 0 {
		// Cache the database results in Redis
		gecoderService.Cache.CacheGecodeResponse(cachedKey, cachedGeocoder)
		log.Printf("Retreived from cache db")
		response(cachedGeocoder, c)
		return
	}
	if err != nil {
		log.Printf("Error fetching gecoder from cache DB")
	}

	geocoding, err := gecoderService.FetchAndParseGeocoding(query, lat, lng, country, lang, limit, radius, categorySet)
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

	response(geocoding, c)
}

func response(results []models.GeocodingResult, c *gin.Context) {
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

		// Add the modified result to the new slice
		modifiedResults[i] = resultMap
	}

	c.JSON(http.StatusOK, modifiedResults)
}

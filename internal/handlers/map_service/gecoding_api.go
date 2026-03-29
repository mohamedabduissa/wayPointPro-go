package map_service

import (
	"WayPointPro/internal/app/services"
	"WayPointPro/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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
	sessionToken := c.Query("session_token")
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

	// Log request parameters
	log.Printf("[GEOCODE] Request received - query: %q, latStr: %q, lngStr: %q, country: %q, lang: %q, limit: %d, radius: %d, categorySet: %d, sessionToken: %q",
		query, latStr, lngStr, country, lang, limit, radius, categorySet, sessionToken)

	// Validate inputs
	if query == "" && (latStr == "" || lngStr == "") {
		log.Printf("[GEOCODE] ERROR - Validation failed: missing query or lat/lng")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request: Provide either 'query' or 'lat' and 'lng'"})
		return
	}
	log.Printf("[GEOCODE] Validation passed - proceeding with geocoding request")
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
	var keyQuery string
	var lowerQuery string
	lowerQuery = strings.ToLower(query)
	//if categorySet != 0 {
	//	keyQuery = strconv.Itoa(categorySet)
	//	if lat != 0 && lng != 0 {
	//		// Append lat & lng to differentiate cache keys for same category but different locations
	//		keyQuery += fmt.Sprintf("_%.4f_%.4f", lat, lng)
	//	}
	//} else {
	//	keyQuery = query
	//}

	// Detect if query is one of the 3 special categories
	var placeType string
	switch {
	case lowerQuery == "airport":
		placeType = "airport"
	case lowerQuery == "mall":
		placeType = "shopping_mall"
	case lowerQuery == "tourist attraction":
		placeType = "tourist_attraction"
	}
	if placeType == "" {
		lat = 0
		lng = 0
	}
	if sessionToken != "" {
		keyQuery = lowerQuery
		keyQuery += "_google"
	}
	cachedKey := gecoderService.Cache.GenerateGecodeCacheKey(keyQuery, lat, lng, country, lang, limit)
	log.Printf("[GEOCODE] Generated cache key: %s for query: %q, lat: %f, lng: %f, country: %q, lang: %q", cachedKey, query, lat, lng, country, lang)

	if c.Query("reset") != "" {
		gecoderService.Cache.RedisClient.FlushDB(gecoderService.Cache.CTX)
		log.Printf("flashed: %s", "redis")
		log.Printf("lat & lng: %s%s", lat, lng)

		cachedKeys := []string{
			gecoderService.Cache.GenerateGecodeCacheKey("airport_google", lat, lng, "SA", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("airport_google", lat, lng, "KW", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("airport_google", lat, lng, "EG", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("airport_google", lat, lng, "KW", "ar", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("airport_google", lat, lng, "SA", "ar", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("airport_google", lat, lng, "EG", "ar", 10),

			gecoderService.Cache.GenerateGecodeCacheKey("shopping_mall_google", lat, lng, "SA", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("shopping_mall_google", lat, lng, "KW", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("shopping_mall_google", lat, lng, "EG", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("shopping_mall_google", lat, lng, "KW", "ar", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("shopping_mall_google", lat, lng, "SA", "ar", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("shopping_mall_google", lat, lng, "EG", "ar", 10),

			gecoderService.Cache.GenerateGecodeCacheKey("tourist_attraction_google", lat, lng, "SA", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("tourist_attraction_google", lat, lng, "KW", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("tourist_attraction_google", lat, lng, "EG", "en", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("tourist_attraction_google", lat, lng, "KW", "ar", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("tourist_attraction_google", lat, lng, "SA", "ar", 10),
			gecoderService.Cache.GenerateGecodeCacheKey("tourist_attraction_google", lat, lng, "EG", "ar", 10),
		}

		// Build placeholder string: $1, $2, ..., $n
		placeholders := make([]string, len(cachedKeys))
		args := make([]interface{}, len(cachedKeys))

		for i, key := range cachedKeys {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = key
		}

		//query := "DELETE FROM geocoding_results;"
		//
		query := fmt.Sprintf(
			`DELETE FROM geocoding_results WHERE cached_key IN (%s);`,
			strings.Join(placeholders, ", "),
		)

		_, _ = gecoderService.Cache.DB.Exec(gecoderService.Cache.CTX, query, args...)

	}

	// Check Redis cache
	cachedData, err := gecoderService.Cache.GetFromRedis(cachedKey)
	if err == nil {
		log.Printf("[GEOCODE] Cache HIT (Redis) - key: %s", cachedKey)
		var cachedGeocoder []models.GeocodingResult
		err := json.Unmarshal(cachedData, &cachedGeocoder)
		if err != nil {
			log.Printf("[GEOCODE] ERROR - Failed to unmarshal Redis cache data: %v", err)
			return
		}
		log.Printf("[GEOCODE] Returning %d results from Redis cache", len(cachedGeocoder))
		response(cachedGeocoder, c)
		return
	}
	log.Printf("[GEOCODE] Cache MISS (Redis) - key: %s, error: %v", cachedKey, err)

	//Check the database for the cached_key
	cachedGeocoder, err := gecoderService.Cache.GetGecodeData(cachedKey)
	if err == nil && len(cachedGeocoder) > 0 {
		// Cache the database results in Redis
		gecoderService.Cache.CacheGecodeResponse(cachedKey, cachedGeocoder)
		log.Printf("[GEOCODE] Cache HIT (DB) - key: %s, results: %d", cachedKey, len(cachedGeocoder))
		response(cachedGeocoder, c)
		return
	}
	if err != nil {
		log.Printf("[GEOCODE] Cache MISS (DB) - key: %s, error: %v", cachedKey, err)
	}

	log.Printf("[GEOCODE] Cache MISS - proceeding to fetch from external API")
	geocoding, err := gecoderService.FetchAndParseGeocoding(query, lat, lng, country, lang, limit, radius, categorySet, sessionToken)
	if err != nil {
		log.Printf("[GEOCODE] ERROR - Failed to fetch geocoding data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Cache and store the results
	if geocoding == nil {
		log.Printf("[GEOCODE] WARNING - No results returned from external API")
		geocoding = []models.GeocodingResult{}
	} else {
		log.Printf("[GEOCODE] Caching %d results - key: %s", len(geocoding), cachedKey)
		gecoderService.Cache.CacheGecodeResponse(cachedKey, geocoding)
		gecoderService.Cache.SaveGecodeData(cachedKey, geocoding)
	}

	log.Printf("[GEOCODE] Returning %d results to client", len(geocoding))
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

		// Check if "place_id" key exists and its value is null
		//if val, ok := resultMap["place_id"]; ok {
		//	if val == nil { // If place_id is null, skip adding this result to modifiedResults
		//		continue // Skip to the next iteration of the loop
		//	}
		//}

		// Add the modified result to the new slice
		modifiedResults[i] = resultMap
	}

	c.JSON(http.StatusOK, modifiedResults)
}

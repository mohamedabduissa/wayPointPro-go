package traffic

import (
	"WayPointPro/internal/config"
	"WayPointPro/internal/db"
	"WayPointPro/internal/models"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"sync"
	"time"
)

type Cache struct {
	DB          *pgxpool.Pool
	RedisClient *redis.Client
	CTX         context.Context
}

var (
	cacheInstance *Cache
	once          sync.Once
)

// NewCache initializes a new Cache instance
func NewCache() *Cache {
	once.Do(func() {
		// Initialize the Cache instance only once
		redisClient := redis.NewClient(&redis.Options{
			Addr: config.LoadConfig().REDIS, // Redis server address
		})
		
		cacheInstance = &Cache{
			DB:          db.Connect(),
			RedisClient: redisClient,
			CTX:         context.Background(),
		}
		log.Println("Cache instance initialized")
	})

	return cacheInstance
}

// SaveTrafficData saves traffic data to the PostgreSQL database
func (c *Cache) SaveTrafficData(trafficData []byte, z, x, y int) error {
	dayOfWeek := time.Now().Weekday().String()
	hour := time.Now().Hour()
	minute := time.Now().Minute() / 15 * 15 // Round to the nearest 15-minute interval

	query := `
		INSERT INTO traffic_data (tile_z, tile_x, tile_y, day_of_week, hour, minute, traffic_data, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, NOW())
		ON CONFLICT (tile_z, tile_x, tile_y, day_of_week, hour, minute)
		DO UPDATE SET
			traffic_data = EXCLUDED.traffic_data,
			updated_at = NOW()
	`
	_, err := c.DB.Exec(c.CTX, query, z, x, y, dayOfWeek, hour, minute, string(trafficData))
	return err
}

// GetTrafficData retrieves traffic data from the PostgreSQL database
func (c *Cache) GetTrafficData(z, x, y, rangeTiles int) ([]byte, error) {
	dayOfWeek := time.Now().Weekday().String()
	hour := time.Now().Hour()
	minute := time.Now().Minute() / 15 * 15 // Round to the nearest 15-minute interval

	query := `
		SELECT traffic_data::text
		FROM traffic_data
		WHERE tile_z = $1 AND tile_x BETWEEN $2 AND $3 AND tile_y BETWEEN $4 AND $5
		AND day_of_week = $6 AND hour = $7 AND minute = $8
	`
	row := c.DB.QueryRow(c.CTX, query, z, x-rangeTiles, x+rangeTiles, y-rangeTiles, y+rangeTiles, dayOfWeek, hour, minute)

	var trafficData string
	err := row.Scan(&trafficData)
	if err == sql.ErrNoRows {
		log.Printf("No traffic data found for tile (%d, %d, %d)", z, x, y)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve traffic data for tile (%d, %d, %d): %w", z, x, y, err)
	}

	return []byte(trafficData), nil
}

func (c *Cache) GetFromRedis(cachedKey string) ([]byte, error) {
	cachedData, err := c.RedisClient.Get(c.CTX, cachedKey).Result()
	if err == nil {
		return []byte(cachedData), nil
	}
	return nil, err
}

// generateCacheKey generates a unique cache key for the request
func (c *Cache) GenerateGecodeCacheKey(query string, lat, lng float64, country, lang string, limit int) string {
	rawKey := ""
	if query != "" {
		rawKey = fmt.Sprintf("geocode:%s:%s:%s:%d", query, country, lang, limit)
	} else {
		rawKey = fmt.Sprintf("geocode:latlng:%f:%f:%s:%s:%d", lat, lng, country, lang, limit)
	}
	// Optional: Use hashing for consistent length and encoding safety
	hasher := sha256.New()
	hasher.Write([]byte(rawKey))
	cachedKey := hex.EncodeToString(hasher.Sum(nil))
	return cachedKey
}

// generateCacheKey generates a unique cache key for the request
func (c *Cache) GenerateRouteCacheKey(coordinates string) string {
	rawKey := ""
	rawKey = fmt.Sprintf("route:%s", coordinates)
	// Optional: Use hashing for consistent length and encoding safety
	hasher := sha256.New()
	hasher.Write([]byte(rawKey))
	cachedKey := hex.EncodeToString(hasher.Sum(nil))
	return cachedKey
}

// cacheResponse caches the response in Redis
func (c *Cache) CacheGecodeResponse(cachedKey string, results []models.GeocodingResult) {
	data, err := json.Marshal(results)
	if err != nil {
		log.Printf("Failed to marshal geocoding results: %v", err)
		return
	}

	err = c.RedisClient.Set(c.CTX, cachedKey, data, 3*time.Hour).Err()
	if err != nil {
		log.Printf("Failed to cache geocoding result in Redis: %v", err)
	}

}

// cacheResponse caches the response in Redis
func (c *Cache) CacheRouteResponse(cachedKey string, results models.TransformedRoute) {
	data, _ := json.Marshal(results)
	c.RedisClient.Set(c.CTX, cachedKey, data, 3*time.Hour)
}

// storeInDatabase stores the geocoding results in the database
func (c *Cache) SaveGecodeData(cachedKey string, results []models.GeocodingResult) {
	for _, result := range results {
		_, err := c.DB.Exec(c.CTX, `
			INSERT INTO geocoding_results 
			(platform, name, address, latitude, longitude, country, country_code, bbox_top_left_lat, bbox_top_left_lon, bbox_bottom_right_lat, bbox_bottom_right_lon, cached_key)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, result.Platform, result.Name, result.Address, result.Latitude, result.Longitude,
			result.Country, result.CountryCode, result.BoundingBoxTopLeftLat, result.BoundingBoxTopLeftLon,
			result.BoundingBoxBottomRightLat, result.BoundingBoxBottomRightLon, cachedKey)
		if err != nil {
			log.Printf("Failed to store geocoding result in database: %v", err)
		}
	}
}

// fetchFromDatabaseByCacheKey retrieves results from the database by cached_key
func (c *Cache) GetGecodeData(cachedKey string) ([]models.GeocodingResult, error) {
	//var results []models.GeocodingResult
	//err := c.DB.Select(&results, `
	//	SELECT platform, name, address, latitude, longitude, country, country_code, bbox_top_left_lat, bbox_top_left_lon, bbox_bottom_right_lat, bbox_bottom_right_lon
	//	FROM geocoding_results
	//	WHERE cached_key = $1
	//`, cachedKey)
	//return results, err

	// Query the database
	rows, err := c.DB.Query(c.CTX, `
		SELECT platform, name, address, latitude, longitude, country, country_code, 
		       bbox_top_left_lat, bbox_top_left_lon, bbox_bottom_right_lat, bbox_bottom_right_lon
		FROM geocoding_results
		WHERE cached_key = $1
	`, cachedKey)
	if err != nil {
		log.Printf("❌ Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	// Prepare results slice
	var results []models.GeocodingResult

	// Loop through rows and scan into struct
	for rows.Next() {
		var result models.GeocodingResult
		err := rows.Scan(
			&result.Platform, &result.Name, &result.Address, &result.Latitude, &result.Longitude,
			&result.Country, &result.CountryCode,
			&result.BoundingBoxTopLeftLat, &result.BoundingBoxTopLeftLon,
			&result.BoundingBoxBottomRightLat, &result.BoundingBoxBottomRightLon,
		)
		if err != nil {
			log.Printf("❌ Error scanning row: %v", err)
			return nil, err
		}
		results = append(results, result)
	}

	// Check for errors from iteration
	if err := rows.Err(); err != nil {
		log.Printf("❌ Row iteration error: %v", err)
		return nil, err
	}

	return results, nil
}

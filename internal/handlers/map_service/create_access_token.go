package map_service

import (
	"WayPointPro/pkg/traffic"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

// Define the struct for access_tokens
type AccessToken struct {
	ID           int    `json:"id"`
	Platform     string `json:"platform"`
	AccessToken  string `json:"access_token"`
	RequestLimit int    `json:"request_limit"`
	RequestCount int    `json:"request_count"`
	ResetDate    string `json:"reset_date"` // Use time.Time for a parsed time representation
}

// CreateAccessTokenHandler handles requests for route information
func CreateAccessTokenHandler(c *gin.Context) {

	// Parse query parameters
	accessToken := c.Query("access_token")
	platform := c.Query("platform")
	request_limit := c.Query("request_limit")

	// Validate inputs
	if accessToken == "" && (platform == "" || request_limit == "") {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request: Provide either 'query' or 'lat' and 'lng'"})
		return
	}
	cache := traffic.NewCache()

	_, err := cache.DB.Exec(cache.CTX, `
			INSERT INTO access_tokens (platform, access_token, request_limit, reset_date)
			VALUES ($1, $2, $3, $4)
		`, platform, accessToken, request_limit, time.Now())
	if err != nil {
		log.Printf("Failed to store geocoding result in database: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully stored access token."})
}

// ListAccessTokenHandler handles requests for route information
func ListAccessTokenHandler(c *gin.Context) {

	cache := traffic.NewCache()

	rows, err := cache.DB.Query(cache.CTX, `
			SELECT * FROM access_tokens 
		`)
	if err != nil {
		log.Printf("Failed to store geocoding result in database: %v", err)
	}
	defer rows.Close()

	// Slice to hold the results
	var accessTokens []AccessToken

	// Loop through rows
	for rows.Next() {
		var token AccessToken
		err := rows.Scan(&token.ID, &token.Platform, &token.AccessToken, &token.RequestLimit, &token.RequestCount, &token.ResetDate)
		if err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		accessTokens = append(accessTokens, token)
	}

	// Check for any error encountered during iteration
	if err = rows.Err(); err != nil {
		log.Fatalf("Error during row iteration: %v", err)
	}

	c.JSON(http.StatusOK, accessTokens)
}

// DeleteAccessTokenHandler handles requests for route information
func DeleteAccessTokenHandler(c *gin.Context) {

	// Parse query parameters
	accessToken := c.Query("access_token")

	// Validate inputs
	if accessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request: Provide either access_token"})
		return
	}
	cache := traffic.NewCache()

	_, err := cache.DB.Exec(cache.CTX, `
			DELETE FROM access_tokens where access_token = $1;
		`, accessToken)
	if err != nil {
		log.Printf("Failed to store geocoding result in database: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted access token."})
}

package traffic

import (
	"fmt"
	"net/http"
	"time"
)

// Service contains the HTTP client and cache
type Service struct {
	HTTPClient  *http.Client
	Cache       *Cache
	accessToken string
}

// NewService initializes a Service instance
func NewService() *Service {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &Service{
		HTTPClient: httpClient,
		Cache:      NewCache(),
	}
}

// Choose platform and access token dynamically
func (s *Service) choosePlatformAndToken(requiredRequests int) (string, string, error) {
	query := `
		SELECT platform, access_token, request_limit, request_count
		FROM access_tokens
		WHERE request_limit - request_count >= $1
		ORDER BY RANDOM()
		LIMIT 1
	`

	row := s.Cache.DB.QueryRow(query, requiredRequests)

	var platform, accessToken string
	var requestLimit, requestCount int

	if err := row.Scan(&platform, &accessToken, &requestLimit, &requestCount); err != nil {
		return "", "", fmt.Errorf("no available access tokens or failed to fetch: %v", err)
	}

	return platform, accessToken, nil
}

// Increment the request count for an access token
func (s *Service) updateAccessTokenRequestCount(accessToken string, count int) error {
	query := `
		UPDATE access_tokens
		SET request_count = request_count + $2
		WHERE access_token = $1
	`
	_, err := s.Cache.DB.Exec(query, accessToken, count)
	return err
}

// Log request details in the database
func (s *Service) logRequestToDB(platform, accessToken string, zoom, x, y int) error {
	query := `
		INSERT INTO request_logs (platform, access_token, zoom, x, y, timestamp)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err := s.Cache.DB.Exec(query, platform, accessToken, zoom, x, y)
	return err
}

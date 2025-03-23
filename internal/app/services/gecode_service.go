package services

import (
	"WayPointPro/internal/models"
	"WayPointPro/pkg/traffic"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type GecodeService struct {
	HTTPClient *http.Client
	Cache      *traffic.Cache
}

// NewService initializes a Service instance
func NewGecodeService() *GecodeService {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &GecodeService{
		HTTPClient: httpClient,
		Cache:      traffic.NewCache(),
	}
}

// Choose platform and access token dynamically
func (s *GecodeService) choosePlatformAndToken(requiredRequests int) (string, string, error) {
	query := `
		SELECT platform, access_token, request_limit, request_count
		FROM access_tokens
		WHERE request_limit - request_count >= $1
		ORDER BY RANDOM()
		LIMIT 1
	`

	row := s.Cache.DB.QueryRow(s.Cache.CTX, query, requiredRequests)

	var platform, accessToken string
	var requestLimit, requestCount int

	if err := row.Scan(&platform, &accessToken, &requestLimit, &requestCount); err != nil {
		return "", "", fmt.Errorf("no available access tokens or failed to fetch: %v", err)
	}

	return platform, accessToken, nil
}

// Increment the request count for an access token
func (s *GecodeService) updateAccessTokenRequestCount(accessToken string, count int) error {
	query := `
		UPDATE access_tokens
		SET request_count = request_count + $2
		WHERE access_token = $1
	`
	_, err := s.Cache.DB.Exec(s.Cache.CTX, query, accessToken, count)
	return err
}

// Log request details in the database
func (s *GecodeService) logRequestToDB(platform, accessToken string, zoom, x, y int) error {
	query := `
		INSERT INTO request_logs (platform, access_token, zoom, x, y, timestamp)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err := s.Cache.DB.Exec(s.Cache.CTX, query, platform, accessToken, zoom, x, y)
	return err
}

// Main logic for fetching and parsing geocoding data
func (s *GecodeService) FetchAndParseGeocoding(query string, lat, lng float64, country, lang string, limit int) ([]models.GeocodingResult, error) {
	platform, token, err := s.choosePlatformAndToken(1)
	if err != nil {
		return nil, err
	}

	url := s.BuildGeocodingURL(platform, query, lat, lng, country, lang, limit, token)
	s.updateAccessTokenRequestCount(token, 1)
	log.Printf("Fetching geocoding url: %s", url)
	body, err := s.FetchGeocodingData(url)
	if err != nil {
		return nil, err
	}

	switch platform {
	case "mapbox":
		return models.ParseMapboxResponse(body)
	case "tomtom":
		return models.ParseTomTomResponse(body)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

// Build URL dynamically
func (s *GecodeService) BuildGeocodingURL(platform, query string, lat, lng float64, country, lang string, limit int, token string) string {
	encodedQuery := url.QueryEscape(query)

	switch platform {
	case "mapbox":
		queryStrings := ""
		queryStrings = "?access_token=" + token
		if country != "" {
			queryStrings += "&country=" + country
		}
		if lang != "" {
			queryStrings += "&language=" + lang
		}
		//queryStrings += "&limit=" + strconv.Itoa(limit)

		if query != "" {
			queryStrings += "&q=" + encodedQuery
			return fmt.Sprintf("https://api.mapbox.com/search/geocode/v6/forward%s",
				queryStrings)
		}
		return fmt.Sprintf("https://api.mapbox.com/search/geocode/v6/reverse%s&longitude=%f&latitude=%f",
			queryStrings, lng, lat)
	case "tomtom":
		queryStrings := ""
		queryStrings = "?key=" + token
		if country != "" {
			queryStrings += "&countrySet=" + country
		}
		if lang != "" && lang != "en" {
			queryStrings += "&language=" + lang
		}
		queryStrings += "&limit=" + strconv.Itoa(limit)

		if query != "" {
			return fmt.Sprintf("https://api.tomtom.com/search/2/geocode/%s.json%s",
				encodedQuery, queryStrings)
		}
		return fmt.Sprintf("https://api.tomtom.com/search/2/reverseGeocode/%f,%f.json%s",
			lat, lng, queryStrings)
	default:
		return ""
	}
}

// Fetch data from the API
func (s *GecodeService) FetchGeocodingData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch geocoding data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	return body, nil
}

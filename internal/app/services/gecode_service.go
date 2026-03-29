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
	"strings"
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
func (s *GecodeService) choosePlatformAndToken(requiredRequests int, requiredPlatform string) (string, string, error) {
	log.Printf("[GEOCODE] Choosing platform/token - requiredRequests: %d, requiredPlatform: %q", requiredRequests, requiredPlatform)

	query := `
		SELECT platform, access_token, request_limit, request_count
		FROM access_tokens
		WHERE platform = $2
		AND request_limit - request_count >= $1
		ORDER BY RANDOM()
		LIMIT 1
	`

	row := s.Cache.DB.QueryRow(s.Cache.CTX, query, requiredRequests, requiredPlatform)

	var platform, accessToken string
	var requestLimit, requestCount int

	if err := row.Scan(&platform, &accessToken, &requestLimit, &requestCount); err != nil {
		log.Printf("[GEOCODE] ERROR - Failed to select platform/token: %v", err)
		return "", "", fmt.Errorf("no available access tokens or failed to fetch: %v", err)
	}

	// Log token with last 4 chars only for security
	tokenPreview := "***"
	if len(accessToken) >= 4 {
		tokenPreview = accessToken[len(accessToken)-4:]
	}
	log.Printf("[GEOCODE] Selected platform: %q, token: ***%s, requestLimit: %d, requestCount: %d", platform, tokenPreview, requestLimit, requestCount)

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
func (s *GecodeService) FetchAndParseGeocoding(query string, lat, lng float64, country, lang string, limit int, radius int, categorySet int, sessionToken string) ([]models.GeocodingResult, error) {
	log.Printf("[GEOCODE] FetchAndParseGeocoding - query: %q, lat: %f, lng: %f, country: %q, lang: %q, limit: %d, radius: %d, categorySet: %d, sessionToken: %q",
		query, lat, lng, country, lang, limit, radius, categorySet, sessionToken)

	requiredPlatform := "tomtom"
	if sessionToken != "" {
		requiredPlatform = "google"
	}
	log.Printf("[GEOCODE] Determined required platform: %q (sessionToken provided: %v)", requiredPlatform, sessionToken != "")

	platform, token, err := s.choosePlatformAndToken(1, requiredPlatform)
	if err != nil {
		log.Printf("[GEOCODE] ERROR - Platform/token selection failed: %v", err)
		return nil, err
	}

	url := s.BuildGeocodingURL(platform, query, lat, lng, country, lang, limit, radius, categorySet, token, sessionToken)
	log.Printf("[GEOCODE] Built URL for platform %q: %s", platform, url)

	s.updateAccessTokenRequestCount(token, 1)
	log.Printf("[GEOCODE] Updated request count for token")

	body, err := s.FetchGeocodingData(url)
	if err != nil {
		log.Printf("[GEOCODE] ERROR - Failed to fetch geocoding data: %v", err)
		return nil, err
	}

	log.Printf("[GEOCODE] Parsing response for platform: %q, body length: %d bytes", platform, len(body))

	var results []models.GeocodingResult
	switch platform {
	case "mapbox":
		results, err = models.ParseMapboxResponse(body)
	case "tomtom":
		results, err = models.ParseTomTomResponse(body)
	case "google":
		results, err = models.ParseGoogleAutocompleteResponse(body)
	default:
		log.Printf("[GEOCODE] ERROR - Unsupported platform: %s", platform)
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}

	if err != nil {
		log.Printf("[GEOCODE] ERROR - Failed to parse %s response: %v", platform, err)
		// Log response body preview for debugging
		previewLen := 500
		if len(body) < previewLen {
			previewLen = len(body)
		}
		log.Printf("[GEOCODE] Response body preview: %s", string(body[:previewLen]))
		return nil, err
	}

	log.Printf("[GEOCODE] Successfully parsed %d results from %s platform", len(results), platform)
	return results, nil
}

// Build URL dynamically
func (s *GecodeService) BuildGeocodingURL(platform, query string, lat, lng float64, country, lang string, limit, radius, categorySet int, token string, sessionToken string) string {
	log.Printf("[GEOCODE] BuildGeocodingURL - platform: %q, query: %q, lat: %f, lng: %f, country: %q, lang: %q, limit: %d, radius: %d, categorySet: %d",
		platform, query, lat, lng, country, lang, limit, radius, categorySet)

	encodedQuery := url.QueryEscape(query)

	var builtURL string

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
			builtURL = fmt.Sprintf("https://api.mapbox.com/search/geocode/v6/forward%s", queryStrings)
		} else {
			builtURL = fmt.Sprintf("https://api.mapbox.com/search/geocode/v6/reverse%s&longitude=%f&latitude=%f", queryStrings, lng, lat)
		}

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

		if radius != 0 {
			queryStrings += "&radius=" + strconv.Itoa(radius)
		}
		if categorySet != 0 {
			queryStrings += "&categorySet=" + strconv.Itoa(categorySet)
			encodedQuery = ""
		}
		if lat != 0 {
			queryStrings += "&lat=" + fmt.Sprintf("%.6f", lat)
		}
		if lng != 0 {
			queryStrings += "&lon=" + fmt.Sprintf("%.6f", lng)
		}
		if query != "" {
			builtURL = fmt.Sprintf("https://api.tomtom.com/search/2/search/%s.json%s", encodedQuery, queryStrings)
		} else {
			builtURL = fmt.Sprintf("https://api.tomtom.com/search/2/reverseGeocode/%f,%f.json%s", lat, lng, queryStrings)
		}

	case "google":
		// Detect if query is one of the 3 special categories
		lowerQuery := strings.ToLower(query)
		var placeType string

		switch {
		case lowerQuery == "airport":
			placeType = "airport"
		case lowerQuery == "mall":
			placeType = "shopping_mall"
		case lowerQuery == "tourist attraction":
			placeType = "tourist_attraction"
		}

		if placeType != "" {
			// Use Google Text Search with type filter
			base := "https://maps.googleapis.com/maps/api/place/textsearch/json"
			params := url.Values{}
			//params.Set("query", query)
			params.Set("type", placeType)
			params.Set("key", token)
			params.Set("sessiontoken", sessionToken)
			params.Set("language", lang)
			if lat != 0 && lng != 0 {
				params.Set("location", fmt.Sprintf("%.6f,%.6f", lat, lng))
			}
			if radius != 0 {
				params.Set("radius", strconv.Itoa(radius))
			}
			builtURL = fmt.Sprintf("%s?%s", base, params.Encode())
			log.Printf("[GEOCODE] Google Text Search URL (placeType: %q): %s", placeType, builtURL)
		} else {
			// Default: use Autocomplete
			base := "https://maps.googleapis.com/maps/api/place/autocomplete/json"
			params := url.Values{}
			params.Set("input", query)
			params.Set("key", token)
			params.Set("sessiontoken", sessionToken)
			params.Set("language", lang)
			if country != "" {
				params.Set("components", "country:"+country)
			}
			builtURL = fmt.Sprintf("%s?%s", base, params.Encode())
			log.Printf("[GEOCODE] Google Autocomplete URL: %s", builtURL)
		}

		// Google-specific: session token (optional: manage externally)
		//sessionToken := uuid.New().String()
		//params.Set("sessiontoken", sessionToken)
	default:
		log.Printf("[GEOCODE] ERROR - Unknown platform: %q", platform)
		return ""
	}

	log.Printf("[GEOCODE] Final built URL for %q: %s", platform, builtURL)
	return builtURL
}

// Fetch and parse Google Place Details by Place ID only
func (s *GecodeService) FetchGooglePlaceDetails(placeID string, lang string, sessionToken string) ([]models.GeocodingResult, error) {
	_, token, err := s.choosePlatformAndToken(1, "google")
	if err != nil {
		return nil, err
	}

	if placeID == "" {
		return nil, fmt.Errorf("place_id is required")
	}

	base := "https://maps.googleapis.com/maps/api/place/details/json"
	params := url.Values{}
	params.Set("place_id", placeID)
	params.Set("key", token)
	params.Set("sessiontoken", sessionToken)
	params.Set("fields", "geometry")
	if lang != "" {
		params.Set("language", lang)
	}

	url := fmt.Sprintf("%s?%s", base, params.Encode())
	log.Printf("Fetching place details from: %s", url)

	body, err := s.FetchGeocodingData(url)
	if err != nil {
		return nil, err
	}

	return models.ParseGooglePlaceDetailsResponse(body)
}

// Fetch data from the API
func (s *GecodeService) FetchGeocodingData(url string) ([]byte, error) {
	log.Printf("[GEOCODE] Fetching data from URL: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[GEOCODE] ERROR - HTTP request failed: %v", err)
		return nil, fmt.Errorf("failed to fetch geocoding data: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("[GEOCODE] API Response - Status: %d (%s)", resp.StatusCode, resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[GEOCODE] ERROR - Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	log.Printf("[GEOCODE] Response body length: %d bytes", len(body))

	if resp.StatusCode != http.StatusOK {
		// Log response body preview for non-200 responses
		previewLen := 500
		if len(body) < previewLen {
			previewLen = len(body)
		}
		log.Printf("[GEOCODE] ERROR - Non-200 response body preview: %s", string(body[:previewLen]))
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	return body, nil
}

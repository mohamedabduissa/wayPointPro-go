package traffic

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
)

// FetchAndAnalyzeTraffic fetches traffic data and analyzes it for a bounding box
func (s *Service) FetchAndAnalyzeTraffic(boundingBox map[string]float64, zoom int) ([]map[string]interface{}, error) {
	var trafficData []map[string]interface{}
	//tileRange := s.getTileRange(boundingBox, zoom)
	tileRange := s.FullGetTileRange(boundingBox, zoom)
	batches := s.batchTileRange(tileRange, 5000)

	var wg sync.WaitGroup
	var mu sync.Mutex
	requestedTiles := make(map[string]bool)

	requiredRequests := 0
	calledRequests := 0

	for _, batch := range batches {
		xLen := len(batch["x"].([]int))
		yLen := len(batch["y"].([]int))

		// Calculate the number of iterations for this batch
		requiredRequests += xLen * yLen
		// Choose platform and token dynamically
		platform, accessTokenDB, err := s.choosePlatformAndToken(requiredRequests)
		s.accessToken = accessTokenDB
		if err != nil {
			log.Printf("Failed to choose platform and token: %v", err)
			return nil, err
		}
		for _, x := range batch["x"].([]int) {
			for _, y := range batch["y"].([]int) {
				// Add a 2-second delay before processing the request
				//time.Sleep(500 * time.Millisecond)
				wg.Add(1)
				go func(x, y int) {
					defer wg.Done()

					tileKey := fmt.Sprintf("%d_%d_%d", zoom, x, y)

					// Ensure tile is only requested once
					if !s.markTileRequested(&mu, requestedTiles, tileKey) {
						return
					}

					// Fetch and process data for the tile
					features, isRequest := s.fetchAndProcessTileData(platform, s.accessToken, zoom, x, y)
					if features == nil {
						return
					}
					if isRequest {
						calledRequests += 1
					}

					// Append features to the result
					mu.Lock()
					trafficData = append(trafficData, features...)
					mu.Unlock()
				}(x, y)
			}
		}
	}
	wg.Wait()
	// Increment request count for the access token
	if err := s.updateAccessTokenRequestCount(s.accessToken, calledRequests); err != nil {
		log.Printf("Failed to update request count for access token: %v", err)
	}
	return trafficData, nil
}

func (s *Service) markTileRequested(mu *sync.Mutex, requestedTiles map[string]bool, tileKey string) bool {
	mu.Lock()
	defer mu.Unlock()

	if requestedTiles[tileKey] {
		return false
	}
	requestedTiles[tileKey] = true
	return true
}
func (s *Service) fetchAndProcessTileData(platform, accessToken string, zoom, x, y int) ([]map[string]interface{}, bool) {
	// Check cache first
	cachedData, _ := s.Cache.GetTrafficData(zoom, x, y, 0)
	if cachedData != nil {
		log.Printf("Cache hit for tile (%d, %d, %d)", zoom, x, y)
		return s.parseTrafficData(cachedData), false
	}

	// Fetch data from the server
	url := fmt.Sprintf("http://localhost:6000/decode-tile?z=%d&x=%d&y=%d&accessToken=%s&platform=%s", zoom, x, y, accessToken, platform)
	resp, err := s.HTTPClient.Get(url)
	if err != nil {
		log.Printf("Failed to fetch tile (%d, %d, %d): %v", zoom, x, y, err)
		return nil, true
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body for tile (%d, %d, %d): %v", zoom, x, y, err)
		return nil, true
	}

	// Cache the data
	if err := s.Cache.SaveTrafficData(body, zoom, x, y); err != nil {
		log.Printf("Failed to save traffic data to cache for tile (%d, %d, %d): %v", zoom, x, y, err)
	}

	// Parse and return traffic data
	return s.parseTrafficData(body), true
}
func (s *Service) parseTrafficData(data []byte) []map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("Failed to parse traffic data: %v", err)
		return nil
	}

	features, ok := result["features"].([]interface{})
	if !ok || len(features) == 0 {
		log.Println("No features available in the response")
		return nil
	}

	var parsedFeatures []map[string]interface{}
	for _, feature := range features {
		featureMap, ok := feature.(map[string]interface{})
		if !ok {
			log.Println("Invalid feature, skipping")
			continue
		}
		parsedFeatures = append(parsedFeatures, featureMap)
	}
	return parsedFeatures
}

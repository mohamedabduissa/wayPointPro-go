package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Coordinates represent latitude and longitude
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// BBox represents a bounding box with top-left and bottom-right points
type BBox struct {
	TopLeft     Coordinates `json:"top_left"`
	BottomRight Coordinates `json:"bottom_right"`
}

// GeocodingResult is the common structure for geocoding results
type GeocodingResult struct {
	Platform                  string  `json:"platform" db:"platform"`
	Name                      string  `json:"name" db:"name"`
	Address                   string  `json:"address" db:"address"`
	Latitude                  float64 `json:"latitude" db:"latitude"`
	Longitude                 float64 `json:"longitude" db:"longitude"`
	Country                   string  `json:"country" db:"country"`
	CountryCode               string  `json:"country_code" db:"country_code"`
	BoundingBoxTopLeftLat     float64 `json:"bbox_top_left_lat,omitempty" db:"bbox_top_left_lat"`
	BoundingBoxTopLeftLon     float64 `json:"bbox_top_left_lon,omitempty" db:"bbox_top_left_lon"`
	BoundingBoxBottomRightLat float64 `json:"bbox_bottom_right_lat,omitempty" db:"bbox_bottom_right_lat"`
	BoundingBoxBottomRightLon float64 `json:"bbox_bottom_right_lon,omitempty" db:"bbox_bottom_right_lon"`
	CachedKey                 string  `json:"cached_key" db:"cached_key"`
	PlaceID                   string  `json:"place_id" db:"place_id"`
}

// Parse Mapbox response
func ParseMapboxResponse(body []byte) ([]GeocodingResult, error) {
	var response struct {
		Type     string `json:"type"`
		Features []struct {
			Type     string `json:"type"`
			ID       string `json:"id"`
			Geometry struct {
				Type        string     `json:"type"`
				Coordinates [2]float64 `json:"coordinates"`
			} `json:"geometry"`
			Properties struct {
				MapboxID      string `json:"mapbox_id"`
				FeatureType   string `json:"feature_type"`
				FullAddress   string `json:"full_address"`
				Name          string `json:"name"`
				NamePreferred string `json:"name_preferred"`
				Coordinates   struct {
					Longitude float64 `json:"longitude"`
					Latitude  float64 `json:"latitude"`
				} `json:"coordinates"`
				BBox    [4]float64 `json:"bbox"`
				Context struct {
					Country struct {
						MapboxID         string `json:"mapbox_id"`
						Name             string `json:"name"`
						CountryCode      string `json:"country_code"`
						CountryCodeAlpha string `json:"country_code_alpha_3"`
						WikidataID       string `json:"wikidata_id"`
					} `json:"country"`
				} `json:"context"`
			} `json:"properties"`
		} `json:"features"`
	}

	err := json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Mapbox response: %v", err)
	}

	var results []GeocodingResult
	for _, feature := range response.Features {
		results = append(results, GeocodingResult{
			Platform:                  "mapbox",
			Name:                      feature.Properties.Name,
			Address:                   feature.Properties.FullAddress,
			Latitude:                  feature.Geometry.Coordinates[1],
			Longitude:                 feature.Geometry.Coordinates[0],
			Country:                   feature.Properties.Context.Country.Name,
			CountryCode:               feature.Properties.Context.Country.CountryCode,
			BoundingBoxTopLeftLat:     feature.Geometry.Coordinates[1],
			BoundingBoxTopLeftLon:     feature.Geometry.Coordinates[0],
			BoundingBoxBottomRightLat: feature.Geometry.Coordinates[1],
			BoundingBoxBottomRightLon: feature.Geometry.Coordinates[0],
		})
	}
	return results, nil
}

// Parse TomTom response
func ParseTomTomResponse(body []byte) ([]GeocodingResult, error) {
	var response struct {
		Results []struct {
			Poi struct {
				Name string `json:"name"`
			} `json:"poi"`
			Address struct {
				Country         string `json:"country"`
				CountryCode     string `json:"countryCode"`
				FreeformAddress string `json:"freeformAddress"`
			} `json:"address"`
			Position struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"position"`
			BoundingBox struct {
				TopLeftPoint struct {
					Lat float64 `json:"lat"`
					Lon float64 `json:"lon"`
				} `json:"topLeftPoint"`
				BottomRightPoint struct {
					Lat float64 `json:"lat"`
					Lon float64 `json:"lon"`
				} `json:"btmRightPoint"`
			} `json:"boundingBox"`
		} `json:"results"`
		Addresses []struct {
			Address struct {
				Country         string `json:"country"`
				CountryCode     string `json:"countryCode"`
				FreeformAddress string `json:"freeformAddress"`
			} `json:"address"`
			Position string `json:"position"`
		} `json:"addresses"`
	}

	err := json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TomTom response: %v", err)
	}

	var results []GeocodingResult
	for _, result := range response.Results {
		var addressName = result.Address.FreeformAddress
		if result.Poi.Name != "" {
			addressName = result.Poi.Name
		}
		results = append(results, GeocodingResult{
			Platform:                  "tomtom",
			Name:                      addressName, // TomTom often omits names
			Address:                   result.Address.FreeformAddress,
			Latitude:                  result.Position.Lat,
			Longitude:                 result.Position.Lon,
			Country:                   result.Address.Country,
			CountryCode:               result.Address.CountryCode,
			BoundingBoxTopLeftLat:     result.Position.Lat,
			BoundingBoxTopLeftLon:     result.Position.Lon,
			BoundingBoxBottomRightLat: result.Position.Lat,
			BoundingBoxBottomRightLon: result.Position.Lon,
		})
	}
	for _, result := range response.Addresses {
		// Split the string into latitude and longitude
		coords := strings.Split(result.Position, ",")
		if len(coords) != 2 {
			fmt.Println("Invalid position format")
		}

		// Convert latitude and longitude to float64
		lat, _ := strconv.ParseFloat(coords[0], 64)
		lng, _ := strconv.ParseFloat(coords[1], 64)

		results = append(results, GeocodingResult{
			Platform:    "tomtom",
			Name:        result.Address.FreeformAddress, // TomTom often omits names
			Address:     result.Address.FreeformAddress,
			Latitude:    lat,
			Longitude:   lng,
			Country:     result.Address.Country,
			CountryCode: result.Address.CountryCode,
		})
	}
	return results, nil
}

// ParseGoogleAutocompleteResponse parses either Google Autocomplete or Text Search response
func ParseGoogleAutocompleteResponse(body []byte) ([]GeocodingResult, error) {
	var autoResp struct {
		Predictions []struct {
			Description          string `json:"description"`
			PlaceID              string `json:"place_id"`
			StructuredFormatting struct {
				MainText      string `json:"main_text"`
				SecondaryText string `json:"secondary_text"`
			} `json:"structured_formatting"`
			Terms []struct {
				Value string `json:"value"`
			} `json:"terms"`
		} `json:"predictions"`
	}

	var textResp struct {
		Results []struct {
			Name             string `json:"name"`
			FormattedAddress string `json:"formatted_address"`
			PlaceID          string `json:"place_id"`
			Geometry         struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
		} `json:"results"`
	}

	var results []GeocodingResult
	var autoErr, textErr error

	// Try unmarshaling autocomplete
	autoErr = json.Unmarshal(body, &autoResp)
	if autoErr == nil && len(autoResp.Predictions) > 0 {
		for _, item := range autoResp.Predictions {
			country := ""
			if len(item.Terms) > 0 {
				country = item.Terms[len(item.Terms)-1].Value
			}
			results = append(results, GeocodingResult{
				Platform:    "google",
				Name:        item.StructuredFormatting.MainText,
				Address:     item.Description,
				Country:     country,
				CountryCode: "",
				Latitude:    0, // No lat/lng in autocomplete
				Longitude:   0,
				PlaceID:     item.PlaceID,
			})
		}
		return results, nil
	}

	// Try unmarshaling text search
	textErr = json.Unmarshal(body, &textResp)
	if textErr == nil && len(textResp.Results) > 0 {
		for _, item := range textResp.Results {
			//country := ""
			//parts := strings.Split(item.FormattedAddress, ",")
			//if len(parts) > 0 {
			//	country = strings.TrimSpace(parts[len(parts)-1])
			//}
			results = append(results, GeocodingResult{
				Platform:    "google",
				Name:        item.Name,
				Address:     item.FormattedAddress,
				Latitude:    item.Geometry.Location.Lat,
				Longitude:   item.Geometry.Location.Lng,
				Country:     "",
				CountryCode: "",
				PlaceID:     item.PlaceID,
			})
		}
		return results, nil
	}

	// Both failed
	return nil, fmt.Errorf("failed to parse response: autoErr=%v, textErr=%v", autoErr, textErr)
}

// ParseGooglePlaceDetailsResponse parses either Google Autocomplete or Text Search response
func ParseGooglePlaceDetailsResponse(body []byte) ([]GeocodingResult, error) {

	var response struct {
		Result struct {
			Geometry struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
		} `json:"result"`
	}

	var results []GeocodingResult
	var autoErr, textErr error

	textErr = json.Unmarshal(body, &response)
	if response.Result.Geometry.Location.Lng != 0 || response.Result.Geometry.Location.Lat != 0 {
		results = append(results, GeocodingResult{
			Platform:  "google",
			Latitude:  response.Result.Geometry.Location.Lat,
			Longitude: response.Result.Geometry.Location.Lng,
		})
		return results, nil
	}

	// Both failed
	return nil, fmt.Errorf("failed to parse response: autoErr=%v, textErr=%v", autoErr, textErr)
}

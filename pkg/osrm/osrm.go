package osrm

import (
	"WayPointPro/internal/config"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type OSRMService struct {
	BaseURL string
}

type RouteResponse struct {
	Code      string     `json:"code"`
	Routes    []Route    `json:"routes"`
	Waypoints []Waypoint `json:"waypoints"`
}

type Route struct {
	Geometry        Geometry `json:"geometry"`
	Legs            []Leg    `json:"legs"`
	Distance        float64  `json:"distance"`
	Duration        float64  `json:"duration"`
	TrafficDuration float64  `json:"TrafficDuration"`
	WeightName      string   `json:"weight_name"`
	Weight          float64  `json:"weight"`
}

type Geometry struct {
	Coordinates [][]float64 `json:"coordinates"`
	Type        string      `json:"type"`
}

// Leg structure
type Leg struct {
	Steps    []Step  `json:"steps"`
	Distance float64 `json:"distance"`
	Duration float64 `json:"duration"`
	Summary  string  `json:"summary"`
	Weight   float64 `json:"weight"`
}

// Step structure
type Step struct {
	Intersections []Intersection `json:"intersections"`
	DrivingSide   string         `json:"driving_side"`
	Geometry      Geometry       `json:"geometry"`
	Mode          string         `json:"mode"`
	Duration      float64        `json:"duration"`
	Maneuver      Maneuver       `json:"maneuver"`
	Weight        float64        `json:"weight"`
	Distance      float64        `json:"distance"`
	Name          string         `json:"name"`
}

// Intersection structure
type Intersection struct {
	Out      int       `json:"out,omitempty"`
	Entry    []bool    `json:"entry"`
	Bearings []int     `json:"bearings"`
	Location []float64 `json:"location"`
	In       int       `json:"in,omitempty"`
}

// Maneuver structure
type Maneuver struct {
	BearingAfter  int       `json:"bearing_after"`
	Type          string    `json:"type"`
	Modifier      string    `json:"modifier"`
	BearingBefore int       `json:"bearing_before"`
	Location      []float64 `json:"location"`
}

type Waypoint struct {
	Hint     string    `json:"hint"`
	Distance float64   `json:"distance"`
	Name     string    `json:"name"`
	Location []float64 `json:"location"`
}

func NewOSRMService() *OSRMService {
	return &OSRMService{BaseURL: config.LoadConfig().OSRMHost}
}

func (s *OSRMService) GetRoute(coordinates string, options map[string]string) (*RouteResponse, error) {
	url := fmt.Sprintf("%s/route/v1/driving/%s", s.BaseURL, coordinates)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	for key, value := range options {
		if key == "overview" || key == "geometries" || key == "steps" {
			query.Add(key, value)
		}
	}
	req.URL.RawQuery = query.Encode()
	//log.Printf("reqesut url: %s, Legs: %s", req.URL)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Log raw response
	bodyBytes, _ := io.ReadAll(resp.Body)
	//log.Printf("Raw response: %s", string(bodyBytes))

	// Decode the response
	var routeResponse RouteResponse
	if err := json.Unmarshal(bodyBytes, &routeResponse); err != nil {
		log.Printf("Error decoding response: %v", err)
		return nil, err
	}

	return &routeResponse, nil
}

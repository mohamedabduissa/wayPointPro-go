package osrm

import (
	"WayPointPro/internal/config"
	"WayPointPro/pkg/valhalla"
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
	Code      string           `json:"code"`
	Routes    []Route          `json:"routes"`
	Waypoints []Waypoint       `json:"waypoints"`
	Summary   valhalla.Summary `json:"summary"`
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
	Steps     []Step              `json:"steps"`
	Distance  float64             `json:"distance"`
	Duration  float64             `json:"duration"`
	Weight    float64             `json:"weight"`
	Maneuvers []valhalla.Maneuver `json:"maneuvers"`
	Summary   valhalla.Summary    `json:"summary"`
	Shape     string              `json:"shape,omitempty"`
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

// Function to convert Valhalla response to OSRM format
func (s *OSRMService) ConvertToOSRM(valhallaResponse *valhalla.RouteResponse) (*RouteResponse, error) {
	var osrmResponse RouteResponse
	osrmResponse.Code = "Ok"
	osrmResponse.Summary = valhallaResponse.Trip.Summary
	// Convert waypoints
	for _, loc := range valhallaResponse.Trip.Locations {
		osrmResponse.Waypoints = append(osrmResponse.Waypoints, Waypoint{
			Name:     loc.Type,
			Distance: 0,
			Location: []float64{loc.Lon, loc.Lat},
		})
	}

	// Convert legs and geometry
	var route Route
	route.WeightName = "routability"

	for _, leg := range valhallaResponse.Trip.Legs {
		var osrmLeg Leg
		osrmLeg.Distance = leg.Summary.Length
		osrmLeg.Duration = leg.Summary.Time
		osrmLeg.Weight = leg.Summary.Cost
		osrmLeg.Summary = leg.Summary
		osrmLeg.Maneuvers = leg.Maneuvers
		osrmLeg.Shape = leg.Shape
		route.Geometry = Geometry{}
		route.Legs = append(route.Legs, osrmLeg)
	}

	// Compute overall distance and duration
	for _, leg := range route.Legs {
		route.Distance += leg.Distance
		route.Duration += leg.Duration
	}

	osrmResponse.Routes = append(osrmResponse.Routes, route)
	return &osrmResponse, nil
}

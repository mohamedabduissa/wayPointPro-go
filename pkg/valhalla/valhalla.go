package valhalla

import (
	"WayPointPro/internal/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type ValhallaService struct {
	BaseURL string
}

// Request Struct for Valhalla API
type RouteRequest struct {
	Locations []Location `json:"locations"`
	Costing   string     `json:"costing"`
	//CostingOptions    CostingOptions    `json:"costing_options,omitempty"`
	DirectionsOptions DirectionsOptions `json:"directions_options,omitempty"`
	Alternates        int               `json:"alternates,omitempty"`
}

// Response Struct for Valhalla API
type RouteResponse struct {
	Trip Trip `json:"trip"`
}

type Trip struct {
	Locations []Location `json:"locations"`
	Legs      []Leg      `json:"legs"`
	Summary   Summary    `json:"summary"`
}

type Location struct {
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
	Type         string  `json:"type,omitempty"`
	SideOfStreet string  `json:"side_of_street,omitempty"`
}

type Leg struct {
	Maneuvers []Maneuver `json:"maneuvers"`
	Summary   Summary    `json:"summary"`
	Shape     string     `json:"shape,omitempty"`
}

type Summary struct {
	HasTimeRestrictions bool    `json:"has_time_restrictions"`
	HasToll             bool    `json:"has_toll"`
	HasHighway          bool    `json:"has_highway"`
	HasFerry            bool    `json:"has_ferry"`
	MinLat              float64 `json:"min_lat"`
	MinLon              float64 `json:"min_lon"`
	MaxLat              float64 `json:"max_lat"`
	MaxLon              float64 `json:"max_lon"`
	Time                float64 `json:"time"`
	Length              float64 `json:"length"`
	Cost                float64 `json:"cost"`
}

type Maneuver struct {
	Type                                int      `json:"type"`
	Instruction                         string   `json:"instruction"`
	VerbalSuccinctTransitionInstruction string   `json:"verbal_succinct_transition_instruction"`
	VerbalPreTransitionInstruction      string   `json:"verbal_pre_transition_instruction"`
	VerbalPostTransitionInstruction     string   `json:"verbal_post_transition_instruction"`
	StreetNames                         []string `json:"street_names"`
	BearingAfter                        int      `json:"bearing_after"`
	Time                                float64  `json:"time"`
	Length                              float64  `json:"length"`
	Cost                                float64  `json:"cost"`
	BeginShapeIndex                     int      `json:"begin_shape_index"`
	EndShapeIndex                       int      `json:"end_shape_index"`
	TravelMode                          string   `json:"travel_mode"`
	TravelType                          string   `json:"travel_type"`
}

type CostingOptions struct {
	Auto struct {
		UseHighways float64 `json:"use_highways,omitempty"`
		UseTolls    float64 `json:"use_tolls,omitempty"`
		UseFerry    float64 `json:"use_ferry,omitempty"`
	} `json:"auto,omitempty"`
}

type DirectionsOptions struct {
	Units string `json:"units,omitempty"`
}

// Create a new ValhallaService instance
func NewValhallaService() *ValhallaService {
	return &ValhallaService{BaseURL: config.LoadConfig().ValhallaHost}
}

// Function to request a route from Valhalla API
func (s *ValhallaService) GetRoute(locations []Location, options map[string]string) (*RouteResponse, error) {
	url := fmt.Sprintf("%s/route", s.BaseURL)

	// Construct the request payload
	requestData := RouteRequest{
		Locations: locations,
		//Costing:   costing,
		DirectionsOptions: DirectionsOptions{
			Units: "kilometers",
		},
	}

	for key, value := range options {
		if key == "costing" {
			requestData.Costing = value
		}
		if key == "alternates" {
			requestData.Alternates = 3
		}
	}

	// Set optional costing parameters
	//if costing == "auto" {
	//	requestData.CostingOptions.Auto.UseHighways = options["use_highways"]
	//	requestData.CostingOptions.Auto.UseTolls = options["use_tolls"]
	//	requestData.CostingOptions.Auto.UseFerry = options["use_ferry"]
	//}

	// Convert request to JSON
	jsonData, err := json.MarshalIndent(requestData, "", "  ") // Pretty print JSON
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and parse response
	bodyBytes, _ := io.ReadAll(resp.Body)

	var routeResponse RouteResponse
	if err := json.Unmarshal(bodyBytes, &routeResponse); err != nil {
		log.Printf("Error decoding response: %v", err)
		return nil, err
	}

	return &routeResponse, nil
}

package models

import (
	"WayPointPro/pkg/osrm"
	"fmt"
	"math"
	"math/rand"
)

type DirectionsValueObject struct {
	Value float64 `json:"value"`
	Text  string  `json:"text"`
}

type TransformedRoute struct {
	Status          bool                  `json:"status"`
	Message         string                `json:"message"`
	Distance        DirectionsValueObject `json:"distance"`
	TrafficDuration DirectionsValueObject `json:"traffic_duration"`
	Duration        DirectionsValueObject `json:"duration"`
	Geometry        interface{}           `json:"geometry"`
	Legs            []osrm.Leg            `json:"legs"`
	Waypoints       []osrm.Waypoint       `json:"waypoints"`
}

// TransformRoute transforms the OSRM route data into the desired format
func TransformRoute(route *osrm.RouteResponse) TransformedRoute {
	// Create DurationObject
	randomNumber := float64(rand.Intn(61) + 120) // 61 = (180 - 120 + 1)
	duration := route.Routes[0].Duration + randomNumber
	durationObject := DirectionsValueObject{
		Value: duration,
		Text:  fmt.Sprintf("%.1f mins", math.Round(duration/60*10)/10),
	}

	// Create TrafficDuration
	trafficDuration := route.Routes[0].TrafficDuration
	trafficDurationObject := DirectionsValueObject{
		Value: math.Round(trafficDuration*10) / 10,
		Text:  fmt.Sprintf("%.1f mins", math.Round(trafficDuration/60*10)/10),
	}

	// Create Distance
	distance := route.Routes[0].Distance
	//if distance > 1500 {
	//	distance = distance / 1000 * 10
	//} else {
	//	distance = distance * 10
	//}
	distance = distance * 10

	distanceObject := DirectionsValueObject{
		Value: distance / 10,
		Text:  fmt.Sprintf("%.1f km", math.Round(distance)/10),
	}

	// Transform and return the route
	return TransformedRoute{
		Status:          true,
		Message:         "Fetched route successfully!",
		Distance:        distanceObject,
		TrafficDuration: trafficDurationObject,
		Duration:        durationObject,
		Geometry:        route.Routes[0].Geometry.Coordinates,
		Legs:            route.Routes[0].Legs,
		Waypoints:       route.Waypoints,
	}
}

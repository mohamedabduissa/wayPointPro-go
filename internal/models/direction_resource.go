package models

import (
	"WayPointPro/pkg/osrm"
	"fmt"
	"math"
)

type DirectionsValueObject struct {
	Value float64 `json:"value"`
	Text  string  `json:"text"`
}

type TransformedRoute struct {
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
	duration := route.Routes[0].Duration
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
	distanceObject := DirectionsValueObject{
		Value: distance,
		Text:  fmt.Sprintf("%.1f km", math.Round(distance/1000*10)/10),
	}

	// Transform and return the route
	return TransformedRoute{
		Distance:        distanceObject,
		TrafficDuration: trafficDurationObject,
		Duration:        durationObject,
		Geometry:        route.Routes[0].Geometry.Coordinates,
		Legs:            route.Routes[0].Legs,
		Waypoints:       route.Waypoints,
	}
}

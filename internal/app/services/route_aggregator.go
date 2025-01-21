package services

import (
	"WayPointPro/pkg/osrm"
	"WayPointPro/pkg/traffic"
	"fmt"
	"log"
	"time"
)

type RouteAggregatorService struct {
	OSRMService      *osrm.OSRMService
	TrafficService   *traffic.Service
	TrafficOptimizer *traffic.Optimizer
}

func NewRouteAggregatorService(osrmService *osrm.OSRMService, trafficService *traffic.Service) *RouteAggregatorService {
	return &RouteAggregatorService{
		OSRMService:    osrmService,
		TrafficService: trafficService,
	}
}

func (s *RouteAggregatorService) GetAggregatedRoute(coordinates string, options map[string]string) (*osrm.RouteResponse, error) {
	startTime := time.Now()

	//log.Printf("coordinates: %v", coordinates)
	//log.Printf("options: %v", options)

	route, err := s.OSRMService.GetRoute(coordinates, options)
	if err != nil {
		log.Printf("Error fetching OSRM route: %v", err)
		return nil, err
	}

	duration := time.Since(startTime)
	log.Printf("OSRM API execution time: %v", duration)

	useTraffic := false
	for key, value := range options {
		if key == "traffic" && value == "true" {
			useTraffic = true
		}
	}

	if route == nil || len(route.Routes) == 0 {
		return nil, fmt.Errorf("invalid route")
	}

	if useTraffic {
		// 1. Fetch bounding box
		geometry := route.Routes[0].Geometry.Coordinates
		boundingBox := s.TrafficOptimizer.GetBoundingBox(geometry)

		// 2. Fetch traffic data
		trafficStartTime := time.Now()
		trafficData, err := s.TrafficService.FetchAndAnalyzeTraffic(boundingBox, 11)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch traffic data: %v", err)
		}
		log.Printf("Execution Time for fetching traffic: %v seconds", time.Since(trafficStartTime).Seconds())

		// 3. Adjust route based on traffic data
		adjustStartTime := time.Now()
		adjustedRoute := s.TrafficOptimizer.AdjustRouteTime(route.Routes[0], trafficData)
		log.Printf("Execution Time for analyze traffic: %v seconds", time.Since(adjustStartTime).Seconds())

		// 4. Update route duration
		route.Routes[0].TrafficDuration = adjustedRoute.TrafficDuration
	}

	return route, nil
}

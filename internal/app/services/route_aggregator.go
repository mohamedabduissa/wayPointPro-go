package services

import (
	"WayPointPro/internal/config"
	"WayPointPro/pkg/osrm"
	"WayPointPro/pkg/traffic"
	"WayPointPro/pkg/valhalla"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type RouteAggregatorService struct {
	OSRMService      *osrm.OSRMService
	ValhallaService  *valhalla.ValhallaService
	TrafficService   *traffic.Service
	TrafficOptimizer *traffic.Optimizer
}

func NewRouteAggregatorService(trafficService *traffic.Service) *RouteAggregatorService {
	return &RouteAggregatorService{
		OSRMService:     osrm.NewOSRMService(),
		ValhallaService: valhalla.NewValhallaService(),
		TrafficService:  trafficService,
	}
}
func (s *RouteAggregatorService) GetAggregatedRoute(coordinates string, options map[string]string) (*osrm.RouteResponse, error) {
	if config.LoadConfig().PLATFORM == "OSRM" {
		return s.GetOSRMService(coordinates, options)
	} else {
		return s.GetValhallaService(coordinates, options)
	}
}
func (s *RouteAggregatorService) GetValhallaService(coordinates string, options map[string]string) (*osrm.RouteResponse, error) {
	startTime := time.Now()

	//log.Printf("coordinates: %v", coordinates)
	//log.Printf("options: %v", options)
	coordinatesList, err := convertCoordinatesToValhalla(coordinates)
	if err != nil {
		log.Printf("Error fetching route: %v", err)
		return nil, err
	}

	route, err := s.ValhallaService.GetRoute(coordinatesList, "auto")
	if err != nil {
		log.Printf("Error fetching route: %v", err)
		return nil, err
	}

	duration := time.Since(startTime)
	log.Printf("Routing API execution time: %v", duration)

	if route == nil || len(route.Trip.Legs) == 0 {
		return nil, fmt.Errorf("invalid route")
	}
	startTime = time.Now()
	validRoute, err := s.OSRMService.ConvertToOSRM(route)
	duration = time.Since(startTime)
	log.Printf("Convert modeling API execution time: %v", duration)
	return validRoute, nil
}

func (s *RouteAggregatorService) GetOSRMService(coordinates string, options map[string]string) (*osrm.RouteResponse, error) {
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
		trafficData, err := s.TrafficService.FetchAndAnalyzeTraffic(boundingBox, 11, false)
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
	} else {
		route.Routes[0].TrafficDuration = route.Routes[0].Duration
	}

	return route, nil
}

// Function to convert coordinate string to Valhalla locations
func convertCoordinatesToValhalla(coordStr string) ([]valhalla.Location, error) {
	coords := strings.Split(coordStr, ";") // Split by ';'
	var locations []valhalla.Location

	for _, coord := range coords {
		parts := strings.Split(coord, ",") // Split lat/lon
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid coordinate format: %s", coord)
		}

		lat, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid latitude: %s", parts[1])
		}

		lon, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid longitude: %s", parts[0])
		}

		// Append to locations with "break" type (waypoints)
		locations = append(locations, valhalla.Location{Lat: lat, Lon: lon, Type: "break"})
	}

	return locations, nil
}

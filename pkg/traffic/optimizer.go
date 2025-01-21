package traffic

import (
	"WayPointPro/pkg/osrm"
	"fmt"
	"log"
	"math"
	_ "math/rand"
	_ "time"
)

// Optimizer handles route optimization and traffic adjustments
type Optimizer struct{}

func NewOptimizer() *Optimizer {
	return &Optimizer{}
}

// AdjustRouteTime adjusts the route time based on traffic data
func (o *Optimizer) AdjustRouteTime(route osrm.Route, trafficData []map[string]interface{}) osrm.Route {
	totalTime := route.Duration // Original travel time
	geometry := route.Geometry.Coordinates

	// Step 2: Simplify route geometry to reduce the number of segments
	simplifiedGeometry := o.SimplifyRoute(geometry, 0.00000000000001) // Adjust tolerance as needed

	// Step 3: Pre-compute congestion weights to avoid redundant calculations
	congestionWeights := o.PrecomputeCongestionWeights()

	// Step 4: Process each segment
	for i := 0; i < len(simplifiedGeometry)-1; i++ {
		segment := [2][]float64{simplifiedGeometry[i], simplifiedGeometry[i+1]}
		segmentTime := o.CalculateSegmentTime(segment)

		for _, trafficFeature := range trafficData {
			if o.IsSegmentInTraffic(segment, trafficFeature) {
				congestionLevel := trafficFeature["properties"].(map[string]interface{})["congestion"].(string)
				adjustedTime := congestionWeights[congestionLevel] * segmentTime
				totalTime += adjustedTime
				break
			}
		}
	}

	// Step 5: Add constant buffer time and return adjusted route
	totalTime += 60 // Add buffer time
	route.TrafficDuration = totalTime
	return route
}

// PrecomputeCongestionWeights generates weights for congestion levels
func (o *Optimizer) PrecomputeCongestionWeights() map[string]float64 {
	//congestionLevels := []string{"unknown", "low", "moderate", "heavy", "severe"}
	weights := map[string]float64{
		"unknown":  1.0,
		"low":      1.0,
		"moderate": 1.1,
		"heavy":    1.3,
		"severe":   1.5,
	}
	return weights
}

// SimplifyRoute simplifies a route geometry using Douglas-Peucker algorithm
func (o *Optimizer) SimplifyRoute(geometry [][]float64, tolerance float64) [][]float64 {
	if len(geometry) < 3 {
		return geometry // No simplification needed
	}

	preSimplified := o.PreSimplify(geometry, 0.00001)
	return o.DouglasPeucker(preSimplified, tolerance)
}

// DouglasPeucker recursively simplifies a line using the Douglas-Peucker algorithm
func (o *Optimizer) DouglasPeucker(points [][]float64, tolerance float64) [][]float64 {
	maxDistance := 0.0
	index := 0

	for i := 1; i < len(points)-1; i++ {
		distance := o.PerpendicularDistance(points[i], points[0], points[len(points)-1])
		if distance > maxDistance {
			index = i
			maxDistance = distance
		}
	}

	if maxDistance > tolerance {
		left := o.DouglasPeucker(points[:index+1], tolerance)
		right := o.DouglasPeucker(points[index:], tolerance)
		return append(left[:len(left)-1], right...)
	}

	return [][]float64{points[0], points[len(points)-1]}
}

// PerpendicularDistance calculates the perpendicular distance of a point from a line
func (o *Optimizer) PerpendicularDistance(point, lineStart, lineEnd []float64) float64 {
	x0, y0 := point[0], point[1]
	x1, y1 := lineStart[0], lineStart[1]
	x2, y2 := lineEnd[0], lineEnd[1]

	dx := x2 - x1
	dy := y2 - y1

	if dx == 0 && dy == 0 {
		return math.Hypot(x0-x1, y0-y1)
	}

	numerator := math.Abs(dy*x0 - dx*y0 + x2*y1 - y2*x1)
	denominator := math.Sqrt(dx*dx + dy*dy)
	return numerator / denominator
}

// PreSimplify reduces the number of points using a grid-based simplification
func (o *Optimizer) PreSimplify(geometry [][]float64, gridSize float64) [][]float64 {
	grid := make(map[string]bool)

	// Preallocate slice with the same capacity as geometry for slight optimization
	simplified := make([][]float64, 0, len(geometry))

	for _, point := range geometry {
		gridX := math.Round(point[0] / gridSize)
		gridY := math.Round(point[1] / gridSize)
		gridKey := fmt.Sprintf("%f,%f", gridX, gridY)

		if !grid[gridKey] {
			grid[gridKey] = true
			simplified = append(simplified, point)
		}
	}
	return simplified
}

// CalculateSegmentTime estimates the time for a route segment
func (o *Optimizer) CalculateSegmentTime(segment [2][]float64) float64 {
	distance := o.CalculateDistance(segment[0], segment[1])
	averageSpeed := 50.0 * 1000.0 / 3600.0 // 50 km/h in m/s
	return distance / averageSpeed
}

// CalculateDistance computes the Haversine distance between two points
func (o *Optimizer) CalculateDistance(point1, point2 []float64) float64 {
	lat1, lon1 := degreesToRadians(point1[1]), degreesToRadians(point1[0])
	lat2, lon2 := degreesToRadians(point2[1]), degreesToRadians(point2[0])

	dLat := lat2 - lat1
	dLon := lon2 - lon1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return 6371000 * c // Radius of Earth in meters
}

// IsSegmentInTraffic checks if a segment intersects with traffic data
func (o *Optimizer) IsSegmentInTraffic(segment [2][]float64, trafficFeature map[string]interface{}) bool {
	// Safely assert trafficFeature["geometry"] as map[string]interface{}
	geometry, ok := trafficFeature["geometry"].(map[string]interface{})
	if !ok {
		log.Println("Invalid geometry in traffic feature")
		return false
	}

	// Safely assert geometry["coordinates"] as []interface{}
	rawCoordinates, ok := geometry["coordinates"].([]interface{})
	if !ok {
		log.Println("Invalid coordinates in traffic feature")
		return false
	}

	// Convert rawCoordinates ([][][]float64)
	// Preallocate with potential max capacity for slight optimization
	trafficGeometry := make([][][]float64, 0, len(rawCoordinates))

	for _, rawLine := range rawCoordinates {
		rawLineSlice, ok := rawLine.([]interface{})
		if !ok {
			log.Println("Invalid MultiLineString line")
			continue
		}

		// Preallocate each line with length of rawLineSlice
		lineGeometry := make([][]float64, 0, len(rawLineSlice))
		for _, rawCoord := range rawLineSlice {
			coordPair, ok := rawCoord.([]interface{})
			if !ok || len(coordPair) != 2 {
				//log.Println("Invalid coordinate pair in traffic feature")
				continue
			}

			lat, latOk := coordPair[0].(float64)
			lon, lonOk := coordPair[1].(float64)
			if !latOk || !lonOk {
				//log.Println("Invalid coordinate values in traffic feature")
				continue
			}

			lineGeometry = append(lineGeometry, []float64{lat, lon})
		}

		trafficGeometry = append(trafficGeometry, lineGeometry)
	}

	// Check if the segment intersects with any traffic segment
	for _, line := range trafficGeometry {
		for i := 0; i < len(line)-1; i++ {
			trafficSegment := [2][]float64{line[i], line[i+1]}
			if o.AreSegmentsIntersecting(segment, trafficSegment) {
				return true
			}
		}
	}

	return false
}

// AreSegmentsIntersecting checks if two line segments intersect
func (o *Optimizer) AreSegmentsIntersecting(segment1 [2][]float64, segment2 [2][]float64) bool {
	// Extract points
	p1, q1 := segment1[0], segment1[1]
	p2, q2 := segment2[0], segment2[1]

	// Validate all points
	if !isValidPoint(p1) || !isValidPoint(q1) || !isValidPoint(p2) || !isValidPoint(q2) {
		return false
	}

	// Orientation function
	orientation := func(p, q, r []float64) int {
		val := (q[1]-p[1])*(r[0]-q[0]) - (q[0]-p[0])*(r[1]-q[1])
		if math.Abs(val) < 1e-10 {
			return 0 // Collinear
		}
		if val > 0 {
			return 1 // Clockwise
		}
		return 2 // Counterclockwise
	}

	// Check if point lies on a segment
	onSegment := func(p, q, r []float64) bool {
		return q[0] <= math.Max(p[0], r[0]) && q[0] >= math.Min(p[0], r[0]) &&
			q[1] <= math.Max(p[1], r[1]) && q[1] >= math.Min(p[1], r[1])
	}

	// Find orientations
	o1 := orientation(p1, q1, p2)
	o2 := orientation(p1, q1, q2)
	o3 := orientation(p2, q2, p1)
	o4 := orientation(p2, q2, q1)

	// General case: Segments intersect if the orientations differ
	if o1 != o2 && o3 != o4 {
		return true
	}

	// Special cases
	if o1 == 0 && onSegment(p1, p2, q1) {
		return true // p2 lies on p1-q1
	}
	if o2 == 0 && onSegment(p1, q2, q1) {
		return true // q2 lies on p1-q1
	}
	if o3 == 0 && onSegment(p2, p1, q2) {
		return true // p1 lies on p2-q2
	}
	if o4 == 0 && onSegment(p2, q1, q2) {
		return true // q1 lies on p2-q2
	}

	return false // No intersection
}

// isValidPoint checks if a point has valid coordinates
func isValidPoint(point []float64) bool {
	return len(point) == 2 && !math.IsNaN(point[0]) && !math.IsNaN(point[1])
}

// degreesToRadians converts degrees to radians
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// GetBoundingBox calculates the bounding box for a given geometry (set of coordinates).
func (o *Optimizer) GetBoundingBox(geometry [][]float64) map[string]float64 {
	// Initialize min and max latitude and longitude
	minLat, maxLat := geometry[0][1], geometry[0][1]
	minLng, maxLng := geometry[0][0], geometry[0][0]

	// Iterate through all points in the geometry
	for _, point := range geometry {
		if point[1] < minLat {
			minLat = point[1]
		}
		if point[1] > maxLat {
			maxLat = point[1]
		}
		if point[0] < minLng {
			minLng = point[0]
		}
		if point[0] > maxLng {
			maxLng = point[0]
		}
	}

	// Return the bounding box as a map
	return map[string]float64{
		"north": maxLat,
		"south": minLat,
		"east":  maxLng,
		"west":  minLng,
	}
}

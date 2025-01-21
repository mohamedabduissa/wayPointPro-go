package traffic

import (
	"fmt"
	"log"
	"math"
)

// Tile represents a map tile
type Tile struct {
	X    int
	Y    int
	Zoom int
}

// latLonToTile converts latitude and longitude to tile coordinates
func latLonToTile(lat, lon float64, zoom int) map[string]int {
	x := int(math.Floor((lon + 180.0) / 360.0 * math.Pow(2.0, float64(zoom))))
	y := int(math.Floor((1.0 - math.Log(math.Tan(lat*math.Pi/180.0)+1.0/math.Cos(lat*math.Pi/180.0))/math.Pi) / 2.0 * math.Pow(2.0, float64(zoom))))
	return map[string]int{"x": x, "y": y}
}

// parseTileKey parses a tile key string into a Tile struct
func parseTileKey(key string) *Tile {
	var z, x, y int
	_, err := fmt.Sscanf(key, "tile_%d_%d_%d", &z, &x, &y)
	if err != nil {
		log.Printf("Failed to parse tile key: %s", key)
		return nil
	}
	return &Tile{X: x, Y: y, Zoom: z}
}

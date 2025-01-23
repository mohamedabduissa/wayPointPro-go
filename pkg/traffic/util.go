package traffic

import "math"

// chunkArray splits an array into chunks
func chunkArray(arr []int, size int) [][]int {
	var chunks [][]int
	for i := 0; i < len(arr); i += size {
		end := i + size
		if end > len(arr) {
			end = len(arr)
		}
		chunks = append(chunks, arr[i:end])
	}
	return chunks
}

// generateRange generates a range of integers
func generateRange(start, end int) []int {
	var rangeArray []int
	for i := start; i <= end; i++ {
		rangeArray = append(rangeArray, i)
	}
	return rangeArray
}

// batchTileRange splits tile ranges into batches
func (s *Service) batchTileRange(tileRange map[string][]int, batchSize int) []map[string]interface{} {
	var batches []map[string]interface{}

	xChunks := chunkArray(tileRange["x"], batchSize)
	yChunks := chunkArray(tileRange["y"], batchSize)

	for _, xChunk := range xChunks {
		for _, yChunk := range yChunks {
			batches = append(batches, map[string]interface{}{
				"x": xChunk,
				"y": yChunk,
			})
		}
	}

	return batches
}

// GetTileRange calculates the range of tiles for a bounding box, constrained by maximum allowed tile dimensions
func (s *Service) getTileRange(boundingBox map[string]float64, zoom int) map[string][]int {
	// Calculate the northwest and southeast tile coordinates
	northWest := latLonToTile(boundingBox["north"], boundingBox["west"], zoom)
	southEast := latLonToTile(boundingBox["south"], boundingBox["east"], zoom)

	// Calculate the number of tiles in x and y directions
	tileWidth := southEast["x"] - northWest["x"] + 1
	tileHeight := southEast["y"] - northWest["y"] + 1

	// Define maximum allowed tiles
	maxTilesWide := 4
	maxTilesHigh := 5

	// Check if the tile range exceeds the maximum allowed tiles
	if tileWidth > maxTilesWide || tileHeight > maxTilesHigh {
		// Center of the bounding box
		centerLat := (boundingBox["north"] + boundingBox["south"]) / 2
		centerLon := (boundingBox["west"] + boundingBox["east"]) / 2

		// Calculate new tile extents
		tileSize := 360.0 / math.Pow(2, float64(zoom)) // Degrees per tile at the given zoom
		newHalfTileWidth := float64(maxTilesWide) * tileSize / 2
		newHalfTileHeight := float64(maxTilesHigh) * tileSize / 2

		// Adjust the bounding box to fit within the max tile range
		boundingBox["north"] = centerLat + (newHalfTileHeight / 2)
		boundingBox["south"] = centerLat - (newHalfTileHeight / 2)
		boundingBox["west"] = centerLon - (newHalfTileWidth / 2)
		boundingBox["east"] = centerLon + (newHalfTileWidth / 2)

		// Recalculate tile coordinates after adjusting the bounding box
		northWest = latLonToTile(boundingBox["north"], boundingBox["west"], zoom)
		southEast = latLonToTile(boundingBox["south"], boundingBox["east"], zoom)
	}

	// Generate ranges for x and y tile indices
	return map[string][]int{
		"x": generateRange(northWest["x"], southEast["x"]),
		"y": generateRange(northWest["y"], southEast["y"]),
	}
}

// GetTileRange calculates the range of tiles for a bounding box, constrained by maximum allowed tile dimensions
func (s *Service) FullGetTileRange(boundingBox map[string]float64, zoom int) map[string][]int {
	// Calculate the northwest and southeast tile coordinates
	northWest := latLonToTile(boundingBox["north"], boundingBox["west"], zoom)
	southEast := latLonToTile(boundingBox["south"], boundingBox["east"], zoom)

	// Calculate the number of tiles in x and y directions
	tileWidth := southEast["x"] - northWest["x"] + 1
	tileHeight := southEast["y"] - northWest["y"] + 1

	// Define maximum allowed tiles
	maxTilesWide := 40000
	maxTilesHigh := 50000

	// Check if the tile range exceeds the maximum allowed tiles
	if tileWidth > maxTilesWide || tileHeight > maxTilesHigh {
		// Center of the bounding box
		centerLat := (boundingBox["north"] + boundingBox["south"]) / 2
		centerLon := (boundingBox["west"] + boundingBox["east"]) / 2

		// Calculate new tile extents
		tileSize := 360.0 / math.Pow(2, float64(zoom)) // Degrees per tile at the given zoom
		newHalfTileWidth := float64(maxTilesWide) * tileSize / 2
		newHalfTileHeight := float64(maxTilesHigh) * tileSize / 2

		// Adjust the bounding box to fit within the max tile range
		boundingBox["north"] = centerLat + (newHalfTileHeight / 2)
		boundingBox["south"] = centerLat - (newHalfTileHeight / 2)
		boundingBox["west"] = centerLon - (newHalfTileWidth / 2)
		boundingBox["east"] = centerLon + (newHalfTileWidth / 2)

		// Recalculate tile coordinates after adjusting the bounding box
		northWest = latLonToTile(boundingBox["north"], boundingBox["west"], zoom)
		southEast = latLonToTile(boundingBox["south"], boundingBox["east"], zoom)
	}

	// Generate ranges for x and y tile indices
	return map[string][]int{
		"x": generateRange(northWest["x"], southEast["x"]),
		"y": generateRange(northWest["y"], southEast["y"]),
	}
}

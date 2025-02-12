package jobs

import (
	"WayPointPro/pkg/logging"
	"WayPointPro/pkg/traffic"
	"log"
)

// CollectTrafficJobHandle job
func CollectTrafficJobHandle() {
	log.Printf("collect traffic job start")
	return
	// Define the bounding box for jeddah
	boundingBox := map[string]float64{
		"north": 21.67, // Northernmost latitude
		"south": 21.27, // Southernmost latitude
		"west":  39.07, // Westernmost longitude
		"east":  39.32, // Easternmost longitude
	}

	service := traffic.NewService()
	_, err := service.FetchAndAnalyzeTraffic(boundingBox, 11, true)

	// Define the bounding box for maccah
	boundingBox = map[string]float64{
		"north": 21.52, // Northernmost latitude
		"south": 21.23, // Southernmost latitude
		"west":  39.62, // Westernmost longitude
		"east":  40.03, // Easternmost longitude
	}

	_, err = service.FetchAndAnalyzeTraffic(boundingBox, 11, true)

	// Define the bounding box for ryadih
	boundingBox = map[string]float64{
		"north": 25.00, // Northernmost latitude
		"south": 24.56, // Southernmost latitude
		"west":  46.55, // Westernmost longitude
		"east":  47.03, // Easternmost longitude
	}

	_, err = service.FetchAndAnalyzeTraffic(boundingBox, 11, true)

	// Define the bounding box for Madinah
	boundingBox = map[string]float64{
		"north": 24.67, // Northernmost latitude
		"south": 24.33, // Southernmost latitude
		"west":  39.45, // Westernmost longitude
		"east":  39.75, // Easternmost longitude
	}

	_, err = service.FetchAndAnalyzeTraffic(boundingBox, 11, true)

	// Define the bounding box for Dammam
	boundingBox = map[string]float64{
		"north": 26.60, // Northernmost latitude
		"south": 26.30, // Southernmost latitude
		"west":  50.00, // Westernmost longitude
		"east":  50.20, // Easternmost longitude
	}

	_, err = service.FetchAndAnalyzeTraffic(boundingBox, 11, true)

	// Define the bounding box for Tabuk
	boundingBox = map[string]float64{
		"north": 28.60, // Northernmost latitude
		"south": 28.30, // Southernmost latitude
		"west":  36.50, // Westernmost longitude
		"east":  36.90, // Easternmost longitude
	}

	_, err = service.FetchAndAnalyzeTraffic(boundingBox, 11, true)

	// Define the bounding box for Buraydah
	boundingBox = map[string]float64{
		"north": 26.43, // Northernmost latitude
		"south": 26.33, // Southernmost latitude
		"west":  43.95, // Westernmost longitude
		"east":  44.10, // Easternmost longitude
	}

	_, err = service.FetchAndAnalyzeTraffic(boundingBox, 11, true)

	// Define the bounding box for Abha
	boundingBox = map[string]float64{
		"north": 18.30, // Northernmost latitude
		"south": 18.20, // Southernmost latitude
		"west":  42.45, // Westernmost longitude
		"east":  42.55, // Easternmost longitude
	}

	_, err = service.FetchAndAnalyzeTraffic(boundingBox, 11, true)

	// Define the bounding box for Taif
	boundingBox = map[string]float64{
		"north": 21.30, // Northernmost latitude
		"south": 21.15, // Southernmost latitude
		"west":  40.35, // Westernmost longitude
		"east":  40.45, // Easternmost longitude
	}

	_, err = service.FetchAndAnalyzeTraffic(boundingBox, 11, true)

	// Define the bounding box for Hofuf
	boundingBox = map[string]float64{
		"north": 25.40, // Northernmost latitude
		"south": 25.30, // Southernmost latitude
		"west":  49.55, // Westernmost longitude
		"east":  49.65, // Easternmost longitude
	}

	_, err = service.FetchAndAnalyzeTraffic(boundingBox, 11, true)

	// Define the bounding box for Qatif
	boundingBox = map[string]float64{
		"north": 26.70, // Northernmost latitude
		"south": 26.50, // Southernmost latitude
		"west":  50.00, // Westernmost longitude
		"east":  50.20, // Easternmost longitude
	}

	_, err = service.FetchAndAnalyzeTraffic(boundingBox, 11, true)

	// Define the bounding box for Khobar
	boundingBox = map[string]float64{
		"north": 26.30, // Northernmost latitude
		"south": 26.20, // Southernmost latitude
		"west":  50.10, // Westernmost longitude
		"east":  50.20, // Easternmost longitude
	}

	_, err = service.FetchAndAnalyzeTraffic(boundingBox, 11, true)

	logging.Logger.Println(err)
	//Initialize the database singleton
}

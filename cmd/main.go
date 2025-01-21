package main

import (
	"WayPointPro/internal/config"
	"WayPointPro/internal/db"
	"WayPointPro/internal/routes"
	"WayPointPro/pkg/logging"
	"log"
	"net/http"
)

func main() {
	// Initialize logger
	logging.InitLogger()

	// Load configuration
	cfg := config.LoadConfig()
	logging.Logger.Println("Configuration loaded")

	// Initialize the database singleton
	database := db.Connect()
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	router := routes.InitRoutes()

	log.Printf("Server running on http://localhost:%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
	log.Printf("exceution time")

}

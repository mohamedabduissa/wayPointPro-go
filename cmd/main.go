package main

import (
	"WayPointPro/internal/config"
	"WayPointPro/internal/routes"
	"WayPointPro/pkg/queue"
	"WayPointPro/pkg/queue/jobs"
	"WayPointPro/pkg/queue/scheduler"
	"log"
	"net/http"
	"time"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	router := routes.InitRoutes()

	runQueue()
	log.Printf("Server running on http://localhost:%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
	log.Printf("exceution time")

}

func runQueue() {
	// Create a job queue with 3 workers
	q := queue.NewQueue(3)
	go q.Start()

	// Create a scheduler
	s := scheduler.NewScheduler()

	// Add tasks to the scheduler
	s.AddTask(15*time.Minute, queue.Job{
		ID:      1,
		Name:    "Collect traffic",
		Execute: jobs.CollectTrafficJobHandle,
	})

	// Add tasks to the scheduler
	s.AddTask(24*time.Hour, queue.Job{
		ID:      1,
		Name:    "Reset Request Limit",
		Execute: jobs.ResetRequestLimit,
	})

	// Start the scheduler
	s.Start(q)

}

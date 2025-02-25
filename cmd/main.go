package main

import (
	"WayPointPro/internal/config"
	"WayPointPro/internal/routes"
	"WayPointPro/pkg/queue"
	"WayPointPro/pkg/queue/jobs"
	"WayPointPro/pkg/queue/scheduler"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	router := routes.InitRoutes()

	// Create HTTP Server
	srv := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Create a context to handle graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start Queue Processing
	runQueue()

	// Start HTTP Server in a Goroutine
	go func() {
		log.Printf("🚀 Server running on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// Wait for shutdown signal (Ctrl+C)
	<-ctx.Done()
	log.Println("🛑 Shutdown signal received")

	// Measure execution time
	startTime := time.Now()

	// Stop processing queue safely
	stop()

	// Gracefully shutdown HTTP server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	// Log execution time
	executionTime := time.Since(startTime)
	log.Printf("✅ Server exited properly. Execution time: %v", executionTime)

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

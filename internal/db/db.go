package db

import (
	"WayPointPro/internal/config"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"log"
	"strconv"
	"sync"
	"time"
)

var (
	instance *sqlx.DB
	once     sync.Once
)

// GetDB returns the singleton database instance
func Connect() *sqlx.DB {
	once.Do(func() {
		cfg := config.LoadConfig()
		// Build the PostgreSQL connection string (DSN)
		dbPortStr := cfg.DBPort

		// Convert the string to an integer
		dbPort, err := strconv.Atoi(dbPortStr)
		if err != nil {
			fmt.Printf("Error converting DB_PORT to int: %v\n", err)
			return
		}

		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost, dbPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
		)

		// Connect to the database
		db, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Set connection pool configurations
		db.SetMaxOpenConns(48)                  // Maximum open connections
		db.SetMaxIdleConns(24)                  // Maximum idle connections
		db.SetConnMaxLifetime(30 * time.Minute) // Reuse connections for up to 30 minutes

		log.Println("Database connection initialized")
		instance = db
	})
	return instance
}
func CloseDB() {
	if instance != nil {
		instance.Close()
		log.Println("Database connection closed")
	}
}

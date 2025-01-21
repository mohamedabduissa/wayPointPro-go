package db

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"log"
	"sync"
	"time"
)

var (
	instance *sqlx.DB
	once     sync.Once
)

// Config holds the database configuration details.
type Config struct {
	Host     string // Database host
	Port     int    // Database port
	User     string // Database username
	Password string // Database password
	DBName   string // Database name
}

// GetDB returns the singleton database instance
func Connect() *sqlx.DB {
	once.Do(func() {
		cfg := Initialize()
		// Build the PostgreSQL connection string (DSN)
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
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
func Initialize() Config {
	dbCfg := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "waypointpro_user",
		Password: "Dh3hMMjzhaLq5VL7RT",
		DBName:   "waypointpro_1",
	}
	return dbCfg
}

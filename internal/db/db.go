package db

import (
	"WayPointPro/internal/config"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool" // pgx (better PostgreSQL driver)
	_ "github.com/lib/pq"             // PostgreSQL driver
	"log"
	"strconv"
	"sync"
	"time"
)

var (
	instance *pgxpool.Pool
	once     sync.Once
)

// GetDB returns the singleton database instance
func Connect() *pgxpool.Pool {
	once.Do(func() {
		cfg := config.LoadConfig()
		// Build the PostgreSQL connection string (DSN)
		// Convert the string to an integer
		dbPort, err := strconv.Atoi(cfg.DBPort)
		if err != nil {
			fmt.Printf("Error converting DB_PORT to int: %v\n", err)
			return
		}

		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost, dbPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
		)

		// Connect to the database
		// Create a connection pool
		poolConfig, poolErr := pgxpool.ParseConfig(dsn)
		if poolErr != nil {
			log.Printf("Failed to parse database config: %v\n", poolErr)
			err = poolErr
			return
		}

		// Optimize connection pool settings
		poolConfig.MaxConns = 48                      // Max open connections
		poolConfig.MinConns = 10                      // Min idle connections
		poolConfig.MaxConnLifetime = 30 * time.Minute // Max lifetime of a connection
		poolConfig.MaxConnIdleTime = 5 * time.Minute  // Max idle time before closing

		// Connect to PostgreSQL
		dbPool, connectErr := pgxpool.NewWithConfig(context.Background(), poolConfig)
		if connectErr != nil {
			log.Printf("Failed to connect to database: %v\n", connectErr)
			err = connectErr
			return
		}

		instance = dbPool
		log.Println("✅ Database connection initialized")
	})
	return instance
}
func CloseDB() {
	if instance != nil {
		instance.Close()
		log.Println("Database connection closed")
	}
}

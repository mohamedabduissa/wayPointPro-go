package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	OSRMHost   string
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

var (
	instance *Config
	once     sync.Once
)

func LoadConfig() *Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using defaults")
		}

		config := &Config{
			OSRMHost:   getEnv("OSRM_HOST", "http://207.127.98.226:5000"),
			Port:       getEnv("PORT", "8080"),
			DBHost:     getEnv("DB_HOST", "localhost"),
			DBUser:     getEnv("DB_USERNAME", "postgres"),
			DBPassword: getEnv("DB_PASSWORD", "123456"),
			DBName:     getEnv("DB_NAME", "waypointpro_1"),
			DBPort:     getEnv("DB_PORT", "5432"),
		}
		instance = config
	})
	return instance
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

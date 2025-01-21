package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	OSRMHost string
	Port     string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults")
	}

	return &Config{
		OSRMHost: getEnv("OSRM_HOST", "http://207.127.98.226:5000"),
		Port:     getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

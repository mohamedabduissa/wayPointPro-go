package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	ValhallaHost string
	OSRMHost     string
	Port         string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	REDIS        string
	PLATFORM     string
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
			ValhallaHost: getEnv("VALHALLA_HOST", ""),
			OSRMHost:     getEnv("OSRM_HOST", ""),
			Port:         getEnv("PORT", ""),
			DBHost:       getEnv("DB_HOST", ""),
			DBUser:       getEnv("DB_USER", ""),
			DBPassword:   getEnv("DB_PASSWORD", ""),
			DBName:       getEnv("DB_NAME", ""),
			DBPort:       getEnv("DB_PORT", "5432"),
			REDIS:        getEnv("REDIS", ""),
			PLATFORM:     getEnv("PLATFORM", ""),
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

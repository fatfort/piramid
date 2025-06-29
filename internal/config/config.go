package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port        string
	DatabaseURL string
	NATSUrl     string
	JWTSecret   string
	Environment string
	GeoIPDBPath string
	HomeNet     string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "65605"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://piramid:piramid@postgres:5432/piramid?sslmode=disable"),
		NATSUrl:     getEnv("NATS_URL", "nats://nats:4222"),
		JWTSecret:   getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		Environment: getEnv("ENVIRONMENT", "development"),
		GeoIPDBPath: getEnv("GEOIP_DB_PATH", "/usr/share/GeoIP/GeoLite2-City.mmdb"),
		HomeNet:     getEnv("HOME_NET", "any"),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// getEnvAsBool gets an environment variable as boolean with a fallback value
func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}

package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// App
	Port string

	// JWT
	JWTSecret string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func Load() *Config {
	_ = godotenv.Load() // safe in prod (no-op if missing)

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		JWTSecret:   mustEnv("JWT_SECRET"),
		DBHost:      mustEnv("DB_HOST"),
		DBPort:      mustEnv("DB_PORT"),
		DBUser:      mustEnv("DB_USER"),
		DBPassword: mustEnv("DB_PASSWORD"),
		DBName:      mustEnv("DB_NAME"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
	}

	return cfg
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("❌ Missing required env var: %s", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

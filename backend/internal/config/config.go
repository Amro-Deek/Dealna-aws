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

	SMTP SMTPConfig

	KeycloakBaseURL string
	KeycloakRealm   string
	KeycloakClientID string
	KeycloakAdminClientID string
	KeycloakAdminClientSecret string
	SQSQueueURL               string
	AWSRegion                 string
	LambdaFunctionName        string
	QdrantURL                 string
	QdrantAPIKey              string
}
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func Load() *Config {
	_ = godotenv.Load() // safe in prod (no-op if missing)

	cfg := &Config{
		Port:       getEnv("PORT", "8080"),
		JWTSecret:  mustEnv("JWT_SECRET"),
		DBHost:     mustEnv("DB_HOST"),
		DBPort:     mustEnv("DB_PORT"),
		DBUser:     mustEnv("DB_USER"),
		DBPassword: mustEnv("DB_PASSWORD"),
		DBName:     mustEnv("DB_NAME"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		KeycloakBaseURL: getEnv("KEYCLOAK_BASE_URL", ""),
		KeycloakRealm:   getEnv("KEYCLOAK_REALM", ""),
		KeycloakClientID: getEnv("KEYCLOAK_CLIENT_ID", ""),
		KeycloakAdminClientID: getEnv("KEYCLOAK_ADMIN_CLIENT_ID", ""),
		KeycloakAdminClientSecret: getEnv("KEYCLOAK_ADMIN_CLIENT_SECRET", ""),
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnv("SMTP_PORT", ""),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
		},
		SQSQueueURL:   getEnv("SQS_QUEUE_URL", ""),
		AWSRegion:     getEnv("AWS_REGION", "us-east-1"),
		LambdaFunctionName: getEnv("LAMBDA_FUNCTION_NAME", "dealna-search-worker"),
		QdrantURL:     getEnv("QDRANT_URL", ""),
		QdrantAPIKey:  getEnv("QDRANT_API_KEY", ""),
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

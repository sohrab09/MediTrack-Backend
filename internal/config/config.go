package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DBUrl     string
	JWTSecret string
}

func LoadConfig() (*Config, error) {
	// .env is optional; deployment environment variables take precedence.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	cfg := &Config{
		Port:      getEnv("PORT", "8080"),
		DBUrl:     getEnv("DATABASE_URL", ""),
		JWTSecret: strings.TrimSpace(os.Getenv("JWT_SECRET")),
	}

	if cfg.DBUrl == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// HS256 should use a strong, randomly generated secret.
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

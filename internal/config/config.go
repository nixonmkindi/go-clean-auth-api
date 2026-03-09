package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config contains runtime configuration loaded from environment variables.
type Config struct {
	AppEnv          string
	AppPort         string
	LogLevel        string
	DatabaseURL     string
	JWTSecret       string
	JWTTTLMinutes   int
	DefaultPageSize int
	MaxPageSize     int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		AppPort:         getEnv("APP_PORT", "8080"),
		LogLevel:        getEnv("LOG_LEVEL", "INFO"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		JWTTTLMinutes:   getInt("JWT_TTL_MINUTES", 60),
		DefaultPageSize: getInt("DEFAULT_PAGE_SIZE", 20),
		MaxPageSize:     getInt("MAX_PAGE_SIZE", 100),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.JWTTTLMinutes <= 0 {
		return nil, fmt.Errorf("JWT_TTL_MINUTES must be positive")
	}
	if cfg.DefaultPageSize <= 0 {
		cfg.DefaultPageSize = 20
	}
	if cfg.MaxPageSize < cfg.DefaultPageSize {
		cfg.MaxPageSize = cfg.DefaultPageSize
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

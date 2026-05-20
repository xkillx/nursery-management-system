package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv      string
	APIPort     string
	APIBasePath string
	DatabaseURL string

	JWTAccessSecret    string
	JWTRefreshSecret   string
	JWTAccessTTLMin    int
	JWTRefreshTTLHours int
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:      getEnv("APP_ENV", "local"),
		APIPort:     getEnv("API_PORT", "8080"),
		APIBasePath: getEnv("API_BASE_PATH", "/api/v1"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),

		JWTAccessSecret:    strings.TrimSpace(os.Getenv("JWT_ACCESS_SECRET")),
		JWTRefreshSecret:   strings.TrimSpace(os.Getenv("JWT_REFRESH_SECRET")),
		JWTAccessTTLMin:    getEnvInt("JWT_ACCESS_TTL_MINUTES", 15),
		JWTRefreshTTLHours: getEnvInt("JWT_REFRESH_TTL_HOURS", 720),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if cfg.JWTAccessSecret == "" {
		return Config{}, errors.New("JWT_ACCESS_SECRET is required")
	}

	if cfg.JWTRefreshSecret == "" {
		return Config{}, errors.New("JWT_REFRESH_SECRET is required")
	}

	if !strings.HasPrefix(cfg.APIBasePath, "/") {
		return Config{}, fmt.Errorf("API_BASE_PATH must start with '/': %q", cfg.APIBasePath)
	}

	if cfg.JWTAccessTTLMin <= 0 {
		return Config{}, errors.New("JWT_ACCESS_TTL_MINUTES must be > 0")
	}

	if cfg.JWTRefreshTTLHours <= 0 {
		return Config{}, errors.New("JWT_REFRESH_TTL_HOURS must be > 0")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return parsed
}

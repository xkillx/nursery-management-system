package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AppEnv      string
	APIPort     string
	APIBasePath string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:      getEnv("APP_ENV", "local"),
		APIPort:     getEnv("API_PORT", "8080"),
		APIBasePath: getEnv("API_BASE_PATH", "/api/v1"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if !strings.HasPrefix(cfg.APIBasePath, "/") {
		return Config{}, fmt.Errorf("API_BASE_PATH must start with '/': %q", cfg.APIBasePath)
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

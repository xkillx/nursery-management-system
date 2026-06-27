package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv      string
	APIPort     string
	APIBasePath string
	DatabaseURL string

	JWTAccessSecret         string
	JWTRefreshSecret        string
	JWTAccessTTLMin         int
	JWTRefreshTTLHours      int
	JWTRefreshShortTTLHours int

	WebBaseURL                   string
	EmailProvider                string
	SMTPHost                     string
	SMTPPort                     int
	SMTPUser                     string
	SMTPPass                     string
	SMTPFrom                     string
	PasswordResetTokenSecret     string
	PasswordResetTokenTTLMinutes int
	InviteTokenSecret            string
	InviteTokenTTLHours          int

	SchedulerOwner bool

	StripeSecretKey      string
	StripeWebhookSecret  string
	StripePublishableKey string

	LogLevel       string
	MetricsEnabled bool
}

func Load() (Config, error) {
	accessTTLMin, err := getEnvInt("JWT_ACCESS_TTL_MINUTES", 15)
	if err != nil {
		return Config{}, err
	}
	refreshTTLHours, err := getEnvInt("JWT_REFRESH_TTL_HOURS", 720)
	if err != nil {
		return Config{}, err
	}
	refreshShortTTLHours, err := getEnvInt("JWT_REFRESH_SHORT_TTL_HOURS", 24)
	if err != nil {
		return Config{}, err
	}
	resetTTLMin, err := getEnvInt("PASSWORD_RESET_TOKEN_TTL_MINUTES", 60)
	if err != nil {
		return Config{}, err
	}
	smtpPort, err := getEnvInt("SMTP_PORT", 1025)
	if err != nil {
		return Config{}, err
	}
	inviteTTLMHours, err := getEnvInt("INVITE_TOKEN_TTL_HOURS", 168)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnv:      getEnv("APP_ENV", "local"),
		APIPort:     getEnv("API_PORT", "8080"),
		APIBasePath: getEnv("API_BASE_PATH", "/api/v1"),
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),

		JWTAccessSecret:         strings.TrimSpace(os.Getenv("JWT_ACCESS_SECRET")),
		JWTRefreshSecret:        strings.TrimSpace(os.Getenv("JWT_REFRESH_SECRET")),
		JWTAccessTTLMin:         accessTTLMin,
		JWTRefreshTTLHours:      refreshTTLHours,
		JWTRefreshShortTTLHours: refreshShortTTLHours,

		WebBaseURL:                   strings.TrimSpace(os.Getenv("WEB_BASE_URL")),
		EmailProvider:                getEnv("EMAIL_PROVIDER", "smtp"),
		SMTPHost:                     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:                     smtpPort,
		SMTPUser:                     strings.TrimSpace(os.Getenv("SMTP_USER")),
		SMTPPass:                     strings.TrimSpace(os.Getenv("SMTP_PASS")),
		SMTPFrom:                     strings.TrimSpace(os.Getenv("SMTP_FROM")),
		PasswordResetTokenSecret:     strings.TrimSpace(os.Getenv("PASSWORD_RESET_TOKEN_SECRET")),
		PasswordResetTokenTTLMinutes: resetTTLMin,
		InviteTokenSecret:            strings.TrimSpace(os.Getenv("INVITE_TOKEN_SECRET")),
		InviteTokenTTLHours:          inviteTTLMHours,

		SchedulerOwner: strings.EqualFold(strings.TrimSpace(os.Getenv("SCHEDULER_OWNER")), "true"),

		StripeSecretKey:      strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY")),
		StripeWebhookSecret:  strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET")),
		StripePublishableKey: strings.TrimSpace(os.Getenv("STRIPE_PUBLISHABLE_KEY")),

		LogLevel:       getEnv("LOG_LEVEL", "info"),
		MetricsEnabled: resolveMetricsEnabled(os.Getenv("METRICS_ENABLED"), getEnv("APP_ENV", "local")),
	}

	if !isAllowedAppEnv(cfg.AppEnv) {
		return Config{}, fmt.Errorf("APP_ENV must be one of local, staging, prod: %q", cfg.AppEnv)
	}

	if err := validatePort(cfg.APIPort); err != nil {
		return Config{}, err
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

	if cfg.APIBasePath != "/api/v1" {
		return Config{}, fmt.Errorf("API_BASE_PATH must be /api/v1: %q", cfg.APIBasePath)
	}

	if cfg.JWTAccessTTLMin <= 0 {
		return Config{}, errors.New("JWT_ACCESS_TTL_MINUTES must be > 0")
	}

	if cfg.JWTRefreshTTLHours <= 0 {
		return Config{}, errors.New("JWT_REFRESH_TTL_HOURS must be > 0")
	}
	if cfg.JWTRefreshShortTTLHours <= 0 {
		return Config{}, errors.New("JWT_REFRESH_SHORT_TTL_HOURS must be > 0")
	}

	if cfg.WebBaseURL == "" {
		return Config{}, errors.New("WEB_BASE_URL is required")
	}
	webURL, err := url.Parse(cfg.WebBaseURL)
	if err != nil || !(webURL.Scheme == "http" || webURL.Scheme == "https") || webURL.Host == "" {
		return Config{}, fmt.Errorf("WEB_BASE_URL must be a valid absolute http or https URL: %q", cfg.WebBaseURL)
	}

	if cfg.EmailProvider != "smtp" {
		return Config{}, fmt.Errorf("EMAIL_PROVIDER must be smtp: %q", cfg.EmailProvider)
	}

	if cfg.SMTPHost == "" {
		return Config{}, errors.New("SMTP_HOST is required when EMAIL_PROVIDER is smtp")
	}
	if cfg.SMTPPort < 1 || cfg.SMTPPort > 65535 {
		return Config{}, errors.New("SMTP_PORT must be between 1 and 65535")
	}
	if cfg.SMTPFrom == "" {
		return Config{}, errors.New("SMTP_FROM is required when EMAIL_PROVIDER is smtp")
	}

	if cfg.PasswordResetTokenSecret == "" {
		return Config{}, errors.New("PASSWORD_RESET_TOKEN_SECRET is required")
	}
	if cfg.PasswordResetTokenTTLMinutes <= 0 {
		return Config{}, errors.New("PASSWORD_RESET_TOKEN_TTL_MINUTES must be > 0")
	}

	if cfg.InviteTokenSecret == "" {
		return Config{}, errors.New("INVITE_TOKEN_SECRET is required")
	}
	if cfg.InviteTokenTTLHours <= 0 {
		return Config{}, errors.New("INVITE_TOKEN_TTL_HOURS must be > 0")
	}

	if cfg.AppEnv != "local" && cfg.StripeSecretKey == "" {
		return Config{}, errors.New("STRIPE_SECRET_KEY is required when APP_ENV is not local")
	}

	if cfg.AppEnv != "local" && cfg.StripeSecretKey != "" && cfg.StripeWebhookSecret == "" {
		return Config{}, errors.New("STRIPE_WEBHOOK_SECRET is required when APP_ENV is not local")
	}

	if !isAllowedLogLevel(cfg.LogLevel) {
		return Config{}, fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error: %q", cfg.LogLevel)
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

func getEnvInt(key string, fallback int) (int, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}

	return parsed, nil
}

func isAllowedAppEnv(v string) bool {
	switch v {
	case "local", "staging", "prod":
		return true
	default:
		return false
	}
}

func validatePort(v string) error {
	port, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("API_PORT must be an integer: %q", v)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("API_PORT must be between 1 and 65535: %q", v)
	}
	return nil
}

func isAllowedLogLevel(v string) bool {
	switch v {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

func resolveMetricsEnabled(raw, appEnv string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return appEnv == "local" || appEnv == "staging"
	}
	return strings.EqualFold(raw, "true")
}

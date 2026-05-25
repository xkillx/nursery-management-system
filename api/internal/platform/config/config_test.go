package config

import (
	"strings"
	"testing"
)

func TestLoadDefaultsNonSecretAPIConfig(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("APP_ENV", "")
	t.Setenv("API_PORT", "")
	t.Setenv("API_BASE_PATH", "")
	t.Setenv("JWT_ACCESS_TTL_MINUTES", "")
	t.Setenv("JWT_REFRESH_TTL_HOURS", "")
	t.Setenv("PASSWORD_RESET_TOKEN_TTL_MINUTES", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("EMAIL_PROVIDER", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != "local" {
		t.Fatalf("expected default APP_ENV local, got %q", cfg.AppEnv)
	}
	if cfg.APIPort != "8080" {
		t.Fatalf("expected default API_PORT 8080, got %q", cfg.APIPort)
	}
	if cfg.APIBasePath != "/api/v1" {
		t.Fatalf("expected default API_BASE_PATH /api/v1, got %q", cfg.APIBasePath)
	}
	if cfg.JWTAccessTTLMin != 15 {
		t.Fatalf("expected default access TTL 15, got %d", cfg.JWTAccessTTLMin)
	}
	if cfg.JWTRefreshTTLHours != 720 {
		t.Fatalf("expected default refresh TTL 720, got %d", cfg.JWTRefreshTTLHours)
	}
	if cfg.PasswordResetTokenTTLMinutes != 60 {
		t.Fatalf("expected default reset TTL 60, got %d", cfg.PasswordResetTokenTTLMinutes)
	}
	if cfg.EmailProvider != "smtp" {
		t.Fatalf("expected default EMAIL_PROVIDER smtp, got %q", cfg.EmailProvider)
	}
	if cfg.SMTPPort != 1025 {
		t.Fatalf("expected default SMTP_PORT 1025, got %d", cfg.SMTPPort)
	}
}

func TestLoadFailsFastForMissingCriticalEnvVars(t *testing.T) {
	for _, key := range []string{
		"DATABASE_URL", "JWT_ACCESS_SECRET", "JWT_REFRESH_SECRET",
		"WEB_BASE_URL", "PASSWORD_RESET_TOKEN_SECRET",
	} {
		t.Run(key, func(t *testing.T) {
			setBaseEnv(t)
			t.Setenv(key, "")

			_, err := Load()
			if err == nil {
				t.Fatalf("expected Load() to fail for missing %s", key)
			}
			if !strings.Contains(err.Error(), key) {
				t.Fatalf("expected error to mention %s, got %v", key, err)
			}
		})
	}

	t.Run("SMTP_HOST required for smtp", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("SMTP_HOST", "")
		t.Setenv("SMTP_FROM", "")

		_, err := Load()
		if err == nil {
			t.Fatalf("expected Load() to fail for missing SMTP_HOST")
		}
		if !strings.Contains(err.Error(), "SMTP_HOST") {
			t.Fatalf("expected error to mention SMTP_HOST, got %v", err)
		}
	})

	t.Run("SMTP_FROM required for smtp", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("SMTP_FROM", "")

		_, err := Load()
		if err == nil {
			t.Fatalf("expected Load() to fail for missing SMTP_FROM")
		}
		if !strings.Contains(err.Error(), "SMTP_FROM") {
			t.Fatalf("expected error to mention SMTP_FROM, got %v", err)
		}
	})
}

func TestLoadRejectsInvalidConfigValues(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr string
	}{
		{name: "invalid app env", key: "APP_ENV", value: "dev", wantErr: "APP_ENV"},
		{name: "non numeric port", key: "API_PORT", value: "http", wantErr: "API_PORT"},
		{name: "out of range port", key: "API_PORT", value: "70000", wantErr: "API_PORT"},
		{name: "non fixed base path", key: "API_BASE_PATH", value: "/v2", wantErr: "API_BASE_PATH"},
		{name: "invalid access ttl", key: "JWT_ACCESS_TTL_MINUTES", value: "many", wantErr: "JWT_ACCESS_TTL_MINUTES"},
		{name: "non positive access ttl", key: "JWT_ACCESS_TTL_MINUTES", value: "0", wantErr: "JWT_ACCESS_TTL_MINUTES"},
		{name: "invalid refresh ttl", key: "JWT_REFRESH_TTL_HOURS", value: "many", wantErr: "JWT_REFRESH_TTL_HOURS"},
		{name: "non positive refresh ttl", key: "JWT_REFRESH_TTL_HOURS", value: "-1", wantErr: "JWT_REFRESH_TTL_HOURS"},
		{name: "invalid web base url", key: "WEB_BASE_URL", value: "not-a-url", wantErr: "WEB_BASE_URL"},
		{name: "unsupported email provider", key: "EMAIL_PROVIDER", value: "ses", wantErr: "EMAIL_PROVIDER"},
		{name: "invalid smtp port", key: "SMTP_PORT", value: "99999", wantErr: "SMTP_PORT"},
		{name: "non positive reset ttl", key: "PASSWORD_RESET_TOKEN_TTL_MINUTES", value: "0", wantErr: "PASSWORD_RESET_TOKEN_TTL_MINUTES"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setBaseEnv(t)
			t.Setenv(tc.key, tc.value)

			_, err := Load()
			if err == nil {
				t.Fatalf("expected Load() to fail")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error to mention %s, got %v", tc.wantErr, err)
			}
		})
	}
}

func setBaseEnv(t *testing.T) {
	t.Helper()

	t.Setenv("APP_ENV", "local")
	t.Setenv("API_PORT", "8080")
	t.Setenv("API_BASE_PATH", "/api/v1")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/nursery?sslmode=disable")
	t.Setenv("JWT_ACCESS_SECRET", "access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "refresh-secret")
	t.Setenv("JWT_ACCESS_TTL_MINUTES", "15")
	t.Setenv("JWT_REFRESH_TTL_HOURS", "720")
	t.Setenv("WEB_BASE_URL", "http://localhost:4200")
	t.Setenv("EMAIL_PROVIDER", "smtp")
	t.Setenv("SMTP_HOST", "localhost")
	t.Setenv("SMTP_PORT", "1025")
	t.Setenv("SMTP_FROM", "no-reply@example.local")
	t.Setenv("PASSWORD_RESET_TOKEN_SECRET", "reset-secret")
	t.Setenv("PASSWORD_RESET_TOKEN_TTL_MINUTES", "60")
}

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
	if cfg.InviteTokenTTLHours != 168 {
		t.Fatalf("expected default INVITE_TOKEN_TTL_HOURS 168, got %d", cfg.InviteTokenTTLHours)
	}
	if cfg.SchedulerOwner {
		t.Fatalf("expected default SchedulerOwner false, got true")
	}
}

func TestLoadFailsFastForMissingCriticalEnvVars(t *testing.T) {
	for _, key := range []string{
		"DATABASE_URL", "JWT_ACCESS_SECRET", "JWT_REFRESH_SECRET",
		"WEB_BASE_URL", "PASSWORD_RESET_TOKEN_SECRET", "INVITE_TOKEN_SECRET",
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
		{name: "non positive invite ttl", key: "INVITE_TOKEN_TTL_HOURS", value: "0", wantErr: "INVITE_TOKEN_TTL_HOURS"},
		{name: "invalid invite ttl", key: "INVITE_TOKEN_TTL_HOURS", value: "many", wantErr: "INVITE_TOKEN_TTL_HOURS"},
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
	t.Setenv("INVITE_TOKEN_SECRET", "invite-secret")
	t.Setenv("INVITE_TOKEN_TTL_HOURS", "168")
	t.Setenv("SCHEDULER_OWNER", "false")
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_abc")
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_abc")
	t.Setenv("STRIPE_PUBLISHABLE_KEY", "pk_test_abc")
}

func TestSchedulerOwnerConfig(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "true enables", value: "true", want: true},
		{name: "TRUE enables", value: "TRUE", want: true},
		{name: "false stays disabled", value: "false", want: false},
		{name: "empty stays disabled", value: "", want: false},
		{name: "yes stays disabled", value: "yes", want: false},
		{name: "1 stays disabled", value: "1", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setBaseEnv(t)
			t.Setenv("SCHEDULER_OWNER", tc.value)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.SchedulerOwner != tc.want {
				t.Fatalf("expected SchedulerOwner %v, got %v", tc.want, cfg.SchedulerOwner)
			}
		})
	}
}

func TestStripeConfig(t *testing.T) {
	t.Run("loads Stripe fields from env", func(t *testing.T) {
		setBaseEnv(t)
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.StripeSecretKey != "sk_test_abc" {
			t.Fatalf("expected StripeSecretKey sk_test_abc, got %q", cfg.StripeSecretKey)
		}
		if cfg.StripeWebhookSecret != "whsec_abc" {
			t.Fatalf("expected StripeWebhookSecret whsec_abc, got %q", cfg.StripeWebhookSecret)
		}
		if cfg.StripePublishableKey != "pk_test_abc" {
			t.Fatalf("expected StripePublishableKey pk_test_abc, got %q", cfg.StripePublishableKey)
		}
	})

	t.Run("staging without STRIPE_SECRET_KEY fails", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("APP_ENV", "staging")
		t.Setenv("STRIPE_SECRET_KEY", "")
		_, err := Load()
		if err == nil {
			t.Fatal("expected Load() to fail for missing STRIPE_SECRET_KEY in staging")
		}
		if !strings.Contains(err.Error(), "STRIPE_SECRET_KEY") {
			t.Fatalf("expected error to mention STRIPE_SECRET_KEY, got %v", err)
		}
	})

	t.Run("local without STRIPE_SECRET_KEY succeeds", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("STRIPE_SECRET_KEY", "")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.StripeSecretKey != "" {
			t.Fatalf("expected empty StripeSecretKey, got %q", cfg.StripeSecretKey)
		}
	})

	t.Run("staging without STRIPE_WEBHOOK_SECRET fails", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("APP_ENV", "staging")
		t.Setenv("STRIPE_WEBHOOK_SECRET", "")
		_, err := Load()
		if err == nil {
			t.Fatal("expected Load() to fail for missing STRIPE_WEBHOOK_SECRET in staging")
		}
		if !strings.Contains(err.Error(), "STRIPE_WEBHOOK_SECRET") {
			t.Fatalf("expected error to mention STRIPE_WEBHOOK_SECRET, got %v", err)
		}
	})

	t.Run("local without STRIPE_WEBHOOK_SECRET succeeds", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("STRIPE_WEBHOOK_SECRET", "")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.StripeWebhookSecret != "" {
			t.Fatalf("expected empty StripeWebhookSecret, got %q", cfg.StripeWebhookSecret)
		}
	})
}

func TestLogLevelConfig(t *testing.T) {
	t.Run("default is info", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("LOG_LEVEL", "")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.LogLevel != "info" {
			t.Fatalf("expected default LogLevel info, got %q", cfg.LogLevel)
		}
	})

	for _, level := range []string{"debug", "info", "warn", "error"} {
		t.Run("allowed level "+level, func(t *testing.T) {
			setBaseEnv(t)
			t.Setenv("LOG_LEVEL", level)
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.LogLevel != level {
				t.Fatalf("expected LogLevel %q, got %q", level, cfg.LogLevel)
			}
		})
	}

	t.Run("invalid level rejected", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("LOG_LEVEL", "verbose")
		_, err := Load()
		if err == nil {
			t.Fatal("expected Load() to fail for invalid LOG_LEVEL")
		}
		if !strings.Contains(err.Error(), "LOG_LEVEL") {
			t.Fatalf("expected error to mention LOG_LEVEL, got %v", err)
		}
	})
}

func TestRunMigrationsConfig(t *testing.T) {
	t.Run("default is true", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("RUN_MIGRATIONS", "")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if !cfg.RunMigrations {
			t.Fatal("expected RunMigrations true by default")
		}
	})

	t.Run("true enables", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("RUN_MIGRATIONS", "true")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if !cfg.RunMigrations {
			t.Fatal("expected RunMigrations true")
		}
	})

	t.Run("TRUE enables", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("RUN_MIGRATIONS", "TRUE")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if !cfg.RunMigrations {
			t.Fatal("expected RunMigrations true for TRUE")
		}
	})

	t.Run("1 enables", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("RUN_MIGRATIONS", "1")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if !cfg.RunMigrations {
			t.Fatal("expected RunMigrations true for 1")
		}
	})

	t.Run("false disables", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("RUN_MIGRATIONS", "false")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.RunMigrations {
			t.Fatal("expected RunMigrations false")
		}
	})

	t.Run("FALSE disables", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("RUN_MIGRATIONS", "FALSE")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.RunMigrations {
			t.Fatal("expected RunMigrations false for FALSE")
		}
	})
}

func TestMetricsEnabledConfig(t *testing.T) {
	t.Run("default local is true", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("METRICS_ENABLED", "")
		t.Setenv("APP_ENV", "local")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if !cfg.MetricsEnabled {
			t.Fatal("expected MetricsEnabled true for local")
		}
	})

	t.Run("default staging is true", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("METRICS_ENABLED", "")
		t.Setenv("APP_ENV", "staging")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if !cfg.MetricsEnabled {
			t.Fatal("expected MetricsEnabled true for staging")
		}
	})

	t.Run("default prod is false", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("METRICS_ENABLED", "")
		t.Setenv("APP_ENV", "prod")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.MetricsEnabled {
			t.Fatal("expected MetricsEnabled false for prod")
		}
	})

	t.Run("explicit true overrides prod default", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("METRICS_ENABLED", "true")
		t.Setenv("APP_ENV", "prod")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if !cfg.MetricsEnabled {
			t.Fatal("expected MetricsEnabled true when explicitly set")
		}
	})

	t.Run("explicit false overrides local default", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("METRICS_ENABLED", "false")
		t.Setenv("APP_ENV", "local")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.MetricsEnabled {
			t.Fatal("expected MetricsEnabled false when explicitly set")
		}
	})

	t.Run("non-boolean value treated as false", func(t *testing.T) {
		setBaseEnv(t)
		t.Setenv("METRICS_ENABLED", "yes")
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.MetricsEnabled {
			t.Fatal("expected MetricsEnabled false for non-boolean value")
		}
	})
}

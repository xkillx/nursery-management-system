package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"nursery-management-system/api/internal/platform/db"
)

type options struct {
	DatabaseURL string
	TenantName  string
	BranchName  string
	Email       string
	Password    string
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	opt := parseFlags()
	if err := validateOptions(opt); err != nil {
		logger.Error("invalid options", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPostgres(ctx, opt.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(opt.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to hash password", "error", err)
		os.Exit(1)
	}

	emailNormalized := normalizeEmail(opt.Email)

	seedCtx, seedCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer seedCancel()

	result, err := seedManager(seedCtx, pool, seedPayload{
		TenantName:      opt.TenantName,
		BranchName:      opt.BranchName,
		Email:           opt.Email,
		EmailNormalized: emailNormalized,
		PasswordHash:    string(passwordHash),
	})
	if err != nil {
		logger.Error("seed failed", "error", err)
		os.Exit(1)
	}

	logger.Info("seed complete",
		"tenant_id", result.TenantID,
		"branch_id", result.BranchID,
		"user_id", result.UserID,
		"email", opt.Email,
	)
}

func parseFlags() options {
	opt := options{}
	flag.StringVar(&opt.DatabaseURL, "database-url", strings.TrimSpace(os.Getenv("DATABASE_URL")), "Postgres DSN (or set DATABASE_URL)")
	flag.StringVar(&opt.TenantName, "tenant", "Pilot Nursery", "Tenant name")
	flag.StringVar(&opt.BranchName, "branch", "Main", "Branch name")
	flag.StringVar(&opt.Email, "email", "", "Manager email address")
	flag.StringVar(&opt.Password, "password", "", "Manager password (plain text)")
	flag.Parse()
	return opt
}

func validateOptions(opt options) error {
	if opt.DatabaseURL == "" {
		return errors.New("database-url is required")
	}
	if strings.TrimSpace(opt.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(opt.Password) == "" {
		return errors.New("password is required")
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

type seedPayload struct {
	TenantName      string
	BranchName      string
	Email           string
	EmailNormalized string
	PasswordHash    string
}

type seedResult struct {
	TenantID string
	BranchID string
	UserID   string
}

func seedManager(ctx context.Context, pool *pgxpool.Pool, payload seedPayload) (seedResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return seedResult{}, err
	}
	defer tx.Rollback(ctx)

	tenantID, err := findOrCreateTenant(ctx, tx, payload.TenantName)
	if err != nil {
		return seedResult{}, err
	}

	branchID, err := findOrCreateBranch(ctx, tx, tenantID, payload.BranchName)
	if err != nil {
		return seedResult{}, err
	}

	userID, err := upsertUser(ctx, tx, payload.Email, payload.EmailNormalized, payload.PasswordHash)
	if err != nil {
		return seedResult{}, err
	}

	if err := ensureMembership(ctx, tx, tenantID, branchID, userID); err != nil {
		return seedResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return seedResult{}, err
	}

	return seedResult{
		TenantID: tenantID.String(),
		BranchID: branchID.String(),
		UserID:   userID.String(),
	}, nil
}

func newUUID() uuid.UUID {
	if id, err := uuid.NewV7(); err == nil {
		return id
	}
	return uuid.New()
}

func findOrCreateTenant(ctx context.Context, tx pgx.Tx, name string) (uuid.UUID, error) {
	const selectQ = `SELECT id FROM tenants WHERE name = $1 LIMIT 1`
	var id uuid.UUID
	err := tx.QueryRow(ctx, selectQ, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.UUID{}, err
	}

	id = newUUID()
	const insertQ = `INSERT INTO tenants (id, name) VALUES ($1, $2)`
	if _, err := tx.Exec(ctx, insertQ, id, name); err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}

func findOrCreateBranch(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, name string) (uuid.UUID, error) {
	const q = `
INSERT INTO branches (id, tenant_id, name)
VALUES ($1, $2, $3)
ON CONFLICT (tenant_id, name)
DO UPDATE SET updated_at = now()
RETURNING id`

	id := newUUID()
	if err := tx.QueryRow(ctx, q, id, tenantID, name).Scan(&id); err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}

func upsertUser(ctx context.Context, tx pgx.Tx, email, emailNormalized, passwordHash string) (uuid.UUID, error) {
	const q = `
INSERT INTO users (id, email, email_normalized, password_hash, is_active)
VALUES ($1, $2, $3, $4, true)
ON CONFLICT (email_normalized)
DO UPDATE SET email = EXCLUDED.email,
              password_hash = EXCLUDED.password_hash,
              is_active = true,
              updated_at = now()
RETURNING id`

	id := newUUID()
	if err := tx.QueryRow(ctx, q, id, email, emailNormalized, passwordHash).Scan(&id); err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}

func ensureMembership(ctx context.Context, tx pgx.Tx, tenantID, branchID, userID uuid.UUID) error {
	const q = `
INSERT INTO memberships (id, tenant_id, branch_id, user_id, role, is_active, ended_at)
VALUES ($1, $2, $3, $4, 'manager', true, NULL)
ON CONFLICT (tenant_id, branch_id, user_id)
DO UPDATE SET role = 'manager', is_active = true, ended_at = NULL, updated_at = now()`

	membershipID := newUUID()
	_, err := tx.Exec(ctx, q, membershipID, tenantID, branchID, userID)
	return err
}

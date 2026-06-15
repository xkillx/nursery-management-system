package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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
	DatabaseURL  string
	TenantName   string
	BranchName   string
	Email        string
	Password     string
	Local        bool
	ManagerEmail string
	StaffEmail   string
	ParentEmail  string
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	opt := parseFlags()
	if err := validateOptions(opt); err != nil {
		logger.Error("invalid options", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	tenantID, branchID, err := ensureTenantAndBranch(ctx, pool, opt.TenantName, opt.BranchName)
	if err != nil {
		logger.Error("tenant/branch setup failed", "error", err)
		os.Exit(1)
	}

	type seedAccount struct {
		Role      string
		Email     string
		BranchID  *uuid.UUID
		PrintRole string
	}

	var accounts []seedAccount
	accounts = append(accounts, seedAccount{"owner", opt.Email, nil, "Owner"})

	if opt.Local {
		accounts = append(accounts,
			seedAccount{"manager", opt.ManagerEmail, &branchID, "Manager"},
			seedAccount{"practitioner", opt.StaffEmail, &branchID, "Staff"},
			seedAccount{"parent", opt.ParentEmail, &branchID, "Parent"},
		)
	}

	for _, acc := range accounts {
		en := normalizeEmail(acc.Email)
		uid, err := upsertUser(ctx, pool, acc.Email, en, string(passwordHash))
		if err != nil {
			logger.Error("upsert user failed", "role", acc.Role, "error", err)
			os.Exit(1)
		}
		if err := ensureMembershipRole(ctx, pool, tenantID, acc.BranchID, uid, acc.Role); err != nil {
			logger.Error("membership failed", "role", acc.Role, "error", err)
			os.Exit(1)
		}
		logger.Info("seeded", "role", acc.PrintRole, "email", acc.Email, "user_id", uid.String())
	}

	fmt.Println()
	if opt.Local {
		fmt.Println("=== Seed Accounts (local) ===")
		fmt.Printf("  Owner:    %s\n", accounts[0].Email)
		fmt.Printf("  Manager:  %s\n", accounts[1].Email)
		fmt.Printf("  Staff:    %s\n", accounts[2].Email)
		fmt.Printf("  Parent:   %s\n", accounts[3].Email)
		fmt.Printf("  Password: %s (all accounts)\n", opt.Password)
	} else {
		fmt.Println("=== Seed Account ===")
		fmt.Printf("  Owner:    %s\n", accounts[0].Email)
		fmt.Printf("  Password: %s\n", opt.Password)
	}
	fmt.Println()
	fmt.Printf("  tenant_id: %s\n", tenantID.String())
	fmt.Printf("  branch_id: %s\n", branchID.String())
}

func parseFlags() options {
	opt := options{}
	flag.StringVar(&opt.DatabaseURL, "database-url", strings.TrimSpace(os.Getenv("DATABASE_URL")), "Postgres DSN (or set DATABASE_URL)")
	flag.StringVar(&opt.TenantName, "tenant", "Pilot Nursery", "Tenant name")
	flag.StringVar(&opt.BranchName, "branch", "Main", "Branch name")
	flag.StringVar(&opt.Email, "email", "", "Owner email address")
	flag.StringVar(&opt.Password, "password", "", "Password (plain text)")
	flag.BoolVar(&opt.Local, "local", false, "Also seed manager, staff, and parent accounts for local testing")
	flag.StringVar(&opt.ManagerEmail, "manager-email", "", "Manager email (required with -local)")
	flag.StringVar(&opt.StaffEmail, "staff-email", "", "Staff email (required with -local)")
	flag.StringVar(&opt.ParentEmail, "parent-email", "", "Parent email (required with -local)")
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
	if opt.Local {
		if strings.TrimSpace(opt.ManagerEmail) == "" {
			return errors.New("manager-email is required with -local")
		}
		if strings.TrimSpace(opt.StaffEmail) == "" {
			return errors.New("staff-email is required with -local")
		}
		if strings.TrimSpace(opt.ParentEmail) == "" {
			return errors.New("parent-email is required with -local")
		}
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func ensureTenantAndBranch(ctx context.Context, pool *pgxpool.Pool, tenantName, branchName string) (uuid.UUID, uuid.UUID, error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}
	defer tx.Rollback(ctx)

	tenantID, err := findOrCreateTenant(ctx, tx, tenantName)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}

	branchID, err := findOrCreateBranch(ctx, tx, tenantID, branchName)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}

	return tenantID, branchID, nil
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

func upsertUser(ctx context.Context, pool *pgxpool.Pool, email, emailNormalized, passwordHash string) (uuid.UUID, error) {
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
	if err := pool.QueryRow(ctx, q, id, email, emailNormalized, passwordHash).Scan(&id); err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}

func ensureMembershipRole(ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID, branchID *uuid.UUID, userID uuid.UUID, role string) error {
	if branchID == nil {
		tag, err := pool.Exec(ctx, `
UPDATE memberships SET role = $1, is_active = true, ended_at = NULL, updated_at = now()
WHERE tenant_id = $2 AND user_id = $3 AND role = 'owner'`,
			role, tenantID, userID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() > 0 {
			return nil
		}
		membershipID := newUUID()
		_, err = pool.Exec(ctx, `
INSERT INTO memberships (id, tenant_id, branch_id, user_id, role, is_active, ended_at)
VALUES ($1, $2, NULL, $3, $4, true, NULL)`,
			membershipID, tenantID, userID, role)
		return err
	}

	tag, err := pool.Exec(ctx, `
UPDATE memberships SET role = $1, is_active = true, ended_at = NULL, updated_at = now()
WHERE tenant_id = $2 AND branch_id = $3 AND user_id = $4`,
		role, tenantID, *branchID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	membershipID := newUUID()
	_, err = pool.Exec(ctx, `
INSERT INTO memberships (id, tenant_id, branch_id, user_id, role, is_active, ended_at)
VALUES ($1, $2, $3, $4, $5, true, NULL)`,
		membershipID, tenantID, *branchID, userID, role)
	return err
}

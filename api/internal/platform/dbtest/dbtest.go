package dbtest

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var blockedDBNames = []string{
	"nursery_management", "postgres", "production", "prod", "staging", "dev",
}

func RequirePostgres(t testing.TB) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping repository integration test")
	}

	if err := validateURL(databaseURL); err != nil {
		t.Fatalf("TEST_DATABASE_URL validation failed: %s", err)
	}

	schemaName := fmt.Sprintf("repo_test_%s_%d", sanitizePackageName(t.Name()), randomSuffix())

	repoRoot := findRepoRoot(t)
	migrationsDir := filepath.Join(repoRoot, "api", "db", "migrations")

	createSchema(t, databaseURL, schemaName)
	migrateURL := buildMigrateURL(t, databaseURL, schemaName)

	migratePath, err := exec.LookPath("migrate")
	if err != nil {
		t.Fatal("migrate binary not found in PATH; install golang-migrate CLI")
	}

	cmd := exec.Command(migratePath, "-path", migrationsDir, "-database", migrateURL, "up")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("migrate up failed for schema %s: %v", schemaName, err)
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse pool config: %v", err)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName
	poolConfig.MaxConns = 5
	poolConfig.MinConns = 1

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		t.Fatalf("ping pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		dropSchema(t, databaseURL, schemaName)
	})

	return pool
}

func validateURL(databaseURL string) error {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}

	dbName := strings.TrimPrefix(u.Path, "/")
	lower := strings.ToLower(dbName)

	allowed := strings.Contains(lower, "test") || strings.Contains(lower, "repository")
	if !allowed {
		return fmt.Errorf("database name %q must contain 'test' or 'repository'", dbName)
	}

	for _, blocked := range blockedDBNames {
		if lower == blocked {
			return fmt.Errorf("database name %q is blocked", dbName)
		}
	}

	return nil
}

func sanitizePackageName(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

func randomSuffix() int64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return n.Int64() + 100000
}

func findRepoRoot(t testing.TB) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root")
		}
		dir = parent
	}
}

func buildMigrateURL(t testing.TB, databaseURL, schemaName string) string {
	t.Helper()
	u, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse database URL for migration: %v", err)
	}
	q := u.Query()
	q.Set("search_path", schemaName)
	q.Set("x-migrations-table", "schema_migrations")
	u.RawQuery = q.Encode()
	return u.String()
}

func dropSchema(t testing.TB, databaseURL, schemaName string) {
	t.Helper()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Logf("drop schema: parse config: %v", err)
		return
	}
	config.MaxConns = 1
	config.MinConns = 0

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Logf("drop schema: create pool: %v", err)
		return
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", pgx.Identifier{schemaName}.Sanitize()))
	if err != nil {
		t.Logf("drop schema %s: %v", schemaName, err)
	}
}

func createSchema(t testing.TB, databaseURL, schemaName string) {
	t.Helper()
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("create schema: parse config: %v", err)
	}
	config.MaxConns = 1
	config.MinConns = 0

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("create schema: create pool: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, fmt.Sprintf("CREATE SCHEMA %s", pgx.Identifier{schemaName}.Sanitize()))
	if err != nil {
		t.Fatalf("create schema %s: %v", schemaName, err)
	}

	// Pre-create enum types that migrations reference with schema-blind IF NOT EXISTS checks.
	// Without this, concurrent test schemas trigger false "already exists" in pg_type lookup.
	enumTypes := []string{
		fmt.Sprintf(
			"CREATE TYPE %s.lifecycle_reason_code AS ENUM ('duplicate_record','entered_in_error','left_nursery','safeguarding_direction','contact_update','access_revoked','other')",
			pgx.Identifier{schemaName}.Sanitize(),
		),
	}
	for _, ddl := range enumTypes {
		_, err = pool.Exec(ctx, ddl)
		if err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				t.Fatalf("pre-create enum in schema %s: %v", schemaName, err)
			}
		}
	}
}

// Reset truncates all application tables in dependency-safe order within the test schema.
func Reset(t testing.TB, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	tables := []string{
		"payment_attempts",
		"invoice_lines",
		"invoices",
		"invoice_runs",
		"invoice_number_sequences",
		"funding_profiles",
		"absence_markers",
		"attendance_events",
		"attendance_sessions",
		"manager_invites",
		"parent_membership_guardians",
		"guardian_child_links",
		"guardians",
		"children",
		"refresh_tokens",
		"audit_logs",
		"password_reset_tokens",
		"memberships",
		"users",
		"branches",
		"tenants",
	}

	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}
}

// BeginTx starts a transaction and registers rollback cleanup.
func BeginTx(t testing.TB, pool *pgxpool.Pool) pgx.Tx {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})
	return tx
}

// CommitTx commits a transaction and fails the test on error.
func CommitTx(t testing.TB, tx pgx.Tx) {
	t.Helper()
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatalf("commit tx: %v", err)
	}
}

// InsertTenant inserts a tenant and returns its ID.
func InsertTenant(t testing.TB, pool *pgxpool.Pool, id uuid.UUID, name string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO tenants (id, name) VALUES ($1, $2)", id, name)
	if err != nil {
		t.Fatalf("insert tenant: %v", err)
	}
}

// InsertBranch inserts a branch scoped to a tenant.
func InsertBranch(t testing.TB, pool *pgxpool.Pool, tenantID, branchID uuid.UUID, name string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO branches (id, tenant_id, name) VALUES ($1, $2, $3)", branchID, tenantID, name)
	if err != nil {
		t.Fatalf("insert branch: %v", err)
	}
}

// InsertUser inserts a user and returns their ID.
func InsertUser(t testing.TB, pool *pgxpool.Pool, id uuid.UUID, email, passwordHash string, isActive bool) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO users (id, email, email_normalized, password_hash, is_active) VALUES ($1, $2, lower($3), $4, $5)",
		id, email, email, passwordHash, isActive)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
}

// InsertMembership inserts a membership scoped to tenant+branch.
func InsertMembership(t testing.TB, pool *pgxpool.Pool, id, tenantID, branchID, userID uuid.UUID, role string, isActive bool) {
	t.Helper()
	var endedAt *time.Time
	if !isActive {
		now := time.Now().UTC()
		endedAt = &now
	}
	_, err := pool.Exec(context.Background(),
		"INSERT INTO memberships (id, tenant_id, branch_id, user_id, role, is_active, ended_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		id, tenantID, branchID, userID, role, isActive, endedAt)
	if err != nil {
		t.Fatalf("insert membership: %v", err)
	}
}

// InsertChild inserts a child scoped to tenant+branch.
func InsertChild(t testing.TB, pool *pgxpool.Pool, id, tenantID, branchID uuid.UUID, fullName string, dob, startDate time.Time, hourlyRate int, isActive bool) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO children (id, tenant_id, branch_id, full_name, date_of_birth, start_date, core_hourly_rate_minor, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		id, tenantID, branchID, fullName, dob, startDate, hourlyRate, isActive)
	if err != nil {
		t.Fatalf("insert child: %v", err)
	}
}

// InsertChildWithNotes inserts a child with notes.
func InsertChildWithNotes(t testing.TB, pool *pgxpool.Pool, id, tenantID, branchID uuid.UUID, fullName string, dob, startDate time.Time, hourlyRate int, isActive bool, notes *string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO children (id, tenant_id, branch_id, full_name, date_of_birth, start_date, core_hourly_rate_minor, is_active, notes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		id, tenantID, branchID, fullName, dob, startDate, hourlyRate, isActive, notes)
	if err != nil {
		t.Fatalf("insert child with notes: %v", err)
	}
}

// InsertGuardian inserts a guardian scoped to tenant+branch.
func InsertGuardian(t testing.TB, pool *pgxpool.Pool, id, tenantID, branchID uuid.UUID, fullName string, isActive bool) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO guardians (id, tenant_id, branch_id, full_name, is_active) VALUES ($1, $2, $3, $4, $5)",
		id, tenantID, branchID, fullName, isActive)
	if err != nil {
		t.Fatalf("insert guardian: %v", err)
	}
}

// InsertGuardianLink inserts a guardian-child link.
func InsertGuardianLink(t testing.TB, pool *pgxpool.Pool, id, tenantID, branchID, guardianID, childID uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO guardian_child_links (id, tenant_id, branch_id, guardian_id, child_id) VALUES ($1, $2, $3, $4, $5)",
		id, tenantID, branchID, guardianID, childID)
	if err != nil {
		t.Fatalf("insert guardian link: %v", err)
	}
}

// InsertParentMapping inserts a parent membership guardian mapping.
func InsertParentMapping(t testing.TB, pool *pgxpool.Pool, id, tenantID, branchID, membershipID, guardianID uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO parent_membership_guardians (id, tenant_id, branch_id, membership_id, guardian_id) VALUES ($1, $2, $3, $4, $5)",
		id, tenantID, branchID, membershipID, guardianID)
	if err != nil {
		t.Fatalf("insert parent mapping: %v", err)
	}
}

// InsertInvite inserts a manager invite.
func InsertInvite(t testing.TB, pool *pgxpool.Pool, id, tenantID, branchID uuid.UUID, email, emailNorm, role, tokenHash string, expiresAt time.Time, createdByUserID, createdByMembershipID uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO manager_invites (id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at, created_by_user_id, created_by_membership_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		id, tenantID, branchID, email, emailNorm, role, tokenHash, expiresAt, createdByUserID, createdByMembershipID)
	if err != nil {
		t.Fatalf("insert invite: %v", err)
	}
}

// AcceptInvite marks an invite as accepted. Creates the user and membership first.
func AcceptInvite(t testing.TB, pool *pgxpool.Pool, inviteID, userID, membershipID, tenantID, branchID uuid.UUID, email string) {
	t.Helper()
	InsertUser(t, pool, userID, email, "hash", true)
	InsertMembership(t, pool, membershipID, tenantID, branchID, userID, "practitioner", true)
	_, err := pool.Exec(context.Background(),
		"UPDATE manager_invites SET accepted_at = now(), accepted_user_id = $2, accepted_membership_id = $3 WHERE id = $1",
		inviteID, userID, membershipID)
	if err != nil {
		t.Fatalf("accept invite: %v", err)
	}
}

// RevokeInvite marks an invite as revoked.
func RevokeInviteDB(t testing.TB, pool *pgxpool.Pool, inviteID, revokedByUserID, revokedByMembershipID uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"UPDATE manager_invites SET revoked_at = now(), revoked_by_user_id = $2, revoked_by_membership_id = $3 WHERE id = $1",
		inviteID, revokedByUserID, revokedByMembershipID)
	if err != nil {
		t.Fatalf("revoke invite: %v", err)
	}
}

// InsertRefreshToken inserts a refresh token.
func InsertRefreshToken(t testing.TB, pool *pgxpool.Pool, id, userID, membershipID uuid.UUID, tokenHash string, expiresAt time.Time, userAgent, ipAddress *string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO refresh_tokens (id, user_id, membership_id, token_hash, expires_at, user_agent, ip_address) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		id, userID, membershipID, tokenHash, expiresAt, userAgent, ipAddress)
	if err != nil {
		t.Fatalf("insert refresh token: %v", err)
	}
}

// RevokeRefreshToken sets revoked_at on the given token.
func RevokeRefreshToken(t testing.TB, pool *pgxpool.Pool, tokenID uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1", tokenID)
	if err != nil {
		t.Fatalf("revoke refresh token: %v", err)
	}
}

// InsertAttendanceSession inserts an attendance session.
func InsertAttendanceSession(t testing.TB, pool *pgxpool.Pool, id, tenantID, branchID, childID uuid.UUID, status string, checkInAt time.Time, checkInLocalDate time.Time) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_in_local_date) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		id, tenantID, branchID, childID, status, checkInAt, checkInLocalDate)
	if err != nil {
		t.Fatalf("insert attendance session: %v", err)
	}
}

// CompleteAttendanceSession marks an attendance session complete.
func CompleteAttendanceSession(t testing.TB, pool *pgxpool.Pool, tenantID, branchID, sessionID uuid.UUID, checkOutAt time.Time, checkOutLocalDate time.Time) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"UPDATE attendance_sessions SET status = 'complete', check_out_at = $1, check_out_local_date = $2 WHERE tenant_id = $3 AND branch_id = $4 AND id = $5",
		checkOutAt, checkOutLocalDate, tenantID, branchID, sessionID)
	if err != nil {
		t.Fatalf("complete attendance session: %v", err)
	}
}

// InsertAbsenceMarker inserts an absence marker. Pass nil for clearedAt/clearedBy fields to create an active marker.
func InsertAbsenceMarker(t testing.TB, pool *pgxpool.Pool, id, tenantID, branchID, childID, markedByUserID, markedByMembershipID uuid.UUID, localDate, markedAt time.Time, clearedAt *time.Time, clearedByUserID, clearedByMembershipID *uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO absence_markers (id, tenant_id, branch_id, child_id, local_date, marked_at, marked_by_user_id, marked_by_membership_id, cleared_at, cleared_by_user_id, cleared_by_membership_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		id, tenantID, branchID, childID, localDate, markedAt, markedByUserID, markedByMembershipID, clearedAt, clearedByUserID, clearedByMembershipID)
	if err != nil {
		t.Fatalf("insert absence marker: %v", err)
	}
}

// StrPtr returns a pointer to the given string.
func StrPtr(s string) *string { return &s }

// TimePtr returns a pointer to the given time.
func TimePtr(t time.Time) *time.Time { return &t }

// DateAt returns a time.Time set to midnight UTC for the given date components.
func DateAt(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// TimestampAt returns a UTC time for the given components.
func TimestampAt(year, month, day, hour, min int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, 0, 0, time.UTC)
}

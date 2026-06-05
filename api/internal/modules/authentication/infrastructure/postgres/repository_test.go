package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	authpostgres "nursery-management-system/api/internal/modules/authentication/infrastructure/postgres"
	"nursery-management-system/api/internal/modules/authentication/domain"
	"nursery-management-system/api/internal/platform/dbtest"
)

func TestFindUserByEmail_NoRows(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	repo := authpostgres.NewRepository(pool)

	_, err := repo.FindUserByEmail(context.Background(), "nobody@example.com")
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestFindUserByEmail_ActiveUser(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	userID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	dbtest.InsertUser(t, pool, userID, "alice@example.com", "hash123", true)

	repo := authpostgres.NewRepository(pool)
	user, err := repo.FindUserByEmail(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != userID {
		t.Errorf("ID = %s, want %s", user.ID, userID)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("Email = %s, want alice@example.com", user.Email)
	}
	if user.PasswordHash != "hash123" {
		t.Errorf("PasswordHash = %s, want hash123", user.PasswordHash)
	}
	if !user.IsActive {
		t.Error("IsActive = false, want true")
	}
}

func TestListMembershipsByUserID_ActiveOnly(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	branchID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	branchID2 := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000003")
	userID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	activeMID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	endedMID := uuid.MustParse("cccccccc-0000-0000-0000-000000000002")

	dbtest.InsertTenant(t, pool, tenantID, "Test Tenant")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Test Branch")
	dbtest.InsertBranch(t, pool, tenantID, branchID2, "Test Branch 2")
	dbtest.InsertUser(t, pool, userID, "user@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, activeMID, tenantID, branchID, userID, "manager", true)
	dbtest.InsertMembership(t, pool, endedMID, tenantID, branchID2, userID, "practitioner", false)

	repo := authpostgres.NewRepository(pool)
	memberships, err := repo.ListMembershipsByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(memberships) != 1 {
		t.Fatalf("expected 1 membership, got %d", len(memberships))
	}
	if memberships[0].ID != activeMID {
		t.Errorf("ID = %s, want %s", memberships[0].ID, activeMID)
	}
	if memberships[0].Role != "manager" {
		t.Errorf("Role = %s, want manager", memberships[0].Role)
	}
	if !memberships[0].IsActive {
		t.Error("IsActive = false, want true")
	}
}

func TestCreateRefreshToken(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	branchID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	branchID2 := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000003")
	userID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	membershipID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	tokenID := uuid.MustParse("dddddddd-0000-0000-0000-000000000001")

	dbtest.InsertTenant(t, pool, tenantID, "Test Tenant")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Test Branch")
	dbtest.InsertBranch(t, pool, tenantID, branchID2, "Test Branch 2")
	dbtest.InsertUser(t, pool, userID, "user@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, membershipID, tenantID, branchID, userID, "manager", true)

	expiresAt := time.Now().Add(7 * 24 * time.Hour).UTC()
	token := domain.RefreshToken{
		ID:           tokenID,
		UserID:       userID,
		MembershipID: membershipID,
		TokenHash:    "abc123hash",
		ExpiresAt:    expiresAt,
	}

	repo := authpostgres.NewRepository(pool)
	if err := repo.CreateRefreshToken(ctx, token, "TestAgent/1.0", "192.168.1.1"); err != nil {
		t.Fatalf("CreateRefreshToken: %v", err)
	}

	var userAgent, ipAddress string
	err := pool.QueryRow(ctx,
		"SELECT user_agent, ip_address FROM refresh_tokens WHERE id = $1", tokenID).Scan(&userAgent, &ipAddress)
	if err != nil {
		t.Fatalf("query refresh token: %v", err)
	}
	if userAgent != "TestAgent/1.0" {
		t.Errorf("user_agent = %s, want TestAgent/1.0", userAgent)
	}
	if ipAddress != "192.168.1.1" {
		t.Errorf("ip_address = %s, want 192.168.1.1", ipAddress)
	}
}

func TestFindActiveRefreshToken_Success(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	branchID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	userID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	membershipID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	tokenID := uuid.MustParse("dddddddd-0000-0000-0000-000000000001")

	dbtest.InsertTenant(t, pool, tenantID, "Test Tenant")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Test Branch")
	dbtest.InsertUser(t, pool, userID, "user@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, membershipID, tenantID, branchID, userID, "manager", true)
	dbtest.InsertRefreshToken(t, pool, tokenID, userID, membershipID, "hashABC",
		time.Now().Add(24*time.Hour).UTC(), dbtest.StrPtr("Agent/1.0"), dbtest.StrPtr("10.0.0.1"))

	repo := authpostgres.NewRepository(pool)
	token, user, membership, err := repo.FindActiveRefreshToken(ctx, "hashABC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.ID != tokenID {
		t.Errorf("token ID = %s, want %s", token.ID, tokenID)
	}
	if user.ID != userID {
		t.Errorf("user ID = %s, want %s", user.ID, userID)
	}
	if membership.ID != membershipID {
		t.Errorf("membership ID = %s, want %s", membership.ID, membershipID)
	}
	if !membership.IsActive {
		t.Error("membership IsActive = false, want true")
	}
}

func TestFindActiveRefreshToken_RevokedToken(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	branchID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	userID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	membershipID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	tokenID := uuid.MustParse("dddddddd-0000-0000-0000-000000000001")

	dbtest.InsertTenant(t, pool, tenantID, "Test Tenant")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Test Branch")
	dbtest.InsertUser(t, pool, userID, "user@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, membershipID, tenantID, branchID, userID, "manager", true)
	dbtest.InsertRefreshToken(t, pool, tokenID, userID, membershipID, "hashREV",
		time.Now().Add(24*time.Hour).UTC(), nil, nil)
	dbtest.RevokeRefreshToken(t, pool, tokenID)

	repo := authpostgres.NewRepository(pool)
	_, _, _, err := repo.FindActiveRefreshToken(ctx, "hashREV")
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound for revoked token, got %v", err)
	}
}

func TestFindActiveRefreshToken_ExpiredToken(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	branchID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	userID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	membershipID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	tokenID := uuid.MustParse("dddddddd-0000-0000-0000-000000000001")

	dbtest.InsertTenant(t, pool, tenantID, "Test Tenant")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Test Branch")
	dbtest.InsertUser(t, pool, userID, "user@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, membershipID, tenantID, branchID, userID, "manager", true)
	dbtest.InsertRefreshToken(t, pool, tokenID, userID, membershipID, "hashEXP",
		time.Now().Add(-1*time.Hour).UTC(), nil, nil)

	repo := authpostgres.NewRepository(pool)
	_, _, _, err := repo.FindActiveRefreshToken(ctx, "hashEXP")
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound for expired token, got %v", err)
	}
}

func TestFindActiveRefreshToken_InactiveUser(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	branchID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	userID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	membershipID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	tokenID := uuid.MustParse("dddddddd-0000-0000-0000-000000000001")

	dbtest.InsertTenant(t, pool, tenantID, "Test Tenant")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Test Branch")
	dbtest.InsertUser(t, pool, userID, "user@example.com", "hash", false)
	dbtest.InsertMembership(t, pool, membershipID, tenantID, branchID, userID, "manager", true)
	dbtest.InsertRefreshToken(t, pool, tokenID, userID, membershipID, "hashINA",
		time.Now().Add(24*time.Hour).UTC(), nil, nil)

	repo := authpostgres.NewRepository(pool)
	_, _, _, err := repo.FindActiveRefreshToken(ctx, "hashINA")
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound for inactive user, got %v", err)
	}
}

func TestRotateRefreshToken(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	branchID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	userID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	membershipID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	oldTokenID := uuid.MustParse("dddddddd-0000-0000-0000-000000000001")
	newTokenID := uuid.MustParse("dddddddd-0000-0000-0000-000000000002")

	dbtest.InsertTenant(t, pool, tenantID, "Test Tenant")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Test Branch")
	dbtest.InsertUser(t, pool, userID, "user@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, membershipID, tenantID, branchID, userID, "manager", true)
	dbtest.InsertRefreshToken(t, pool, oldTokenID, userID, membershipID, "oldHash",
		time.Now().Add(24*time.Hour).UTC(), nil, nil)

	repo := authpostgres.NewRepository(pool)
	replacement := domain.RefreshToken{
		ID:           newTokenID,
		UserID:       userID,
		MembershipID: membershipID,
		TokenHash:    "newHash",
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour).UTC(),
	}
	if err := repo.RotateRefreshToken(ctx, oldTokenID, replacement, "Agent/2.0", "10.0.0.2"); err != nil {
		t.Fatalf("RotateRefreshToken: %v", err)
	}

	var revokedAt *time.Time
	err := pool.QueryRow(ctx,
		"SELECT revoked_at FROM refresh_tokens WHERE id = $1", oldTokenID).Scan(&revokedAt)
	if err != nil {
		t.Fatalf("query old token: %v", err)
	}
	if revokedAt == nil {
		t.Error("old token revoked_at = nil, want non-nil")
	}

	var newUA string
	err = pool.QueryRow(ctx,
		"SELECT user_agent FROM refresh_tokens WHERE id = $1", newTokenID).Scan(&newUA)
	if err != nil {
		t.Fatalf("query new token: %v", err)
	}
	if newUA != "Agent/2.0" {
		t.Errorf("new token user_agent = %s, want Agent/2.0", newUA)
	}
}

func TestRevokeByTokenHash(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	branchID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	userID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	membershipID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	tokenID := uuid.MustParse("dddddddd-0000-0000-0000-000000000001")

	dbtest.InsertTenant(t, pool, tenantID, "Test Tenant")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Test Branch")
	dbtest.InsertUser(t, pool, userID, "user@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, membershipID, tenantID, branchID, userID, "manager", true)
	dbtest.InsertRefreshToken(t, pool, tokenID, userID, membershipID, "revokeMe",
		time.Now().Add(24*time.Hour).UTC(), nil, nil)

	repo := authpostgres.NewRepository(pool)
	if err := repo.RevokeByTokenHash(ctx, "revokeMe"); err != nil {
		t.Fatalf("RevokeByTokenHash: %v", err)
	}

	var revokedAt *time.Time
	err := pool.QueryRow(ctx,
		"SELECT revoked_at FROM refresh_tokens WHERE id = $1", tokenID).Scan(&revokedAt)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if revokedAt == nil {
		t.Error("revoked_at = nil, want non-nil")
	}
}

func TestRevokeByTokenHash_Idempotent(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	repo := authpostgres.NewRepository(pool)
	if err := repo.RevokeByTokenHash(ctx, "nonexistent_hash"); err != nil {
		t.Fatalf("RevokeByTokenHash on missing hash: %v", err)
	}
}

func TestCreateScopeSwitchAuditLog(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	branchID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	branchID2 := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000003")
	userID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	fromMID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	toMID := uuid.MustParse("cccccccc-0000-0000-0000-000000000002")

	dbtest.InsertTenant(t, pool, tenantID, "Test Tenant")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Test Branch")
	dbtest.InsertBranch(t, pool, tenantID, branchID2, "Test Branch 2")
	dbtest.InsertUser(t, pool, userID, "user@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, fromMID, tenantID, branchID, userID, "manager", true)
	dbtest.InsertMembership(t, pool, toMID, tenantID, branchID2, userID, "practitioner", true)

	repo := authpostgres.NewRepository(pool)
	fromM := domain.Membership{ID: fromMID, TenantID: tenantID, TenantName: "Test Tenant", BranchID: branchID, BranchName: "Test Branch", Role: "manager", IsActive: true}
	toM := domain.Membership{ID: toMID, TenantID: tenantID, TenantName: "Test Tenant", BranchID: branchID2, BranchName: "Test Branch 2", Role: "practitioner", IsActive: true}
	if err := repo.CreateScopeSwitchAuditLog(ctx, userID, fromM, toM, "req-123"); err != nil {
		t.Fatalf("CreateScopeSwitchAuditLog: %v", err)
	}

	var actionType, reqID string
	var details []byte
	err := pool.QueryRow(ctx,
		"SELECT action_type, request_id, details FROM audit_logs WHERE actor_user_id = $1 LIMIT 1", userID,
	).Scan(&actionType, &reqID, &details)
	if err != nil {
		t.Fatalf("query audit log: %v", err)
	}
	if actionType != "session_scope_switched" {
		t.Errorf("action_type = %s, want session_scope_switched", actionType)
	}
	if reqID != "req-123" {
		t.Errorf("request_id = %s, want req-123", reqID)
	}
	if len(details) == 0 {
		t.Error("details = empty, want JSON object")
	}
}

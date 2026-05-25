package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/invites/domain"
	invitepostgres "nursery-management-system/api/internal/modules/invites/infrastructure/postgres"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/dbtest"
	"nursery-management-system/api/internal/platform/tenant"
)

var (
	invTenantID = uuid.MustParse("b1000000-0000-0000-0000-000000000001")
	invBranchID = uuid.MustParse("b2000000-0000-0000-0000-000000000001")
	invUserID   = uuid.MustParse("b3000000-0000-0000-0000-000000000001")
	invMemID    = uuid.MustParse("b4000000-0000-0000-0000-000000000001")
)

func setupInviteRepo(t *testing.T) (*invitepostgres.Repository, *pgxpool.Pool) {
	t.Helper()
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	dbtest.InsertTenant(t, pool, invTenantID, "Invite Tenant")
	dbtest.InsertBranch(t, pool, invTenantID, invBranchID, "Invite Branch")
	dbtest.InsertUser(t, pool, invUserID, "manager@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, invMemID, invTenantID, invBranchID, invUserID, "manager", true)

	auditWriter := audit.NewWriter()
	return invitepostgres.NewRepository(pool, auditWriter), pool
}

func baseActor() tenant.ActorContext {
	return tenant.ActorContext{
		TenantID:     invTenantID,
		BranchID:     invBranchID,
		UserID:       invUserID,
		MembershipID: invMemID,
	}
}

func futureExpiry() time.Time {
	return time.Now().UTC().Add(168 * time.Hour)
}

func pastExpiry() time.Time {
	return time.Now().UTC().Add(-1 * time.Hour)
}

func TestCreateInvite_NewInvite(t *testing.T) {
	repo, _ := setupInviteRepo(t)
	ctx := context.Background()

	inv := domain.Invite{
		ID:                    uuid.MustParse("c1000000-0000-0000-0000-000000000001"),
		TenantID:              invTenantID,
		BranchID:              invBranchID,
		Email:                 "prac@example.com",
		EmailNormalized:       "prac@example.com",
		Role:                  "practitioner",
		TokenHash:             "hash1",
		ExpiresAt:             futureExpiry(),
		CreatedByUserID:       invUserID,
		CreatedByMembershipID: invMemID,
		SendCount:             1,
	}

	emailSent := false
	created, isNew, err := repo.CreateInvite(ctx, baseActor(), inv, func() error {
		emailSent = true
		return nil
	})
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if !isNew {
		t.Fatal("expected isNew=true")
	}
	if !emailSent {
		t.Fatal("email not sent")
	}
	if created.ID != inv.ID {
		t.Errorf("ID = %s, want %s", created.ID, inv.ID)
	}
	if created.Email != inv.Email {
		t.Errorf("Email = %s, want %s", created.Email, inv.Email)
	}
	if created.Role != "practitioner" {
		t.Errorf("Role = %s, want practitioner", created.Role)
	}
}

func TestCreateInvite_DuplicateRefreshes(t *testing.T) {
	repo, _ := setupInviteRepo(t)
	ctx := context.Background()

	inv := domain.Invite{
		ID:                    uuid.MustParse("c1000000-0000-0000-0000-000000000001"),
		TenantID:              invTenantID,
		BranchID:              invBranchID,
		Email:                 "prac@example.com",
		EmailNormalized:       "prac@example.com",
		Role:                  "practitioner",
		TokenHash:             "hash1",
		ExpiresAt:             futureExpiry(),
		CreatedByUserID:       invUserID,
		CreatedByMembershipID: invMemID,
		SendCount:             1,
	}

	_, _, err := repo.CreateInvite(ctx, baseActor(), inv, func() error { return nil })
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	refreshed := domain.Invite{
		ID:                    uuid.MustParse("c2000000-0000-0000-0000-000000000001"),
		TenantID:              invTenantID,
		BranchID:              invBranchID,
		Email:                 "prac@example.com",
		EmailNormalized:       "prac@example.com",
		Role:                  "practitioner",
		TokenHash:             "hash2",
		ExpiresAt:             futureExpiry(),
		CreatedByUserID:       invUserID,
		CreatedByMembershipID: invMemID,
		SendCount:             1,
	}

	result, isNew, err := repo.CreateInvite(ctx, baseActor(), refreshed, func() error { return nil })
	if err != nil {
		t.Fatalf("duplicate create: %v", err)
	}
	if isNew {
		t.Fatal("expected isNew=false for refresh")
	}
	if result.TokenHash != "hash2" {
		t.Errorf("TokenHash = %s, want hash2", result.TokenHash)
	}
	if result.SendCount != 2 {
		t.Errorf("SendCount = %d, want 2", result.SendCount)
	}
}

func TestCreateInvite_ScopeConflict(t *testing.T) {
	repo, _ := setupInviteRepo(t)
	ctx := context.Background()

	inv := domain.Invite{
		ID:                    uuid.MustParse("c1000000-0000-0000-0000-000000000001"),
		TenantID:              invTenantID,
		BranchID:              invBranchID,
		Email:                 "user@example.com",
		EmailNormalized:       "user@example.com",
		Role:                  "practitioner",
		TokenHash:             "hash1",
		ExpiresAt:             futureExpiry(),
		CreatedByUserID:       invUserID,
		CreatedByMembershipID: invMemID,
		SendCount:             1,
	}

	_, _, err := repo.CreateInvite(ctx, baseActor(), inv, func() error { return nil })
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	conflict := domain.Invite{
		ID:                    uuid.MustParse("c2000000-0000-0000-0000-000000000001"),
		TenantID:              invTenantID,
		BranchID:              invBranchID,
		Email:                 "user@example.com",
		EmailNormalized:       "user@example.com",
		Role:                  "parent",
		TokenHash:             "hash2",
		ExpiresAt:             futureExpiry(),
		CreatedByUserID:       invUserID,
		CreatedByMembershipID: invMemID,
		SendCount:             1,
	}

	_, _, err = repo.CreateInvite(ctx, baseActor(), conflict, func() error { return nil })
	if err != domain.ErrScopeConflict {
		t.Fatalf("expected ErrScopeConflict, got %v", err)
	}
}

func TestCreateInvite_EmailAlreadyRegistered(t *testing.T) {
	repo, _ := setupInviteRepo(t)
	ctx := context.Background()

	inv := domain.Invite{
		ID:                    uuid.MustParse("c1000000-0000-0000-0000-000000000001"),
		TenantID:              invTenantID,
		BranchID:              invBranchID,
		Email:                 "manager@example.com",
		EmailNormalized:       "manager@example.com",
		Role:                  "practitioner",
		TokenHash:             "hash1",
		ExpiresAt:             futureExpiry(),
		CreatedByUserID:       invUserID,
		CreatedByMembershipID: invMemID,
		SendCount:             1,
	}

	_, _, err := repo.CreateInvite(ctx, baseActor(), inv, func() error { return nil })
	if err != domain.ErrEmailAlreadyRegistered {
		t.Fatalf("expected ErrEmailAlreadyRegistered, got %v", err)
	}
}

func TestCreateInvite_EmailRollbackOnSendFailure(t *testing.T) {
	repo, _ := setupInviteRepo(t)
	ctx := context.Background()

	inv := domain.Invite{
		ID:                    uuid.MustParse("c1000000-0000-0000-0000-000000000001"),
		TenantID:              invTenantID,
		BranchID:              invBranchID,
		Email:                 "new@example.com",
		EmailNormalized:       "new@example.com",
		Role:                  "practitioner",
		TokenHash:             "hash1",
		ExpiresAt:             futureExpiry(),
		CreatedByUserID:       invUserID,
		CreatedByMembershipID: invMemID,
		SendCount:             1,
	}

	_, _, err := repo.CreateInvite(ctx, baseActor(), inv, func() error {
		return fmt.Errorf("smtp down")
	})
	if err == nil {
		t.Fatal("expected error on send failure")
	}
}

func TestListInvites_StatusFilters(t *testing.T) {
	repo, pool := setupInviteRepo(t)
	ctx := context.Background()

	pendingID := uuid.MustParse("c1000000-0000-0000-0000-000000000001")
	acceptedID := uuid.MustParse("c1000000-0000-0000-0000-000000000002")
	revokedID := uuid.MustParse("c1000000-0000-0000-0000-000000000003")
	expiredID := uuid.MustParse("c1000000-0000-0000-0000-000000000004")

	dbtest.InsertInvite(t, pool, pendingID, invTenantID, invBranchID,
		"pending@test.com", "pending@test.com", "practitioner", "ph", futureExpiry(), invUserID, invMemID)
	dbtest.InsertInvite(t, pool, acceptedID, invTenantID, invBranchID,
		"accepted@test.com", "accepted@test.com", "parent", "ah", futureExpiry(), invUserID, invMemID)
	dbtest.InsertInvite(t, pool, revokedID, invTenantID, invBranchID,
		"revoked@test.com", "revoked@test.com", "practitioner", "rh", futureExpiry(), invUserID, invMemID)
	dbtest.InsertInvite(t, pool, expiredID, invTenantID, invBranchID,
		"expired@test.com", "expired@test.com", "parent", "eh", pastExpiry(), invUserID, invMemID)

	acceptUID := uuid.MustParse("d1000000-0000-0000-0000-000000000001")
	acceptMID := uuid.MustParse("d2000000-0000-0000-0000-000000000001")
	dbtest.AcceptInvite(t, pool, acceptedID, acceptUID, acceptMID, invTenantID, invBranchID, "accepted@test.com")
	dbtest.RevokeInviteDB(t, pool, revokedID, invUserID, invMemID)

	pending, err := repo.ListInvites(ctx, invTenantID, invBranchID, domain.StatusPending)
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(pending) != 1 || pending[0].Email != "pending@test.com" {
		t.Fatalf("expected 1 pending, got %v", pending)
	}

	accepted, err := repo.ListInvites(ctx, invTenantID, invBranchID, domain.StatusAccepted)
	if err != nil {
		t.Fatalf("ListAccepted: %v", err)
	}
	if len(accepted) != 1 || accepted[0].Email != "accepted@test.com" {
		t.Fatalf("expected 1 accepted, got %v", accepted)
	}

	revoked, err := repo.ListInvites(ctx, invTenantID, invBranchID, domain.StatusRevoked)
	if err != nil {
		t.Fatalf("ListRevoked: %v", err)
	}
	if len(revoked) != 1 || revoked[0].Email != "revoked@test.com" {
		t.Fatalf("expected 1 revoked, got %v", revoked)
	}

	expired, err := repo.ListInvites(ctx, invTenantID, invBranchID, domain.StatusExpired)
	if err != nil {
		t.Fatalf("ListExpired: %v", err)
	}
	if len(expired) != 1 || expired[0].Email != "expired@test.com" {
		t.Fatalf("expected 1 expired, got %v", expired)
	}

	all, err := repo.ListInvites(ctx, invTenantID, invBranchID, domain.StatusAll)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) != 4 {
		t.Fatalf("expected 4 total, got %d", len(all))
	}
}

func TestGetInviteForUpdate_NotFound(t *testing.T) {
	repo, _ := setupInviteRepo(t)
	ctx := context.Background()

	_, err := repo.GetInviteForUpdate(ctx, invTenantID, invBranchID, uuid.New())
	if err != domain.ErrInviteNotFound {
		t.Fatalf("expected ErrInviteNotFound, got %v", err)
	}
}

func TestGetInviteForUpdate_WrongScope(t *testing.T) {
	repo, pool := setupInviteRepo(t)
	ctx := context.Background()

	invID := uuid.MustParse("c1000000-0000-0000-0000-000000000001")
	dbtest.InsertInvite(t, pool, invID, invTenantID, invBranchID,
		"test@test.com", "test@test.com", "practitioner", "h", futureExpiry(), invUserID, invMemID)

	_, err := repo.GetInviteForUpdate(ctx, uuid.New(), invBranchID, invID)
	if err != domain.ErrInviteNotFound {
		t.Fatalf("expected ErrInviteNotFound for wrong tenant, got %v", err)
	}
}

func TestGetInviteByTokenHashForUpdate_NotFound(t *testing.T) {
	repo, _ := setupInviteRepo(t)
	ctx := context.Background()

	_, err := repo.GetInviteByTokenHashForUpdate(ctx, "nonexistent")
	if err != domain.ErrTokenInvalid {
		t.Fatalf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestGetInviteByTokenHashForUpdate_Found(t *testing.T) {
	repo, pool := setupInviteRepo(t)
	ctx := context.Background()

	invID := uuid.MustParse("c1000000-0000-0000-0000-000000000001")
	dbtest.InsertInvite(t, pool, invID, invTenantID, invBranchID,
		"test@test.com", "test@test.com", "practitioner", "known-hash", futureExpiry(), invUserID, invMemID)

	inv, err := repo.GetInviteByTokenHashForUpdate(ctx, "known-hash")
	if err != nil {
		t.Fatalf("GetByTokenHash: %v", err)
	}
	if inv.Email != "test@test.com" {
		t.Errorf("Email = %s, want test@test.com", inv.Email)
	}
}

func TestRevokeInvite_Success(t *testing.T) {
	repo, pool := setupInviteRepo(t)
	ctx := context.Background()

	invID := uuid.MustParse("c1000000-0000-0000-0000-000000000001")
	dbtest.InsertInvite(t, pool, invID, invTenantID, invBranchID,
		"test@test.com", "test@test.com", "practitioner", "h", futureExpiry(), invUserID, invMemID)

	revoked, err := repo.RevokeInvite(ctx, baseActor(), invID)
	if err != nil {
		t.Fatalf("RevokeInvite: %v", err)
	}
	if revoked.RevokedAt == nil {
		t.Fatal("RevokedAt is nil")
	}
}

func TestRevokeInvite_AlreadyAccepted(t *testing.T) {
	repo, pool := setupInviteRepo(t)
	ctx := context.Background()

	invID := uuid.MustParse("c1000000-0000-0000-0000-000000000001")
	dbtest.InsertInvite(t, pool, invID, invTenantID, invBranchID,
		"test@test.com", "test@test.com", "practitioner", "h", futureExpiry(), invUserID, invMemID)
	acceptUID := uuid.MustParse("d1000000-0000-0000-0000-000000000001")
	acceptMID := uuid.MustParse("d2000000-0000-0000-0000-000000000001")
	dbtest.AcceptInvite(t, pool, invID, acceptUID, acceptMID, invTenantID, invBranchID, "test@test.com")

	_, err := repo.RevokeInvite(ctx, baseActor(), invID)
	if err != domain.ErrInviteAccepted {
		t.Fatalf("expected ErrInviteAccepted, got %v", err)
	}
}

func TestRevokeInvite_Idempotent(t *testing.T) {
	repo, pool := setupInviteRepo(t)
	ctx := context.Background()

	invID := uuid.MustParse("c1000000-0000-0000-0000-000000000001")
	dbtest.InsertInvite(t, pool, invID, invTenantID, invBranchID,
		"test@test.com", "test@test.com", "practitioner", "h", futureExpiry(), invUserID, invMemID)
	dbtest.RevokeInviteDB(t, pool, invID, invUserID, invMemID)

	revoked, err := repo.RevokeInvite(ctx, baseActor(), invID)
	if err != nil {
		t.Fatalf("idempotent revoke: %v", err)
	}
	if revoked.RevokedAt == nil {
		t.Fatal("RevokedAt is nil")
	}
}

func TestRevokeInvite_ExpiredInvite(t *testing.T) {
	repo, pool := setupInviteRepo(t)
	ctx := context.Background()

	invID := uuid.MustParse("c1000000-0000-0000-0000-000000000001")
	dbtest.InsertInvite(t, pool, invID, invTenantID, invBranchID,
		"test@test.com", "test@test.com", "practitioner", "h", pastExpiry(), invUserID, invMemID)

	revoked, err := repo.RevokeInvite(ctx, baseActor(), invID)
	if err != nil {
		t.Fatalf("revoke expired: %v", err)
	}
	if revoked.RevokedAt == nil {
		t.Fatal("RevokedAt is nil")
	}
}

func TestAcceptInvite_Success(t *testing.T) {
	repo, pool := setupInviteRepo(t)
	ctx := context.Background()

	invID := uuid.MustParse("c1000000-0000-0000-0000-000000000001")
	dbtest.InsertInvite(t, pool, invID, invTenantID, invBranchID,
		"new@test.com", "new@test.com", "practitioner", "h", futureExpiry(), invUserID, invMemID)

	userID := uuid.MustParse("d1000000-0000-0000-0000-000000000001")
	memID := uuid.MustParse("d2000000-0000-0000-0000-000000000001")

	inv := domain.Invite{
		ID:              invID,
		TenantID:        invTenantID,
		BranchID:        invBranchID,
		Email:           "new@test.com",
		EmailNormalized: "new@test.com",
		Role:            "practitioner",
	}

	user := domain.CreatedUser{
		ID:           userID,
		Email:        "new@test.com",
		PasswordHash: "bcrypt-hash",
	}

	membership := domain.CreatedMembership{
		ID:       memID,
		TenantID: invTenantID,
		BranchID: invBranchID,
		UserID:   userID,
		Role:     "practitioner",
	}

	if err := repo.AcceptInvite(ctx, inv, user, membership); err != nil {
		t.Fatalf("AcceptInvite: %v", err)
	}

	var acceptedAt *time.Time
	err := pool.QueryRow(ctx, "SELECT accepted_at FROM manager_invites WHERE id = $1", invID).Scan(&acceptedAt)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if acceptedAt == nil {
		t.Fatal("accepted_at is nil")
	}

	var count int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM users WHERE email_normalized = 'new@test.com'").Scan(&count)
	if err != nil {
		t.Fatalf("query users: %v", err)
	}
	if count != 1 {
		t.Errorf("users count = %d, want 1", count)
	}

	err = pool.QueryRow(ctx, "SELECT count(*) FROM memberships WHERE user_id = $1 AND role = 'practitioner'", userID).Scan(&count)
	if err != nil {
		t.Fatalf("query memberships: %v", err)
	}
	if count != 1 {
		t.Errorf("memberships count = %d, want 1", count)
	}
}

func TestAcceptInvite_EmailBecameRegistered(t *testing.T) {
	repo, pool := setupInviteRepo(t)
	ctx := context.Background()

	invID := uuid.MustParse("c1000000-0000-0000-0000-000000000001")
	dbtest.InsertInvite(t, pool, invID, invTenantID, invBranchID,
		"taken@test.com", "taken@test.com", "practitioner", "h", futureExpiry(), invUserID, invMemID)

	anotherUID := uuid.MustParse("e1000000-0000-0000-0000-000000000001")
	dbtest.InsertUser(t, pool, anotherUID, "taken@test.com", "hash", true)

	inv := domain.Invite{
		ID:              invID,
		TenantID:        invTenantID,
		BranchID:        invBranchID,
		Email:           "taken@test.com",
		EmailNormalized: "taken@test.com",
		Role:            "practitioner",
	}

	user := domain.CreatedUser{
		ID:           uuid.MustParse("d1000000-0000-0000-0000-000000000001"),
		Email:        "taken@test.com",
		PasswordHash: "hash",
	}

	membership := domain.CreatedMembership{
		ID:       uuid.MustParse("d2000000-0000-0000-0000-000000000001"),
		TenantID: invTenantID,
		BranchID: invBranchID,
		UserID:   user.ID,
		Role:     "practitioner",
	}

	err := repo.AcceptInvite(ctx, inv, user, membership)
	if err != domain.ErrEmailAlreadyRegistered {
		t.Fatalf("expected ErrEmailAlreadyRegistered, got %v", err)
	}
}

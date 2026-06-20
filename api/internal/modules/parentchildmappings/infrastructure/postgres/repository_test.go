package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	pcmdomain "nursery-management-system/api/internal/modules/parentchildmappings/domain"
	pcmpostgres "nursery-management-system/api/internal/modules/parentchildmappings/infrastructure/postgres"
	"nursery-management-system/api/internal/platform/dbtest"
)

var (
	pcmTenantID = uuid.MustParse("c1000000-0000-0000-0000-000000000001")
	pcmBranchID = uuid.MustParse("c2000000-0000-0000-0000-000000000001")
	pcmUserID   = uuid.MustParse("c3000000-0000-0000-0000-000000000001")
)

func setupPCMRepo(t *testing.T) (*pcmpostgres.ParentChildMappingRepository, *pgxpool.Pool) {
	t.Helper()
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	dbtest.InsertTenant(t, pool, pcmTenantID, "PCM Tenant")
	dbtest.InsertBranch(t, pool, pcmTenantID, pcmBranchID, "PCM Branch")
	dbtest.InsertUser(t, pool, pcmUserID, "staff@example.com", "hash", true)

	return pcmpostgres.NewParentChildMappingRepository(pool), pool
}

func seedPCMEntities(t testing.TB, pool *pgxpool.Pool) (membershipID, childID uuid.UUID) {
	t.Helper()
	membershipID = uuid.MustParse("c4000000-0000-0000-0000-000000000001")
	childID = uuid.MustParse("c5000000-0000-0000-0000-000000000001")

	dbtest.InsertMembership(t, pool, membershipID, pcmTenantID, pcmBranchID, pcmUserID, "parent", true)
	dbtest.InsertChild(t, pool, childID, pcmTenantID, pcmBranchID, "PCM Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	return membershipID, childID
}

func TestPCMFindActiveByPair_NoRow(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()
	membershipID, childID := seedPCMEntities(t, pool)

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.FindActiveByPair(ctx, tx, pcmTenantID, pcmBranchID, membershipID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true when no mapping exists")
	}
}

func TestPCMFindActiveByPair_EndedRow(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()
	membershipID, childID := seedPCMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	dbtest.InsertParentMapping(t, pool, mappingID, pcmTenantID, pcmBranchID, membershipID, childID)

	_, err := pool.Exec(ctx,
		"UPDATE parent_membership_children SET ended_at = now(), ended_reason_code = 'access_revoked' WHERE id = $1", mappingID)
	if err != nil {
		t.Fatalf("end mapping: %v", err)
	}

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.FindActiveByPair(ctx, tx, pcmTenantID, pcmBranchID, membershipID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for ended mapping")
	}
}

func TestPCMFindActiveByPair_WrongScope(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()
	membershipID, childID := seedPCMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	dbtest.InsertParentMapping(t, pool, mappingID, pcmTenantID, pcmBranchID, membershipID, childID)

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.FindActiveByPair(ctx, tx, uuid.New(), pcmBranchID, membershipID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for wrong tenant")
	}
}

func TestPCMFindActiveByPair_Success(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()
	membershipID, childID := seedPCMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	dbtest.InsertParentMapping(t, pool, mappingID, pcmTenantID, pcmBranchID, membershipID, childID)

	tx := dbtest.BeginTx(t, pool)
	mapping, found, err := repo.FindActiveByPair(ctx, tx, pcmTenantID, pcmBranchID, membershipID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if mapping.ID != mappingID {
		t.Errorf("ID = %s, want %s", mapping.ID, mappingID)
	}
	if mapping.ChildID != childID {
		t.Errorf("ChildID = %s, want %s", mapping.ChildID, childID)
	}
}

func TestPCMListActiveByMembership(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()
	membershipID, childID := seedPCMEntities(t, pool)

	childID2 := uuid.MustParse("c5000000-0000-0000-0000-000000000002")
	dbtest.InsertChild(t, pool, childID2, pcmTenantID, pcmBranchID, "PCM Child 2",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)

	mappingID1 := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	mappingID2 := uuid.MustParse("c6000000-0000-0000-0000-000000000002")
	dbtest.InsertParentMapping(t, pool, mappingID1, pcmTenantID, pcmBranchID, membershipID, childID)
	dbtest.InsertParentMapping(t, pool, mappingID2, pcmTenantID, pcmBranchID, membershipID, childID2)

	tx := dbtest.BeginTx(t, pool)
	mappings, err := repo.ListActiveByMembership(ctx, tx, pcmTenantID, pcmBranchID, membershipID)
	if err != nil {
		t.Fatalf("ListActiveByMembership: %v", err)
	}
	if len(mappings) != 2 {
		t.Errorf("len(mappings) = %d, want 2", len(mappings))
	}
}

func TestPCMCreate(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()
	membershipID, childID := seedPCMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	mapping := pcmdomain.ParentChildMapping{
		ID:           mappingID,
		TenantID:     pcmTenantID,
		BranchID:     pcmBranchID,
		MembershipID: membershipID,
		ChildID:      childID,
	}

	tx := dbtest.BeginTx(t, pool)
	if err := repo.Create(ctx, tx, mapping); err != nil {
		t.Fatalf("Create: %v", err)
	}
	dbtest.CommitTx(t, tx)

	tx2 := dbtest.BeginTx(t, pool)
	got, found, err := repo.FindActiveByPair(ctx, tx2, pcmTenantID, pcmBranchID, membershipID, childID)
	if err != nil {
		t.Fatalf("FindActiveByPair: %v", err)
	}
	if !found {
		t.Fatal("mapping not found after create")
	}
	if got.ID != mappingID {
		t.Errorf("ID = %s, want %s", got.ID, mappingID)
	}
	dbtest.CommitTx(t, tx2)
}

func TestPCMCreate_NonParentRole(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()

	managerMID := uuid.MustParse("c4000000-0000-0000-0000-000000000002")
	childID := uuid.MustParse("c5000000-0000-0000-0000-000000000001")

	dbtest.InsertMembership(t, pool, managerMID, pcmTenantID, pcmBranchID, pcmUserID, "manager", true)
	dbtest.InsertChild(t, pool, childID, pcmTenantID, pcmBranchID, "PCM Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000002")
	mapping := pcmdomain.ParentChildMapping{
		ID:           mappingID,
		TenantID:     pcmTenantID,
		BranchID:     pcmBranchID,
		MembershipID: managerMID,
		ChildID:      childID,
	}

	tx := dbtest.BeginTx(t, pool)
	err := repo.Create(ctx, tx, mapping)
	if err == nil {
		t.Fatal("expected error for non-parent role, got nil")
	}
}

func TestPCMCreate_InactiveMembership(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()

	inactiveMID := uuid.MustParse("c4000000-0000-0000-0000-000000000002")
	childID := uuid.MustParse("c5000000-0000-0000-0000-000000000001")

	dbtest.InsertMembership(t, pool, inactiveMID, pcmTenantID, pcmBranchID, pcmUserID, "parent", false)
	dbtest.InsertChild(t, pool, childID, pcmTenantID, pcmBranchID, "PCM Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000002")
	mapping := pcmdomain.ParentChildMapping{
		ID:           mappingID,
		TenantID:     pcmTenantID,
		BranchID:     pcmBranchID,
		MembershipID: inactiveMID,
		ChildID:      childID,
	}

	tx := dbtest.BeginTx(t, pool)
	err := repo.Create(ctx, tx, mapping)
	if err == nil {
		t.Fatal("expected error for inactive membership, got nil")
	}
}

func TestPCMCreate_ChildNotInScope(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()

	membershipID := uuid.MustParse("c4000000-0000-0000-0000-000000000001")
	otherChildID := uuid.MustParse("c5000000-0000-0000-0000-000000000002")

	// Child lives in a different tenant; the trigger should reject the insert.
	dbtest.InsertChild(t, pool, otherChildID, uuid.New(), pcmBranchID, "Other Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000002")
	mapping := pcmdomain.ParentChildMapping{
		ID:           mappingID,
		TenantID:     pcmTenantID,
		BranchID:     pcmBranchID,
		MembershipID: membershipID,
		ChildID:      otherChildID,
	}

	tx := dbtest.BeginTx(t, pool)
	err := repo.Create(ctx, tx, mapping)
	if err == nil {
		t.Fatal("expected error for out-of-scope child, got nil")
	}
}

func TestPCMGetByIDForUpdate(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()
	membershipID, childID := seedPCMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	dbtest.InsertParentMapping(t, pool, mappingID, pcmTenantID, pcmBranchID, membershipID, childID)

	tx := dbtest.BeginTx(t, pool)
	got, found, err := repo.GetByIDForUpdate(ctx, tx, pcmTenantID, pcmBranchID, mappingID)
	if err != nil {
		t.Fatalf("GetByIDForUpdate: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if got.MembershipID != membershipID {
		t.Errorf("MembershipID = %s, want %s", got.MembershipID, membershipID)
	}
	if got.ChildID != childID {
		t.Errorf("ChildID = %s, want %s", got.ChildID, childID)
	}
}

func TestPCMEnd(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()
	membershipID, childID := seedPCMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	dbtest.InsertParentMapping(t, pool, mappingID, pcmTenantID, pcmBranchID, membershipID, childID)

	tx := dbtest.BeginTx(t, pool)
	if err := repo.End(ctx, tx, pcmTenantID, pcmBranchID, mappingID, "access_revoked", "test reason"); err != nil {
		t.Fatalf("End: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var reasonCode, reasonNote *string
	var endedAt *string
	err := pool.QueryRow(ctx,
		"SELECT ended_reason_code::text, ended_reason_note, ended_at::text FROM parent_membership_children WHERE id = $1",
		mappingID).Scan(&reasonCode, &reasonNote, &endedAt)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if endedAt == nil {
		t.Error("ended_at = nil, want non-nil")
	}
	if reasonCode == nil || *reasonCode != "access_revoked" {
		t.Errorf("reason_code = %v, want access_revoked", reasonCode)
	}
	if reasonNote == nil || *reasonNote != "test reason" {
		t.Errorf("reason_note = %v, want test reason", reasonNote)
	}
}

func TestPCMGetMembershipForScope(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()
	membershipID, _ := seedPCMEntities(t, pool)

	tx := dbtest.BeginTx(t, pool)
	info, found, err := repo.GetMembershipForScope(ctx, tx, pcmTenantID, pcmBranchID, membershipID)
	if err != nil {
		t.Fatalf("GetMembershipForScope: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if info.Role != "parent" {
		t.Errorf("Role = %s, want parent", info.Role)
	}
	if !info.IsActive {
		t.Error("IsActive = false, want true")
	}
}

func TestPCMGetMembershipForScope_NotFound(t *testing.T) {
	repo, pool := setupPCMRepo(t)

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.GetMembershipForScope(context.Background(), tx, pcmTenantID, pcmBranchID, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for nonexistent membership")
	}
}

func TestPCMActivePairUniqueness(t *testing.T) {
	repo, pool := setupPCMRepo(t)
	ctx := context.Background()
	membershipID, childID := seedPCMEntities(t, pool)

	mappingID1 := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	dbtest.InsertParentMapping(t, pool, mappingID1, pcmTenantID, pcmBranchID, membershipID, childID)

	tx := dbtest.BeginTx(t, pool)
	err := repo.Create(ctx, tx, pcmdomain.ParentChildMapping{
		ID:           uuid.MustParse("c6000000-0000-0000-0000-000000000002"),
		TenantID:     pcmTenantID,
		BranchID:     pcmBranchID,
		MembershipID: membershipID,
		ChildID:      childID,
	})
	if err == nil {
		t.Fatal("expected error for duplicate active pair mapping, got nil")
	}
}

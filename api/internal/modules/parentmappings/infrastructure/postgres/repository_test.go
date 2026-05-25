package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	pmdomain "nursery-management-system/api/internal/modules/parentmappings/domain"
	pmpostgres "nursery-management-system/api/internal/modules/parentmappings/infrastructure/postgres"
	"nursery-management-system/api/internal/platform/dbtest"
)

var (
	pmTenantID = uuid.MustParse("c1000000-0000-0000-0000-000000000001")
	pmBranchID = uuid.MustParse("c2000000-0000-0000-0000-000000000001")
	pmUserID   = uuid.MustParse("c3000000-0000-0000-0000-000000000001")
)

func setupPMRepo(t *testing.T) (*pmpostgres.ParentMappingRepository, *pgxpool.Pool) {
	t.Helper()
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	dbtest.InsertTenant(t, pool, pmTenantID, "PM Tenant")
	dbtest.InsertBranch(t, pool, pmTenantID, pmBranchID, "PM Branch")
	dbtest.InsertUser(t, pool, pmUserID, "staff@example.com", "hash", true)

	return pmpostgres.NewParentMappingRepository(pool), pool
}

func seedPMEntities(t testing.TB, pool *pgxpool.Pool) (membershipID, guardianID uuid.UUID) {
	t.Helper()
	membershipID = uuid.MustParse("c4000000-0000-0000-0000-000000000001")
	guardianID = uuid.MustParse("c5000000-0000-0000-0000-000000000001")

	dbtest.InsertMembership(t, pool, membershipID, pmTenantID, pmBranchID, pmUserID, "parent", true)
	dbtest.InsertGuardian(t, pool, guardianID, pmTenantID, pmBranchID, "Guardian G", true)

	return membershipID, guardianID
}

func TestPMFindActiveByMembership_NoRow(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()
	membershipID, _ := seedPMEntities(t, pool)

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.FindActiveByMembership(ctx, tx, pmTenantID, pmBranchID, membershipID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true when no mapping exists")
	}
}

func TestPMFindActiveByMembership_EndedRow(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()
	membershipID, guardianID := seedPMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	dbtest.InsertParentMapping(t, pool, mappingID, pmTenantID, pmBranchID, membershipID, guardianID)

	_, err := pool.Exec(ctx,
		"UPDATE parent_membership_guardians SET ended_at = now(), ended_reason_code = 'access_revoked' WHERE id = $1", mappingID)
	if err != nil {
		t.Fatalf("end mapping: %v", err)
	}

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.FindActiveByMembership(ctx, tx, pmTenantID, pmBranchID, membershipID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for ended mapping")
	}
}

func TestPMFindActiveByMembership_WrongScope(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()
	membershipID, guardianID := seedPMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	dbtest.InsertParentMapping(t, pool, mappingID, pmTenantID, pmBranchID, membershipID, guardianID)

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.FindActiveByMembership(ctx, tx, uuid.New(), pmBranchID, membershipID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for wrong tenant")
	}
}

func TestPMFindActiveByMembership_Success(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()
	membershipID, guardianID := seedPMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	dbtest.InsertParentMapping(t, pool, mappingID, pmTenantID, pmBranchID, membershipID, guardianID)

	tx := dbtest.BeginTx(t, pool)
	mapping, found, err := repo.FindActiveByMembership(ctx, tx, pmTenantID, pmBranchID, membershipID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if mapping.ID != mappingID {
		t.Errorf("ID = %s, want %s", mapping.ID, mappingID)
	}
	if mapping.GuardianID != guardianID {
		t.Errorf("GuardianID = %s, want %s", mapping.GuardianID, guardianID)
	}
}

func TestPMCreate(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()
	membershipID, guardianID := seedPMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	mapping := pmdomain.ParentMapping{
		ID:           mappingID,
		TenantID:     pmTenantID,
		BranchID:     pmBranchID,
		MembershipID: membershipID,
		GuardianID:   guardianID,
	}

	tx := dbtest.BeginTx(t, pool)
	if err := repo.Create(ctx, tx, mapping); err != nil {
		t.Fatalf("Create: %v", err)
	}
	dbtest.CommitTx(t, tx)

	tx2 := dbtest.BeginTx(t, pool)
	got, found, err := repo.FindActiveByMembership(ctx, tx2, pmTenantID, pmBranchID, membershipID)
	if err != nil {
		t.Fatalf("FindActiveByMembership: %v", err)
	}
	if !found {
		t.Fatal("mapping not found after create")
	}
	if got.ID != mappingID {
		t.Errorf("ID = %s, want %s", got.ID, mappingID)
	}
	dbtest.CommitTx(t, tx2)
}

func TestPMCreate_NonParentRole(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()

	managerMID := uuid.MustParse("c4000000-0000-0000-0000-000000000002")
	guardianID := uuid.MustParse("c5000000-0000-0000-0000-000000000001")

	dbtest.InsertMembership(t, pool, managerMID, pmTenantID, pmBranchID, pmUserID, "manager", true)
	dbtest.InsertGuardian(t, pool, guardianID, pmTenantID, pmBranchID, "Guardian G", true)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000002")
	mapping := pmdomain.ParentMapping{
		ID:           mappingID,
		TenantID:     pmTenantID,
		BranchID:     pmBranchID,
		MembershipID: managerMID,
		GuardianID:   guardianID,
	}

	tx := dbtest.BeginTx(t, pool)
	err := repo.Create(ctx, tx, mapping)
	if err == nil {
		t.Fatal("expected error for non-parent role, got nil")
	}
}

func TestPMCreate_InactiveMembership(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()

	inactiveMID := uuid.MustParse("c4000000-0000-0000-0000-000000000002")
	guardianID := uuid.MustParse("c5000000-0000-0000-0000-000000000001")

	dbtest.InsertMembership(t, pool, inactiveMID, pmTenantID, pmBranchID, pmUserID, "parent", false)
	dbtest.InsertGuardian(t, pool, guardianID, pmTenantID, pmBranchID, "Guardian G", true)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000002")
	mapping := pmdomain.ParentMapping{
		ID:           mappingID,
		TenantID:     pmTenantID,
		BranchID:     pmBranchID,
		MembershipID: inactiveMID,
		GuardianID:   guardianID,
	}

	tx := dbtest.BeginTx(t, pool)
	err := repo.Create(ctx, tx, mapping)
	if err == nil {
		t.Fatal("expected error for inactive membership, got nil")
	}
}

func TestPMCreate_InactiveGuardian(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()

	membershipID := uuid.MustParse("c4000000-0000-0000-0000-000000000001")
	inactiveGuardID := uuid.MustParse("c5000000-0000-0000-0000-000000000002")

	dbtest.InsertMembership(t, pool, membershipID, pmTenantID, pmBranchID, pmUserID, "parent", true)
	_, err := pool.Exec(ctx,
		"INSERT INTO guardians (id, tenant_id, branch_id, full_name, is_active, deactivated_at, deactivation_reason_code) VALUES ($1, $2, $3, $4, false, now(), 'access_revoked')",
		inactiveGuardID, pmTenantID, pmBranchID, "Inactive G")
	if err != nil {
		t.Fatalf("insert inactive guardian: %v", err)
	}

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000002")
	mapping := pmdomain.ParentMapping{
		ID:           mappingID,
		TenantID:     pmTenantID,
		BranchID:     pmBranchID,
		MembershipID: membershipID,
		GuardianID:   inactiveGuardID,
	}

	tx := dbtest.BeginTx(t, pool)
	err = repo.Create(ctx, tx, mapping)
	if err == nil {
		t.Fatal("expected error for inactive guardian, got nil")
	}
}

func TestPMGetByIDForUpdate(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()
	membershipID, guardianID := seedPMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	dbtest.InsertParentMapping(t, pool, mappingID, pmTenantID, pmBranchID, membershipID, guardianID)

	tx := dbtest.BeginTx(t, pool)
	got, found, err := repo.GetByIDForUpdate(ctx, tx, pmTenantID, pmBranchID, mappingID)
	if err != nil {
		t.Fatalf("GetByIDForUpdate: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if got.MembershipID != membershipID {
		t.Errorf("MembershipID = %s, want %s", got.MembershipID, membershipID)
	}
	if got.GuardianID != guardianID {
		t.Errorf("GuardianID = %s, want %s", got.GuardianID, guardianID)
	}
}

func TestPMEnd(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()
	membershipID, guardianID := seedPMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	dbtest.InsertParentMapping(t, pool, mappingID, pmTenantID, pmBranchID, membershipID, guardianID)

	tx := dbtest.BeginTx(t, pool)
	if err := repo.End(ctx, tx, pmTenantID, pmBranchID, mappingID, "access_revoked", "test reason"); err != nil {
		t.Fatalf("End: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var reasonCode, reasonNote *string
	var endedAt *string
	err := pool.QueryRow(ctx,
		"SELECT ended_reason_code::text, ended_reason_note, ended_at::text FROM parent_membership_guardians WHERE id = $1",
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

func TestPMEnd_WrongScope(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()
	membershipID, guardianID := seedPMEntities(t, pool)

	mappingID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	dbtest.InsertParentMapping(t, pool, mappingID, pmTenantID, pmBranchID, membershipID, guardianID)

	tx := dbtest.BeginTx(t, pool)
	if err := repo.End(ctx, tx, uuid.New(), pmBranchID, mappingID, "access_revoked", ""); err != nil {
				t.Fatalf("End wrong scope: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var endedAt *string
	err := pool.QueryRow(ctx,
		"SELECT ended_at::text FROM parent_membership_guardians WHERE id = $1", mappingID).Scan(&endedAt)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if endedAt != nil {
		t.Error("ended_at should remain nil for wrong scope")
	}
}

func TestPMGetMembershipForScope(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()
	membershipID, _ := seedPMEntities(t, pool)

	tx := dbtest.BeginTx(t, pool)
	info, found, err := repo.GetMembershipForScope(ctx, tx, pmTenantID, pmBranchID, membershipID)
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

func TestPMGetMembershipForScope_NotFound(t *testing.T) {
	repo, pool := setupPMRepo(t)

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.GetMembershipForScope(context.Background(), tx, pmTenantID, pmBranchID, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for nonexistent membership")
	}
}

func TestPMActiveMappingUniqueness(t *testing.T) {
	repo, pool := setupPMRepo(t)
	ctx := context.Background()
	membershipID, guardianID := seedPMEntities(t, pool)

	mappingID1 := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	mappingID2 := uuid.MustParse("c6000000-0000-0000-0000-000000000002")
	dbtest.InsertParentMapping(t, pool, mappingID1, pmTenantID, pmBranchID, membershipID, guardianID)

	tx := dbtest.BeginTx(t, pool)
	err := repo.Create(ctx, tx, pmdomain.ParentMapping{
		ID:           mappingID2,
		TenantID:     pmTenantID,
		BranchID:     pmBranchID,
		MembershipID: membershipID,
		GuardianID:   guardianID,
	})
	if err == nil {
		t.Fatal("expected error for duplicate active mapping per membership, got nil")
	}
}

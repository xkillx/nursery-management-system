package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	guardiandomain "nursery-management-system/api/internal/modules/guardians/domain"
	guardianpostgres "nursery-management-system/api/internal/modules/guardians/infrastructure/postgres"
	"nursery-management-system/api/internal/platform/dbtest"
)

var (
	guardTenantID = uuid.MustParse("a1000000-0000-0000-0000-000000000001")
	guardBranchID = uuid.MustParse("a2000000-0000-0000-0000-000000000001")
	guardUserID   = uuid.MustParse("a3000000-0000-0000-0000-000000000001")
)

func setupGuardianRepo(t *testing.T) (*guardianpostgres.GuardianRepository, *pgxpool.Pool) {
	t.Helper()
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	dbtest.InsertTenant(t, pool, guardTenantID, "Guard Tenant")
	dbtest.InsertBranch(t, pool, guardTenantID, guardBranchID, "Guard Branch")
	dbtest.InsertUser(t, pool, guardUserID, "staff@example.com", "hash", true)

	return guardianpostgres.NewGuardianRepository(pool), pool
}

func TestGuardianGetByID_NotFound(t *testing.T) {
	repo, _ := setupGuardianRepo(t)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, guardTenantID, guardBranchID, uuid.New())
	if err != guardiandomain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGuardianGetByID_WrongScope(t *testing.T) {
	repo, pool := setupGuardianRepo(t)
	ctx := context.Background()

	guardID := uuid.MustParse("a4000000-0000-0000-0000-000000000001")
	dbtest.InsertGuardian(t, pool, guardID, guardTenantID, guardBranchID, "John", true)

	_, err := repo.GetByID(ctx, uuid.New(), guardBranchID, guardID)
	if err != guardiandomain.ErrNotFound {
		t.Fatalf("expected ErrNotFound for wrong tenant, got %v", err)
	}
}

func TestGuardianGetByID_Success(t *testing.T) {
	repo, pool := setupGuardianRepo(t)
	ctx := context.Background()

	guardID := uuid.MustParse("a4000000-0000-0000-0000-000000000001")
	dbtest.InsertGuardian(t, pool, guardID, guardTenantID, guardBranchID, "John Smith", true)

	got, err := repo.GetByID(ctx, guardTenantID, guardBranchID, guardID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.FullName != "John Smith" {
		t.Errorf("FullName = %s, want John Smith", got.FullName)
	}
	if !got.IsActive {
		t.Error("IsActive = false, want true")
	}
}

func TestGuardianList_Filters(t *testing.T) {
	repo, pool := setupGuardianRepo(t)
	ctx := context.Background()

	activeID := uuid.MustParse("a4000000-0000-0000-0000-000000000001")
	inactiveID := uuid.MustParse("a4000000-0000-0000-0000-000000000002")
	dbtest.InsertGuardian(t, pool, activeID, guardTenantID, guardBranchID, "Active G", true)

	_, err := pool.Exec(ctx,
		"INSERT INTO guardians (id, tenant_id, branch_id, full_name, is_active, deactivated_at, deactivation_reason_code) VALUES ($1, $2, $3, $4, false, now(), 'access_revoked')",
		inactiveID, guardTenantID, guardBranchID, "Inactive G")
	if err != nil {
		t.Fatalf("insert inactive: %v", err)
	}

	active, err := repo.List(ctx, guardTenantID, guardBranchID, guardiandomain.StatusActive, 10, 0)
	if err != nil {
		t.Fatalf("List active: %v", err)
	}
	if len(active) != 1 || active[0].FullName != "Active G" {
		t.Fatalf("expected 1 active, got %v", active)
	}

	inactive, err := repo.List(ctx, guardTenantID, guardBranchID, guardiandomain.StatusInactive, 10, 0)
	if err != nil {
		t.Fatalf("List inactive: %v", err)
	}
	if len(inactive) != 1 || inactive[0].FullName != "Inactive G" {
		t.Fatalf("expected 1 inactive, got %v", inactive)
	}

	all, err := repo.List(ctx, guardTenantID, guardBranchID, guardiandomain.StatusAll, 10, 0)
	if err != nil {
		t.Fatalf("List all: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 total, got %d", len(all))
	}
}

func TestGuardianCreate_NullOptionals(t *testing.T) {
	repo, pool := setupGuardianRepo(t)
	ctx := context.Background()

	guardID := uuid.MustParse("a4000000-0000-0000-0000-000000000001")
	g := guardiandomain.Guardian{
		ID:       guardID,
		TenantID: guardTenantID,
		BranchID: guardBranchID,
		FullName: "No Contacts",
		Email:    nil,
		Phone:    nil,
		Notes:    nil,
	}

	if err := repo.Create(ctx, g); err != nil {
		t.Fatalf("Create: %v", err)
	}

	var email, phone, notes *string
	err := pool.QueryRow(ctx,
		"SELECT email, phone, notes FROM guardians WHERE id = $1", guardID).Scan(&email, &phone, &notes)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if email != nil {
		t.Errorf("email = %v, want nil", email)
	}
	if phone != nil {
		t.Errorf("phone = %v, want nil", phone)
	}
	if notes != nil {
		t.Errorf("notes = %v, want nil", notes)
	}
}

func TestGuardianUpdate_SelectiveFields(t *testing.T) {
	repo, pool := setupGuardianRepo(t)
	ctx := context.Background()

	guardID := uuid.MustParse("a4000000-0000-0000-0000-000000000001")
	dbtest.InsertGuardian(t, pool, guardID, guardTenantID, guardBranchID, "Original", true)

	ct, err := repo.Update(ctx, guardTenantID, guardBranchID, guardID, map[string]any{
		"full_name": "Updated Name",
		"email":     "test@example.com",
		"phone":     "",
		"notes":     "",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if ct != 1 {
		t.Errorf("rows affected = %d, want 1", ct)
	}

	var email, phone, notes *string
	err = pool.QueryRow(ctx,
		"SELECT email, phone, notes FROM guardians WHERE id = $1", guardID).Scan(&email, &phone, &notes)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if email == nil || *email != "test@example.com" {
		t.Errorf("email = %v, want test@example.com", email)
	}
	if phone != nil {
		t.Errorf("phone = %v, want nil (cleared)", phone)
	}
	if notes != nil {
		t.Errorf("notes = %v, want nil (cleared)", notes)
	}
}

func TestGuardianUpdate_WrongScope(t *testing.T) {
	repo, _ := setupGuardianRepo(t)

	ct, err := repo.Update(context.Background(), uuid.New(), guardBranchID, uuid.New(), map[string]any{
		"full_name": "X",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if ct != 0 {
		t.Errorf("rows affected = %d, want 0", ct)
	}
}

func TestGuardianGetByIDForUpdate(t *testing.T) {
	repo, pool := setupGuardianRepo(t)
	ctx := context.Background()

	guardID := uuid.MustParse("a4000000-0000-0000-0000-000000000001")
	dbtest.InsertGuardian(t, pool, guardID, guardTenantID, guardBranchID, "John", true)

	tx := dbtest.BeginTx(t, pool)
	got, err := repo.GetByIDForUpdate(ctx, tx, guardTenantID, guardBranchID, guardID)
	if err != nil {
		t.Fatalf("GetByIDForUpdate: %v", err)
	}
	if got.FullName != "John" {
		t.Errorf("FullName = %s, want John", got.FullName)
	}

	_, err = repo.GetByIDForUpdate(ctx, tx, uuid.New(), guardBranchID, guardID)
	if err != guardiandomain.ErrNotFound {
		t.Fatalf("expected ErrNotFound for wrong scope, got %v", err)
	}
}

func TestGuardianGetActive(t *testing.T) {
	repo, pool := setupGuardianRepo(t)
	ctx := context.Background()

	guardID := uuid.MustParse("a4000000-0000-0000-0000-000000000001")
	dbtest.InsertGuardian(t, pool, guardID, guardTenantID, guardBranchID, "John", true)

	tx := dbtest.BeginTx(t, pool)
	isActive, found, err := repo.GetActive(ctx, tx, guardTenantID, guardBranchID, guardID)
	if err != nil {
		t.Fatalf("GetActive: %v", err)
	}
	if !found {
		t.Error("found = false")
	}
	if !isActive {
		t.Error("isActive = false, want true")
	}

	_, found, err = repo.GetActive(ctx, tx, uuid.New(), guardBranchID, guardID)
	if err != nil {
		t.Fatalf("GetActive wrong scope: %v", err)
	}
	if found {
		t.Error("found = true for wrong scope")
	}
}

func TestGuardianDeactivate(t *testing.T) {
	repo, pool := setupGuardianRepo(t)
	ctx := context.Background()

	guardID := uuid.MustParse("a4000000-0000-0000-0000-000000000001")
	dbtest.InsertGuardian(t, pool, guardID, guardTenantID, guardBranchID, "John", true)

	tx := dbtest.BeginTx(t, pool)
	if err := repo.Deactivate(ctx, tx, guardTenantID, guardBranchID, guardID, "access_revoked", "left nursery"); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var isActive bool
	var reasonCode, reasonNote *string
	var deactivatedAt *string
	err := pool.QueryRow(ctx,
		"SELECT is_active, deactivation_reason_code::text, deactivation_reason_note, deactivated_at::text FROM guardians WHERE id = $1",
		guardID).Scan(&isActive, &reasonCode, &reasonNote, &deactivatedAt)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if isActive {
		t.Error("is_active = true, want false")
	}
	if reasonCode == nil || *reasonCode != "access_revoked" {
		t.Errorf("reason_code = %v, want access_revoked", reasonCode)
	}
	if reasonNote == nil || *reasonNote != "left nursery" {
		t.Errorf("reason_note = %v, want left nursery", reasonNote)
	}
	if deactivatedAt == nil {
		t.Error("deactivated_at = nil, want non-nil")
	}
}

func TestGuardianCascadeLinks(t *testing.T) {
	repo, pool := setupGuardianRepo(t)
	ctx := context.Background()

	guardID := uuid.MustParse("a4000000-0000-0000-0000-000000000001")
	childID := uuid.MustParse("a5000000-0000-0000-0000-000000000001")
	linkID := uuid.MustParse("a6000000-0000-0000-0000-000000000001")
	otherChildID := uuid.MustParse("a5000000-0000-0000-0000-000000000002")
	otherLinkID := uuid.MustParse("a6000000-0000-0000-0000-000000000002")

	dbtest.InsertChild(t, pool, childID, guardTenantID, guardBranchID, "Child 1",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), true)
	dbtest.InsertChild(t, pool, otherChildID, guardTenantID, guardBranchID, "Child 2",
		dbtest.DateAt(2023, 3, 20), dbtest.DateAt(2024, 9, 1), true)
	dbtest.InsertGuardian(t, pool, guardID, guardTenantID, guardBranchID, "John", true)
	dbtest.InsertGuardianLink(t, pool, linkID, guardTenantID, guardBranchID, guardID, childID)
	dbtest.InsertGuardianLink(t, pool, otherLinkID, guardTenantID, guardBranchID, guardID, otherChildID)

	tx := dbtest.BeginTx(t, pool)
	if err := repo.CascadeLinks(ctx, tx, guardTenantID, guardBranchID, guardID, "access_revoked", ""); err != nil {
		t.Fatalf("CascadeLinks: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var count int
	err := pool.QueryRow(ctx,
		"SELECT count(*) FROM guardian_child_links WHERE guardian_id = $1 AND ended_at IS NULL", guardID).Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 0 {
		t.Errorf("active links = %d, want 0", count)
	}
}

func TestGuardianCascadeMappings(t *testing.T) {
	repo, pool := setupGuardianRepo(t)
	ctx := context.Background()

	guardID := uuid.MustParse("a4000000-0000-0000-0000-000000000001")
	membershipID := uuid.MustParse("a7000000-0000-0000-0000-000000000001")
	mappingID := uuid.MustParse("a8000000-0000-0000-0000-000000000001")

	dbtest.InsertMembership(t, pool, membershipID, guardTenantID, guardBranchID, guardUserID, "parent", true)
	dbtest.InsertGuardian(t, pool, guardID, guardTenantID, guardBranchID, "John", true)
	dbtest.InsertParentMapping(t, pool, mappingID, guardTenantID, guardBranchID, membershipID, guardID)

	tx := dbtest.BeginTx(t, pool)
	if err := repo.CascadeMappings(ctx, tx, guardTenantID, guardBranchID, guardID, "access_revoked", ""); err != nil {
		t.Fatalf("CascadeMappings: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var count int
	err := pool.QueryRow(ctx,
		"SELECT count(*) FROM parent_membership_guardians WHERE guardian_id = $1 AND ended_at IS NULL", guardID).Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 0 {
		t.Errorf("active mappings = %d, want 0", count)
	}
}

func TestGuardianReactivate(t *testing.T) {
	repo, pool := setupGuardianRepo(t)
	ctx := context.Background()

	guardID := uuid.MustParse("a4000000-0000-0000-0000-000000000001")
	_, err := pool.Exec(ctx,
		"INSERT INTO guardians (id, tenant_id, branch_id, full_name, is_active, deactivated_at, deactivation_reason_code) VALUES ($1, $2, $3, $4, false, now(), 'access_revoked')",
		guardID, guardTenantID, guardBranchID, "John")
	if err != nil {
		t.Fatalf("insert deactivated: %v", err)
	}

	tx := dbtest.BeginTx(t, pool)
	if err := repo.Reactivate(ctx, tx, guardTenantID, guardBranchID, guardID); err != nil {
		t.Fatalf("Reactivate: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var isActive bool
	var deactivatedAt, reasonCode, reasonNote *string
	err = pool.QueryRow(ctx,
		"SELECT is_active, deactivated_at::text, deactivation_reason_code::text, deactivation_reason_note FROM guardians WHERE id = $1",
		guardID).Scan(&isActive, &deactivatedAt, &reasonCode, &reasonNote)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if !isActive {
		t.Error("is_active = false, want true")
	}
	if deactivatedAt != nil {
		t.Errorf("deactivated_at = %s, want nil", *deactivatedAt)
	}
	if reasonCode != nil {
		t.Errorf("reason_code = %s, want nil", *reasonCode)
	}
	if reasonNote != nil {
		t.Errorf("reason_note = %s, want nil", *reasonNote)
	}
}

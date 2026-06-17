package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	linkdomain "nursery-management-system/api/internal/modules/guardianlinks/domain"
	linkpostgres "nursery-management-system/api/internal/modules/guardianlinks/infrastructure/postgres"
	"nursery-management-system/api/internal/platform/dbtest"
)

var (
	linkTenantID = uuid.MustParse("b1000000-0000-0000-0000-000000000001")
	linkBranchID = uuid.MustParse("b2000000-0000-0000-0000-000000000001")
	linkUserID   = uuid.MustParse("b3000000-0000-0000-0000-000000000001")
)

func setupLinkRepo(t *testing.T) (*linkpostgres.GuardianChildLinkRepository, *pgxpool.Pool) {
	t.Helper()
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	dbtest.InsertTenant(t, pool, linkTenantID, "Link Tenant")
	dbtest.InsertBranch(t, pool, linkTenantID, linkBranchID, "Link Branch")
	dbtest.InsertUser(t, pool, linkUserID, "staff@example.com", "hash", true)

	return linkpostgres.NewGuardianChildLinkRepository(pool), pool
}

func seedLinkEntities(t testing.TB, pool *pgxpool.Pool) (guardianID, childID uuid.UUID) {
	t.Helper()
	guardianID = uuid.MustParse("b4000000-0000-0000-0000-000000000001")
	childID = uuid.MustParse("b5000000-0000-0000-0000-000000000001")

	dbtest.InsertGuardian(t, pool, guardianID, linkTenantID, linkBranchID, "Parent G", true)
	dbtest.InsertChild(t, pool, childID, linkTenantID, linkBranchID, "Child C",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), true)

	return guardianID, childID
}

func TestLinkFindActiveByPair_NoRow(t *testing.T) {
	repo, pool := setupLinkRepo(t)
	ctx := context.Background()
	guardID, childID := seedLinkEntities(t, pool)

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.FindActiveByPair(ctx, tx, linkTenantID, linkBranchID, guardID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true when no link exists")
	}
}

func TestLinkFindActiveByPair_EndedRow(t *testing.T) {
	repo, pool := setupLinkRepo(t)
	ctx := context.Background()
	guardID, childID := seedLinkEntities(t, pool)

	linkID := uuid.MustParse("b6000000-0000-0000-0000-000000000001")
	dbtest.InsertGuardianLink(t, pool, linkID, linkTenantID, linkBranchID, guardID, childID)

	_, err := pool.Exec(ctx,
		"UPDATE guardian_child_links SET ended_at = now(), ended_reason_code = 'access_revoked' WHERE id = $1", linkID)
	if err != nil {
		t.Fatalf("end link: %v", err)
	}

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.FindActiveByPair(ctx, tx, linkTenantID, linkBranchID, guardID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for ended link")
	}
}

func TestLinkFindActiveByPair_WrongTenant(t *testing.T) {
	repo, pool := setupLinkRepo(t)
	ctx := context.Background()
	guardID, childID := seedLinkEntities(t, pool)

	linkID := uuid.MustParse("b6000000-0000-0000-0000-000000000001")
	dbtest.InsertGuardianLink(t, pool, linkID, linkTenantID, linkBranchID, guardID, childID)

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.FindActiveByPair(ctx, tx, uuid.New(), linkBranchID, guardID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for wrong tenant")
	}
}

func TestLinkFindActiveByPair_Success(t *testing.T) {
	repo, pool := setupLinkRepo(t)
	ctx := context.Background()
	guardID, childID := seedLinkEntities(t, pool)

	linkID := uuid.MustParse("b6000000-0000-0000-0000-000000000001")
	dbtest.InsertGuardianLink(t, pool, linkID, linkTenantID, linkBranchID, guardID, childID)

	tx := dbtest.BeginTx(t, pool)
	link, found, err := repo.FindActiveByPair(ctx, tx, linkTenantID, linkBranchID, guardID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if link.ID != linkID {
		t.Errorf("ID = %s, want %s", link.ID, linkID)
	}
	if link.EndedAt != nil {
		t.Error("EndedAt = non-nil, want nil for active link")
	}
}

func TestLinkCreate(t *testing.T) {
	repo, pool := setupLinkRepo(t)
	ctx := context.Background()
	guardID, childID := seedLinkEntities(t, pool)

	linkID := uuid.MustParse("b6000000-0000-0000-0000-000000000001")
	link := linkdomain.GuardianChildLink{
		ID:         linkID,
		GuardianID: guardID,
		ChildID:    childID,
		TenantID:   linkTenantID,
		BranchID:   linkBranchID,
	}

	tx := dbtest.BeginTx(t, pool)
	if err := repo.Create(ctx, tx, link); err != nil {
		t.Fatalf("Create: %v", err)
	}
	dbtest.CommitTx(t, tx)

	tx2 := dbtest.BeginTx(t, pool)
	got, found, err := repo.FindActiveByPair(ctx, tx2, linkTenantID, linkBranchID, guardID, childID)
	if err != nil {
		t.Fatalf("FindActiveByPair: %v", err)
	}
	if !found {
		t.Fatal("link not found after create")
	}
	if got.ID != linkID {
		t.Errorf("ID = %s, want %s", got.ID, linkID)
	}
}

func TestLinkGetByIDForUpdate(t *testing.T) {
	repo, pool := setupLinkRepo(t)
	ctx := context.Background()
	guardID, childID := seedLinkEntities(t, pool)

	linkID := uuid.MustParse("b6000000-0000-0000-0000-000000000001")
	dbtest.InsertGuardianLink(t, pool, linkID, linkTenantID, linkBranchID, guardID, childID)

	tx := dbtest.BeginTx(t, pool)
	got, found, err := repo.GetByIDForUpdate(ctx, tx, linkTenantID, linkBranchID, linkID)
	if err != nil {
		t.Fatalf("GetByIDForUpdate: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if got.GuardianID != guardID {
		t.Errorf("GuardianID = %s, want %s", got.GuardianID, guardID)
	}
	if got.ChildID != childID {
		t.Errorf("ChildID = %s, want %s", got.ChildID, childID)
	}
	if got.EndedAt != nil {
		t.Error("EndedAt should be nil for active link")
	}
}

func TestLinkEnd(t *testing.T) {
	repo, pool := setupLinkRepo(t)
	ctx := context.Background()
	guardID, childID := seedLinkEntities(t, pool)

	linkID := uuid.MustParse("b6000000-0000-0000-0000-000000000001")
	dbtest.InsertGuardianLink(t, pool, linkID, linkTenantID, linkBranchID, guardID, childID)

	tx := dbtest.BeginTx(t, pool)
	if err := repo.End(ctx, tx, linkTenantID, linkBranchID, linkID, "access_revoked", "guardian left"); err != nil {
		t.Fatalf("End: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var endedReasonCode, endedReasonNote *string
	var endedAt *string
	err := pool.QueryRow(ctx,
		"SELECT ended_reason_code::text, ended_reason_note, ended_at::text FROM guardian_child_links WHERE id = $1",
		linkID).Scan(&endedReasonCode, &endedReasonNote, &endedAt)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if endedAt == nil {
		t.Error("ended_at = nil, want non-nil")
	}
	if endedReasonCode == nil || *endedReasonCode != "access_revoked" {
		t.Errorf("reason_code = %v, want access_revoked", endedReasonCode)
	}
	if endedReasonNote == nil || *endedReasonNote != "guardian left" {
		t.Errorf("reason_note = %v, want guardian left", endedReasonNote)
	}
}

func TestLinkEnd_WrongScope_NoOp(t *testing.T) {
	repo, pool := setupLinkRepo(t)
	ctx := context.Background()
	guardID, childID := seedLinkEntities(t, pool)

	linkID := uuid.MustParse("b6000000-0000-0000-0000-000000000001")
	dbtest.InsertGuardianLink(t, pool, linkID, linkTenantID, linkBranchID, guardID, childID)

	tx := dbtest.BeginTx(t, pool)
	if err := repo.End(ctx, tx, uuid.New(), linkBranchID, linkID, "access_revoked", ""); err != nil {
		t.Fatalf("End wrong scope: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var endedAt *string
	err := pool.QueryRow(ctx,
		"SELECT ended_at::text FROM guardian_child_links WHERE id = $1", linkID).Scan(&endedAt)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if endedAt != nil {
		t.Error("ended_at should remain nil for wrong scope")
	}
}

func TestLinkDuplicateActivePair(t *testing.T) {
	repo, pool := setupLinkRepo(t)
	ctx := context.Background()
	guardID, childID := seedLinkEntities(t, pool)

	linkID1 := uuid.MustParse("b6000000-0000-0000-0000-000000000001")
	linkID2 := uuid.MustParse("b6000000-0000-0000-0000-000000000002")
	dbtest.InsertGuardianLink(t, pool, linkID1, linkTenantID, linkBranchID, guardID, childID)

	link := linkdomain.GuardianChildLink{
		ID:         linkID2,
		GuardianID: guardID,
		ChildID:    childID,
		TenantID:   linkTenantID,
		BranchID:   linkBranchID,
	}
	tx := dbtest.BeginTx(t, pool)
	err := repo.Create(ctx, tx, link)
	if err == nil {
		t.Fatal("expected error for duplicate active pair, got nil")
	}
}

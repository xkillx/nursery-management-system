package postgres

import (
	"context"
	"testing"
	"time"

	"nursery-management-system/api/internal/modules/funding/domain"
	"nursery-management-system/api/internal/platform/dbtest"
	"nursery-management-system/api/internal/platform/uid"
)

func TestRepository_CreateAndGet(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uid.NewUUID()
	branchID := uid.NewUUID()
	childID := uid.NewUUID()

	dbtest.InsertTenant(t, pool, tenantID, "Test Nursery")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Main Branch")
	dbtest.InsertChild(t, pool, childID, tenantID, branchID, "Test Child",
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), true)

	repo := NewRepository(pool)
	tx := dbtest.BeginTx(t, pool)

	profile := domain.FundingProfile{
		ID:                     uid.NewUUID(),
		TenantID:               tenantID,
		BranchID:               branchID,
		ChildID:                childID,
		BillingMonth:           time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		FundedAllowanceMinutes: 570,
	}

	created, err := repo.Create(ctx, tx, profile)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got, found, err := repo.Get(ctx, tenantID, branchID, childID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !found {
		t.Fatal("expected profile to be found")
	}
	if got.FundedAllowanceMinutes != 570 {
		t.Errorf("allowance = %d, want 570", got.FundedAllowanceMinutes)
	}
	if got.ID != created.ID {
		t.Errorf("id mismatch")
	}
}

func TestRepository_WrongTenantReturnsNotFound(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uid.NewUUID()
	otherTenantID := uid.NewUUID()
	branchID := uid.NewUUID()
	childID := uid.NewUUID()

	dbtest.InsertTenant(t, pool, tenantID, "Tenant A")
	dbtest.InsertTenant(t, pool, otherTenantID, "Tenant B")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Branch")
	dbtest.InsertChild(t, pool, childID, tenantID, branchID, "Child",
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), true)

	repo := NewRepository(pool)
	tx := dbtest.BeginTx(t, pool)

	profile := domain.FundingProfile{
		ID:                     uid.NewUUID(),
		TenantID:               tenantID,
		BranchID:               branchID,
		ChildID:                childID,
		BillingMonth:           time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		FundedAllowanceMinutes: 300,
	}
	_, err := repo.Create(ctx, tx, profile)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	dbtest.CommitTx(t, tx)

	_, found, err := repo.Get(ctx, otherTenantID, branchID, childID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if found {
		t.Error("expected not found for wrong tenant")
	}
}

func TestRepository_ExplicitZeroAllowance(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uid.NewUUID()
	branchID := uid.NewUUID()
	childID := uid.NewUUID()

	dbtest.InsertTenant(t, pool, tenantID, "Test")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Branch")
	dbtest.InsertChild(t, pool, childID, tenantID, branchID, "Child",
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), true)

	repo := NewRepository(pool)
	tx := dbtest.BeginTx(t, pool)

	profile := domain.FundingProfile{
		ID:                     uid.NewUUID(),
		TenantID:               tenantID,
		BranchID:               branchID,
		ChildID:                childID,
		BillingMonth:           time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		FundedAllowanceMinutes: 0,
	}
	_, err := repo.Create(ctx, tx, profile)
	if err != nil {
		t.Fatalf("create zero allowance: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got, found, err := repo.Get(ctx, tenantID, branchID, childID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !found {
		t.Fatal("expected profile to be found")
	}
	if got.FundedAllowanceMinutes != 0 {
		t.Errorf("allowance = %d, want 0", got.FundedAllowanceMinutes)
	}
}

func TestRepository_ChildEnrollmentLookup(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID := uid.NewUUID()
	branchID := uid.NewUUID()
	childID := uid.NewUUID()

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	dbtest.InsertTenant(t, pool, tenantID, "Test")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Branch")
	dbtest.InsertChild(t, pool, childID, tenantID, branchID, "Child",
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		startDate, true)

	repo := NewRepository(pool)
	tx := dbtest.BeginTx(t, pool)

	enrollment, found, err := repo.GetChildEnrollmentForUpdate(ctx, tx, tenantID, branchID, childID)
	if err != nil {
		t.Fatalf("get enrollment: %v", err)
	}
	if !found {
		t.Fatal("expected enrollment to be found")
	}
	if enrollment.StartDate != startDate {
		t.Errorf("start_date = %v, want %v", enrollment.StartDate, startDate)
	}
	if enrollment.EndDate != nil {
		t.Errorf("end_date should be nil, got %v", enrollment.EndDate)
	}
}

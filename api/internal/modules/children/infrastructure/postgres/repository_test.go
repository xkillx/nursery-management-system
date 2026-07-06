package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	childdomain "nursery-management-system/api/internal/modules/children/domain"
	childpostgres "nursery-management-system/api/internal/modules/children/infrastructure/postgres"
	"nursery-management-system/api/internal/platform/dbtest"
)

var (
	childTenantID = uuid.MustParse("10000000-0000-0000-0000-000000000001")
	childBranchID = uuid.MustParse("20000000-0000-0000-0000-000000000001")
	childUserID   = uuid.MustParse("30000000-0000-0000-0000-000000000001")
)

func setupChildRepo(t *testing.T) (*childpostgres.ChildRepository, *pgxpool.Pool) {
	t.Helper()
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	dbtest.InsertTenant(t, pool, childTenantID, "Child Tenant")
	dbtest.InsertBranch(t, pool, childTenantID, childBranchID, "Child Branch")
	dbtest.InsertUser(t, pool, childUserID, "staff@example.com", "hash", true)

	return childpostgres.NewChildRepository(pool), pool
}

func TestChildGetByID_NotFound(t *testing.T) {
	repo, _ := setupChildRepo(t)

	_, found, err := repo.GetByID(context.Background(), childTenantID, childBranchID, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for nonexistent child")
	}
}

func TestChildGetByID_WrongTenant(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	dbtest.InsertChild(t, pool, childID, childTenantID, childBranchID, "Alice",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), true)

	_, found, err := repo.GetByID(ctx, uuid.New(), childBranchID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for wrong tenant")
	}
}

func TestChildGetByID_Success(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	dbtest.InsertChild(t, pool, childID, childTenantID, childBranchID, "Alice",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), true)

	child, found, err := repo.GetByID(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if child.FirstName != "Alice" {
		t.Errorf("FirstName = %s, want Alice", child.FirstName)
	}
	if !child.IsActive {
		t.Error("IsActive = false, want true")
	}
}

func TestChildList_FilterActive(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	activeID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	inactiveID := uuid.MustParse("40000000-0000-0000-0000-000000000002")
	dbtest.InsertChild(t, pool, activeID, childTenantID, childBranchID, "Active Child",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), true)

	_, err := pool.Exec(ctx,
		"INSERT INTO children (id, tenant_id, branch_id, first_name, date_of_birth, start_date, is_active) VALUES ($1, $2, $3, $4, $5, $6, false)",
		inactiveID, childTenantID, childBranchID, "Inactive Child",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1),
	)
	if err != nil {
		t.Fatalf("insert inactive: %v", err)
	}

	rows, err := repo.List(ctx, childTenantID, childBranchID, childdomain.StatusActive, 50, 0, nil)
	if err != nil {
		t.Fatalf("List active: %v", err)
	}
	if len(rows) != 1 {
		t.Errorf("len(rows) = %d, want 1", len(rows))
	}

	rows, err = repo.List(ctx, childTenantID, childBranchID, childdomain.StatusAll, 50, 0, nil)
	if err != nil {
		t.Fatalf("List all: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("len(rows) = %d, want 2", len(rows))
	}
}

func TestChildCreate(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.New()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin tx: %v", err)
	}
	defer tx.Rollback(ctx)
	if err := repo.Create(ctx, tx, childdomain.Child{
		ID:          childID,
		FirstName:   "Bob",
		DateOfBirth: dbtest.DateAt(2021, 5, 1),
		StartDate:   dbtest.DateAt(2024, 9, 1),
		IsActive:    true,
	}, "my notes", childTenantID, childBranchID); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	got, found, err := repo.GetByID(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !found {
		t.Fatal("not found")
	}
	if got.FirstName != "Bob" {
		t.Errorf("FirstName = %s, want Bob", got.FirstName)
	}
}

func TestChildUpdate_FirstName(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	dbtest.InsertChild(t, pool, childID, childTenantID, childBranchID, "Alice",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), true)

	rows, err := repo.Update(ctx, childTenantID, childBranchID, childID, map[string]any{
		"first_name": "Alicia",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if rows == 0 {
		t.Fatal("Update affected 0 rows")
	}

	got, _, err := repo.GetByID(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.FirstName != "Alicia" {
		t.Errorf("FirstName = %s, want Alicia", got.FirstName)
	}
}

func TestChildUpdate_NullNotes(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	notes := "some notes"
	dbtest.InsertChildWithNotes(t, pool, childID, childTenantID, childBranchID, "Child",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), true, &notes)

	_, err := repo.Update(ctx, childTenantID, childBranchID, childID, map[string]any{
		"notes": "",
	})
	if err != nil {
		t.Fatalf("Update notes: %v", err)
	}

	var gotNotes *string
	err = pool.QueryRow(ctx, "SELECT notes FROM children WHERE id = $1", childID).Scan(&gotNotes)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if gotNotes != nil {
		t.Errorf("notes = %v, want nil after clearing", gotNotes)
	}
}

func TestChildMarkInactive(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	dbtest.InsertChild(t, pool, childID, childTenantID, childBranchID, "Alice",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), true)

	tx := dbtest.BeginTx(t, pool)
	if err := repo.MarkInactive(ctx, tx, childTenantID, childBranchID, childID); err != nil {
		t.Fatalf("MarkInactive: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var isActive bool
	err := pool.QueryRow(ctx,
		"SELECT is_active FROM children WHERE id = $1", childID,
	).Scan(&isActive)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if isActive {
		t.Error("is_active = true, want false")
	}
}

func TestChildGetByIDForUpdate_ScopeCheck(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	dbtest.InsertChild(t, pool, childID, childTenantID, childBranchID, "Alice",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), true)

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.GetByIDForUpdate(ctx, tx, uuid.New(), childBranchID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for wrong tenant")
	}
	dbtest.CommitTx(t, tx)
}

func TestChildExistsInScope(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	dbtest.InsertChild(t, pool, childID, childTenantID, childBranchID, "Alice",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), true)

	tx := dbtest.BeginTx(t, pool)
	exists, err := repo.ExistsInScope(ctx, tx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("ExistsInScope: %v", err)
	}
	if !exists {
		t.Error("exists = false, want true")
	}

	exists, err = repo.ExistsInScope(ctx, tx, uuid.New(), childBranchID, childID)
	if err != nil {
		t.Fatalf("ExistsInScope wrong scope: %v", err)
	}
	if exists {
		t.Error("exists = true, want false (wrong tenant)")
	}
	dbtest.CommitTx(t, tx)
}

package postgres_test

import (
	"context"
	"fmt"
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
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)

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
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)

	child, found, err := repo.GetByID(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if child.FullName != "Alice" {
		t.Errorf("FullName = %s, want Alice", child.FullName)
	}
	if !child.IsActive {
		t.Error("IsActive = false, want true")
	}
	if child.CoreHourlyRateMinor != 500 {
		t.Errorf("CoreHourlyRateMinor = %d, want 500", child.CoreHourlyRateMinor)
	}
}

func TestChildList_FilterActive(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	activeID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	inactiveID := uuid.MustParse("40000000-0000-0000-0000-000000000002")
	dbtest.InsertChild(t, pool, activeID, childTenantID, childBranchID, "Active Child",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)

	_, err := pool.Exec(ctx,
		"INSERT INTO children (id, tenant_id, branch_id, full_name, date_of_birth, start_date, core_hourly_rate_minor, is_active, left_at, left_reason_code) VALUES ($1, $2, $3, $4, $5, $6, $7, false, now(), 'left_nursery')",
		inactiveID, childTenantID, childBranchID, "Inactive Child",
		dbtest.DateAt(2023, 3, 20), dbtest.DateAt(2024, 9, 1), 500)
	if err != nil {
		t.Fatalf("insert inactive child: %v", err)
	}

	children, err := repo.List(ctx, childTenantID, childBranchID, childdomain.StatusActive, 10, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(children) != 1 {
		t.Fatalf("expected 1 active child, got %d", len(children))
	}
	if children[0].FullName != "Active Child" {
		t.Errorf("FullName = %s, want Active Child", children[0].FullName)
	}
}

func TestChildList_FilterInactive(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	activeID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	inactiveID := uuid.MustParse("40000000-0000-0000-0000-000000000002")
	dbtest.InsertChild(t, pool, activeID, childTenantID, childBranchID, "Active",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)
	_, err := pool.Exec(ctx,
		"INSERT INTO children (id, tenant_id, branch_id, full_name, date_of_birth, start_date, core_hourly_rate_minor, is_active, left_at, left_reason_code) VALUES ($1, $2, $3, $4, $5, $6, $7, false, now(), 'left_nursery')",
		inactiveID, childTenantID, childBranchID, "Inactive",
		dbtest.DateAt(2023, 3, 20), dbtest.DateAt(2024, 9, 1), 500)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	children, err := repo.List(ctx, childTenantID, childBranchID, childdomain.StatusInactive, 10, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(children) != 1 {
		t.Fatalf("expected 1 inactive, got %d", len(children))
	}
	if children[0].FullName != "Inactive" {
		t.Errorf("FullName = %s, want Inactive", children[0].FullName)
	}
}

func TestChildList_Pagination(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		id := uuid.MustParse(fmt.Sprintf("40000000-0000-0000-%04d-000000000001", i+1))
		dbtest.InsertChild(t, pool, id, childTenantID, childBranchID, fmt.Sprintf("Child %d", i),
			dbtest.DateAt(2022, 1, 15+i), dbtest.DateAt(2024, 9, 1), 500, true)
	}

	children, err := repo.List(ctx, childTenantID, childBranchID, childdomain.StatusAll, 3, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(children) != 3 {
		t.Fatalf("expected 3, got %d", len(children))
	}

	children2, err := repo.List(ctx, childTenantID, childBranchID, childdomain.StatusAll, 3, 3)
	if err != nil {
		t.Fatalf("List offset: %v", err)
	}
	if len(children2) != 2 {
		t.Fatalf("expected 2, got %d", len(children2))
	}
}

func TestChildCreate_NullNotes(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	child := childdomain.Child{
		ID:                  childID,
		FullName:            "NoNotes",
		DateOfBirth:         dbtest.DateAt(2022, 6, 10),
		StartDate:           dbtest.DateAt(2024, 9, 1),
		CoreHourlyRateMinor: 400,
	}

	if err := repo.Create(ctx, child, "", childTenantID, childBranchID); err != nil {
		t.Fatalf("Create: %v", err)
	}

	var notes *string
	err := pool.QueryRow(ctx, "SELECT notes FROM children WHERE id = $1", childID).Scan(&notes)
	if err != nil {
		t.Fatalf("query notes: %v", err)
	}
	if notes != nil {
		t.Errorf("notes = %v, want nil", notes)
	}
}

func TestChildCreate_WithEndDate(t *testing.T) {
	repo, _ := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	endDate := dbtest.DateAt(2025, 7, 31)
	child := childdomain.Child{
		ID:                  childID,
		FullName:            "WithEnd",
		DateOfBirth:         dbtest.DateAt(2022, 6, 10),
		StartDate:           dbtest.DateAt(2024, 9, 1),
		EndDate:             &endDate,
		CoreHourlyRateMinor: 400,
	}

	if err := repo.Create(ctx, child, "", childTenantID, childBranchID); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, found, err := repo.GetByID(ctx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !found {
		t.Fatal("not found")
	}
	if got.EndDate == nil {
		t.Fatal("EndDate = nil, want non-nil")
	}
}

func TestChildUpdate_SelectiveFields(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	dbtest.InsertChild(t, pool, childID, childTenantID, childBranchID, "Original",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)

	ct, err := repo.Update(ctx, childTenantID, childBranchID, childID, map[string]any{
		"full_name": "Updated",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if ct != 1 {
		t.Errorf("rows affected = %d, want 1", ct)
	}

	got, _, _ := repo.GetByID(ctx, childTenantID, childBranchID, childID)
	if got.FullName != "Updated" {
		t.Errorf("FullName = %s, want Updated", got.FullName)
	}
	if got.CoreHourlyRateMinor != 500 {
		t.Errorf("CoreHourlyRateMinor = %d, want 500 (unchanged)", got.CoreHourlyRateMinor)
	}
}

func TestChildUpdate_WrongScope(t *testing.T) {
	repo, _ := setupChildRepo(t)

	ct, err := repo.Update(context.Background(), uuid.New(), childBranchID, uuid.New(), map[string]any{
		"full_name": "X",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if ct != 0 {
		t.Errorf("rows affected = %d, want 0", ct)
	}
}

func TestChildUpdate_NullNotes(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	notes := "some notes"
	dbtest.InsertChildWithNotes(t, pool, childID, childTenantID, childBranchID, "Child",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true, &notes)

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
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)

	tx := dbtest.BeginTx(t, pool)
	if err := repo.MarkInactive(ctx, tx, childTenantID, childBranchID, childID, "left_nursery", "family moved"); err != nil {
		t.Fatalf("MarkInactive: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var isActive bool
	var leftReasonCode, leftReasonNote *string
	err := pool.QueryRow(ctx,
		"SELECT is_active, left_reason_code::text, left_reason_note FROM children WHERE id = $1", childID,
	).Scan(&isActive, &leftReasonCode, &leftReasonNote)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if isActive {
		t.Error("is_active = true, want false")
	}
	if leftReasonCode == nil || *leftReasonCode != "left_nursery" {
		t.Errorf("left_reason_code = %v, want left_nursery", leftReasonCode)
	}
	if leftReasonNote == nil || *leftReasonNote != "family moved" {
		t.Errorf("left_reason_note = %v, want family moved", leftReasonNote)
	}
}

func TestChildGetByIDForUpdate_ScopeCheck(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	dbtest.InsertChild(t, pool, childID, childTenantID, childBranchID, "Alice",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)

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
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)

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
		t.Error("exists = true for wrong tenant")
	}
	dbtest.CommitTx(t, tx)
}

func TestChildGetForCorrection(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	endDate := dbtest.DateAt(2025, 8, 31)
	dbtest.InsertChildWithNotes(t, pool, childID, childTenantID, childBranchID, "Alice",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true, nil)

	_, err := pool.Exec(ctx,
		"UPDATE children SET end_date = $1 WHERE id = $2", endDate, childID)
	if err != nil {
		t.Fatalf("update end_date: %v", err)
	}

	tx := dbtest.BeginTx(t, pool)
	info, found, err := repo.GetChildForCorrection(ctx, tx, childTenantID, childBranchID, childID)
	if err != nil {
		t.Fatalf("GetChildForCorrection: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if info.ID != childID {
		t.Errorf("ID = %s, want %s", info.ID, childID)
	}
	if info.EndDate == nil {
		t.Fatal("EndDate = nil")
	}
	dbtest.CommitTx(t, tx)
}

func TestChildGetForCorrection_WrongScope(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.GetChildForCorrection(ctx, tx, uuid.New(), childBranchID, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for wrong scope")
	}
	dbtest.CommitTx(t, tx)
}

func TestChildListAttendance(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("40000000-0000-0000-0000-000000000001")
	guardianID := uuid.MustParse("50000000-0000-0000-0000-000000000001")
	linkID := uuid.MustParse("60000000-0000-0000-0000-000000000001")

	dbtest.InsertChild(t, pool, childID, childTenantID, childBranchID, "Alice",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)
	dbtest.InsertGuardian(t, pool, guardianID, childTenantID, childBranchID, "Parent", true)
	dbtest.InsertGuardianLink(t, pool, linkID, childTenantID, childBranchID, guardianID, childID)

	localDate := dbtest.DateAt(2025, 5, 15)
	children, err := repo.ListAttendance(ctx, childTenantID, childBranchID, localDate)
	if err != nil {
		t.Fatalf("ListAttendance: %v", err)
	}
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	c := children[0]
	if c.FullName != "Alice" {
		t.Errorf("FullName = %s, want Alice", c.FullName)
	}
	if c.AttendanceState != "not_checked_in" {
		t.Errorf("AttendanceState = %s, want not_checked_in", c.AttendanceState)
	}
	if c.OpenSessionID != nil {
		t.Errorf("OpenSessionID = %v, want nil", c.OpenSessionID)
	}
	if !c.EnrollmentComplete {
		t.Error("EnrollmentComplete = false, want true")
	}
}

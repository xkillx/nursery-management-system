package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/absence/domain"
	absencepostgres "nursery-management-system/api/internal/modules/absence/infrastructure/postgres"
	"nursery-management-system/api/internal/platform/dbtest"
)

var (
	absTenantID     = uuid.MustParse("10000000-0000-0000-0000-000000000100")
	absBranchID     = uuid.MustParse("20000000-0000-0000-0000-000000000100")
	absUserID       = uuid.MustParse("30000000-0000-0000-0000-000000000100")
	absMembershipID = uuid.MustParse("40000000-0000-0000-0000-000000000100")
)

func setupAbsenceRepo(t *testing.T) (*absencepostgres.AbsenceRepository, *pgxpool.Pool) {
	t.Helper()
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	dbtest.InsertTenant(t, pool, absTenantID, "Absence Tenant")
	dbtest.InsertBranch(t, pool, absTenantID, absBranchID, "Absence Branch")
	dbtest.InsertUser(t, pool, absUserID, "absence@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, absMembershipID, absTenantID, absBranchID, absUserID, "manager", true)

	return absencepostgres.NewAbsenceRepository(pool), pool
}

func TestAbsenceRepo_CreateAndFind(t *testing.T) {
	repo, pool := setupAbsenceRepo(t)
	ctx := context.Background()
	childID := uuid.New()
	markerID := uuid.New()
	localDate := dbtest.DateAt(2026, 5, 27)

	dbtest.InsertChild(t, pool, childID, absTenantID, absBranchID, "Test Child",
		dbtest.DateAt(2022, 1, 1), dbtest.DateAt(2024, 9, 1), true)

	tx := dbtest.BeginTx(t, pool)
	defer dbtest.CommitTx(t, tx)

	created, err := repo.Create(ctx, tx, domain.AbsenceMarker{
		ID:                   markerID,
		TenantID:             absTenantID,
		BranchID:             absBranchID,
		ChildID:              childID,
		LocalDate:            localDate,
		MarkedAt:             time.Now().UTC(),
		MarkedByUserID:       absUserID,
		MarkedByMembershipID: absMembershipID,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID != markerID {
		t.Errorf("ID = %s, want %s", created.ID, markerID)
	}

	found, ok, err := repo.FindActiveByChildDate(ctx, tx, absTenantID, absBranchID, childID, localDate)
	if err != nil {
		t.Fatalf("FindActive: %v", err)
	}
	if !ok {
		t.Fatal("expected found=true")
	}
	if found.ID != markerID {
		t.Errorf("found ID = %s, want %s", found.ID, markerID)
	}
}

func TestAbsenceRepo_FindActive_NotFound(t *testing.T) {
	repo, pool := setupAbsenceRepo(t)
	ctx := context.Background()
	tx := dbtest.BeginTx(t, pool)
	defer dbtest.CommitTx(t, tx)

	_, ok, err := repo.FindActiveByChildDate(ctx, tx, absTenantID, absBranchID, uuid.New(), dbtest.DateAt(2026, 5, 27))
	if err != nil {
		t.Fatalf("FindActive: %v", err)
	}
	if ok {
		t.Error("expected found=false")
	}
}

func TestAbsenceRepo_PartialUniqueIndex_OneActivePerChildDate(t *testing.T) {
	repo, pool := setupAbsenceRepo(t)
	ctx := context.Background()
	childID := uuid.New()
	localDate := dbtest.DateAt(2026, 5, 27)

	dbtest.InsertChild(t, pool, childID, absTenantID, absBranchID, "Test Child",
		dbtest.DateAt(2022, 1, 1), dbtest.DateAt(2024, 9, 1), true)

	dbtest.InsertAbsenceMarker(t, pool, uuid.New(), absTenantID, absBranchID, childID, absUserID, absMembershipID,
		localDate, time.Now().UTC(), nil, nil, nil)

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = repo.Create(ctx, tx, domain.AbsenceMarker{
		ID:                   uuid.New(),
		TenantID:             absTenantID,
		BranchID:             absBranchID,
		ChildID:              childID,
		LocalDate:            localDate,
		MarkedAt:             time.Now().UTC(),
		MarkedByUserID:       absUserID,
		MarkedByMembershipID: absMembershipID,
	})
	if err == nil {
		t.Fatal("expected unique constraint violation for duplicate active marker")
	}
}

func TestAbsenceRepo_ClearedMarker_AllowsNewActive(t *testing.T) {
	repo, pool := setupAbsenceRepo(t)
	ctx := context.Background()
	childID := uuid.New()
	localDate := dbtest.DateAt(2026, 5, 27)

	dbtest.InsertChild(t, pool, childID, absTenantID, absBranchID, "Test Child",
		dbtest.DateAt(2022, 1, 1), dbtest.DateAt(2024, 9, 1), true)

	clearedAt := time.Now().UTC()
	dbtest.InsertAbsenceMarker(t, pool, uuid.New(), absTenantID, absBranchID, childID, absUserID, absMembershipID,
		localDate, time.Now().UTC().Add(-time.Hour), &clearedAt, &absUserID, &absMembershipID)

	tx := dbtest.BeginTx(t, pool)
	defer dbtest.CommitTx(t, tx)

	newMarker, err := repo.Create(ctx, tx, domain.AbsenceMarker{
		ID:                   uuid.New(),
		TenantID:             absTenantID,
		BranchID:             absBranchID,
		ChildID:              childID,
		LocalDate:            localDate,
		MarkedAt:             time.Now().UTC(),
		MarkedByUserID:       absUserID,
		MarkedByMembershipID: absMembershipID,
	})
	if err != nil {
		t.Fatalf("Create after clear: %v", err)
	}

	found, ok, err := repo.FindActiveByChildDate(ctx, tx, absTenantID, absBranchID, childID, localDate)
	if err != nil {
		t.Fatalf("FindActive: %v", err)
	}
	if !ok {
		t.Fatal("expected found=true for new active marker")
	}
	if found.ID != newMarker.ID {
		t.Errorf("found ID = %s, want new marker %s", found.ID, newMarker.ID)
	}
}

func TestAbsenceRepo_Clear_ActiveMarker(t *testing.T) {
	repo, pool := setupAbsenceRepo(t)
	ctx := context.Background()
	childID := uuid.New()
	markerID := uuid.New()
	localDate := dbtest.DateAt(2026, 5, 27)

	dbtest.InsertChild(t, pool, childID, absTenantID, absBranchID, "Test Child",
		dbtest.DateAt(2022, 1, 1), dbtest.DateAt(2024, 9, 1), true)

	dbtest.InsertAbsenceMarker(t, pool, markerID, absTenantID, absBranchID, childID, absUserID, absMembershipID,
		localDate, time.Now().UTC(), nil, nil, nil)

	tx := dbtest.BeginTx(t, pool)
	defer dbtest.CommitTx(t, tx)

	cleared, ok, err := repo.Clear(ctx, tx, absTenantID, absBranchID, markerID, time.Now().UTC(), absUserID, absMembershipID)
	if err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true")
	}
	if cleared.ClearedAt == nil {
		t.Fatal("expected cleared_at to be set")
	}

	_, ok, _ = repo.FindActiveByChildDate(ctx, tx, absTenantID, absBranchID, childID, localDate)
	if ok {
		t.Error("expected no active marker after clear")
	}
}

func TestAbsenceRepo_Clear_AlreadyCleared(t *testing.T) {
	repo, pool := setupAbsenceRepo(t)
	ctx := context.Background()
	childID := uuid.New()
	markerID := uuid.New()
	localDate := dbtest.DateAt(2026, 5, 27)

	dbtest.InsertChild(t, pool, childID, absTenantID, absBranchID, "Test Child",
		dbtest.DateAt(2022, 1, 1), dbtest.DateAt(2024, 9, 1), true)

	clearedAt := time.Now().UTC()
	dbtest.InsertAbsenceMarker(t, pool, markerID, absTenantID, absBranchID, childID, absUserID, absMembershipID,
		localDate, time.Now().UTC().Add(-time.Hour), &clearedAt, &absUserID, &absMembershipID)

	tx := dbtest.BeginTx(t, pool)
	defer dbtest.CommitTx(t, tx)

	_, ok, err := repo.Clear(ctx, tx, absTenantID, absBranchID, markerID, time.Now().UTC(), absUserID, absMembershipID)
	if err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if ok {
		t.Error("expected ok=false for already cleared marker")
	}
}

func TestAbsenceRepo_HasAttendance_True(t *testing.T) {
	repo, pool := setupAbsenceRepo(t)
	ctx := context.Background()
	childID := uuid.New()
	sessionID := uuid.New()
	localDate := dbtest.DateAt(2026, 5, 27)

	dbtest.InsertChild(t, pool, childID, absTenantID, absBranchID, "Test Child",
		dbtest.DateAt(2022, 1, 1), dbtest.DateAt(2024, 9, 1), true)
	dbtest.InsertAttendanceSession(t, pool, sessionID, absTenantID, absBranchID, childID, "open",
		time.Now().UTC(), localDate)

	tx := dbtest.BeginTx(t, pool)
	defer dbtest.CommitTx(t, tx)

	has, err := repo.HasAttendanceForChildDate(ctx, tx, absTenantID, absBranchID, childID, localDate)
	if err != nil {
		t.Fatalf("HasAttendance: %v", err)
	}
	if !has {
		t.Error("expected has=true when attendance exists")
	}
}

func TestAbsenceRepo_HasAttendance_False(t *testing.T) {
	repo, pool := setupAbsenceRepo(t)
	ctx := context.Background()

	tx := dbtest.BeginTx(t, pool)
	defer dbtest.CommitTx(t, tx)

	has, err := repo.HasAttendanceForChildDate(ctx, tx, absTenantID, absBranchID, uuid.New(), dbtest.DateAt(2026, 5, 27))
	if err != nil {
		t.Fatalf("HasAttendance: %v", err)
	}
	if has {
		t.Error("expected has=false when no attendance")
	}
}

func TestAbsenceRepo_GetByID_WrongScope(t *testing.T) {
	repo, pool := setupAbsenceRepo(t)
	ctx := context.Background()
	childID := uuid.New()
	markerID := uuid.New()

	dbtest.InsertChild(t, pool, childID, absTenantID, absBranchID, "Test Child",
		dbtest.DateAt(2022, 1, 1), dbtest.DateAt(2024, 9, 1), true)
	dbtest.InsertAbsenceMarker(t, pool, markerID, absTenantID, absBranchID, childID, absUserID, absMembershipID,
		dbtest.DateAt(2026, 5, 27), time.Now().UTC(), nil, nil, nil)

	tx := dbtest.BeginTx(t, pool)
	defer dbtest.CommitTx(t, tx)

	_, ok, err := repo.GetByID(ctx, tx, uuid.New(), absBranchID, markerID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if ok {
		t.Error("expected found=false for wrong tenant")
	}
}

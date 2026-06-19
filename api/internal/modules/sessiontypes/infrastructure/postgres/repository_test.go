package postgres_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	sessiontypepostgres "nursery-management-system/api/internal/modules/sessiontypes/infrastructure/postgres"
	"nursery-management-system/api/internal/platform/dbtest"
)

var (
	stTenantID = uuid.MustParse("60000000-0000-0000-0000-000000000001")
	stBranchID = uuid.MustParse("60000000-0000-0000-0000-000000000002")
)

func setupSessionTypeRepo(t *testing.T) (*sessiontypepostgres.SessionTypeRepository, *pgxpool.Pool) {
	t.Helper()
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	dbtest.InsertTenant(t, pool, stTenantID, "ST Tenant")
	dbtest.InsertBranch(t, pool, stTenantID, stBranchID, "ST Branch")

	return sessiontypepostgres.NewRepository(pool), pool
}

func TestSessionType_TimeCheckConstraint(t *testing.T) {
	_, pool := setupSessionTypeRepo(t)
	ctx := context.Background()

	// Insert with start == end — should violate CHECK (start_time < end_time).
	_, err := pool.Exec(ctx,
		`INSERT INTO session_types (id, tenant_id, branch_id, name, start_time, end_time, is_active)
		 VALUES ($1, $2, $3, $4, '08:00'::time, '08:00'::time, true)`,
		uuid.New(), stTenantID, stBranchID, "Bad")
	if err == nil {
		t.Fatal("expected CHECK violation")
	}
	if !strings.Contains(err.Error(), "session_types_time_check") {
		t.Errorf("got err %v, want session_types_time_check", err)
	}
}

func TestSessionType_ActiveNameUnique(t *testing.T) {
	_, pool := setupSessionTypeRepo(t)
	ctx := context.Background()

	dbtest.InsertSessionType(t, pool, uuid.New(), stTenantID, stBranchID, "Morning", 8*60, 13*60, true)

	// Insert another active with the same name.
	_, err := pool.Exec(ctx,
		`INSERT INTO session_types (id, tenant_id, branch_id, name, start_time, end_time, is_active)
		 VALUES ($1, $2, $3, $4, '13:00'::time, '18:00'::time, true)`,
		uuid.New(), stTenantID, stBranchID, "Morning")
	if err == nil {
		t.Fatal("expected unique violation on active name")
	}
}

func TestSessionType_ArchivedNameReused(t *testing.T) {
	_, pool := setupSessionTypeRepo(t)
	ctx := context.Background()

	dbtest.InsertSessionType(t, pool, uuid.New(), stTenantID, stBranchID, "Morning", 8*60, 13*60, false)

	// Re-using the same name while another is archived should be allowed.
	_, err := pool.Exec(ctx,
		`INSERT INTO session_types (id, tenant_id, branch_id, name, start_time, end_time, is_active)
		 VALUES ($1, $2, $3, $4, '08:00'::time, '13:00'::time, true)`,
		uuid.New(), stTenantID, stBranchID, "Morning")
	if err != nil {
		t.Fatalf("re-using archived name should be allowed: %v", err)
	}
}

func TestSessionType_RepoListArchiveReactivate(t *testing.T) {
	repo, pool := setupSessionTypeRepo(t)
	ctx := context.Background()

	stID := uuid.New()
	dbtest.InsertSessionType(t, pool, stID, stTenantID, stBranchID, "X", 8*60, 13*60, true)

	// Default list (no archived) returns the type.
	types, err := repo.ListByBranch(ctx, stTenantID, stBranchID, false)
	if err != nil {
		t.Fatalf("ListByBranch: %v", err)
	}
	if len(types) != 1 {
		t.Errorf("got %d, want 1", len(types))
	}

	// Archive.
	tx := dbtest.BeginTx(t, pool)
	if err := repo.Archive(ctx, tx, stTenantID, stBranchID, stID); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	dbtest.CommitTx(t, tx)

	// Now default list (no archived) is empty.
	types, err = repo.ListByBranch(ctx, stTenantID, stBranchID, false)
	if err != nil {
		t.Fatalf("ListByBranch after archive: %v", err)
	}
	if len(types) != 0 {
		t.Errorf("got %d active after archive, want 0", len(types))
	}

	// With includeArchived, we see it again.
	types, err = repo.ListByBranch(ctx, stTenantID, stBranchID, true)
	if err != nil {
		t.Fatalf("ListByBranch with includeArchived: %v", err)
	}
	if len(types) != 1 {
		t.Errorf("got %d with includeArchived, want 1", len(types))
	}

	// Reactivate.
	tx = dbtest.BeginTx(t, pool)
	if err := repo.Reactivate(ctx, tx, stTenantID, stBranchID, stID); err != nil {
		t.Fatalf("Reactivate: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got, err := repo.GetByID(ctx, stTenantID, stBranchID, stID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !got.IsActive {
		t.Error("expected IsActive true after reactivate")
	}
}

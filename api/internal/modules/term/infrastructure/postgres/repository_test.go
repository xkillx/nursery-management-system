package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/term/domain"
	termpostgres "nursery-management-system/api/internal/modules/term/infrastructure/postgres"
	"nursery-management-system/api/internal/platform/dbtest"
)

var (
	termTenantID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	termBranchID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func newTermTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return dbtest.RequirePostgres(t)
}

func setupTermTestData(t *testing.T, pool *pgxpool.Pool) (tenantID, branchID, childID, patternID, membershipID uuid.UUID) {
	t.Helper()
	dbtest.Reset(t, pool)
	tenantID = termTenantID
	branchID = termBranchID
	dbtest.InsertTenant(t, pool, tenantID, "Term Test Tenant")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Term Test Branch")
	dbtest.SetBranchCoreHourlyRate(t, pool, branchID, 750)

	membershipID = uuid.New()
	userID := uuid.New()
	dbtest.InsertUser(t, pool, userID, "termtest@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, membershipID, tenantID, branchID, userID, "manager", true)

	childID = uuid.New()
	dbtest.InsertChild(t, pool, childID, tenantID, branchID, "TermKid", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), true)

	// A booking pattern is a child_booking_patterns row. The application layer's
	// BookingPatternLookup just checks existence; we don't need entries for this
	// repo test, but the FK requires a child_booking_patterns row to exist.
	patternID = uuid.New()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO child_booking_patterns (id, tenant_id, branch_id, child_id, effective_from) VALUES ($1, $2, $3, $4, $5)",
		patternID, tenantID, branchID, childID, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("insert booking pattern: %v", err)
	}
	return
}

func TestTermRepository_InsertAndGetByID(t *testing.T) {
	pool := newTermTestPool(t)
	tenantID, branchID, childID, patternID, membershipID := setupTermTestData(t, pool)
	repo := termpostgres.NewTermRepository(pool)

	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	term, err := domain.NewTerm(uuid.New(), tenantID, branchID, childID, start, patternID, 750, membershipID)
	if err != nil {
		t.Fatalf("NewTerm: %v", err)
	}

	tx := dbtest.BeginTx(t, pool)
	saved, err := repo.Insert(context.Background(), tx, term)
	dbtest.CommitTx(t, tx)
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if saved.ID != term.ID {
		t.Errorf("ID mismatch: got %v want %v", saved.ID, term.ID)
	}

	got, found, err := repo.GetByID(context.Background(), tenantID, branchID, saved.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !found {
		t.Fatal("not found")
	}
	if got.ChildID != childID {
		t.Errorf("ChildID: got %v want %v", got.ChildID, childID)
	}
	wantEnd := time.Date(2027, 6, 30, 0, 0, 0, 0, time.UTC)
	if !got.TermEndDate.Equal(wantEnd) {
		t.Errorf("TermEndDate: got %v want %v", got.TermEndDate, wantEnd)
	}
}

func TestTermRepository_GetActiveForChild(t *testing.T) {
	pool := newTermTestPool(t)
	tenantID, branchID, childID, patternID, membershipID := setupTermTestData(t, pool)
	repo := termpostgres.NewTermRepository(pool)

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	term, err := domain.NewTerm(uuid.New(), tenantID, branchID, childID, start, patternID, 750, membershipID)
	if err != nil {
		t.Fatal(err)
	}
	tx := dbtest.BeginTx(t, pool)
	_, err = repo.Insert(context.Background(), tx, term)
	dbtest.CommitTx(t, tx)
	if err != nil {
		t.Fatal(err)
	}

	got, found, err := repo.GetActiveForChild(context.Background(), tenantID, branchID, childID)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("expected found")
	}
	if got.ID != term.ID {
		t.Errorf("ID mismatch")
	}
}

func TestTermRepository_Terminate(t *testing.T) {
	pool := newTermTestPool(t)
	tenantID, branchID, childID, patternID, membershipID := setupTermTestData(t, pool)
	repo := termpostgres.NewTermRepository(pool)

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	term, _ := domain.NewTerm(uuid.New(), tenantID, branchID, childID, start, patternID, 750, membershipID)

	tx := dbtest.BeginTx(t, pool)
	_, _ = repo.Insert(context.Background(), tx, term)
	now := time.Now().UTC()
	rows, err := repo.Terminate(context.Background(), tx, tenantID, branchID, term.ID, now, "left_nursery", "test note")
	dbtest.CommitTx(t, tx)
	if err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	got, _, _ := repo.GetByID(context.Background(), tenantID, branchID, term.ID)
	if got.Status != domain.TermStatusTerminated {
		t.Errorf("status: got %s, want terminated", got.Status)
	}
	if got.TerminationReasonCode == nil || *got.TerminationReasonCode != "left_nursery" {
		t.Errorf("reason code: %v", got.TerminationReasonCode)
	}
}

func TestTermRepository_ListExpiringWithin(t *testing.T) {
	pool := newTermTestPool(t)
	tenantID, branchID, childID, patternID, membershipID := setupTermTestData(t, pool)
	repo := termpostgres.NewTermRepository(pool)

	// Term ending 2025-12-31
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	term, _ := domain.NewTerm(uuid.New(), tenantID, branchID, childID, start, patternID, 750, membershipID)
	tx := dbtest.BeginTx(t, pool)
	_, _ = repo.Insert(context.Background(), tx, term)
	dbtest.CommitTx(t, tx)

	// Lookahead 60 days from 2025-11-01 should include 2025-12-31.
	today := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	cutoff := today.AddDate(0, 0, 60)
	got, err := repo.ListExpiringWithin(context.Background(), tenantID, branchID, cutoff)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 term, got %d", len(got))
	}
}

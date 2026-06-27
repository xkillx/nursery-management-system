package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	childdomain "nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/dbtest"
)

// newBookingPatternChildID returns a fresh child ID in the test fixture's
// tenant/branch (childTenantID, childBranchID from setupChildRepo).
func newBookingPatternChildID(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	return newChildID(t, pool, childTenantID, childBranchID, "BP Kid")
}

func insertSessionTypesForBP(t *testing.T, pool *pgxpool.Pool, stID1, stID2 uuid.UUID) {
	t.Helper()
	dbtest.InsertSessionType(t, pool, stID1, childTenantID, childBranchID, "Morning", 8*60, 13*60, true)
	dbtest.InsertSessionType(t, pool, stID2, childTenantID, childBranchID, "Afternoon", 13*60, 18*60, true)
}

func TestBookingPattern_InsertAndGetByID(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newBookingPatternChildID(t, pool)
	stID1 := uuid.New()
	stID2 := uuid.New()
	insertSessionTypesForBP(t, pool, stID1, stID2)

	tx := dbtest.BeginTx(t, pool)
	pattern := &childdomain.BookingPattern{
		ID:            uuid.New(),
		TenantID:      childTenantID,
		BranchID:      childBranchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	entries := []childdomain.BookingPatternEntry{
		{
			ID:        uuid.New(),
			DayOfWeek: 1,
			SessionType: &childdomain.EntrySessionType{
				ID: stID1,
			},
		},
	}
	saved, err := repo.InsertPattern(ctx, tx, pattern, entries)
	if err != nil {
		t.Fatalf("InsertPattern: %v", err)
	}
	if !saved.IsCurrent {
		t.Error("expected saved.IsCurrent true")
	}
	if len(saved.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(saved.Entries))
	}
	if saved.Entries[0].SessionType == nil {
		t.Fatal("expected entry.SessionType populated")
	}
	if saved.Entries[0].SessionType.Name != "Morning" {
		t.Errorf("expected session type name Morning, got %q", saved.Entries[0].SessionType.Name)
	}
	dbtest.CommitTx(t, tx)

	got, found, err := repo.GetPatternByID(ctx, childTenantID, childBranchID, saved.ID)
	if err != nil {
		t.Fatalf("GetPatternByID: %v", err)
	}
	if !found {
		t.Fatal("expected found true")
	}
	if got.ChildID != childID {
		t.Errorf("got childID %s, want %s", got.ChildID, childID)
	}
	if len(got.Entries) != 1 {
		t.Errorf("got %d entries, want 1", len(got.Entries))
	}
}

func TestBookingPattern_CloseCurrentAndGetActiveForDate(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newBookingPatternChildID(t, pool)
	stID1 := uuid.New()
	stID2 := uuid.New()
	insertSessionTypesForBP(t, pool, stID1, stID2)

	// Create first pattern (open) starting 2026-07-01.
	tx := dbtest.BeginTx(t, pool)
	first := &childdomain.BookingPattern{
		ID:            uuid.New(),
		TenantID:      childTenantID,
		BranchID:      childBranchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	firstEntries := []childdomain.BookingPatternEntry{
		{ID: uuid.New(), DayOfWeek: 1, SessionType: &childdomain.EntrySessionType{ID: stID1}},
	}
	_, err := repo.InsertPattern(ctx, tx, first, firstEntries)
	if err != nil {
		t.Fatalf("InsertPattern first: %v", err)
	}
	dbtest.CommitTx(t, tx)

	// Close current pattern at 2026-07-31, then create new from 2026-08-01.
	tx = dbtest.BeginTx(t, pool)
	if err := repo.CloseCurrentPattern(ctx, tx, childTenantID, childBranchID, childID, time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("CloseCurrentPattern: %v", err)
	}
	second := &childdomain.BookingPattern{
		ID:            uuid.New(),
		TenantID:      childTenantID,
		BranchID:      childBranchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	}
	secondEntries := []childdomain.BookingPatternEntry{
		{ID: uuid.New(), DayOfWeek: 2, SessionType: &childdomain.EntrySessionType{ID: stID2}},
	}
	_, err = repo.InsertPattern(ctx, tx, second, secondEntries)
	if err != nil {
		t.Fatalf("InsertPattern second: %v", err)
	}
	dbtest.CommitTx(t, tx)

	// ActiveForDate 2026-07-15 should return first.
	firstGot, found, err := repo.GetActiveForDate(ctx, childTenantID, childBranchID, childID, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetActiveForDate 07-15: %v", err)
	}
	if !found {
		t.Fatal("expected first found for 2026-07-15")
	}
	if firstGot.ID != first.ID {
		t.Errorf("expected first.ID, got %s", firstGot.ID)
	}

	// ActiveForDate 2026-08-15 should return second.
	secondGot, found, err := repo.GetActiveForDate(ctx, childTenantID, childBranchID, childID, time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("GetActiveForDate 08-15: %v", err)
	}
	if !found {
		t.Fatal("expected second found for 2026-08-15")
	}
	if secondGot.ID != second.ID {
		t.Errorf("expected second.ID, got %s", secondGot.ID)
	}
}

func TestBookingPattern_OneOpenPerChild(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newBookingPatternChildID(t, pool)
	stID1 := uuid.New()
	stID2 := uuid.New()
	insertSessionTypesForBP(t, pool, stID1, stID2)

	tx := dbtest.BeginTx(t, pool)
	_, err := repo.InsertPattern(ctx, tx, &childdomain.BookingPattern{
		ID:            uuid.New(),
		TenantID:      childTenantID,
		BranchID:      childBranchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}, []childdomain.BookingPatternEntry{
		{ID: uuid.New(), DayOfWeek: 1, SessionType: &childdomain.EntrySessionType{ID: stID1}},
	})
	if err != nil {
		t.Fatalf("first insert: %v", err)
	}
	dbtest.CommitTx(t, tx)

	// Try to insert a second open pattern without closing the first.
	tx = dbtest.BeginTx(t, pool)
	_, err = repo.InsertPattern(ctx, tx, &childdomain.BookingPattern{
		ID:            uuid.New(),
		TenantID:      childTenantID,
		BranchID:      childBranchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	}, []childdomain.BookingPatternEntry{
		{ID: uuid.New(), DayOfWeek: 2, SessionType: &childdomain.EntrySessionType{ID: stID2}},
	})
	if err == nil {
		t.Fatal("expected unique constraint violation, got nil")
	}
	tx.Rollback(ctx)
}

func TestBookingPattern_DuplicateDayAndSession(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newBookingPatternChildID(t, pool)
	stID1 := uuid.New()
	stID2 := uuid.New()
	insertSessionTypesForBP(t, pool, stID1, stID2)

	tx := dbtest.BeginTx(t, pool)
	_, err := repo.InsertPattern(ctx, tx, &childdomain.BookingPattern{
		ID:            uuid.New(),
		TenantID:      childTenantID,
		BranchID:      childBranchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}, []childdomain.BookingPatternEntry{
		{ID: uuid.New(), DayOfWeek: 1, SessionType: &childdomain.EntrySessionType{ID: stID1}},
		{ID: uuid.New(), DayOfWeek: 1, SessionType: &childdomain.EntrySessionType{ID: stID1}},
	})
	if err == nil {
		t.Fatal("expected duplicate day+session rejection")
	}
	tx.Rollback(ctx)
}

func TestBookingPattern_DuplicateDayRejected(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newBookingPatternChildID(t, pool)
	stID1 := uuid.New()
	stID2 := uuid.New()
	insertSessionTypesForBP(t, pool, stID1, stID2)

	tx := dbtest.BeginTx(t, pool)
	_, err := repo.InsertPattern(ctx, tx, &childdomain.BookingPattern{
		ID:            uuid.New(),
		TenantID:      childTenantID,
		BranchID:      childBranchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}, []childdomain.BookingPatternEntry{
		{ID: uuid.New(), DayOfWeek: 1, SessionType: &childdomain.EntrySessionType{ID: stID1}},
		{ID: uuid.New(), DayOfWeek: 1, SessionType: &childdomain.EntrySessionType{ID: stID2}},
	})
	if err == nil {
		t.Fatal("expected unique constraint violation, got nil")
	}
	tx.Rollback(ctx)
}

func TestBookingPattern_SingleEntryPerDaySuccess(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newBookingPatternChildID(t, pool)
	stID1 := uuid.New()
	stID2 := uuid.New()
	insertSessionTypesForBP(t, pool, stID1, stID2)

	tx := dbtest.BeginTx(t, pool)
	saved, err := repo.InsertPattern(ctx, tx, &childdomain.BookingPattern{
		ID:            uuid.New(),
		TenantID:      childTenantID,
		BranchID:      childBranchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}, []childdomain.BookingPatternEntry{
		{ID: uuid.New(), DayOfWeek: 1, SessionType: &childdomain.EntrySessionType{ID: stID1}},
		{ID: uuid.New(), DayOfWeek: 2, SessionType: &childdomain.EntrySessionType{ID: stID2}},
	})
	if err != nil {
		t.Fatalf("InsertPattern: %v", err)
	}
	if len(saved.Entries) != 2 {
		t.Errorf("got %d entries, want 2", len(saved.Entries))
	}
	dbtest.CommitTx(t, tx)
}

func TestBookingPattern_BranchMismatchEntryRejected(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newBookingPatternChildID(t, pool)
	// Insert a session type scoped to a DIFFERENT branch.
	otherBranchID := uuid.New()
	dbtest.InsertBranch(t, pool, childTenantID, otherBranchID, "Other Branch")
	wrongSTID := uuid.New()
	dbtest.InsertSessionType(t, pool, wrongSTID, childTenantID, otherBranchID, "WrongBranch", 8*60, 13*60, true)

	tx := dbtest.BeginTx(t, pool)
	_, err := repo.InsertPattern(ctx, tx, &childdomain.BookingPattern{
		ID:            uuid.New(),
		TenantID:      childTenantID,
		BranchID:      childBranchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}, []childdomain.BookingPatternEntry{
		{ID: uuid.New(), DayOfWeek: 1, SessionType: &childdomain.EntrySessionType{ID: wrongSTID}},
	})
	if err == nil {
		t.Fatal("expected composite FK rejection")
	}
	tx.Rollback(ctx)
}

func TestBookingPattern_ReplaceEntries(t *testing.T) {
	repo, pool := setupChildRepo(t)
	ctx := context.Background()

	childID := newBookingPatternChildID(t, pool)
	stID1 := uuid.New()
	stID2 := uuid.New()
	insertSessionTypesForBP(t, pool, stID1, stID2)

	tx := dbtest.BeginTx(t, pool)
	saved, err := repo.InsertPattern(ctx, tx, &childdomain.BookingPattern{
		ID:            uuid.New(),
		TenantID:      childTenantID,
		BranchID:      childBranchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}, []childdomain.BookingPatternEntry{
		{ID: uuid.New(), DayOfWeek: 1, SessionType: &childdomain.EntrySessionType{ID: stID1}},
	})
	if err != nil {
		t.Fatalf("InsertPattern: %v", err)
	}
	dbtest.CommitTx(t, tx)

	tx = dbtest.BeginTx(t, pool)
	if err := repo.ReplaceEntries(ctx, tx, childTenantID, childBranchID, saved.ID, []childdomain.BookingPatternEntry{
		{ID: uuid.New(), DayOfWeek: 3, SessionType: &childdomain.EntrySessionType{ID: stID2}},
	}); err != nil {
		t.Fatalf("ReplaceEntries: %v", err)
	}
	dbtest.CommitTx(t, tx)

	got, _, err := repo.GetPatternByID(ctx, childTenantID, childBranchID, saved.ID)
	if err != nil {
		t.Fatalf("GetPatternByID: %v", err)
	}
	if len(got.Entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(got.Entries))
	}
	if got.Entries[0].DayOfWeek != 3 {
		t.Errorf("got day_of_week %d, want 3", got.Entries[0].DayOfWeek)
	}
	if got.Entries[0].SessionType.Name != "Afternoon" {
		t.Errorf("got name %q, want Afternoon", got.Entries[0].SessionType.Name)
	}
}

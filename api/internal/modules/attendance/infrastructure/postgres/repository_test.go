package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	attendomain "nursery-management-system/api/internal/modules/attendance/domain"
	attrepostgres "nursery-management-system/api/internal/modules/attendance/infrastructure/postgres"
	"nursery-management-system/api/internal/platform/dbtest"
)

var (
	attTenantID     = uuid.MustParse("d1000000-0000-0000-0000-000000000001")
	attBranchID     = uuid.MustParse("d2000000-0000-0000-0000-000000000001")
	attUserID       = uuid.MustParse("d3000000-0000-0000-0000-000000000001")
	attMembershipID = uuid.MustParse("d4000000-0000-0000-0000-000000000001")
)

func setupAttRepo(t *testing.T) (*attrepostgres.AttendanceRepository, *pgxpool.Pool) {
	t.Helper()
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	dbtest.InsertTenant(t, pool, attTenantID, "Att Tenant")
	dbtest.InsertBranch(t, pool, attTenantID, attBranchID, "Att Branch")
	dbtest.InsertUser(t, pool, attUserID, "staff@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, attMembershipID, attTenantID, attBranchID, attUserID, "practitioner", true)

	return attrepostgres.NewAttendanceRepository(pool), pool
}

func seedAttChild(t testing.TB, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	childID := uuid.MustParse("d5000000-0000-0000-0000-000000000001")
	guardianID := uuid.MustParse("d6000000-0000-0000-0000-000000000001")
	linkID := uuid.MustParse("d7000000-0000-0000-0000-000000000001")

	dbtest.InsertChild(t, pool, childID, attTenantID, attBranchID, "Att Child",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)
	dbtest.InsertGuardian(t, pool, guardianID, attTenantID, attBranchID, "Att Parent", true)
	dbtest.InsertGuardianLink(t, pool, linkID, attTenantID, attBranchID, guardianID, childID)

	return childID
}

func TestAttCreateOpenSessionWithEvent(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	occurredAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	localDate := dbtest.DateAt(2025, 5, 15)

	tx := dbtest.BeginTx(t, pool)
	session, err := repo.CreateOpenSessionWithEvent(ctx, tx, attTenantID, attBranchID, childID,
		occurredAt, localDate, attUserID, attMembershipID, "req-001")
	if err != nil {
		t.Fatalf("CreateOpenSessionWithEvent: %v", err)
	}
	dbtest.CommitTx(t, tx)
	if session.Status != attendomain.SessionStatusOpen {
		t.Errorf("Status = %s, want open", session.Status)
	}
	if session.ChildID != childID {
		t.Errorf("ChildID = %s, want %s", session.ChildID, childID)
	}

	var status string
	var checkInEventID, checkOutEventID, correctedByEventID *uuid.UUID
	var requestID *string
	err = pool.QueryRow(ctx,
		"SELECT status, check_in_event_id, check_out_event_id, corrected_by_event_id FROM attendance_sessions WHERE id = $1",
		session.ID).Scan(&status, &checkInEventID, &checkOutEventID, &correctedByEventID)
	if err != nil {
		t.Fatalf("query session: %v", err)
	}
	if status != "open" {
		t.Errorf("status = %s, want open", status)
	}
	if checkInEventID == nil {
		t.Error("check_in_event_id = nil, want non-nil")
	}
	if checkOutEventID != nil {
		t.Error("check_out_event_id should be nil")
	}
	if correctedByEventID != nil {
		t.Error("corrected_by_event_id should be nil")
	}

	err = pool.QueryRow(ctx,
		"SELECT request_id FROM attendance_events WHERE session_id = $1 AND event_type = 'check_in'", session.ID).Scan(&requestID)
	if err != nil {
		t.Fatalf("query event: %v", err)
	}
	if requestID == nil || *requestID != "req-001" {
		t.Errorf("request_id = %v, want req-001", requestID)
	}
}

func TestAttCreateOpenSessionWithEvent_EmptyRequestID(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	occurredAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	localDate := dbtest.DateAt(2025, 5, 15)

	tx := dbtest.BeginTx(t, pool)
	session, err := repo.CreateOpenSessionWithEvent(ctx, tx, attTenantID, attBranchID, childID,
		occurredAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("CreateOpenSessionWithEvent: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var requestID *string
	err = pool.QueryRow(ctx,
		"SELECT request_id FROM attendance_events WHERE session_id = $1 AND event_type = 'check_in'", session.ID).Scan(&requestID)
	if err != nil {
		t.Fatalf("query event: %v", err)
	}
	if requestID != nil {
		t.Errorf("request_id = %v, want nil for empty string", requestID)
	}
}

func TestAttCreateOpenSession_DuplicateOpen(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	occurredAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	localDate := dbtest.DateAt(2025, 5, 15)

	tx := dbtest.BeginTx(t, pool)
	_, err := repo.CreateOpenSessionWithEvent(ctx, tx, attTenantID, attBranchID, childID,
		occurredAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	dbtest.CommitTx(t, tx)

	tx2 := dbtest.BeginTx(t, pool)
	_, err = repo.CreateOpenSessionWithEvent(ctx, tx2, attTenantID, attBranchID, childID,
		dbtest.TimestampAt(2025, 5, 15, 10, 0), localDate, attUserID, attMembershipID, "")
	if err != attendomain.ErrSessionAlreadyOpen {
		t.Fatalf("expected ErrSessionAlreadyOpen, got %v", err)
	}
}

func TestAttGetOpenSessionForUpdate_NotFound(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.GetOpenSessionForUpdate(ctx, tx, attTenantID, attBranchID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true when no open session exists")
	}
	dbtest.CommitTx(t, tx)
}

func TestAttGetOpenSessionForUpdate_WrongScope(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	checkInAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	localDate := dbtest.DateAt(2025, 5, 15)

	txSeed := dbtest.BeginTx(t, pool)
	_, err := repo.CreateOpenSessionWithEvent(ctx, txSeed, attTenantID, attBranchID, childID,
		checkInAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("create open session: %v", err)
	}
	dbtest.CommitTx(t, txSeed)

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.GetOpenSessionForUpdate(ctx, tx, uuid.New(), attBranchID, childID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for wrong tenant")
	}
	dbtest.CommitTx(t, tx)
}

func TestAttCompleteSessionWithEvent(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	checkInAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	checkOutAt := dbtest.TimestampAt(2025, 5, 15, 16, 0)
	localDate := dbtest.DateAt(2025, 5, 15)

	tx := dbtest.BeginTx(t, pool)
	session, err := repo.CreateOpenSessionWithEvent(ctx, tx, attTenantID, attBranchID, childID,
		checkInAt, localDate, attUserID, attMembershipID, "req-in")
	if err != nil {
		t.Fatalf("create open session: %v", err)
	}
	dbtest.CommitTx(t, tx)

	tx2 := dbtest.BeginTx(t, pool)
	completed, err := repo.CompleteSessionWithEvent(ctx, tx2, attTenantID, attBranchID, session,
		checkOutAt, localDate, attUserID, attMembershipID, "req-out")
	if err != nil {
		t.Fatalf("CompleteSessionWithEvent: %v", err)
	}
	dbtest.CommitTx(t, tx2)
	if completed.Status != attendomain.SessionStatusComplete {
		t.Errorf("Status = %s, want complete", completed.Status)
	}
	if completed.CheckOutAt == nil {
		t.Fatal("CheckOutAt = nil")
	}
	if completed.DurationMinutes == nil {
		t.Fatal("DurationMinutes = nil")
	}
	expectedDuration := int(checkOutAt.Sub(checkInAt).Minutes())
	if *completed.DurationMinutes != expectedDuration {
		t.Errorf("DurationMinutes = %d, want %d", *completed.DurationMinutes, expectedDuration)
	}

	var status string
	var checkOutEventID *uuid.UUID
	err = pool.QueryRow(ctx,
		"SELECT status, check_out_event_id FROM attendance_sessions WHERE id = $1", session.ID,
	).Scan(&status, &checkOutEventID)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if status != "complete" {
		t.Errorf("status = %s, want complete", status)
	}
	if checkOutEventID == nil {
		t.Error("check_out_event_id = nil")
	}
}

func TestAttCompleteSession_InvalidTimeOrder(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	checkInAt := dbtest.TimestampAt(2025, 5, 15, 16, 0)
	checkOutAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	localDate := dbtest.DateAt(2025, 5, 15)

	tx := dbtest.BeginTx(t, pool)
	session, err := repo.CreateOpenSessionWithEvent(ctx, tx, attTenantID, attBranchID, childID,
		checkInAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	dbtest.CommitTx(t, tx)

	session.CheckInAt = checkInAt
	tx2 := dbtest.BeginTx(t, pool)
	_, err = repo.CompleteSessionWithEvent(ctx, tx2, attTenantID, attBranchID, session,
		checkOutAt, localDate, attUserID, attMembershipID, "")
	if err != attendomain.ErrInvalidTimeOrder {
		t.Fatalf("expected ErrInvalidTimeOrder, got %v", err)
	}
}

func TestAttGetSessionForCorrection(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	checkInAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	checkOutAt := dbtest.TimestampAt(2025, 5, 15, 16, 0)
	localDate := dbtest.DateAt(2025, 5, 15)

	txSeed := dbtest.BeginTx(t, pool)
	sessionSeed, err := repo.CreateOpenSessionWithEvent(ctx, txSeed, attTenantID, attBranchID, childID,
		checkInAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("create open session: %v", err)
	}
	dbtest.CommitTx(t, txSeed)

	txComplete := dbtest.BeginTx(t, pool)
	_, err = repo.CompleteSessionWithEvent(ctx, txComplete, attTenantID, attBranchID, sessionSeed,
		checkOutAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("complete session: %v", err)
	}
	dbtest.CommitTx(t, txComplete)

	tx := dbtest.BeginTx(t, pool)
	session, found, err := repo.GetSessionForCorrection(ctx, tx, attTenantID, attBranchID, sessionSeed.ID)
	if err != nil {
		t.Fatalf("GetSessionForCorrection: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if session.Status != attendomain.SessionStatusComplete {
		t.Errorf("Status = %s, want complete", session.Status)
	}
	if session.CheckOutAt == nil {
		t.Fatal("CheckOutAt = nil")
	}
	dbtest.CommitTx(t, tx)
}

func TestAttGetSessionForCorrection_WrongScope(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	checkInAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	localDate := dbtest.DateAt(2025, 5, 15)

	txSeed := dbtest.BeginTx(t, pool)
	sessionSeed, err := repo.CreateOpenSessionWithEvent(ctx, txSeed, attTenantID, attBranchID, childID,
		checkInAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("create open session: %v", err)
	}
	dbtest.CommitTx(t, txSeed)

	tx := dbtest.BeginTx(t, pool)
	_, found, err := repo.GetSessionForCorrection(ctx, tx, uuid.New(), attBranchID, sessionSeed.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true for wrong scope")
	}
	dbtest.CommitTx(t, tx)
}

func TestAttHasOverlappingSession(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	checkInAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	checkOutAt := dbtest.TimestampAt(2025, 5, 15, 16, 0)
	localDate := dbtest.DateAt(2025, 5, 15)

	txSeed := dbtest.BeginTx(t, pool)
	sessionSeed, err := repo.CreateOpenSessionWithEvent(ctx, txSeed, attTenantID, attBranchID, childID,
		checkInAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("create open session: %v", err)
	}
	dbtest.CommitTx(t, txSeed)

	txComplete := dbtest.BeginTx(t, pool)
	_, err = repo.CompleteSessionWithEvent(ctx, txComplete, attTenantID, attBranchID, sessionSeed,
		checkOutAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("complete session: %v", err)
	}
	dbtest.CommitTx(t, txComplete)

	tx := dbtest.BeginTx(t, pool)

	// Overlapping: 09:00-12:00 overlaps with 8:30-16:00
	overlaps, err := repo.HasOverlappingSession(ctx, tx, attTenantID, attBranchID, childID, nil,
		dbtest.TimestampAt(2025, 5, 15, 9, 0), dbtest.TimestampAt(2025, 5, 15, 12, 0))
	if err != nil {
		t.Fatalf("HasOverlappingSession: %v", err)
	}
	if !overlaps {
		t.Error("expected overlap, got false")
		var status string
		var rowCheckIn time.Time
		var rowCheckOut *time.Time
		err = pool.QueryRow(ctx,
			"SELECT status, check_in_at, check_out_at FROM attendance_sessions WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3 LIMIT 1",
			attTenantID, attBranchID, childID,
		).Scan(&status, &rowCheckIn, &rowCheckOut)
		if err == nil {
			t.Errorf("seeded session status=%s check_in_at=%s check_out_at=%v", status, rowCheckIn.Format(time.RFC3339), rowCheckOut)
		}
	}

	// Non-overlapping: 16:00-17:00 is adjacent (checkout_at = 16:00, not strictly before)
	noOverlap, err := repo.HasOverlappingSession(ctx, tx, attTenantID, attBranchID, childID, nil,
		dbtest.TimestampAt(2025, 5, 15, 16, 0), dbtest.TimestampAt(2025, 5, 15, 17, 0))
	if err != nil {
		t.Fatalf("HasOverlappingSession: %v", err)
	}
	if noOverlap {
		t.Error("adjacent session should not count as overlap")
	}

	// Before: 7:00-8:30 is adjacent at start
	before, err := repo.HasOverlappingSession(ctx, tx, attTenantID, attBranchID, childID, nil,
		dbtest.TimestampAt(2025, 5, 15, 7, 0), dbtest.TimestampAt(2025, 5, 15, 8, 30))
	if err != nil {
		t.Fatalf("HasOverlappingSession: %v", err)
	}
	if before {
		t.Error("adjacent at start should not count as overlap")
	}
	dbtest.CommitTx(t, tx)
}

func TestAttHasOverlappingSession_ExcludeSession(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	checkInAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	checkOutAt := dbtest.TimestampAt(2025, 5, 15, 16, 0)
	localDate := dbtest.DateAt(2025, 5, 15)

	txSeed := dbtest.BeginTx(t, pool)
	sessionSeed, err := repo.CreateOpenSessionWithEvent(ctx, txSeed, attTenantID, attBranchID, childID,
		checkInAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("create open session: %v", err)
	}
	dbtest.CommitTx(t, txSeed)

	txComplete := dbtest.BeginTx(t, pool)
	_, err = repo.CompleteSessionWithEvent(ctx, txComplete, attTenantID, attBranchID, sessionSeed,
		checkOutAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("complete session: %v", err)
	}
	dbtest.CommitTx(t, txComplete)

	tx := dbtest.BeginTx(t, pool)
	overlaps, err := repo.HasOverlappingSession(ctx, tx, attTenantID, attBranchID, childID, &sessionSeed.ID,
		checkInAt, checkOutAt)
	if err != nil {
		t.Fatalf("HasOverlappingSession: %v", err)
	}
	if overlaps {
		t.Error("should not overlap when excluding the same session")
	}
	dbtest.CommitTx(t, tx)
}

func TestAttHasOverlappingSession_OpenSession(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	checkInAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	localDate := dbtest.DateAt(2025, 5, 15)

	txSeed := dbtest.BeginTx(t, pool)
	_, err := repo.CreateOpenSessionWithEvent(ctx, txSeed, attTenantID, attBranchID, childID,
		checkInAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("create open session: %v", err)
	}
	dbtest.CommitTx(t, txSeed)

	tx := dbtest.BeginTx(t, pool)
	overlaps, err := repo.HasOverlappingSession(ctx, tx, attTenantID, attBranchID, childID, nil,
		dbtest.TimestampAt(2025, 5, 15, 9, 0), dbtest.TimestampAt(2025, 5, 15, 10, 0))
	if err != nil {
		t.Fatalf("HasOverlappingSession: %v", err)
	}
	if !overlaps {
		t.Error("expected overlap with open session")
	}
	dbtest.CommitTx(t, tx)
}

func TestAttCorrectSessionWithEvent(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	checkInAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	checkOutAt := dbtest.TimestampAt(2025, 5, 15, 16, 0)
	localDate := dbtest.DateAt(2025, 5, 15)

	txSeed := dbtest.BeginTx(t, pool)
	sessionSeed, err := repo.CreateOpenSessionWithEvent(ctx, txSeed, attTenantID, attBranchID, childID,
		checkInAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("create open session: %v", err)
	}
	dbtest.CommitTx(t, txSeed)

	txComplete := dbtest.BeginTx(t, pool)
	_, err = repo.CompleteSessionWithEvent(ctx, txComplete, attTenantID, attBranchID, sessionSeed,
		checkOutAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("complete session: %v", err)
	}
	dbtest.CommitTx(t, txComplete)

	tx := dbtest.BeginTx(t, pool)
	session := attendomain.Session{
		ID:               sessionSeed.ID,
		ChildID:          childID,
		Status:           attendomain.SessionStatusComplete,
		CheckInAt:        checkInAt,
		CheckOutAt:       &checkOutAt,
		CheckInLocalDate: localDate,
	}

	correctedInAt := dbtest.TimestampAt(2025, 5, 15, 9, 0)
	correctedOutAt := dbtest.TimestampAt(2025, 5, 15, 15, 0)
	occurredAt := dbtest.TimestampAt(2025, 5, 15, 17, 0)

	params := attendomain.CorrectionParams{
		SessionID:  &sessionSeed.ID,
		ChildID:    &childID,
		CheckInAt:  correctedInAt,
		CheckOutAt: correctedOutAt,
		ReasonCode: "incorrect_time",
		ReasonNote: "clock error",
	}

	corrected, err := repo.CorrectSessionWithEvent(ctx, tx, attTenantID, attBranchID, session, params,
		dbtest.DateAt(2025, 5, 15), dbtest.DateAt(2025, 5, 15), dbtest.DateAt(2025, 5, 15), occurredAt, attUserID, attMembershipID, "req-cor")
	if err != nil {
		t.Fatalf("CorrectSessionWithEvent: %v", err)
	}
	dbtest.CommitTx(t, tx)
	if corrected.Status != attendomain.SessionStatusCorrected {
		t.Errorf("Status = %s, want corrected", corrected.Status)
	}
	if corrected.DurationMinutes == nil {
		t.Fatal("DurationMinutes = nil")
	}

	var status string
	var correctedByEventID *uuid.UUID
	err = pool.QueryRow(ctx,
		"SELECT status, corrected_by_event_id FROM attendance_sessions WHERE id = $1", sessionSeed.ID,
	).Scan(&status, &correctedByEventID)
	if err != nil {
		t.Fatalf("query session: %v", err)
	}
	if status != "corrected" {
		t.Errorf("status = %s, want corrected", status)
	}
	if correctedByEventID == nil {
		t.Error("corrected_by_event_id = nil")
	}

	var reasonCode *string
	err = pool.QueryRow(ctx,
		"SELECT reason_code FROM attendance_events WHERE session_id = $1 AND event_type = 'correction'", sessionSeed.ID,
	).Scan(&reasonCode)
	if err != nil {
		t.Fatalf("query event: %v", err)
	}
	if reasonCode == nil || *reasonCode != "incorrect_time" {
		t.Errorf("reason_code = %v, want incorrect_time", reasonCode)
	}
}

func TestAttCreateCorrectedSessionWithEvent(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	correctedInAt := dbtest.TimestampAt(2025, 5, 14, 9, 0)
	correctedOutAt := dbtest.TimestampAt(2025, 5, 14, 15, 0)
	occurredAt := dbtest.TimestampAt(2025, 5, 15, 10, 0)

	params := attendomain.CorrectionParams{
		ChildID:    &childID,
		CheckInAt:  correctedInAt,
		CheckOutAt: correctedOutAt,
		ReasonCode: "missed_check_in",
		ReasonNote: "forgot to check in",
	}

	tx := dbtest.BeginTx(t, pool)
	session, err := repo.CreateCorrectedSessionWithEvent(ctx, tx, attTenantID, attBranchID, params,
		dbtest.DateAt(2025, 5, 14), dbtest.DateAt(2025, 5, 14), dbtest.DateAt(2025, 5, 15), occurredAt, attUserID, attMembershipID, "req-miss")
	if err != nil {
		t.Fatalf("CreateCorrectedSessionWithEvent: %v", err)
	}
	dbtest.CommitTx(t, tx)
	if session.Status != attendomain.SessionStatusCorrected {
		t.Errorf("Status = %s, want corrected", session.Status)
	}
	if session.ChildID != childID {
		t.Errorf("ChildID = %s, want %s", session.ChildID, childID)
	}
	if session.DurationMinutes == nil {
		t.Fatal("DurationMinutes = nil")
	}
	expectedDuration := int(correctedOutAt.Sub(correctedInAt).Minutes())
	if *session.DurationMinutes != expectedDuration {
		t.Errorf("DurationMinutes = %d, want %d", *session.DurationMinutes, expectedDuration)
	}

	var status string
	var correctedByEventID *uuid.UUID
	err = pool.QueryRow(ctx,
		"SELECT status, corrected_by_event_id FROM attendance_sessions WHERE id = $1", session.ID,
	).Scan(&status, &correctedByEventID)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if status != "corrected" {
		t.Errorf("status = %s, want corrected", status)
	}
	if correctedByEventID == nil {
		t.Error("corrected_by_event_id = nil")
	}
}

func TestAttCreateCorrectedSessionWithEvent_NullableFields(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	params := attendomain.CorrectionParams{
		ChildID:    &childID,
		CheckInAt:  dbtest.TimestampAt(2025, 5, 14, 9, 0),
		CheckOutAt: dbtest.TimestampAt(2025, 5, 14, 15, 0),
		ReasonCode: "other",
		ReasonNote: "some note",
	}

	tx := dbtest.BeginTx(t, pool)
	session, err := repo.CreateCorrectedSessionWithEvent(ctx, tx, attTenantID, attBranchID, params,
		dbtest.DateAt(2025, 5, 14), dbtest.DateAt(2025, 5, 14),
		dbtest.DateAt(2025, 5, 15),
		dbtest.TimestampAt(2025, 5, 15, 10, 0), attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("CreateCorrectedSessionWithEvent: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var requestID *string
	err = pool.QueryRow(ctx,
		"SELECT request_id FROM attendance_events WHERE session_id = $1 AND event_type = 'correction'", session.ID,
	).Scan(&requestID)
	if err != nil {
		t.Fatalf("query event: %v", err)
	}
	if requestID != nil {
		t.Errorf("request_id = %v, want nil for empty string", requestID)
	}
}

func TestAttHasOverlappingSession_WrongScope(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	checkInAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	checkOutAt := dbtest.TimestampAt(2025, 5, 15, 16, 0)
	localDate := dbtest.DateAt(2025, 5, 15)

	txSeed := dbtest.BeginTx(t, pool)
	sessionSeed, err := repo.CreateOpenSessionWithEvent(ctx, txSeed, attTenantID, attBranchID, childID,
		checkInAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("create open session: %v", err)
	}
	dbtest.CommitTx(t, txSeed)

	txComplete := dbtest.BeginTx(t, pool)
	_, err = repo.CompleteSessionWithEvent(ctx, txComplete, attTenantID, attBranchID, sessionSeed,
		checkOutAt, localDate, attUserID, attMembershipID, "")
	if err != nil {
		t.Fatalf("complete session: %v", err)
	}
	dbtest.CommitTx(t, txComplete)

	tx := dbtest.BeginTx(t, pool)
	overlaps, err := repo.HasOverlappingSession(ctx, tx, uuid.New(), attBranchID, childID, nil,
		dbtest.TimestampAt(2025, 5, 15, 9, 0), dbtest.TimestampAt(2025, 5, 15, 17, 0))
	if err != nil {
		t.Fatalf("HasOverlappingSession: %v", err)
	}
	if overlaps {
		t.Error("should not overlap for wrong tenant")
	}
	dbtest.CommitTx(t, tx)
}

func TestAttCreateCorrectedSessionWithEvent_NullReasonNote(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	params := attendomain.CorrectionParams{
		ChildID:    &childID,
		CheckInAt:  dbtest.TimestampAt(2025, 5, 14, 9, 0),
		CheckOutAt: dbtest.TimestampAt(2025, 5, 14, 15, 0),
		ReasonCode: "missed_check_in",
		ReasonNote: "",
	}

	tx := dbtest.BeginTx(t, pool)
	session, err := repo.CreateCorrectedSessionWithEvent(ctx, tx, attTenantID, attBranchID, params,
		dbtest.DateAt(2025, 5, 14), dbtest.DateAt(2025, 5, 14),
		dbtest.DateAt(2025, 5, 15),
		dbtest.TimestampAt(2025, 5, 15, 10, 0), attUserID, attMembershipID, "req-test")
	if err != nil {
		t.Fatalf("CreateCorrectedSessionWithEvent: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var reasonNote *string
	err = pool.QueryRow(ctx,
		"SELECT reason_note FROM attendance_events WHERE session_id = $1 AND event_type = 'correction'", session.ID,
	).Scan(&reasonNote)
	if err != nil {
		t.Fatalf("query event: %v", err)
	}
	if reasonNote != nil {
		t.Errorf("reason_note = %v, want nil for empty string", reasonNote)
	}
}

func TestAttRoutineEvents_NullReasonFields(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	checkInAt := dbtest.TimestampAt(2025, 5, 15, 8, 30)
	checkOutAt := dbtest.TimestampAt(2025, 5, 15, 16, 0)
	localDate := dbtest.DateAt(2025, 5, 15)

	tx := dbtest.BeginTx(t, pool)
	session, err := repo.CreateOpenSessionWithEvent(ctx, tx, attTenantID, attBranchID, childID,
		checkInAt, localDate, attUserID, attMembershipID, "req-in")
	if err != nil {
		t.Fatalf("create open session: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var reasonCode *string
	var reasonNote *string
	err = pool.QueryRow(ctx,
		"SELECT reason_code, reason_note FROM attendance_events WHERE session_id = $1 AND event_type = 'check_in'", session.ID,
	).Scan(&reasonCode, &reasonNote)
	if err != nil {
		t.Fatalf("query check-in event: %v", err)
	}
	if reasonCode != nil {
		t.Errorf("check-in reason_code = %v, want nil", reasonCode)
	}
	if reasonNote != nil {
		t.Errorf("check-in reason_note = %v, want nil", reasonNote)
	}

	tx2 := dbtest.BeginTx(t, pool)
	_, err = repo.CompleteSessionWithEvent(ctx, tx2, attTenantID, attBranchID, session,
		checkOutAt, localDate, attUserID, attMembershipID, "req-out")
	if err != nil {
		t.Fatalf("complete session: %v", err)
	}
	dbtest.CommitTx(t, tx2)

	err = pool.QueryRow(ctx,
		"SELECT reason_code, reason_note FROM attendance_events WHERE session_id = $1 AND event_type = 'check_out'", session.ID,
	).Scan(&reasonCode, &reasonNote)
	if err != nil {
		t.Fatalf("query check-out event: %v", err)
	}
	if reasonCode != nil {
		t.Errorf("check-out reason_code = %v, want nil", reasonCode)
	}
	if reasonNote != nil {
		t.Errorf("check-out reason_note = %v, want nil", reasonNote)
	}
}

func TestAttCorrectionEvent_ActionLocalDate(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	actionDate := dbtest.DateAt(2025, 5, 20)
	occurredAt := dbtest.TimestampAt(2025, 5, 20, 14, 0)

	params := attendomain.CorrectionParams{
		ChildID:    &childID,
		CheckInAt:  dbtest.TimestampAt(2025, 5, 15, 8, 0),
		CheckOutAt: dbtest.TimestampAt(2025, 5, 15, 16, 0),
		ReasonCode: "missed_check_in",
	}

	tx := dbtest.BeginTx(t, pool)
	session, err := repo.CreateCorrectedSessionWithEvent(ctx, tx, attTenantID, attBranchID, params,
		dbtest.DateAt(2025, 5, 15), dbtest.DateAt(2025, 5, 15), actionDate, occurredAt, attUserID, attMembershipID, "req-test")
	if err != nil {
		t.Fatalf("CreateCorrectedSessionWithEvent: %v", err)
	}
	dbtest.CommitTx(t, tx)

	var eventLocalDate time.Time
	err = pool.QueryRow(ctx,
		"SELECT local_date FROM attendance_events WHERE session_id = $1 AND event_type = 'correction'", session.ID,
	).Scan(&eventLocalDate)
	if err != nil {
		t.Fatalf("query correction event: %v", err)
	}
	y, m, d := eventLocalDate.Date()
	if y != 2025 || m != 5 || d != 20 {
		t.Errorf("correction event local_date = %d-%02d-%02d, want 2025-05-20 (action day)", y, m, d)
	}

	var sessionCheckInLD time.Time
	err = pool.QueryRow(ctx,
		"SELECT check_in_local_date FROM attendance_sessions WHERE id = $1", session.ID,
	).Scan(&sessionCheckInLD)
	if err != nil {
		t.Fatalf("query session: %v", err)
	}
	y, m, d = sessionCheckInLD.Date()
	if y != 2025 || m != 5 || d != 15 {
		t.Errorf("session check_in_local_date = %d-%02d-%02d, want 2025-05-15 (corrected interval)", y, m, d)
	}
}

func TestAttListIncompleteSessionsForPeriod_ReturnsOpenInPeriod(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	sessionID := uuid.New()
	dbtest.InsertAttendanceSession(t, pool, sessionID, attTenantID, attBranchID, childID, "open",
		dbtest.TimestampAt(2025, 6, 10, 8, 30), dbtest.DateAt(2025, 6, 10))

	results, err := repo.ListIncompleteSessionsForPeriod(ctx, attTenantID, attBranchID,
		dbtest.DateAt(2025, 6, 1), dbtest.DateAt(2025, 7, 1))
	if err != nil {
		t.Fatalf("ListIncompleteSessionsForPeriod: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ChildID != childID {
		t.Errorf("ChildID = %s, want %s", results[0].ChildID, childID)
	}
	if results[0].SessionID != sessionID {
		t.Errorf("SessionID = %s, want %s", results[0].SessionID, sessionID)
	}
	if results[0].ChildFirstName != "Att Child" {
		t.Errorf("ChildFirstName = %s, want Att Child", results[0].ChildFirstName)
	}
}

func TestAttListIncompleteSessionsForPeriod_ExcludesBeforePeriod(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	dbtest.InsertAttendanceSession(t, pool, uuid.New(), attTenantID, attBranchID, childID, "open",
		dbtest.TimestampAt(2025, 5, 31, 8, 30), dbtest.DateAt(2025, 5, 31))

	results, err := repo.ListIncompleteSessionsForPeriod(ctx, attTenantID, attBranchID,
		dbtest.DateAt(2025, 6, 1), dbtest.DateAt(2025, 7, 1))
	if err != nil {
		t.Fatalf("ListIncompleteSessionsForPeriod: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for before-period session, got %d", len(results))
	}
}

func TestAttListIncompleteSessionsForPeriod_ExcludesOnExclusiveEnd(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	dbtest.InsertAttendanceSession(t, pool, uuid.New(), attTenantID, attBranchID, childID, "open",
		dbtest.TimestampAt(2025, 7, 1, 8, 30), dbtest.DateAt(2025, 7, 1))

	results, err := repo.ListIncompleteSessionsForPeriod(ctx, attTenantID, attBranchID,
		dbtest.DateAt(2025, 6, 1), dbtest.DateAt(2025, 7, 1))
	if err != nil {
		t.Fatalf("ListIncompleteSessionsForPeriod: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for on-exclusive-end session, got %d", len(results))
	}
}

func TestAttListIncompleteSessionsForPeriod_ExcludesCompleteSessions(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	sessionID := uuid.New()
	dbtest.InsertAttendanceSession(t, pool, sessionID, attTenantID, attBranchID, childID, "open",
		dbtest.TimestampAt(2025, 6, 10, 8, 30), dbtest.DateAt(2025, 6, 10))
	dbtest.CompleteAttendanceSession(t, pool, attTenantID, attBranchID, sessionID,
		dbtest.TimestampAt(2025, 6, 10, 16, 0), dbtest.DateAt(2025, 6, 10))

	results, err := repo.ListIncompleteSessionsForPeriod(ctx, attTenantID, attBranchID,
		dbtest.DateAt(2025, 6, 1), dbtest.DateAt(2025, 7, 1))
	if err != nil {
		t.Fatalf("ListIncompleteSessionsForPeriod: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for complete session, got %d", len(results))
	}
}

func TestAttListIncompleteSessionsForPeriod_IncludesInactiveChild(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()

	childID := uuid.MustParse("d5000000-0000-0000-0000-000000000099")
	guardianID := uuid.MustParse("d6000000-0000-0000-0000-000000000099")
	linkID := uuid.MustParse("d7000000-0000-0000-0000-000000000099")

	dbtest.InsertChild(t, pool, childID, attTenantID, attBranchID, "Inactive Child",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)
	_, err := pool.Exec(ctx,
		"UPDATE children SET is_active = false, left_at = now(), left_reason_code = 'left_nursery', updated_at = now() WHERE id = $1",
		childID)
	if err != nil {
		t.Fatalf("mark child inactive: %v", err)
	}
	dbtest.InsertGuardian(t, pool, guardianID, attTenantID, attBranchID, "Inactive Parent", true)
	dbtest.InsertGuardianLink(t, pool, linkID, attTenantID, attBranchID, guardianID, childID)

	dbtest.InsertAttendanceSession(t, pool, uuid.New(), attTenantID, attBranchID, childID, "open",
		dbtest.TimestampAt(2025, 6, 10, 8, 30), dbtest.DateAt(2025, 6, 10))

	results, err := repo.ListIncompleteSessionsForPeriod(ctx, attTenantID, attBranchID,
		dbtest.DateAt(2025, 6, 1), dbtest.DateAt(2025, 7, 1))
	if err != nil {
		t.Fatalf("ListIncompleteSessionsForPeriod: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for inactive child open session, got %d", len(results))
	}
	if results[0].ChildFirstName != "Inactive Child" {
		t.Errorf("ChildFirstName = %s, want Inactive Child", results[0].ChildFirstName)
	}
}

func TestAttListIncompleteSessionsForPeriod_SortsStably(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()

	childA := uuid.MustParse("d5000000-0000-0000-0000-0000000000a1")
	childB := uuid.MustParse("d5000000-0000-0000-0000-0000000000b2")
	guardianA := uuid.MustParse("d6000000-0000-0000-0000-0000000000a1")
	guardianB := uuid.MustParse("d6000000-0000-0000-0000-0000000000b2")
	linkA := uuid.MustParse("d7000000-0000-0000-0000-0000000000a1")
	linkB := uuid.MustParse("d7000000-0000-0000-0000-0000000000b2")

	dbtest.InsertChild(t, pool, childA, attTenantID, attBranchID, "Beta Child",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)
	dbtest.InsertChild(t, pool, childB, attTenantID, attBranchID, "Alpha Child",
		dbtest.DateAt(2022, 1, 15), dbtest.DateAt(2024, 9, 1), 500, true)
	dbtest.InsertGuardian(t, pool, guardianA, attTenantID, attBranchID, "Parent A", true)
	dbtest.InsertGuardian(t, pool, guardianB, attTenantID, attBranchID, "Parent B", true)
	dbtest.InsertGuardianLink(t, pool, linkA, attTenantID, attBranchID, guardianA, childA)
	dbtest.InsertGuardianLink(t, pool, linkB, attTenantID, attBranchID, guardianB, childB)

	dbtest.InsertAttendanceSession(t, pool, uuid.New(), attTenantID, attBranchID, childA, "open",
		dbtest.TimestampAt(2025, 6, 11, 8, 30), dbtest.DateAt(2025, 6, 11))
	dbtest.InsertAttendanceSession(t, pool, uuid.New(), attTenantID, attBranchID, childB, "open",
		dbtest.TimestampAt(2025, 6, 10, 8, 30), dbtest.DateAt(2025, 6, 10))

	results, err := repo.ListIncompleteSessionsForPeriod(ctx, attTenantID, attBranchID,
		dbtest.DateAt(2025, 6, 1), dbtest.DateAt(2025, 7, 1))
	if err != nil {
		t.Fatalf("ListIncompleteSessionsForPeriod: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].ChildFirstName != "Alpha Child" {
		t.Errorf("first result ChildFirstName = %s, want Alpha Child (earlier date)", results[0].ChildFirstName)
	}
	if results[1].ChildFirstName != "Beta Child" {
		t.Errorf("second result ChildFirstName = %s, want Beta Child (later date)", results[1].ChildFirstName)
	}
}

func TestAttListIncompleteSessionsForPeriod_RespectsTenantBranchScope(t *testing.T) {
	repo, pool := setupAttRepo(t)
	ctx := context.Background()
	childID := seedAttChild(t, pool)

	dbtest.InsertAttendanceSession(t, pool, uuid.New(), attTenantID, attBranchID, childID, "open",
		dbtest.TimestampAt(2025, 6, 10, 8, 30), dbtest.DateAt(2025, 6, 10))

	wrongTenant := uuid.New()
	results, err := repo.ListIncompleteSessionsForPeriod(ctx, wrongTenant, attBranchID,
		dbtest.DateAt(2025, 6, 1), dbtest.DateAt(2025, 7, 1))
	if err != nil {
		t.Fatalf("ListIncompleteSessionsForPeriod: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for wrong tenant, got %d", len(results))
	}
}

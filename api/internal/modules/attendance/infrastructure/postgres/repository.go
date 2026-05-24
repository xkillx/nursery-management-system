package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/attendance/domain"
	"nursery-management-system/api/internal/platform/uid"
)

type AttendanceRepository struct {
	pool *pgxpool.Pool
}

func NewAttendanceRepository(pool *pgxpool.Pool) *AttendanceRepository {
	return &AttendanceRepository{pool: pool}
}

func (r *AttendanceRepository) CreateOpenSessionWithEvent(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, branchID, childID uuid.UUID,
	occurredAt time.Time,
	localDate time.Time,
	userID, membershipID uuid.UUID,
	requestID string,
) (domain.Session, error) {
	sessionID := uid.NewUUID()
	eventID := uid.NewUUID()

	_, err := tx.Exec(ctx, `
INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_in_local_date)
VALUES ($1, $2, $3, $4, 'open', $5, $6)`,
		sessionID, tenantID, branchID, childID, occurredAt, localDate,
	)
	if err != nil {
		if isOpenSessionUniqueViolation(err) {
			return domain.Session{}, domain.ErrSessionAlreadyOpen
		}
		return domain.Session{}, fmt.Errorf("insert attendance session: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO attendance_events (id, tenant_id, branch_id, child_id, session_id, event_type, occurred_at, local_date, recorded_by_user_id, recorded_by_membership_id, request_id)
VALUES ($1, $2, $3, $4, $5, 'check_in', $6, $7, $8, $9, NULLIF($10, ''))`,
		eventID, tenantID, branchID, childID, sessionID, occurredAt, localDate, userID, membershipID, requestID,
	)
	if err != nil {
		return domain.Session{}, fmt.Errorf("insert check-in event: %w", err)
	}

	_, err = tx.Exec(ctx, `
UPDATE attendance_sessions SET check_in_event_id = $1
WHERE tenant_id = $2 AND branch_id = $3 AND id = $4`,
		eventID, tenantID, branchID, sessionID,
	)
	if err != nil {
		return domain.Session{}, fmt.Errorf("update session check_in_event_id: %w", err)
	}

	return domain.Session{
		ID:               sessionID,
		ChildID:          childID,
		Status:           domain.SessionStatusOpen,
		CheckInAt:        occurredAt,
		CheckInLocalDate: localDate,
		CreatedAt:        occurredAt,
		UpdatedAt:        occurredAt,
	}, nil
}

func (r *AttendanceRepository) GetOpenSessionForUpdate(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, branchID, childID uuid.UUID,
) (domain.Session, bool, error) {
	var s domain.Session
	err := tx.QueryRow(ctx, `
SELECT id, child_id, status, check_in_at, check_in_local_date, created_at, updated_at
FROM attendance_sessions
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3 AND status = 'open'
FOR UPDATE`,
		tenantID, branchID, childID,
	).Scan(&s.ID, &s.ChildID, &s.Status, &s.CheckInAt, &s.CheckInLocalDate, &s.CreatedAt, &s.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Session{}, false, nil
	}
	if err != nil {
		return domain.Session{}, false, fmt.Errorf("select open session for update: %w", err)
	}
	return s, true, nil
}

func (r *AttendanceRepository) CompleteSessionWithEvent(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, branchID uuid.UUID,
	session domain.Session,
	occurredAt time.Time,
	localDate time.Time,
	userID, membershipID uuid.UUID,
	requestID string,
) (domain.Session, error) {
	if !occurredAt.After(session.CheckInAt) {
		return domain.Session{}, domain.ErrInvalidTimeOrder
	}

	eventID := uid.NewUUID()

	_, err := tx.Exec(ctx, `
INSERT INTO attendance_events (id, tenant_id, branch_id, child_id, session_id, event_type, occurred_at, local_date, recorded_by_user_id, recorded_by_membership_id, request_id)
VALUES ($1, $2, $3, $4, $5, 'check_out', $6, $7, $8, $9, NULLIF($10, ''))`,
		eventID, tenantID, branchID, session.ChildID, session.ID, occurredAt, localDate, userID, membershipID, requestID,
	)
	if err != nil {
		return domain.Session{}, fmt.Errorf("insert check-out event: %w", err)
	}

	duration := int(occurredAt.Sub(session.CheckInAt).Minutes())

	_, err = tx.Exec(ctx, `
UPDATE attendance_sessions
SET status = 'complete',
    check_out_at = $1,
    check_out_local_date = $2,
    check_out_event_id = $3,
    updated_at = $4
WHERE tenant_id = $5 AND branch_id = $6 AND id = $7`,
		occurredAt, localDate, eventID, occurredAt, tenantID, branchID, session.ID,
	)
	if err != nil {
		return domain.Session{}, fmt.Errorf("complete attendance session: %w", err)
	}

	return domain.Session{
		ID:                session.ID,
		ChildID:           session.ChildID,
		Status:            domain.SessionStatusComplete,
		CheckInAt:         session.CheckInAt,
		CheckOutAt:        &occurredAt,
		CheckInLocalDate:  session.CheckInLocalDate,
		CheckOutLocalDate: &localDate,
		DurationMinutes:   &duration,
		CreatedAt:         session.CreatedAt,
		UpdatedAt:         occurredAt,
	}, nil
}

func isOpenSessionUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return pgErr.ConstraintName == "idx_attendance_sessions_one_open_child"
	}
	return false
}

func (r *AttendanceRepository) GetSessionForCorrection(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, branchID, sessionID uuid.UUID,
) (domain.Session, bool, error) {
	var s domain.Session
	var checkOutAt *time.Time
	var checkOutLocalDate *time.Time

	err := tx.QueryRow(ctx, `
	SELECT id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date, created_at, updated_at
	FROM attendance_sessions
	WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
	FOR UPDATE`,
		tenantID, branchID, sessionID,
	).Scan(&s.ID, &s.ChildID, &s.Status, &s.CheckInAt, &checkOutAt, &s.CheckInLocalDate, &checkOutLocalDate, &s.CreatedAt, &s.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Session{}, false, nil
	}
	if err != nil {
		return domain.Session{}, false, fmt.Errorf("get session for correction: %w", err)
	}

	s.CheckOutAt = checkOutAt
	s.CheckOutLocalDate = checkOutLocalDate
	return s, true, nil
}

func (r *AttendanceRepository) HasOverlappingSession(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, branchID, childID uuid.UUID,
	excludeSessionID *uuid.UUID,
	checkInAt, checkOutAt time.Time,
) (bool, error) {
	var exists bool

	q := `
	SELECT EXISTS (
	    SELECT 1 FROM attendance_sessions
	    WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
	      AND status IN ('open', 'complete', 'corrected')
	      AND (
	          (check_out_at IS NOT NULL AND check_in_at < $5 AND check_out_at > $4)
	          OR
	          (check_out_at IS NULL AND check_in_at < $5)
	      )`

	args := []any{tenantID, branchID, childID, checkInAt, checkOutAt}
	argPos := 6

	if excludeSessionID != nil {
		q += fmt.Sprintf(" AND id != $%d", argPos)
		args = append(args, *excludeSessionID)
		argPos++
	}

	q += ")"

	err := tx.QueryRow(ctx, q, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check overlapping session: %w", err)
	}
	return exists, nil
}

func (r *AttendanceRepository) CorrectSessionWithEvent(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, branchID uuid.UUID,
	session domain.Session,
	params domain.CorrectionParams,
	checkInLocalDate, checkOutLocalDate time.Time,
	occurredAt time.Time,
	userID, membershipID uuid.UUID,
	requestID string,
) (domain.Session, error) {
	eventID := uid.NewUUID()

	reasonNote := params.ReasonNote
	details := map[string]any{
		"event_type":           "correction",
		"reason_code":          params.ReasonCode,
		"previous_check_in_at": session.CheckInAt.Format(time.RFC3339),
		"corrected_check_in":   params.CheckInAt.Format(time.RFC3339),
		"corrected_check_out":  params.CheckOutAt.Format(time.RFC3339),
	}
	if session.CheckOutAt != nil {
		details["previous_check_out_at"] = session.CheckOutAt.Format(time.RFC3339)
	}
	if reasonNote != "" {
		details["reason_note"] = reasonNote
	}
	detailsJSON, _ := json.Marshal(details)

	_, err := tx.Exec(ctx, `
	INSERT INTO attendance_events (id, tenant_id, branch_id, child_id, session_id, event_type, occurred_at, local_date, recorded_by_user_id, recorded_by_membership_id, request_id, reason_code, reason_note, details)
	VALUES ($1, $2, $3, $4, $5, 'correction', $6, $7, $8, $9, NULLIF($10, ''), $11, NULLIF($12, ''), $13::jsonb)`,
		eventID, tenantID, branchID, session.ChildID, session.ID, occurredAt, checkInLocalDate, userID, membershipID, requestID, params.ReasonCode, reasonNote, string(detailsJSON),
	)
	if err != nil {
		return domain.Session{}, fmt.Errorf("insert correction event: %w", err)
	}

	duration := int(params.CheckOutAt.Sub(params.CheckInAt).Minutes())

	_, err = tx.Exec(ctx, `
	UPDATE attendance_sessions
	SET status = 'corrected',
	    check_in_at = $1,
	    check_out_at = $2,
	    check_in_local_date = $3,
	    check_out_local_date = $4,
	    corrected_by_event_id = $5,
	    updated_at = $6
	WHERE tenant_id = $7 AND branch_id = $8 AND id = $9`,
		params.CheckInAt, params.CheckOutAt, checkInLocalDate, checkOutLocalDate, eventID, occurredAt, tenantID, branchID, session.ID,
	)
	if err != nil {
		return domain.Session{}, fmt.Errorf("correct attendance session: %w", err)
	}

	return domain.Session{
		ID:                session.ID,
		ChildID:           session.ChildID,
		Status:            domain.SessionStatusCorrected,
		CheckInAt:         params.CheckInAt,
		CheckOutAt:        &params.CheckOutAt,
		CheckInLocalDate:  checkInLocalDate,
		CheckOutLocalDate: &checkOutLocalDate,
		DurationMinutes:   &duration,
		CreatedAt:         session.CreatedAt,
		UpdatedAt:         occurredAt,
	}, nil
}

func (r *AttendanceRepository) CreateCorrectedSessionWithEvent(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, branchID uuid.UUID,
	params domain.CorrectionParams,
	checkInLocalDate, checkOutLocalDate time.Time,
	occurredAt time.Time,
	userID, membershipID uuid.UUID,
	requestID string,
) (domain.Session, error) {
	sessionID := uid.NewUUID()
	childID := *params.ChildID

	_, err := tx.Exec(ctx, `
	INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date)
	VALUES ($1, $2, $3, $4, 'corrected', $5, $6, $7, $8)`,
		sessionID, tenantID, branchID, childID, params.CheckInAt, params.CheckOutAt, checkInLocalDate, checkOutLocalDate,
	)
	if err != nil {
		return domain.Session{}, fmt.Errorf("insert corrected session: %w", err)
	}

	eventID := uid.NewUUID()

	reasonNote := params.ReasonNote
	details := map[string]any{
		"event_type":          "correction",
		"reason_code":         params.ReasonCode,
		"corrected_check_in":  params.CheckInAt.Format(time.RFC3339),
		"corrected_check_out": params.CheckOutAt.Format(time.RFC3339),
		"created_by_correction": true,
	}
	if reasonNote != "" {
		details["reason_note"] = reasonNote
	}
	detailsJSON, _ := json.Marshal(details)

	_, err = tx.Exec(ctx, `
	INSERT INTO attendance_events (id, tenant_id, branch_id, child_id, session_id, event_type, occurred_at, local_date, recorded_by_user_id, recorded_by_membership_id, request_id, reason_code, reason_note, details)
	VALUES ($1, $2, $3, $4, $5, 'correction', $6, $7, $8, $9, NULLIF($10, ''), $11, NULLIF($12, ''), $13::jsonb)`,
		eventID, tenantID, branchID, childID, sessionID, occurredAt, checkInLocalDate, userID, membershipID, requestID, params.ReasonCode, reasonNote, string(detailsJSON),
	)
	if err != nil {
		return domain.Session{}, fmt.Errorf("insert correction event: %w", err)
	}

	_, err = tx.Exec(ctx, `
	UPDATE attendance_sessions SET corrected_by_event_id = $1
	WHERE tenant_id = $2 AND branch_id = $3 AND id = $4`,
		eventID, tenantID, branchID, sessionID,
	)
	if err != nil {
		return domain.Session{}, fmt.Errorf("update session corrected_by_event_id: %w", err)
	}

	duration := int(params.CheckOutAt.Sub(params.CheckInAt).Minutes())

	return domain.Session{
		ID:                sessionID,
		ChildID:           childID,
		Status:            domain.SessionStatusCorrected,
		CheckInAt:         params.CheckInAt,
		CheckOutAt:        &params.CheckOutAt,
		CheckInLocalDate:  checkInLocalDate,
		CheckOutLocalDate: &checkOutLocalDate,
		DurationMinutes:   &duration,
		CreatedAt:         occurredAt,
		UpdatedAt:         occurredAt,
	}, nil
}

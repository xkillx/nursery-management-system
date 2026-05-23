package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		if isUniqueViolation(err) {
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

func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}

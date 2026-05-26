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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/attendance/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
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
	q := sqlc.New(tx)

	if err := q.AttendanceInsertOpenSession(ctx, sqlc.AttendanceInsertOpenSessionParams{
		ID:               uuidToPgtype(sessionID),
		TenantID:         uuidToPgtype(tenantID),
		BranchID:         uuidToPgtype(branchID),
		ChildID:          uuidToPgtype(childID),
		CheckInAt:        timeToPgtypeTimestamptz(occurredAt),
		CheckInLocalDate: timeToPgtypeDate(localDate),
	}); err != nil {
		if isOpenSessionUniqueViolation(err) {
			return domain.Session{}, domain.ErrSessionAlreadyOpen
		}
		return domain.Session{}, fmt.Errorf("insert attendance session: %w", err)
	}

	if err := q.AttendanceInsertCheckInEvent(ctx, sqlc.AttendanceInsertCheckInEventParams{
		ID:                     uuidToPgtype(eventID),
		TenantID:               uuidToPgtype(tenantID),
		BranchID:               uuidToPgtype(branchID),
		ChildID:                uuidToPgtype(childID),
		SessionID:              uuidToPgtype(sessionID),
		OccurredAt:             timeToPgtypeTimestamptz(occurredAt),
		LocalDate:              timeToPgtypeDate(localDate),
		RecordedByUserID:       uuidToPgtype(userID),
		RecordedByMembershipID: uuidToPgtype(membershipID),
		Column10:               requestID,
	}); err != nil {
		return domain.Session{}, fmt.Errorf("insert check-in event: %w", err)
	}

	if err := q.AttendanceAttachCheckInEvent(ctx, sqlc.AttendanceAttachCheckInEventParams{
		CheckInEventID: uuidToPgtype(eventID),
		TenantID:       uuidToPgtype(tenantID),
		BranchID:       uuidToPgtype(branchID),
		ID:             uuidToPgtype(sessionID),
	}); err != nil {
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
	q := sqlc.New(tx)
	row, err := q.AttendanceGetOpenSessionForUpdate(ctx, sqlc.AttendanceGetOpenSessionForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Session{}, false, nil
	}
	if err != nil {
		return domain.Session{}, false, fmt.Errorf("select open session for update: %w", err)
	}
	return domain.Session{
		ID:               pgtypeUUIDToUUID(row.ID),
		ChildID:          pgtypeUUIDToUUID(row.ChildID),
		Status:           domain.SessionStatus(row.Status),
		CheckInAt:        pgtypeTimestamptzToTime(row.CheckInAt),
		CheckInLocalDate: pgtypeDateToTime(row.CheckInLocalDate),
		CreatedAt:        pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:        pgtypeTimestamptzToTime(row.UpdatedAt),
	}, true, nil
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
	q := sqlc.New(tx)

	if err := q.AttendanceInsertCheckOutEvent(ctx, sqlc.AttendanceInsertCheckOutEventParams{
		ID:                     uuidToPgtype(eventID),
		TenantID:               uuidToPgtype(tenantID),
		BranchID:               uuidToPgtype(branchID),
		ChildID:                uuidToPgtype(session.ChildID),
		SessionID:              uuidToPgtype(session.ID),
		OccurredAt:             timeToPgtypeTimestamptz(occurredAt),
		LocalDate:              timeToPgtypeDate(localDate),
		RecordedByUserID:       uuidToPgtype(userID),
		RecordedByMembershipID: uuidToPgtype(membershipID),
		Column10:               requestID,
	}); err != nil {
		return domain.Session{}, fmt.Errorf("insert check-out event: %w", err)
	}

	duration := int(occurredAt.Sub(session.CheckInAt).Minutes())

	if err := q.AttendanceCompleteSession(ctx, sqlc.AttendanceCompleteSessionParams{
		CheckOutAt:        timeToPgtypeTimestamptz(occurredAt),
		CheckOutLocalDate: timeToPgtypeDate(localDate),
		CheckOutEventID:   uuidToPgtype(eventID),
		UpdatedAt:         timeToPgtypeTimestamptz(occurredAt),
		TenantID:          uuidToPgtype(tenantID),
		BranchID:          uuidToPgtype(branchID),
		ID:                uuidToPgtype(session.ID),
	}); err != nil {
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
	q := sqlc.New(tx)
	row, err := q.AttendanceGetSessionForCorrection(ctx, sqlc.AttendanceGetSessionForCorrectionParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(sessionID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Session{}, false, nil
	}
	if err != nil {
		return domain.Session{}, false, fmt.Errorf("get session for correction: %w", err)
	}
	return domain.Session{
		ID:                pgtypeUUIDToUUID(row.ID),
		ChildID:           pgtypeUUIDToUUID(row.ChildID),
		Status:            domain.SessionStatus(row.Status),
		CheckInAt:         pgtypeTimestamptzToTime(row.CheckInAt),
		CheckOutAt:        pgtypeTimestamptzToTimePtr(row.CheckOutAt),
		CheckInLocalDate:  pgtypeDateToTime(row.CheckInLocalDate),
		CheckOutLocalDate: pgtypeDateToTimePtr(row.CheckOutLocalDate),
		CreatedAt:         pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:         pgtypeTimestamptzToTime(row.UpdatedAt),
	}, true, nil
}

func (r *AttendanceRepository) HasOverlappingSession(
	ctx context.Context,
	tx pgx.Tx,
	tenantID, branchID, childID uuid.UUID,
	excludeSessionID *uuid.UUID,
	checkInAt, checkOutAt time.Time,
) (bool, error) {
	q := sqlc.New(tx)
	var excludeID pgtype.UUID
	if excludeSessionID != nil {
		excludeID = uuidToPgtype(*excludeSessionID)
	}
	exists, err := q.AttendanceHasOverlappingSession(ctx, sqlc.AttendanceHasOverlappingSessionParams{
		TenantID:   uuidToPgtype(tenantID),
		BranchID:   uuidToPgtype(branchID),
		ChildID:    uuidToPgtype(childID),
		CheckOutAt: timeToPgtypeTimestamptz(checkOutAt),
		CheckInAt:  timeToPgtypeTimestamptz(checkInAt),
		Column6:    excludeID,
	})
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
	checkInLocalDate, checkOutLocalDate, correctionActionLocalDate time.Time,
	occurredAt time.Time,
	userID, membershipID uuid.UUID,
	requestID string,
) (domain.Session, error) {
	eventID := uid.NewUUID()
	q := sqlc.New(tx)

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

	if err := q.AttendanceInsertCorrectionEvent(ctx, sqlc.AttendanceInsertCorrectionEventParams{
		ID:                     uuidToPgtype(eventID),
		TenantID:               uuidToPgtype(tenantID),
		BranchID:               uuidToPgtype(branchID),
		ChildID:                uuidToPgtype(session.ChildID),
		SessionID:              uuidToPgtype(session.ID),
		OccurredAt:             timeToPgtypeTimestamptz(occurredAt),
		LocalDate:              timeToPgtypeDate(correctionActionLocalDate),
		RecordedByUserID:       uuidToPgtype(userID),
		RecordedByMembershipID: uuidToPgtype(membershipID),
		Column10:               requestID,
		ReasonCode:             stringToPgtypeText(params.ReasonCode),
		Column12:               reasonNote,
		Column13:               detailsJSON,
	}); err != nil {
		return domain.Session{}, fmt.Errorf("insert correction event: %w", err)
	}

	duration := int(params.CheckOutAt.Sub(params.CheckInAt).Minutes())

	if err := q.AttendanceCorrectSession(ctx, sqlc.AttendanceCorrectSessionParams{
		CheckInAt:          timeToPgtypeTimestamptz(params.CheckInAt),
		CheckOutAt:         timeToPgtypeTimestamptz(params.CheckOutAt),
		CheckInLocalDate:   timeToPgtypeDate(checkInLocalDate),
		CheckOutLocalDate:  timeToPgtypeDate(checkOutLocalDate),
		CorrectedByEventID: uuidToPgtype(eventID),
		UpdatedAt:          timeToPgtypeTimestamptz(occurredAt),
		TenantID:           uuidToPgtype(tenantID),
		BranchID:           uuidToPgtype(branchID),
		ID:                 uuidToPgtype(session.ID),
	}); err != nil {
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
	checkInLocalDate, checkOutLocalDate, correctionActionLocalDate time.Time,
	occurredAt time.Time,
	userID, membershipID uuid.UUID,
	requestID string,
) (domain.Session, error) {
	sessionID := uid.NewUUID()
	childID := *params.ChildID
	q := sqlc.New(tx)

	if err := q.AttendanceInsertCorrectedSession(ctx, sqlc.AttendanceInsertCorrectedSessionParams{
		ID:                uuidToPgtype(sessionID),
		TenantID:          uuidToPgtype(tenantID),
		BranchID:          uuidToPgtype(branchID),
		ChildID:           uuidToPgtype(childID),
		CheckInAt:         timeToPgtypeTimestamptz(params.CheckInAt),
		CheckOutAt:        timeToPgtypeTimestamptz(params.CheckOutAt),
		CheckInLocalDate:  timeToPgtypeDate(checkInLocalDate),
		CheckOutLocalDate: timeToPgtypeDate(checkOutLocalDate),
	}); err != nil {
		return domain.Session{}, fmt.Errorf("insert corrected session: %w", err)
	}

	eventID := uid.NewUUID()

	reasonNote := params.ReasonNote
	details := map[string]any{
		"event_type":            "correction",
		"reason_code":           params.ReasonCode,
		"corrected_check_in":    params.CheckInAt.Format(time.RFC3339),
		"corrected_check_out":   params.CheckOutAt.Format(time.RFC3339),
		"created_by_correction": true,
	}
	if reasonNote != "" {
		details["reason_note"] = reasonNote
	}
	detailsJSON, _ := json.Marshal(details)

	if err := q.AttendanceInsertCorrectionEvent(ctx, sqlc.AttendanceInsertCorrectionEventParams{
		ID:                     uuidToPgtype(eventID),
		TenantID:               uuidToPgtype(tenantID),
		BranchID:               uuidToPgtype(branchID),
		ChildID:                uuidToPgtype(childID),
		SessionID:              uuidToPgtype(sessionID),
		OccurredAt:             timeToPgtypeTimestamptz(occurredAt),
		LocalDate:              timeToPgtypeDate(correctionActionLocalDate),
		RecordedByUserID:       uuidToPgtype(userID),
		RecordedByMembershipID: uuidToPgtype(membershipID),
		Column10:               requestID,
		ReasonCode:             stringToPgtypeText(params.ReasonCode),
		Column12:               reasonNote,
		Column13:               detailsJSON,
	}); err != nil {
		return domain.Session{}, fmt.Errorf("insert correction event: %w", err)
	}

	if err := q.AttendanceAttachCorrectedEvent(ctx, sqlc.AttendanceAttachCorrectedEventParams{
		CorrectedByEventID: uuidToPgtype(eventID),
		TenantID:           uuidToPgtype(tenantID),
		BranchID:           uuidToPgtype(branchID),
		ID:                 uuidToPgtype(sessionID),
	}); err != nil {
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

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func timeToPgtypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func pgtypeTimestamptzToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func timeToPgtypeDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func pgtypeDateToTime(d pgtype.Date) time.Time {
	return d.Time
}

func pgtypeDateToTimePtr(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

func stringToPgtypeText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

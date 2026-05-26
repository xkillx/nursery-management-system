package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/attendance/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type CorrectionResult struct {
	Session domain.Session
	Created bool
}

type CorrectAttendance struct {
	repo          domain.Repository
	childChecker  domain.ChildCorrectionChecker
	txMgr         *transaction.Manager
	audit         *audit.Writer
	clock         *AttendanceClock
}

func NewCorrectAttendance(
	repo domain.Repository,
	childChecker domain.ChildCorrectionChecker,
	txMgr *transaction.Manager,
	auditWriter *audit.Writer,
	clock *AttendanceClock,
) *CorrectAttendance {
	return &CorrectAttendance{
		repo:         repo,
		childChecker: childChecker,
		txMgr:        txMgr,
		audit:        auditWriter,
		clock:        clock,
	}
}

func (uc *CorrectAttendance) Execute(ctx context.Context, actor tenant.ActorContext, params domain.CorrectionParams) (CorrectionResult, error) {
	if err := validateTargets(params); err != nil {
		return CorrectionResult{}, err
	}

	if err := domain.ValidateCorrectionReason(params.ReasonCode, params.ReasonNote); err != nil {
		return CorrectionResult{}, mapCorrectionError(err)
	}

	if !params.CheckOutAt.After(params.CheckInAt) {
		return CorrectionResult{}, domainerrors.Conflict("attendance_invalid_time_order", "Check-out must be after check-in.")
	}

	now, _ := uc.clock.Now()
	if params.CheckInAt.After(now) || params.CheckOutAt.After(now) {
		return CorrectionResult{}, domainerrors.Conflict("attendance_correction_future_time", "Corrected times must not be in the future.")
	}

	checkInLocalDate := uc.clock.LocalDate(params.CheckInAt)
	checkOutLocalDate := uc.clock.LocalDate(params.CheckOutAt)
	correctionActionLocalDate := uc.clock.LocalDate(now)

	var result CorrectionResult

	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		if params.SessionID != nil {
			session, err := uc.correctExistingSession(ctx, tx, actor, params, checkInLocalDate, checkOutLocalDate, correctionActionLocalDate, now)
			if err != nil {
				return err
			}
			result = CorrectionResult{Session: session, Created: false}
			return uc.writeAudit(ctx, tx, actor, session, params, now, checkInLocalDate, checkOutLocalDate, false)
		}

		session, err := uc.createMissedSession(ctx, tx, actor, params, checkInLocalDate, checkOutLocalDate, correctionActionLocalDate, now)
		if err != nil {
			return err
		}
		result = CorrectionResult{Session: session, Created: true}
		return uc.writeAudit(ctx, tx, actor, session, params, now, checkInLocalDate, checkOutLocalDate, true)
	})

	if err != nil {
		return CorrectionResult{}, err
	}
	return result, nil
}

func (uc *CorrectAttendance) correctExistingSession(
	ctx context.Context, tx pgx.Tx, actor tenant.ActorContext,
	params domain.CorrectionParams, checkInLocalDate, checkOutLocalDate, correctionActionLocalDate, now time.Time,
) (domain.Session, error) {
	session, found, err := uc.repo.GetSessionForCorrection(ctx, tx, actor.TenantID, actor.BranchID, *params.SessionID)
	if err != nil {
		return domain.Session{}, domainerrors.Internal(fmt.Errorf("get session for correction: %w", err))
	}
	if !found {
		return domain.Session{}, domainerrors.New("attendance_session_not_found", "Attendance session not found.")
	}

	params.ChildID = &session.ChildID

	if err := uc.validateChildAndWindow(ctx, tx, actor, *params.ChildID, checkInLocalDate, checkOutLocalDate); err != nil {
		return domain.Session{}, err
	}

	overlaps, err := uc.repo.HasOverlappingSession(ctx, tx, actor.TenantID, actor.BranchID, session.ChildID, params.SessionID, params.CheckInAt, params.CheckOutAt)
	if err != nil {
		return domain.Session{}, domainerrors.Internal(fmt.Errorf("check overlap: %w", err))
	}
	if overlaps {
		return domain.Session{}, domainerrors.Conflict("attendance_session_overlap", "Corrected interval overlaps another session.")
	}

	return uc.repo.CorrectSessionWithEvent(ctx, tx, actor.TenantID, actor.BranchID, session, params, checkInLocalDate, checkOutLocalDate, correctionActionLocalDate, now, actor.UserID, actor.MembershipID, actor.RequestID)
}

func (uc *CorrectAttendance) createMissedSession(
	ctx context.Context, tx pgx.Tx, actor tenant.ActorContext,
	params domain.CorrectionParams, checkInLocalDate, checkOutLocalDate, correctionActionLocalDate, now time.Time,
) (domain.Session, error) {
	if err := uc.validateChildAndWindow(ctx, tx, actor, *params.ChildID, checkInLocalDate, checkOutLocalDate); err != nil {
		return domain.Session{}, err
	}

	overlaps, err := uc.repo.HasOverlappingSession(ctx, tx, actor.TenantID, actor.BranchID, *params.ChildID, nil, params.CheckInAt, params.CheckOutAt)
	if err != nil {
		return domain.Session{}, domainerrors.Internal(fmt.Errorf("check overlap: %w", err))
	}
	if overlaps {
		return domain.Session{}, domainerrors.Conflict("attendance_session_overlap", "Corrected interval overlaps another session.")
	}

	return uc.repo.CreateCorrectedSessionWithEvent(ctx, tx, actor.TenantID, actor.BranchID, params, checkInLocalDate, checkOutLocalDate, correctionActionLocalDate, now, actor.UserID, actor.MembershipID, actor.RequestID)
}

func (uc *CorrectAttendance) validateChildAndWindow(
	ctx context.Context, tx pgx.Tx, actor tenant.ActorContext,
	childID uuid.UUID, checkInLocalDate, checkOutLocalDate time.Time,
) error {
	child, found, err := uc.childChecker.GetChildForCorrection(ctx, tx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return domainerrors.Internal(fmt.Errorf("get child for correction: %w", err))
	}
	if !found {
		return domainerrors.New("child_not_found", "Resource not found.")
	}

	checkInDateOnly := dateOnly(checkInLocalDate)
	checkOutDateOnly := dateOnly(checkOutLocalDate)

	if checkInDateOnly.Before(child.StartDate) || checkOutDateOnly.Before(child.StartDate) {
		return domainerrors.Conflict("attendance_outside_enrollment_window", "Corrected dates are before child start date.")
	}
	if child.EndDate != nil {
		if checkInDateOnly.After(*child.EndDate) || checkOutDateOnly.After(*child.EndDate) {
			return domainerrors.Conflict("attendance_outside_enrollment_window", "Corrected dates are after child end date.")
		}
	}
	return nil
}

func (uc *CorrectAttendance) writeAudit(
	ctx context.Context, tx pgx.Tx, actor tenant.ActorContext,
	session domain.Session, params domain.CorrectionParams,
	now time.Time, checkInLocalDate, checkOutLocalDate time.Time, created bool,
) error {
	details := map[string]any{
		"child_id":             session.ChildID.String(),
		"session_id":           session.ID.String(),
		"event_type":           "correction",
		"reason_code":          params.ReasonCode,
		"corrected_check_in":   params.CheckInAt.Format(time.RFC3339),
		"corrected_check_out":  params.CheckOutAt.Format(time.RFC3339),
		"check_in_local_date":  checkInLocalDate.Format("2006-01-02"),
		"check_out_local_date": checkOutLocalDate.Format("2006-01-02"),
		"created_by_correction": created,
	}
	if params.ReasonNote != "" {
		details["reason_note"] = params.ReasonNote
	}
	if !created {
		details["previous_check_in_at"] = session.CheckInAt.Format(time.RFC3339)
	}

	if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
		ActionType: "attendance_corrected",
		EntityType: "attendance_session",
		EntityID:   session.ID,
		Details:    details,
	}); err != nil {
		return domainerrors.Internal(fmt.Errorf("audit attendance_corrected: %w", err))
	}
	return nil
}

func validateTargets(params domain.CorrectionParams) error {
	hasSession := params.SessionID != nil
	hasChild := params.ChildID != nil

	if hasSession && hasChild {
		return domainerrors.Validation("Specify exactly one of session_id or child_id, not both.", "session_id")
	}
	if !hasSession && !hasChild {
		return domainerrors.Validation("Specify exactly one of session_id or child_id.", "session_id")
	}
	return nil
}

func mapCorrectionError(err error) error {
	switch err {
	case domain.ErrCorrectionReasonRequired:
		return domainerrors.New("attendance_correction_reason_required", "Correction reason is required.")
	case domain.ErrCorrectionReasonInvalid:
		return domainerrors.New("attendance_correction_reason_invalid", "Invalid correction reason code.")
	case domain.ErrReasonNoteRequiredForOther:
		return domainerrors.New("reason_note_required_for_other", "Reason note is required when reason is 'other'.", "reason_note")
	default:
		return domainerrors.Internal(fmt.Errorf("correction: %w", err))
	}
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

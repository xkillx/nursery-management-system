package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/attendance/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type CheckInChild struct {
	repo         domain.Repository
	childChecker domain.ChildEnrollmentChecker
	txMgr        *transaction.Manager
	audit        *audit.Writer
	clock        *AttendanceClock
}

func NewCheckInChild(
	repo domain.Repository,
	childChecker domain.ChildEnrollmentChecker,
	txMgr *transaction.Manager,
	auditWriter *audit.Writer,
	clock *AttendanceClock,
) *CheckInChild {
	return &CheckInChild{
		repo:         repo,
		childChecker: childChecker,
		txMgr:        txMgr,
		audit:        auditWriter,
		clock:        clock,
	}
}

func (uc *CheckInChild) Execute(ctx context.Context, actor tenant.ActorContext, childID uuid.UUID) (domain.Session, error) {
	var result domain.Session

	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		now, localDate := uc.clock.Now()

		if err := uc.childChecker.CheckEnrollmentForAttendance(ctx, tx, actor.TenantID, actor.BranchID, childID, localDate); err != nil {
			return mapCheckInError(err)
		}

		session, err := uc.repo.CreateOpenSessionWithEvent(ctx, tx, actor.TenantID, actor.BranchID, childID, now, localDate, actor.UserID, actor.MembershipID, actor.RequestID)
		if err != nil {
			return mapCheckInError(err)
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "attendance_checked_in",
			EntityType: "attendance_session",
			EntityID:   session.ID,
			Details: map[string]any{
				"child_id":            childID.String(),
				"event_type":          "check_in",
				"check_in_at":         now.Format("2006-01-02T15:04:05Z"),
				"check_in_local_date": localDate.Format("2006-01-02"),
			},
		}); err != nil {
			return domainerrors.Internal(fmt.Errorf("audit attendance_checked_in: %w", err))
		}

		result = session
		return nil
	})

	if err != nil {
		return domain.Session{}, err
	}
	return result, nil
}

func mapCheckInError(err error) error {
	if err == domain.ErrSessionAlreadyOpen {
		return domainerrors.Conflict("attendance_session_already_open", "An open attendance session already exists for this child.")
	}
	if err == domain.ErrChildEnrollmentIncomplete {
		return domainerrors.Conflict("child_enrollment_incomplete", "Child enrollment is not complete.")
	}
	if err == domain.ErrChildNotFound {
		return domainerrors.NotFound("child", "Resource not found.")
	}
	return domainerrors.Internal(fmt.Errorf("check in: %w", err))
}

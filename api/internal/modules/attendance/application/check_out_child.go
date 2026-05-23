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

type CheckOutChild struct {
	repo  domain.Repository
	txMgr *transaction.Manager
	audit *audit.Writer
}

func NewCheckOutChild(
	repo domain.Repository,
	txMgr *transaction.Manager,
	auditWriter *audit.Writer,
) *CheckOutChild {
	return &CheckOutChild{
		repo:   repo,
		txMgr:  txMgr,
		audit:  auditWriter,
	}
}

func (uc *CheckOutChild) Execute(ctx context.Context, actor tenant.ActorContext, childID uuid.UUID) (domain.Session, error) {
	var result domain.Session

	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		session, found, err := uc.repo.GetOpenSessionForUpdate(ctx, tx, actor.TenantID, actor.BranchID, childID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("get open session: %w", err))
		}
		if !found {
			return domainerrors.Conflict("attendance_session_not_open", "No open attendance session found for this child.")
		}

		now, localDate := LondonNow()

		completed, err := uc.repo.CompleteSessionWithEvent(ctx, tx, actor.TenantID, actor.BranchID, session, now, localDate, actor.UserID, actor.MembershipID, actor.RequestID)
		if err != nil {
			return mapCheckOutError(err)
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "attendance_checked_out",
			EntityType: "attendance_session",
			EntityID:   completed.ID,
			Details: map[string]any{
				"child_id":             childID.String(),
				"event_type":           "check_out",
				"check_in_at":          session.CheckInAt.Format("2006-01-02T15:04:05Z"),
				"check_out_at":         now.Format("2006-01-02T15:04:05Z"),
				"check_in_local_date":  session.CheckInLocalDate.Format("2006-01-02"),
				"check_out_local_date": localDate.Format("2006-01-02"),
			},
		}); err != nil {
			return domainerrors.Internal(fmt.Errorf("audit attendance_checked_out: %w", err))
		}

		result = completed
		return nil
	})

	if err != nil {
		return domain.Session{}, err
	}
	return result, nil
}

func mapCheckOutError(err error) error {
	if err == domain.ErrInvalidTimeOrder {
		return domainerrors.Conflict("attendance_invalid_time_order", "Check-out time must be after check-in time.")
	}
	return domainerrors.Internal(fmt.Errorf("check out: %w", err))
}

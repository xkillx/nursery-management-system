package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/absence/domain"
	"nursery-management-system/api/internal/modules/attendance/application"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ClearMarker struct {
	repo  domain.Repository
	txMgr txManager
	audit auditWriter
	clock *application.AttendanceClock
}

func NewClearMarker(
	repo domain.Repository,
	txMgr txManager,
	auditWriter auditWriter,
	clock *application.AttendanceClock,
) *ClearMarker {
	return &ClearMarker{
		repo:  repo,
		txMgr: txMgr,
		audit: auditWriter,
		clock: clock,
	}
}

func (uc *ClearMarker) Execute(ctx context.Context, actor tenant.ActorContext, markerIDRaw string) (domain.AbsenceMarker, error) {
	markerID, err := uuid.Parse(markerIDRaw)
	if err != nil {
		return domain.AbsenceMarker{}, domainerrors.Validation("Invalid marker ID format.", "absence_marker_id")
	}

	var result domain.AbsenceMarker

	err = uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		marker, found, err := uc.repo.GetByID(ctx, tx, actor.TenantID, actor.BranchID, markerID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("get absence marker: %w", err))
		}
		if !found {
			return domainerrors.NotFound("absence_marker", "Resource not found.")
		}

		if marker.ClearedAt != nil {
			result = marker
			return nil
		}

		now, _ := uc.clock.Now()
		cleared, ok, err := uc.repo.Clear(ctx, tx, actor.TenantID, actor.BranchID, markerID, now, actor.UserID, actor.MembershipID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("clear absence marker: %w", err))
		}
		if !ok {
			return domainerrors.NotFound("absence_marker", "Resource not found.")
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "absence_marker_cleared",
			EntityType: "absence_marker",
			EntityID:   cleared.ID,
			Details: map[string]any{
				"child_id":   cleared.ChildID.String(),
				"local_date": cleared.LocalDate.Format("2006-01-02"),
				"cleared_at": now.Format("2006-01-02T15:04:05Z"),
			},
		}); err != nil {
			return domainerrors.Internal(fmt.Errorf("audit absence_marker_cleared: %w", err))
		}

		result = cleared
		return nil
	})

	if err != nil {
		return domain.AbsenceMarker{}, err
	}
	return result, nil
}

package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/term/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type approveOrRejectScheduleChangeDeps struct {
	scheduleRepo domain.ScheduleChangeRepository
	audit        *audit.Writer
	txMgr        *transaction.Manager
}

func (uc *approveOrRejectScheduleChangeDeps) approveOrReject(
	ctx context.Context,
	actor tenant.ActorContext,
	termIDRaw, changeIDRaw string,
	decision domain.ScheduleChangeDecision,
	auditAction string,
) (*domain.TermScheduleChange, error) {
	termID, err := uuid.Parse(termIDRaw)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "term_id")
	}
	changeID, err := uuid.Parse(changeIDRaw)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "change_id")
	}
	if actor.MembershipID == uuid.Nil {
		return nil, domainerrors.Forbidden("manager_only", "Manager role required.")
	}

	var result *domain.TermScheduleChange
	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		change, found, err := uc.scheduleRepo.GetByID(ctx, actor.TenantID, actor.BranchID, changeID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("get schedule change: %w", err))
		}
		if !found || change.TermID != termID {
			return domainerrors.NotFound("term_schedule_change", "Resource not found.")
		}
		if change.ApprovalDecision != nil {
			return domainerrors.Conflict("schedule_change_already_decided", "Schedule change already decided.")
		}
		if change.ChangeKind != domain.ScheduleChangeIncrease {
			return domainerrors.New("decrease_auto_approved", "Only increase changes require approval.", "change_kind")
		}

		var rows int64
		var actErr error
		if decision == domain.ScheduleChangeApproved {
			rows, actErr = uc.scheduleRepo.Approve(ctx, tx, actor.TenantID, actor.BranchID, change.ID, actor.MembershipID)
		} else {
			rows, actErr = uc.scheduleRepo.Reject(ctx, tx, actor.TenantID, actor.BranchID, change.ID, actor.MembershipID)
		}
		if actErr != nil {
			return domainerrors.Internal(fmt.Errorf("apply decision: %w", actErr))
		}
		if rows == 0 {
			return domainerrors.Conflict("schedule_change_already_decided", "Schedule change already decided.")
		}

		reloaded, found, err := uc.scheduleRepo.GetByID(ctx, actor.TenantID, actor.BranchID, change.ID)
		if err != nil || !found {
			return domainerrors.Internal(fmt.Errorf("reload schedule change: %w", err))
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: auditAction,
			EntityType: domain.AuditEntityTermScheduleChange,
			EntityID:   reloaded.ID,
			Details: map[string]any{
				"term_id":             reloaded.TermID.String(),
				"change_kind":         string(reloaded.ChangeKind),
				"decision":            string(decision),
				"effective_from":      reloaded.EffectiveFrom.Format("2006-01-02"),
				"previous_pattern_id": reloaded.PreviousBookingPatternID.String(),
				"new_pattern_id":      reloaded.NewBookingPatternID.String(),
			},
		}); err != nil {
			return domainerrors.Internal(fmt.Errorf("audit decision: %w", err))
		}

		result = reloaded
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	return result, nil
}

type ApproveScheduleChangeUseCase struct {
	deps *approveOrRejectScheduleChangeDeps
}

func NewApproveScheduleChangeUseCase(
	scheduleRepo domain.ScheduleChangeRepository,
	auditWriter *audit.Writer,
	txMgr *transaction.Manager,
) *ApproveScheduleChangeUseCase {
	return &ApproveScheduleChangeUseCase{
		deps: &approveOrRejectScheduleChangeDeps{
			scheduleRepo: scheduleRepo,
			audit:        auditWriter,
			txMgr:        txMgr,
		},
	}
}

func (uc *ApproveScheduleChangeUseCase) Execute(ctx context.Context, actor tenant.ActorContext, termID, changeID string) (*domain.TermScheduleChange, error) {
	return uc.deps.approveOrReject(ctx, actor, termID, changeID, domain.ScheduleChangeApproved, domain.AuditTermScheduleChangeApproved)
}

type RejectScheduleChangeUseCase struct {
	deps *approveOrRejectScheduleChangeDeps
}

func NewRejectScheduleChangeUseCase(
	scheduleRepo domain.ScheduleChangeRepository,
	auditWriter *audit.Writer,
	txMgr *transaction.Manager,
) *RejectScheduleChangeUseCase {
	return &RejectScheduleChangeUseCase{
		deps: &approveOrRejectScheduleChangeDeps{
			scheduleRepo: scheduleRepo,
			audit:        auditWriter,
			txMgr:        txMgr,
		},
	}
}

func (uc *RejectScheduleChangeUseCase) Execute(ctx context.Context, actor tenant.ActorContext, termID, changeID string) (*domain.TermScheduleChange, error) {
	return uc.deps.approveOrReject(ctx, actor, termID, changeID, domain.ScheduleChangeRejected, domain.AuditTermScheduleChangeRejected)
}

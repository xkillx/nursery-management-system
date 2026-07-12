package application

import (
	"context"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type UpdateBranchSettingsUseCase struct {
	repo        domain.BranchSettingsRepository
	auditWriter *audit.Writer
	txMgr       *transaction.Manager
}

func NewUpdateBranchSettingsUseCase(
	repo domain.BranchSettingsRepository,
	auditWriter *audit.Writer,
	txMgr *transaction.Manager,
) *UpdateBranchSettingsUseCase {
	return &UpdateBranchSettingsUseCase{
		repo:        repo,
		auditWriter: auditWriter,
		txMgr:       txMgr,
	}
}

type BranchSettings struct {
	OverdueGraceDays   int `json:"overdue_grace_days"`
	ReminderDaysBefore int `json:"reminder_days_before"`
}

func (uc *UpdateBranchSettingsUseCase) GetSettings(ctx context.Context, actor tenant.ActorContext) (BranchSettings, error) {
	graceDays, err := uc.repo.GetOverdueGraceDays(ctx, actor.TenantID, actor.BranchID)
	if err != nil {
		return BranchSettings{}, err
	}

	reminderDays, err := uc.repo.GetReminderDaysBefore(ctx, actor.TenantID, actor.BranchID)
	if err != nil {
		return BranchSettings{}, err
	}

	return BranchSettings{
		OverdueGraceDays:   graceDays,
		ReminderDaysBefore: reminderDays,
	}, nil
}

func (uc *UpdateBranchSettingsUseCase) Execute(ctx context.Context, actor tenant.ActorContext, settings BranchSettings) error {
	if settings.OverdueGraceDays < 0 || settings.OverdueGraceDays > 30 {
		return &domain.ValidationError{
			Field:   "overdue_grace_days",
			Message: "overdue_grace_days must be between 0 and 30.",
		}
	}

	if settings.ReminderDaysBefore < 1 || settings.ReminderDaysBefore > 30 {
		return &domain.ValidationError{
			Field:   "reminder_days_before",
			Message: "reminder_days_before must be between 1 and 30.",
		}
	}

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		if err := uc.repo.UpdateOverdueGraceDays(ctx, tx, actor.TenantID, actor.BranchID, settings.OverdueGraceDays); err != nil {
			return err
		}

		if err := uc.repo.UpdateReminderDaysBefore(ctx, tx, actor.TenantID, actor.BranchID, settings.ReminderDaysBefore); err != nil {
			return err
		}

		details := map[string]any{
			"overdue_grace_days":   settings.OverdueGraceDays,
			"reminder_days_before": settings.ReminderDaysBefore,
		}

		if auditErr := uc.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "branch_billing_settings_updated",
			EntityType: "branch",
			EntityID:   actor.BranchID,
			Details:    details,
		}); auditErr != nil {
			return auditErr
		}

		return nil
	})

	if txErr != nil {
		return txErr
	}

	return nil
}

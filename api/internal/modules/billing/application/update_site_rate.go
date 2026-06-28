package application

import (
	"context"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type UpdateSiteRateUseCase struct {
	repo        domain.SiteRateRepository
	auditWriter *audit.Writer
	txMgr       *transaction.Manager
}

func NewUpdateSiteRateUseCase(
	repo domain.SiteRateRepository,
	auditWriter *audit.Writer,
	txMgr *transaction.Manager,
) *UpdateSiteRateUseCase {
	return &UpdateSiteRateUseCase{
		repo:        repo,
		auditWriter: auditWriter,
		txMgr:       txMgr,
	}
}

func (uc *UpdateSiteRateUseCase) Execute(ctx context.Context, actor tenant.ActorContext, coreHourlyRateMinor int) error {
	if coreHourlyRateMinor <= 0 {
		return &domain.ValidationError{
			Field:   "core_hourly_rate_minor",
			Message: "core_hourly_rate_minor must be a positive integer.",
		}
	}

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		if err := uc.repo.UpdateCoreHourlyRate(ctx, tx, actor.TenantID, actor.BranchID, coreHourlyRateMinor); err != nil {
			return err
		}

		details := map[string]any{
			"new_core_hourly_rate_minor": coreHourlyRateMinor,
		}

		if auditErr := uc.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "site_core_hourly_rate_updated",
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

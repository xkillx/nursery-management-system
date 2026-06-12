package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/owner/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type UpdateSiteBillingSetupResult struct {
	SiteID                  uuid.UUID
	SiteCoreHourlyRateMinor int
}

type UpdateSiteBillingSetupUseCase struct {
	repo        domain.SummaryRepository
	auditWriter *audit.Writer
	txMgr       *transaction.Manager
}

func NewUpdateSiteBillingSetupUseCase(
	repo domain.SummaryRepository,
	auditWriter *audit.Writer,
	txMgr *transaction.Manager,
) *UpdateSiteBillingSetupUseCase {
	return &UpdateSiteBillingSetupUseCase{
		repo:        repo,
		auditWriter: auditWriter,
		txMgr:       txMgr,
	}
}

func (uc *UpdateSiteBillingSetupUseCase) Execute(ctx context.Context, actor domain.OwnerActor, siteID uuid.UUID, coreHourlyRateMinor int) (UpdateSiteBillingSetupResult, error) {
	if coreHourlyRateMinor <= 0 {
		return UpdateSiteBillingSetupResult{}, &domain.ValidationError{
			Field:   "core_hourly_rate_minor",
			Message: "core_hourly_rate_minor must be a positive integer.",
		}
	}

	var result UpdateSiteBillingSetupResult

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		prev, curr, err := uc.repo.UpdateSiteCoreHourlyRate(ctx, tx, actor.TenantID, siteID, coreHourlyRateMinor)
		if err != nil {
			return err
		}

		auditActor := tenant.ActorContext{
			UserID:       actor.UserID,
			MembershipID: actor.MembershipID,
			TenantID:     actor.TenantID,
			BranchID:     siteID,
		}

		details := map[string]any{
			"new_core_hourly_rate_minor": curr,
		}
		if prev != nil {
			details["previous_core_hourly_rate_minor"] = *prev
		} else {
			details["previous_core_hourly_rate_minor"] = nil
		}

		if auditErr := uc.auditWriter.WriteWithTx(ctx, tx, auditActor, audit.WriteParams{
			ActionType: "site_core_hourly_rate_updated",
			EntityType: "branch",
			EntityID:   siteID,
			Details:    details,
		}); auditErr != nil {
			return auditErr
		}

		result = UpdateSiteBillingSetupResult{
			SiteID:                  siteID,
			SiteCoreHourlyRateMinor: curr,
		}
		return nil
	})

	if txErr != nil {
		return UpdateSiteBillingSetupResult{}, txErr
	}

	return result, nil
}

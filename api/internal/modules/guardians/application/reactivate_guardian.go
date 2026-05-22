package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/guardians/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type ReactivateGuardian struct {
	repo  domain.Repository
	txm   *transaction.Manager
	audit *audit.Writer
}

func NewReactivateGuardian(repo domain.Repository, txm *transaction.Manager, auditWriter *audit.Writer) *ReactivateGuardian {
	return &ReactivateGuardian{repo: repo, txm: txm, audit: auditWriter}
}

func (uc *ReactivateGuardian) Execute(ctx context.Context, actor tenant.ActorContext, guardianID uuid.UUID) (domain.Guardian, error) {
	var result domain.Guardian

	err := uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		guardian, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID, actor.BranchID, guardianID)
		if err != nil {
			return fmt.Errorf("fetch guardian for update: %w", err)
		}

		if guardian.IsActive {
			result = guardian
			return nil
		}

		if err := uc.repo.Reactivate(ctx, tx, actor.TenantID, actor.BranchID, guardianID); err != nil {
			return fmt.Errorf("reactivate guardian: %w", err)
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "guardian_reactivated",
			EntityType: "guardian",
			EntityID:   guardianID,
			Details:    map[string]any{},
		}); err != nil {
			return fmt.Errorf("audit guardian_reactivated: %w", err)
		}

		result, err = uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID, actor.BranchID, guardianID)
		if err != nil {
			return fmt.Errorf("fetch reactivated guardian: %w", err)
		}

		return nil
	})

	if err != nil {
		return domain.Guardian{}, err
	}

	return result, nil
}

package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/parents/domain"
	"nursery-management-system/api/internal/platform/audit"
)

type SoftDeleteParentUseCase struct {
	repo  domain.Repository
	audit *audit.Writer
	txMgr TxManager
}

func NewSoftDeleteParentUseCase(repo domain.Repository, auditWriter *audit.Writer, txMgr TxManager) *SoftDeleteParentUseCase {
	return &SoftDeleteParentUseCase{repo: repo, audit: auditWriter, txMgr: txMgr}
}

func (uc *SoftDeleteParentUseCase) Execute(ctx context.Context, actor ActorContext, parentID uuid.UUID) error {
	return uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		parent, found, err := uc.repo.GetByID(ctx, tx, actor.TenantID, actor.BranchID, parentID)
		if err != nil {
			return err
		}
		if !found {
			return domain.ErrParentNotFound
		}
		if !parent.IsActive {
			return domain.ErrParentInactive
		}

		if err := uc.repo.SoftDelete(ctx, tx, actor.TenantID, actor.BranchID, parentID); err != nil {
			return err
		}

		return uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "parent_deactivated",
			EntityType: "parent",
			EntityID:   parentID,
			Details:    map[string]any{},
		})
	})
}

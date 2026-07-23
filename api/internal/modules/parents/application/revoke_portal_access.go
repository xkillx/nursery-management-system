package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/parents/domain"
	"nursery-management-system/api/internal/platform/audit"
)

type RevokePortalAccessUseCase struct {
	repo  domain.Repository
	audit *audit.Writer
	txMgr TxManager
}

func NewRevokePortalAccessUseCase(repo domain.Repository, auditWriter *audit.Writer, txMgr TxManager) *RevokePortalAccessUseCase {
	return &RevokePortalAccessUseCase{repo: repo, audit: auditWriter, txMgr: txMgr}
}

func (uc *RevokePortalAccessUseCase) Execute(ctx context.Context, actor ActorContext, parentID uuid.UUID) error {
	return uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		parent, found, err := uc.repo.GetByID(ctx, tx, actor.TenantID, actor.BranchID, parentID)
		if err != nil {
			return err
		}
		if !found {
			return domain.ErrParentNotFound
		}
		if parent.UserID == nil {
			return domain.ErrParentNotFound
		}

		if err := uc.repo.SetUserID(ctx, tx, actor.TenantID, actor.BranchID, parentID, nil); err != nil {
			return err
		}

		return uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "parent_portal_access_revoked",
			EntityType: "parent",
			EntityID:   parentID,
			Details: map[string]any{
				"previous_user_id": parent.UserID.String(),
			},
		})
	})
}

package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/parents/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/lifecycle"
)

type UnlinkChildUseCase struct {
	repo  domain.Repository
	audit *audit.Writer
	txMgr TxManager
}

func NewUnlinkChildUseCase(repo domain.Repository, auditWriter *audit.Writer, txMgr TxManager) *UnlinkChildUseCase {
	return &UnlinkChildUseCase{repo: repo, audit: auditWriter, txMgr: txMgr}
}

func (uc *UnlinkChildUseCase) Execute(ctx context.Context, actor ActorContext, parentID, childID uuid.UUID, reasonCode, reasonNote string) error {
	return uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		_, found, err := uc.repo.GetByID(ctx, tx, actor.TenantID, actor.BranchID, parentID)
		if err != nil {
			return err
		}
		if !found {
			return domain.ErrParentNotFound
		}

		link, hasLink, err := uc.repo.FindActiveByPair(ctx, tx, actor.TenantID, actor.BranchID, parentID, childID)
		if err != nil {
			return err
		}
		if !hasLink {
			return domain.ErrLinkNotFound
		}

		if !lifecycle.IsValidReasonCode(reasonCode) {
			return domain.ErrLinkNotFound
		}

		if err := uc.repo.EndLink(ctx, tx, actor.TenantID, actor.BranchID, link.ID, reasonCode, reasonNote); err != nil {
			return err
		}

		return uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "parent_child_unlinked",
			EntityType: "parent_child",
			EntityID:   link.ID,
			ReasonCode: &reasonCode,
			ReasonNote: nullableString(reasonNote),
			Details: map[string]any{
				"parent_id": parentID.String(),
				"child_id":  childID.String(),
			},
		})
	})
}

func nullableString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

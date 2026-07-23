package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/parents/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/uid"
)

type LinkChildUseCase struct {
	repo      domain.Repository
	audit     *audit.Writer
	txMgr     TxManager
	childCheck domain.ChildExistenceChecker
}

func NewLinkChildUseCase(repo domain.Repository, auditWriter *audit.Writer, txMgr TxManager, childCheck domain.ChildExistenceChecker) *LinkChildUseCase {
	return &LinkChildUseCase{repo: repo, audit: auditWriter, txMgr: txMgr, childCheck: childCheck}
}

func (uc *LinkChildUseCase) Execute(ctx context.Context, actor ActorContext, parentID, childID uuid.UUID) (domain.ParentChild, error) {
	var result domain.ParentChild

	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
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

		childExists, err := uc.childCheck.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, childID)
		if err != nil {
			return err
		}
		if !childExists {
			return domain.ErrChildNotFound
		}

		existing, hasLink, err := uc.repo.FindActiveByPair(ctx, tx, actor.TenantID, actor.BranchID, parentID, childID)
		if err != nil {
			return err
		}
		if hasLink {
			result = existing
			return nil
		}

		link := domain.ParentChild{
			ID:       uid.NewUUID(),
			TenantID: actor.TenantID,
			BranchID: actor.BranchID,
			ParentID: parentID,
			ChildID:  childID,
		}

		if err := uc.repo.CreateLink(ctx, tx, link); err != nil {
			return err
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "parent_child_linked",
			EntityType: "parent_child",
			EntityID:   link.ID,
			Details: map[string]any{
				"parent_id": parentID.String(),
				"child_id":  childID.String(),
			},
		}); err != nil {
			return err
		}

		created, found, err := uc.repo.FindActiveByPair(ctx, tx, actor.TenantID, actor.BranchID, parentID, childID)
		if err != nil || !found {
			return err
		}

		result = created
		return nil
	})

	return result, err
}

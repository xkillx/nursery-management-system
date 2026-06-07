package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/guardianlinks/domain"
)

type ListChildLinksUseCase struct {
	repo          domain.Repository
	txMgr         TxManager
	childChecker  ChildExistsChecker
}

func NewListChildLinksUseCase(repo domain.Repository, txMgr TxManager, childChecker ChildExistsChecker) *ListChildLinksUseCase {
	return &ListChildLinksUseCase{repo: repo, txMgr: txMgr, childChecker: childChecker}
}

type ListChildLinksParams struct {
	TenantID uuid.UUID
	BranchID uuid.UUID
	ChildID  uuid.UUID
}

func (uc *ListChildLinksUseCase) Execute(ctx context.Context, actor ActorContext, params ListChildLinksParams) ([]domain.LinkedGuardianChildLink, error) {
	var result []domain.LinkedGuardianChildLink

	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		childExists, err := uc.childChecker.ExistsInScope(ctx, tx, params.TenantID, params.BranchID, params.ChildID)
		if err != nil {
			return err
		}
		if !childExists {
			return ErrChildNotFound
		}

		links, err := uc.repo.ListActiveByChild(ctx, tx, params.TenantID, params.BranchID, params.ChildID)
		if err != nil {
			return err
		}
		result = links
		return nil
	})

	return result, err
}

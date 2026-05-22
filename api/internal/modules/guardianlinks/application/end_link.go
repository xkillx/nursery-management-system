package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/guardianlinks/domain"
	"nursery-management-system/api/internal/platform/audit"
)

var ErrLinkNotFound = errors.New("link not found")

type EndLinkUseCase struct {
	repo  domain.Repository
	audit *audit.Writer
	txMgr TxManager
}

func NewEndLinkUseCase(repo domain.Repository, auditWriter *audit.Writer, txMgr TxManager) *EndLinkUseCase {
	return &EndLinkUseCase{repo: repo, audit: auditWriter, txMgr: txMgr}
}

func (uc *EndLinkUseCase) Execute(ctx context.Context, actor ActorContext, linkID uuid.UUID, reasonCode, reasonNote string) (domain.GuardianChildLink, error) {
	var result domain.GuardianChildLink

	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		row, found, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID, actor.BranchID, linkID)
		if err != nil {
			return err
		}
		if !found {
			return ErrLinkNotFound
		}

		if row.EndedAt == nil {
			if err := uc.repo.End(ctx, tx, actor.TenantID, actor.BranchID, linkID, reasonCode, reasonNote); err != nil {
				return err
			}

			if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: "guardian_child_link_ended",
				EntityType: "guardian_child_link",
				EntityID:   linkID,
				ReasonCode: &reasonCode,
				ReasonNote: nullableString(reasonNote),
				Details:    map[string]any{},
			}); err != nil {
				return err
			}
		}

		updated, found, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID, actor.BranchID, linkID)
		if err != nil || !found {
			return err
		}

		result = updated
		return nil
	})

	return result, err
}

func nullableString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

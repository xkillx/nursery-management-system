package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/parentmappings/domain"
	"nursery-management-system/api/internal/platform/audit"
)

var ErrMappingNotFound = errors.New("parent mapping not found")

type EndMappingUseCase struct {
	repo  domain.Repository
	audit *audit.Writer
	txMgr TxManager
}

func NewEndMappingUseCase(repo domain.Repository, auditWriter *audit.Writer, txMgr TxManager) *EndMappingUseCase {
	return &EndMappingUseCase{repo: repo, audit: auditWriter, txMgr: txMgr}
}

func (uc *EndMappingUseCase) Execute(ctx context.Context, actor ActorContext, mappingID uuid.UUID, reasonCode, reasonNote string) (domain.ParentMapping, error) {
	var result domain.ParentMapping

	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		row, found, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID, actor.BranchID, mappingID)
		if err != nil {
			return err
		}
		if !found {
			return ErrMappingNotFound
		}

		if row.EndedAt == nil {
			if err := uc.repo.End(ctx, tx, actor.TenantID, actor.BranchID, mappingID, reasonCode, reasonNote); err != nil {
				return err
			}

			if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: "parent_mapping_ended",
				EntityType: "parent_membership_guardian_mapping",
				EntityID:   mappingID,
				ReasonCode: &reasonCode,
				ReasonNote: nullableString(reasonNote),
				Details:    map[string]any{},
			}); err != nil {
				return err
			}
		}

		updated, found, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID, actor.BranchID, mappingID)
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

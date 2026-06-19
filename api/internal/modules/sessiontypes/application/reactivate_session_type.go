package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/sessiontypes/domain"
	"nursery-management-system/api/internal/platform/audit"
)

type ReactivateSessionType struct {
	repo  domain.Repository
	txMgr TxManager
	audit *audit.Writer
}

func NewReactivateSessionType(repo domain.Repository, txMgr TxManager, auditWriter *audit.Writer) *ReactivateSessionType {
	return &ReactivateSessionType{repo: repo, txMgr: txMgr, audit: auditWriter}
}

func (uc *ReactivateSessionType) Execute(ctx context.Context, actor SessionTypeActor, siteID, stID uuid.UUID) (domain.SessionType, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.SessionType{}, err
	}

	var st domain.SessionType
	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		existing, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, stID)
		if err != nil {
			return err
		}

		if existing.IsActive {
			st = existing
			return nil
		}

		if err := uc.repo.Reactivate(ctx, tx, actor.TenantID(), siteID, stID); err != nil {
			return internalError(err)
		}

		if uc.audit != nil {
			if err := uc.audit.WriteSystemWithTx(ctx, tx, actor.TenantID(), siteID, actor.RequestID(), audit.WriteParams{
				ActionType: "session_type_reactivated",
				EntityType: "session_type",
				EntityID:   stID,
				Details:    map[string]any{},
			}); err != nil {
				return internalError(err)
			}
		}

		var gerr error
		st, gerr = uc.repo.GetByID(ctx, actor.TenantID(), siteID, stID)
		return gerr
	})
	if err != nil {
		return domain.SessionType{}, err
	}
	return st, nil
}

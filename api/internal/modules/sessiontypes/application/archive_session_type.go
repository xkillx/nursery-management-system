package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/sessiontypes/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type ArchiveSessionType struct {
	repo  domain.Repository
	txMgr TxManager
	audit *audit.Writer
}

func NewArchiveSessionType(repo domain.Repository, txMgr TxManager, auditWriter *audit.Writer) *ArchiveSessionType {
	return &ArchiveSessionType{repo: repo, txMgr: txMgr, audit: auditWriter}
}

func (uc *ArchiveSessionType) Execute(ctx context.Context, actor SessionTypeActor, siteID, stID uuid.UUID) error {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return err
	}

	return uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		existing, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, stID)
		if err != nil {
			return err
		}

		if !existing.IsActive {
			return domainerrors.Conflict("session_type_not_active", "Session type is already archived.")
		}

		if err := uc.repo.Archive(ctx, tx, actor.TenantID(), siteID, stID); err != nil {
			return internalError(err)
		}

		if uc.audit != nil {
			if err := uc.audit.WriteSystemWithTx(ctx, tx, actor.TenantID(), siteID, actor.RequestID(), audit.WriteParams{
				ActionType: "session_type_archived",
				EntityType: "session_type",
				EntityID:   stID,
				Details:    map[string]any{},
			}); err != nil {
				return internalError(err)
			}
		}

		return nil
	})
}

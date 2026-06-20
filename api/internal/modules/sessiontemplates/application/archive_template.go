package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/sessiontemplates/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type ArchiveSessionTemplate struct {
	repo  domain.Repository
	txMgr TxManager
	audit *audit.Writer
}

func NewArchiveSessionTemplate(repo domain.Repository, txMgr TxManager, auditWriter *audit.Writer) *ArchiveSessionTemplate {
	return &ArchiveSessionTemplate{repo: repo, txMgr: txMgr, audit: auditWriter}
}

func (uc *ArchiveSessionTemplate) Execute(ctx context.Context, actor SessionTemplateActor, siteID, templateID uuid.UUID) error {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return err
	}

	return uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		existing, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, templateID)
		if err != nil {
			return err
		}

		if !existing.IsActive {
			return domainerrors.Conflict("session_template_not_active", "Session template is already archived.")
		}

		if err := uc.repo.Archive(ctx, tx, actor.TenantID(), siteID, templateID); err != nil {
			return internalError(err)
		}

		if uc.audit != nil {
			if err := uc.audit.WriteSystemWithTx(ctx, tx, actor.TenantID(), siteID, actor.RequestID(), audit.WriteParams{
				ActionType: "session_template_archived",
				EntityType: "session_template",
				EntityID:   templateID,
				Details:    map[string]any{},
			}); err != nil {
				return internalError(err)
			}
		}

		return nil
	})
}

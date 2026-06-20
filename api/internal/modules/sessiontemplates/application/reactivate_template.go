package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/sessiontemplates/domain"
	"nursery-management-system/api/internal/platform/audit"
)

type ReactivateSessionTemplate struct {
	repo  domain.Repository
	txMgr TxManager
	audit *audit.Writer
}

func NewReactivateSessionTemplate(repo domain.Repository, txMgr TxManager, auditWriter *audit.Writer) *ReactivateSessionTemplate {
	return &ReactivateSessionTemplate{repo: repo, txMgr: txMgr, audit: auditWriter}
}

func (uc *ReactivateSessionTemplate) Execute(ctx context.Context, actor SessionTemplateActor, siteID, templateID uuid.UUID) (domain.SessionTemplate, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.SessionTemplate{}, err
	}

	var result domain.SessionTemplate
	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		existing, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, templateID)
		if err != nil {
			return err
		}

		if existing.IsActive {
			result = existing
			return nil
		}

		if err := uc.repo.Reactivate(ctx, tx, actor.TenantID(), siteID, templateID); err != nil {
			return internalError(err)
		}

		if uc.audit != nil {
			if err := uc.audit.WriteSystemWithTx(ctx, tx, actor.TenantID(), siteID, actor.RequestID(), audit.WriteParams{
				ActionType: "session_template_reactivated",
				EntityType: "session_template",
				EntityID:   templateID,
				Details:    map[string]any{},
			}); err != nil {
				return internalError(err)
			}
		}

		reloaded, rerr := uc.repo.GetByID(ctx, actor.TenantID(), siteID, templateID)
		if rerr != nil {
			return rerr
		}
		result = reloaded
		return nil
	})
	if err != nil {
		return domain.SessionTemplate{}, err
	}
	entries, eerr := uc.repo.EntriesListByTemplate(ctx, actor.TenantID(), siteID, templateID)
	if eerr != nil {
		return domain.SessionTemplate{}, internalError(eerr)
	}
	result.Entries = entries
	return result, nil
}

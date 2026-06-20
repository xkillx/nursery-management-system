package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/sessiontemplates/domain"
)

type GetSessionTemplate struct {
	repo domain.Repository
}

func NewGetSessionTemplate(repo domain.Repository) *GetSessionTemplate {
	return &GetSessionTemplate{repo: repo}
}

func (uc *GetSessionTemplate) Execute(ctx context.Context, actor SessionTemplateActor, siteID, templateID uuid.UUID) (domain.SessionTemplate, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.SessionTemplate{}, err
	}

	t, err := uc.repo.GetByID(ctx, actor.TenantID(), siteID, templateID)
	if err != nil {
		return domain.SessionTemplate{}, err
	}
	entries, eerr := uc.repo.EntriesListByTemplate(ctx, actor.TenantID(), siteID, templateID)
	if eerr != nil {
		return domain.SessionTemplate{}, internalError(eerr)
	}
	t.Entries = entries
	return t, nil
}

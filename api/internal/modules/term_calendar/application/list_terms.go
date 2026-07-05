package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/term_calendar/domain"
)

type ListTerms struct {
	repo domain.Repository
}

func NewListTerms(repo domain.Repository) *ListTerms {
	return &ListTerms{repo: repo}
}

func (uc *ListTerms) Execute(ctx context.Context, actor TermCalendarActor, siteID uuid.UUID, includeArchived bool) ([]domain.AcademicTerm, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, err
	}

	return uc.repo.ListByBranch(ctx, actor.TenantID(), siteID, includeArchived)
}

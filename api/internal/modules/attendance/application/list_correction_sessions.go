package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/attendance/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type ListCorrectionSessions struct {
	repo domain.Repository
}

func NewListCorrectionSessions(repo domain.Repository) *ListCorrectionSessions {
	return &ListCorrectionSessions{repo: repo}
}

func (uc *ListCorrectionSessions) Execute(ctx context.Context, actor tenant.ActorContext, childID uuid.UUID, localDate time.Time) (domain.CorrectionSessionContext, error) {
	return uc.repo.ListSessionsForCorrection(ctx, actor.TenantID, actor.BranchID, childID, localDate)
}

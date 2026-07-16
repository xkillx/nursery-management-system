package application

import (
	"context"
	"time"

	"nursery-management-system/api/internal/modules/attendance/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetRegisterSummary struct {
	repo domain.Repository
}

func NewGetRegisterSummary(repo domain.Repository) *GetRegisterSummary {
	return &GetRegisterSummary{repo: repo}
}

func (uc *GetRegisterSummary) Execute(ctx context.Context, actor tenant.ActorContext, fromDate, toDate time.Time) ([]domain.RegisterSummaryEntry, error) {
	return uc.repo.GetRegisterSummary(ctx, actor.TenantID, actor.BranchID, fromDate, toDate)
}

package application

import (
	"context"
	"time"

	"nursery-management-system/api/internal/modules/attendance/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetRegister struct {
	repo domain.Repository
}

func NewGetRegister(repo domain.Repository) *GetRegister {
	return &GetRegister{repo: repo}
}

func (uc *GetRegister) Execute(ctx context.Context, actor tenant.ActorContext, registerDate time.Time) ([]domain.RegisterEntry, error) {
	dow := []int32{int32(registerDate.Weekday())}
	return uc.repo.GetRegister(ctx, actor.TenantID, actor.BranchID, registerDate, dow)
}

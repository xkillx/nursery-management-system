package application

import (
	"context"
	"fmt"

	"nursery-management-system/api/internal/modules/children/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ListAttendance struct {
	repo domain.Repository
}

func NewListAttendance(repo domain.Repository) *ListAttendance {
	return &ListAttendance{repo: repo}
}

func (uc *ListAttendance) Execute(ctx context.Context, actor tenant.ActorContext) ([]domain.AttendanceChild, error) {
	children, err := uc.repo.ListAttendance(ctx, actor.TenantID, actor.BranchID)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("list attendance children: %w", err))
	}

	return children, nil
}

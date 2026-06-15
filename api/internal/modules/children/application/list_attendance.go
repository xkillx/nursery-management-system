package application

import (
	"context"
	"fmt"
	"time"

	"nursery-management-system/api/internal/modules/children/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ListAttendance struct {
	repo  domain.Repository
	clock func() time.Time
}

func NewListAttendance(repo domain.Repository, clock func() time.Time) *ListAttendance {
	return &ListAttendance{repo: repo, clock: clock}
}

func (uc *ListAttendance) Execute(ctx context.Context, actor tenant.ActorContext) ([]domain.AttendanceChild, error) {
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("load London timezone: %w", err))
	}
	localDate := uc.clock().In(london)

	children, err := uc.repo.ListAttendance(ctx, actor.TenantID, actor.BranchID, localDate)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("list attendance children: %w", err))
	}

	return children, nil
}

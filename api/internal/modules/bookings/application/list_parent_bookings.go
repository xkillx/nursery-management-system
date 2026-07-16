package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/bookings/domain"
)

type ListParentBookings struct {
	repo      domain.Repository
	childLook ParentChildLookup
}

func NewListParentBookings(repo domain.Repository, childLook ParentChildLookup) *ListParentBookings {
	return &ListParentBookings{repo: repo, childLook: childLook}
}

func (uc *ListParentBookings) Execute(ctx context.Context, actor BookingActor, siteID uuid.UUID, from, to time.Time) ([]domain.UnifiedBookingRow, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, err
	}

	childIDs, err := uc.childLook.ListChildIDsForParent(ctx, actor.TenantID(), siteID, actor.MembershipID())
	if err != nil {
		return nil, internalError(err)
	}
	if len(childIDs) == 0 {
		return nil, nil
	}

	var allRows []domain.UnifiedBookingRow
	for _, childID := range childIDs {
		filters := domain.ListFilters{
			ChildID:    &childID,
			From:       &from,
			To:         &to,
			ActiveOnly: true,
		}
		rows, err := uc.repo.ListUnifiedByBranchPaginated(ctx, actor.TenantID(), siteID, filters, 200, 0)
		if err != nil {
			return nil, internalError(err)
		}
		allRows = append(allRows, rows...)
	}

	return allRows, nil
}

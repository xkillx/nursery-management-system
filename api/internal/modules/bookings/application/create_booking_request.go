package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/bookings/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type CreateBookingRequestParams struct {
	ChildID             uuid.UUID
	SessionTemplateID   uuid.UUID
	RoomID              uuid.UUID
	DaysOfWeek          []int32
	EffectiveStartDate  time.Time
	EffectiveEndDate    *time.Time
	FundingType         *string
	FundingHoursPerWeek *float64
	LaReference         *string
}

type CreateBookingRequest struct {
	repo      domain.Repository
	childLook ParentChildLookup
	txm       TxManager
}

func NewCreateBookingRequest(repo domain.Repository, childLook ParentChildLookup, txm TxManager) *CreateBookingRequest {
	return &CreateBookingRequest{repo: repo, childLook: childLook, txm: txm}
}

func (uc *CreateBookingRequest) Execute(ctx context.Context, actor BookingActor, siteID uuid.UUID, params CreateBookingRequestParams) (domain.Booking, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.Booking{}, err
	}

	childIDs, err := uc.childLook.ListChildIDsForParent(ctx, actor.TenantID(), siteID, actor.MembershipID())
	if err != nil {
		return domain.Booking{}, internalError(err)
	}

	childAllowed := false
	for _, cid := range childIDs {
		if cid == params.ChildID {
			childAllowed = true
			break
		}
	}
	if !childAllowed {
		return domain.Booking{}, domainerrors.Forbidden("child_not_in_scope", "Child is not linked to your account.")
	}

	booking := domain.Booking{
		ID:                   uuid.New(),
		TenantID:             actor.TenantID(),
		BranchID:             siteID,
		ChildID:              params.ChildID,
		SessionTemplateID:    params.SessionTemplateID,
		RoomID:               params.RoomID,
		DaysOfWeek:           params.DaysOfWeek,
		EffectiveStartDate:   params.EffectiveStartDate,
		EffectiveEndDate:     params.EffectiveEndDate,
		FundingType:          params.FundingType,
		FundingHoursPerWeek:  params.FundingHoursPerWeek,
		LaReference:          params.LaReference,
		Status:               domain.StatusActive,
		BookedByMembershipID: actor.MembershipID(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := uc.repo.Create(ctx, booking); err != nil {
		return domain.Booking{}, internalError(err)
	}

	return booking, nil
}

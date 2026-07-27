package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/bookings/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type CreateBookingParams struct {
	ChildID             uuid.UUID
	EffectiveStartDate  time.Time
	EffectiveEndDate    *time.Time
	FundingType         *string
	FundingHoursPerWeek *float64
	LaReference         *string
	SessionEntries      []domain.SessionEntry
	TermTimeOnly        bool
}

type CreateBooking struct {
	repo          domain.Repository
	fundingLookup domain.FundingLookup
	calendarQuery domain.CalendarQuery
}

func NewCreateBooking(repo domain.Repository, fundingLookup domain.FundingLookup, calendarQuery domain.CalendarQuery) *CreateBooking {
	return &CreateBooking{repo: repo, fundingLookup: fundingLookup, calendarQuery: calendarQuery}
}

type CreateBookingResult struct {
	Booking  domain.Booking
	Warnings []domain.ClosureWarning
}

func (uc *CreateBooking) Execute(ctx context.Context, actor BookingActor, siteID uuid.UUID, params CreateBookingParams) (CreateBookingResult, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return CreateBookingResult{}, err
	}

	if params.ChildID == uuid.Nil {
		return CreateBookingResult{}, domainerrors.Validation("Child is required.", "child_id")
	}
	if len(params.SessionEntries) == 0 {
		return CreateBookingResult{}, domainerrors.Validation("Session entries are required.", "session_entries")
	}
	if params.EffectiveEndDate != nil && params.EffectiveEndDate.Before(params.EffectiveStartDate) {
		return CreateBookingResult{}, domain.ErrInvalidDateRange
	}
	if params.FundingType != nil && !domain.ValidFundingType(*params.FundingType) {
		return CreateBookingResult{}, domain.ErrInvalidFundingType
	}

	fundingType := params.FundingType
	fundingHours := params.FundingHoursPerWeek
	laReference := params.LaReference
	termTimeOnly := params.TermTimeOnly

	if fundingType == nil {
		fi, err := uc.fundingLookup.GetChildFunding(ctx, actor.TenantID(), siteID, params.ChildID)
		if err != nil {
			return CreateBookingResult{}, internalError(err)
		}
		if fi.HasFunding {
			fundingType = &fi.FundingType
			fundingHours = fi.FundedHoursPerWeek
			laReference = fi.LaReference
			termTimeOnly = fi.TermTimeOnly
		} else {
			none := "none"
			fundingType = &none
		}
	}

	booking := domain.Booking{
		ID:                   uuid.New(),
		TenantID:             actor.TenantID(),
		BranchID:             siteID,
		ChildID:              params.ChildID,
		EffectiveStartDate:   params.EffectiveStartDate,
		EffectiveEndDate:     params.EffectiveEndDate,
		FundingType:          fundingType,
		FundingHoursPerWeek:  fundingHours,
		LaReference:          laReference,
		SessionEntries:       params.SessionEntries,
		TermTimeOnly:         termTimeOnly,
		Status:               domain.StatusActive,
		BookedByMembershipID: actor.MembershipID(),
	}

	var warnings []domain.ClosureWarning
	if uc.calendarQuery != nil {
		warnings = uc.checkClosureWarnings(ctx, actor.TenantID(), siteID, params, termTimeOnly)
	}

	if err := uc.repo.Create(ctx, booking); err != nil {
		return CreateBookingResult{}, internalError(err)
	}

	return CreateBookingResult{Booking: booking, Warnings: warnings}, nil
}

func weekdayToDayOfWeek(w time.Weekday) int {
	// Convert Go's time.Weekday (Sunday=0, Monday=1, ..., Saturday=6)
	// to domain convention (Monday=1, ..., Sunday=7).
	if w == time.Sunday {
		return 7
	}
	return int(w)
}

func (uc *CreateBooking) checkClosureWarnings(ctx context.Context, tenantID, branchID uuid.UUID, params CreateBookingParams, termTimeOnly bool) []domain.ClosureWarning {
	sessionDays := make(map[int]bool)
	for _, se := range params.SessionEntries {
		sessionDays[int(se.DayOfWeek)] = true
	}

	endDate := params.EffectiveStartDate.AddDate(0, 3, 0)
	if params.EffectiveEndDate != nil && params.EffectiveEndDate.Before(endDate) {
		endDate = *params.EffectiveEndDate
	}

	var warnings []domain.ClosureWarning
	for d := params.EffectiveStartDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		if !sessionDays[weekdayToDayOfWeek(d.Weekday())] {
			continue
		}
		isClosed, reason, err := uc.calendarQuery.CheckDate(ctx, tenantID, branchID, d, termTimeOnly)
		if err != nil {
			continue
		}
		if isClosed {
			warnings = append(warnings, domain.ClosureWarning{Date: d, Reason: reason})
		}
	}
	return warnings
}

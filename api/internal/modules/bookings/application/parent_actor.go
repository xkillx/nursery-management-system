package application

import (
	"context"

	"github.com/google/uuid"

	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

// ParentBookingActor implements BookingActor for parent portal requests.
// The parent's branch_id from the JWT token is used as the site scope.
type ParentBookingActor struct {
	actor tenant.ActorContext
}

func NewParentBookingActor(actor tenant.ActorContext) *ParentBookingActor {
	return &ParentBookingActor{actor: actor}
}

func (a *ParentBookingActor) TenantID() uuid.UUID     { return a.actor.TenantID }
func (a *ParentBookingActor) UserID() uuid.UUID       { return a.actor.UserID }
func (a *ParentBookingActor) MembershipID() uuid.UUID { return a.actor.MembershipID }
func (a *ParentBookingActor) RequestID() string       { return a.actor.RequestID }

func (a *ParentBookingActor) ValidateSiteAccess(_ context.Context, siteID uuid.UUID) error {
	if a.actor.BranchID != siteID {
		return domainerrors.Forbidden("forbidden_site_scope", "Access denied.")
	}
	return nil
}

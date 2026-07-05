package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type TxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type AdHocBookingActor interface {
	TenantID() uuid.UUID
	UserID() uuid.UUID
	MembershipID() uuid.UUID
	RequestID() string
	ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error
}

type OwnerAdHocBookingActor struct {
	actor tenant.OwnerActorContext
}

func NewOwnerAdHocBookingActor(actor tenant.OwnerActorContext) *OwnerAdHocBookingActor {
	return &OwnerAdHocBookingActor{actor: actor}
}

func (a *OwnerAdHocBookingActor) TenantID() uuid.UUID     { return a.actor.TenantID }
func (a *OwnerAdHocBookingActor) UserID() uuid.UUID       { return a.actor.UserID }
func (a *OwnerAdHocBookingActor) MembershipID() uuid.UUID { return uuid.Nil }
func (a *OwnerAdHocBookingActor) RequestID() string       { return a.actor.RequestID }
func (a *OwnerAdHocBookingActor) ValidateSiteAccess(_ context.Context, _ uuid.UUID) error {
	return nil
}

type ManagerAdHocBookingActor struct {
	actor tenant.ActorContext
}

func NewManagerAdHocBookingActor(actor tenant.ActorContext) *ManagerAdHocBookingActor {
	return &ManagerAdHocBookingActor{actor: actor}
}

func (a *ManagerAdHocBookingActor) TenantID() uuid.UUID     { return a.actor.TenantID }
func (a *ManagerAdHocBookingActor) UserID() uuid.UUID       { return a.actor.UserID }
func (a *ManagerAdHocBookingActor) MembershipID() uuid.UUID { return a.actor.MembershipID }
func (a *ManagerAdHocBookingActor) RequestID() string       { return a.actor.RequestID }

func (a *ManagerAdHocBookingActor) ValidateSiteAccess(_ context.Context, siteID uuid.UUID) error {
	if a.actor.BranchID != siteID {
		return domainerrors.Forbidden("forbidden_site_scope", "Access denied.")
	}
	return nil
}

func IsOwnerActor(actor AdHocBookingActor) bool {
	_, ok := actor.(*OwnerAdHocBookingActor)
	return ok
}

func internalError(err error) error {
	return domainerrors.Internal(fmt.Errorf("ad_hoc_bookings: %w", err))
}

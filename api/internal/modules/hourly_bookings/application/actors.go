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

type HourlyBookingActor interface {
	TenantID() uuid.UUID
	UserID() uuid.UUID
	MembershipID() uuid.UUID
	RequestID() string
	ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error
}

type OwnerHourlyBookingActor struct {
	actor tenant.OwnerActorContext
}

func NewOwnerHourlyBookingActor(actor tenant.OwnerActorContext) *OwnerHourlyBookingActor {
	return &OwnerHourlyBookingActor{actor: actor}
}

func (a *OwnerHourlyBookingActor) TenantID() uuid.UUID     { return a.actor.TenantID }
func (a *OwnerHourlyBookingActor) UserID() uuid.UUID       { return a.actor.UserID }
func (a *OwnerHourlyBookingActor) MembershipID() uuid.UUID { return uuid.Nil }
func (a *OwnerHourlyBookingActor) RequestID() string       { return a.actor.RequestID }
func (a *OwnerHourlyBookingActor) ValidateSiteAccess(_ context.Context, _ uuid.UUID) error {
	return nil
}

type ManagerHourlyBookingActor struct {
	actor tenant.ActorContext
}

func NewManagerHourlyBookingActor(actor tenant.ActorContext) *ManagerHourlyBookingActor {
	return &ManagerHourlyBookingActor{actor: actor}
}

func (a *ManagerHourlyBookingActor) TenantID() uuid.UUID     { return a.actor.TenantID }
func (a *ManagerHourlyBookingActor) UserID() uuid.UUID       { return a.actor.UserID }
func (a *ManagerHourlyBookingActor) MembershipID() uuid.UUID { return a.actor.MembershipID }
func (a *ManagerHourlyBookingActor) RequestID() string       { return a.actor.RequestID }

func (a *ManagerHourlyBookingActor) ValidateSiteAccess(_ context.Context, siteID uuid.UUID) error {
	if a.actor.BranchID != siteID {
		return domainerrors.Forbidden("forbidden_site_scope", "Access denied.")
	}
	return nil
}

func IsOwnerActor(actor HourlyBookingActor) bool {
	_, ok := actor.(*OwnerHourlyBookingActor)
	return ok
}

func internalError(err error) error {
	return domainerrors.Internal(fmt.Errorf("hourly_bookings: %w", err))
}

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

type BookingActor interface {
	TenantID() uuid.UUID
	UserID() uuid.UUID
	MembershipID() uuid.UUID
	RequestID() string
	ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error
}

type OwnerBookingActor struct {
	actor tenant.OwnerActorContext
}

func NewOwnerBookingActor(actor tenant.OwnerActorContext) *OwnerBookingActor {
	return &OwnerBookingActor{actor: actor}
}

func (a *OwnerBookingActor) TenantID() uuid.UUID     { return a.actor.TenantID }
func (a *OwnerBookingActor) UserID() uuid.UUID       { return a.actor.UserID }
func (a *OwnerBookingActor) MembershipID() uuid.UUID { return uuid.Nil }
func (a *OwnerBookingActor) RequestID() string       { return a.actor.RequestID }
func (a *OwnerBookingActor) ValidateSiteAccess(_ context.Context, _ uuid.UUID) error {
	return nil
}

type ManagerBookingActor struct {
	actor tenant.ActorContext
}

func NewManagerBookingActor(actor tenant.ActorContext) *ManagerBookingActor {
	return &ManagerBookingActor{actor: actor}
}

func (a *ManagerBookingActor) TenantID() uuid.UUID     { return a.actor.TenantID }
func (a *ManagerBookingActor) UserID() uuid.UUID       { return a.actor.UserID }
func (a *ManagerBookingActor) MembershipID() uuid.UUID { return a.actor.MembershipID }
func (a *ManagerBookingActor) RequestID() string       { return a.actor.RequestID }

func (a *ManagerBookingActor) ValidateSiteAccess(_ context.Context, siteID uuid.UUID) error {
	if a.actor.BranchID != siteID {
		return domainerrors.Forbidden("forbidden_site_scope", "Access denied.")
	}
	return nil
}

func IsOwnerActor(actor BookingActor) bool {
	_, ok := actor.(*OwnerBookingActor)
	return ok
}

func internalError(err error) error {
	return domainerrors.Internal(fmt.Errorf("bookings: %w", err))
}

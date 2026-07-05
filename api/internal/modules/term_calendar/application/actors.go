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

type SiteExistsChecker interface {
	SiteExists(ctx context.Context, tenantID, siteID uuid.UUID) (bool, error)
}

type TermCalendarActor interface {
	TenantID() uuid.UUID
	UserID() uuid.UUID
	RequestID() string
	ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error
}

type OwnerTermCalendarActor struct {
	actor tenant.OwnerActorContext
}

func NewOwnerTermCalendarActor(actor tenant.OwnerActorContext) *OwnerTermCalendarActor {
	return &OwnerTermCalendarActor{actor: actor}
}

func (a *OwnerTermCalendarActor) TenantID() uuid.UUID { return a.actor.TenantID }
func (a *OwnerTermCalendarActor) UserID() uuid.UUID   { return a.actor.UserID }
func (a *OwnerTermCalendarActor) RequestID() string   { return a.actor.RequestID }
func (a *OwnerTermCalendarActor) ValidateSiteAccess(_ context.Context, _ uuid.UUID) error {
	return nil
}

type ManagerTermCalendarActor struct {
	actor tenant.ActorContext
}

func NewManagerTermCalendarActor(actor tenant.ActorContext) *ManagerTermCalendarActor {
	return &ManagerTermCalendarActor{actor: actor}
}

func (a *ManagerTermCalendarActor) TenantID() uuid.UUID { return a.actor.TenantID }
func (a *ManagerTermCalendarActor) UserID() uuid.UUID   { return a.actor.UserID }
func (a *ManagerTermCalendarActor) RequestID() string   { return a.actor.RequestID }

func (a *ManagerTermCalendarActor) ValidateSiteAccess(_ context.Context, siteID uuid.UUID) error {
	if a.actor.BranchID != siteID {
		return domainerrors.Forbidden("forbidden_site_scope", "Access denied.")
	}
	return nil
}

func IsOwnerActor(actor TermCalendarActor) bool {
	_, ok := actor.(*OwnerTermCalendarActor)
	return ok
}

func siteIDFromUUIDOrError(siteID string) (uuid.UUID, error) {
	id, err := uuid.Parse(siteID)
	if err != nil {
		return uuid.Nil, domainerrors.Validation("Invalid request payload.", "site_id")
	}
	return id, nil
}

func termIDFromUUIDOrError(id string) (uuid.UUID, error) {
	v, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, domainerrors.Validation("Invalid request payload.", "term_id")
	}
	return v, nil
}

func internalError(err error) error {
	return domainerrors.Internal(fmt.Errorf("term_calendar: %w", err))
}

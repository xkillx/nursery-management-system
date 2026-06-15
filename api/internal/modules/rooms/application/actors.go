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

type RoomActor interface {
	TenantID() uuid.UUID
	UserID() uuid.UUID
	RequestID() string
	ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error
}

type OwnerRoomActor struct {
	actor tenant.OwnerActorContext
}

func NewOwnerRoomActor(actor tenant.OwnerActorContext) *OwnerRoomActor {
	return &OwnerRoomActor{actor: actor}
}

func (a *OwnerRoomActor) TenantID() uuid.UUID { return a.actor.TenantID }
func (a *OwnerRoomActor) UserID() uuid.UUID   { return a.actor.UserID }
func (a *OwnerRoomActor) RequestID() string   { return a.actor.RequestID }

func (a *OwnerRoomActor) ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error {
	return nil
}

type ManagerRoomActor struct {
	actor tenant.ActorContext
}

func NewManagerRoomActor(actor tenant.ActorContext) *ManagerRoomActor {
	return &ManagerRoomActor{actor: actor}
}

func (a *ManagerRoomActor) TenantID() uuid.UUID { return a.actor.TenantID }
func (a *ManagerRoomActor) UserID() uuid.UUID   { return a.actor.UserID }
func (a *ManagerRoomActor) RequestID() string   { return a.actor.RequestID }

func (a *ManagerRoomActor) ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error {
	if a.actor.BranchID != siteID {
		return domainerrors.Forbidden("forbidden_site_scope", "Access denied.")
	}
	return nil
}

type PractitionerRoomActor struct {
	actor tenant.ActorContext
}

func NewPractitionerRoomActor(actor tenant.ActorContext) *PractitionerRoomActor {
	return &PractitionerRoomActor{actor: actor}
}

func (a *PractitionerRoomActor) TenantID() uuid.UUID { return a.actor.TenantID }
func (a *PractitionerRoomActor) UserID() uuid.UUID   { return a.actor.UserID }
func (a *PractitionerRoomActor) RequestID() string   { return a.actor.RequestID }

func (a *PractitionerRoomActor) ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error {
	if a.actor.BranchID != siteID {
		return domainerrors.Forbidden("forbidden_site_scope", "Access denied.")
	}
	return nil
}

func IsOwnerActor(actor RoomActor) bool {
	_, ok := actor.(*OwnerRoomActor)
	return ok
}

func siteIDFromUUIDOrError(siteID string) (uuid.UUID, error) {
	id, err := uuid.Parse(siteID)
	if err != nil {
		return uuid.Nil, domainerrors.Validation("Invalid request payload.", "site_id")
	}
	return id, nil
}

func roomIDFromUUIDOrError(roomID string) (uuid.UUID, error) {
	id, err := uuid.Parse(roomID)
	if err != nil {
		return uuid.Nil, domainerrors.Validation("Invalid request payload.", "room_id")
	}
	return id, nil
}

func internalError(err error) error {
	return domainerrors.Internal(fmt.Errorf("rooms: %w", err))
}

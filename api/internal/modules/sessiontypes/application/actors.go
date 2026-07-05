package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

type SessionTypeActor interface {
	TenantID() uuid.UUID
	UserID() uuid.UUID
	RequestID() string
	ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error
}

type OwnerSessionTypeActor struct {
	actor tenant.OwnerActorContext
}

func NewOwnerSessionTypeActor(actor tenant.OwnerActorContext) *OwnerSessionTypeActor {
	return &OwnerSessionTypeActor{actor: actor}
}

func (a *OwnerSessionTypeActor) TenantID() uuid.UUID { return a.actor.TenantID }
func (a *OwnerSessionTypeActor) UserID() uuid.UUID   { return a.actor.UserID }
func (a *OwnerSessionTypeActor) RequestID() string   { return a.actor.RequestID }

func (a *OwnerSessionTypeActor) ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error {
	return nil
}

type ManagerSessionTypeActor struct {
	actor tenant.ActorContext
}

func NewManagerSessionTypeActor(actor tenant.ActorContext) *ManagerSessionTypeActor {
	return &ManagerSessionTypeActor{actor: actor}
}

func (a *ManagerSessionTypeActor) TenantID() uuid.UUID { return a.actor.TenantID }
func (a *ManagerSessionTypeActor) UserID() uuid.UUID   { return a.actor.UserID }
func (a *ManagerSessionTypeActor) RequestID() string   { return a.actor.RequestID }

func (a *ManagerSessionTypeActor) ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error {
	if a.actor.BranchID != siteID {
		return domainerrors.Forbidden("forbidden_site_scope", "Access denied.")
	}
	return nil
}

type PractitionerSessionTypeActor struct {
	actor tenant.ActorContext
}

func NewPractitionerSessionTypeActor(actor tenant.ActorContext) *PractitionerSessionTypeActor {
	return &PractitionerSessionTypeActor{actor: actor}
}

func (a *PractitionerSessionTypeActor) TenantID() uuid.UUID { return a.actor.TenantID }
func (a *PractitionerSessionTypeActor) UserID() uuid.UUID   { return a.actor.UserID }
func (a *PractitionerSessionTypeActor) RequestID() string   { return a.actor.RequestID }

func (a *PractitionerSessionTypeActor) ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error {
	if a.actor.BranchID != siteID {
		return domainerrors.Forbidden("forbidden_site_scope", "Access denied.")
	}
	return nil
}

func IsOwnerActor(actor SessionTypeActor) bool {
	_, ok := actor.(*OwnerSessionTypeActor)
	return ok
}

func siteIDFromUUIDOrError(siteID string) (uuid.UUID, error) {
	id, err := uuid.Parse(siteID)
	if err != nil {
		return uuid.Nil, domainerrors.Validation("Invalid request payload.", "site_id")
	}
	return id, nil
}

func sessionTypeIDFromUUIDOrError(id string) (uuid.UUID, error) {
	v, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, domainerrors.Validation("Invalid request payload.", "session_type_id")
	}
	return v, nil
}

// parseHHMM parses an "HH:MM" 24-hour string into minutes since midnight.
func parseHHMM(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty time")
	}
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid format")
	}
	hh, err := strconv.Atoi(parts[0])
	if err != nil || hh < 0 || hh > 23 {
		return 0, fmt.Errorf("invalid hour")
	}
	mm, err := strconv.Atoi(parts[1])
	if err != nil || mm < 0 || mm > 59 {
		return 0, fmt.Errorf("invalid minute")
	}
	return hh*60 + mm, nil
}

// minutesToHHMM converts minutes since midnight to "HH:MM" (zero-padded).
func minutesToHHMM(m int) string {
	hh := m / 60
	mm := m % 60
	return fmt.Sprintf("%02d:%02d", hh, mm)
}

func internalError(err error) error {
	return domainerrors.Internal(fmt.Errorf("sessiontypes: %w", err))
}

func validSessionTypeKind(k string) bool {
	switch k {
	case "standard", "wraparound_before", "wraparound_after", "core", "extended":
		return true
	}
	return false
}

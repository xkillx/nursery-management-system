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

// TxManager is the interface satisfied by *transaction.Manager.
type TxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

// SiteExistsChecker is implemented by the owner repository; used to validate
// that a site exists when an owner actor is performing a write.
type SiteExistsChecker interface {
	SiteExists(ctx context.Context, tenantID, siteID uuid.UUID) (bool, error)
}

// SessionTypeLookup is implemented by an adapter in the bootstrap layer. The
// sessiontemplates module depends on this small projection so it does not need
// to import the sessiontypes module directly.
type SessionTypeLookup interface {
	GetActiveInScope(ctx context.Context, tenantID, branchID, sessionTypeID uuid.UUID) (SessionTypeInfo, bool, error)
}

// SessionTypeInfo is a minimal projection of a session type used to validate
// template entries. It mirrors the children-module projection exactly so the
// adapter can satisfy both interfaces.
type SessionTypeInfo struct {
	ID           uuid.UUID
	Name         string
	StartMinutes int
	EndMinutes   int
	IsActive     bool
}

// SessionTemplateActor is the minimal actor surface needed by the use cases.
// Owner, manager, and practitioner actors each implement it.
type SessionTemplateActor interface {
	TenantID() uuid.UUID
	UserID() uuid.UUID
	RequestID() string
	ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error
}

type OwnerSessionTemplateActor struct {
	actor tenant.OwnerActorContext
}

func NewOwnerSessionTemplateActor(actor tenant.OwnerActorContext) *OwnerSessionTemplateActor {
	return &OwnerSessionTemplateActor{actor: actor}
}

func (a *OwnerSessionTemplateActor) TenantID() uuid.UUID { return a.actor.TenantID }
func (a *OwnerSessionTemplateActor) UserID() uuid.UUID   { return a.actor.UserID }
func (a *OwnerSessionTemplateActor) RequestID() string   { return a.actor.RequestID }

func (a *OwnerSessionTemplateActor) ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error {
	return nil
}

type ManagerSessionTemplateActor struct {
	actor tenant.ActorContext
}

func NewManagerSessionTemplateActor(actor tenant.ActorContext) *ManagerSessionTemplateActor {
	return &ManagerSessionTemplateActor{actor: actor}
}

func (a *ManagerSessionTemplateActor) TenantID() uuid.UUID { return a.actor.TenantID }
func (a *ManagerSessionTemplateActor) UserID() uuid.UUID   { return a.actor.UserID }
func (a *ManagerSessionTemplateActor) RequestID() string   { return a.actor.RequestID }

func (a *ManagerSessionTemplateActor) ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error {
	if a.actor.BranchID != siteID {
		return domainerrors.Forbidden("forbidden_site_scope", "Access denied.")
	}
	return nil
}

type PractitionerSessionTemplateActor struct {
	actor tenant.ActorContext
}

func NewPractitionerSessionTemplateActor(actor tenant.ActorContext) *PractitionerSessionTemplateActor {
	return &PractitionerSessionTemplateActor{actor: actor}
}

func (a *PractitionerSessionTemplateActor) TenantID() uuid.UUID { return a.actor.TenantID }
func (a *PractitionerSessionTemplateActor) UserID() uuid.UUID   { return a.actor.UserID }
func (a *PractitionerSessionTemplateActor) RequestID() string   { return a.actor.RequestID }

func (a *PractitionerSessionTemplateActor) ValidateSiteAccess(ctx context.Context, siteID uuid.UUID) error {
	if a.actor.BranchID != siteID {
		return domainerrors.Forbidden("forbidden_site_scope", "Access denied.")
	}
	return nil
}

func IsOwnerActor(actor SessionTemplateActor) bool {
	_, ok := actor.(*OwnerSessionTemplateActor)
	return ok
}

func siteIDFromUUIDOrError(siteID string) (uuid.UUID, error) {
	id, err := uuid.Parse(siteID)
	if err != nil {
		return uuid.Nil, domainerrors.Validation("Invalid request payload.", "site_id")
	}
	return id, nil
}

func templateIDFromUUIDOrError(id string) (uuid.UUID, error) {
	v, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, domainerrors.Validation("Invalid request payload.", "template_id")
	}
	return v, nil
}

func parseName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", domainerrors.Validation("Invalid request payload.", "name")
	}
	if len(s) > 120 {
		return "", domainerrors.Validation("Invalid request payload.", "name")
	}
	return s, nil
}

func parseDescription(s *string) (*string, error) {
	if s == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil, nil
	}
	if len(trimmed) > 1000 {
		return nil, domainerrors.Validation("Invalid request payload.", "description")
	}
	return &trimmed, nil
}

func parseInt(s string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, err
	}
	return n, nil
}

func internalError(err error) error {
	return domainerrors.Internal(fmt.Errorf("sessiontemplates: %w", err))
}

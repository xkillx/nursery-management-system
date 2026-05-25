package domain

import (
	"time"

	"github.com/google/uuid"
)

type Invite struct {
	ID                     uuid.UUID
	TenantID               uuid.UUID
	BranchID               uuid.UUID
	Email                  string
	EmailNormalized        string
	Role                   string
	TokenHash              string
	ExpiresAt              time.Time
	AcceptedAt             *time.Time
	AcceptedUserID         uuid.UUID
	AcceptedMembershipID   uuid.UUID
	RevokedAt              *time.Time
	RevokedByUserID        uuid.UUID
	RevokedByMembershipID  uuid.UUID
	CreatedByUserID        uuid.UUID
	CreatedByMembershipID  uuid.UUID
	ResentAt               *time.Time
	ResentByUserID         uuid.UUID
	ResentByMembershipID   uuid.UUID
	SendCount              int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type InviteStatus string

const (
	StatusPending  InviteStatus = "pending"
	StatusAccepted InviteStatus = "accepted"
	StatusRevoked  InviteStatus = "revoked"
	StatusExpired  InviteStatus = "expired"
	StatusAll      InviteStatus = "all"
)

func (inv *Invite) Status() InviteStatus {
	if inv.AcceptedAt != nil {
		return StatusAccepted
	}
	if inv.RevokedAt != nil {
		return StatusRevoked
	}
	if inv.ExpiresAt.Before(time.Now().UTC()) {
		return StatusExpired
	}
	return StatusPending
}

func (inv *Invite) IsLivePending() bool {
	return inv.AcceptedAt == nil && inv.RevokedAt == nil && inv.ExpiresAt.After(time.Now().UTC())
}

type CreatedUser struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
}

type CreatedMembership struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	BranchID uuid.UUID
	UserID   uuid.UUID
	Role     string
}

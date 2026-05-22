package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	IsActive     bool
}

type Membership struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	BranchID uuid.UUID
	Role     string
	IsActive bool
}

type RefreshToken struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	MembershipID uuid.UUID
	TokenHash    string
	ExpiresAt    time.Time
	RevokedAt    *time.Time
}

type ScopeClaims struct {
	MembershipID string
	TenantID     string
	BranchID     string
	Role         string
}

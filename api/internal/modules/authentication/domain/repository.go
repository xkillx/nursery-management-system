package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	FindUserByEmail(ctx context.Context, email string) (User, error)
	ListMembershipsByUserID(ctx context.Context, userID uuid.UUID) ([]Membership, error)
}

type SessionRepository interface {
	CreateRefreshToken(ctx context.Context, token RefreshToken, userAgent, ipAddress string) error
	FindActiveRefreshToken(ctx context.Context, tokenHash string) (RefreshToken, User, Membership, error)
	RotateRefreshToken(ctx context.Context, oldTokenID uuid.UUID, replacement RefreshToken, userAgent, ipAddress string) error
	RevokeByTokenHash(ctx context.Context, tokenHash string) error
	CreateScopeSwitchAuditLog(ctx context.Context, actorUserID uuid.UUID, fromMembership, toMembership Membership, requestID string) error
}

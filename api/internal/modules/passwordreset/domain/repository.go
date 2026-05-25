package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	FindUserByEmail(ctx context.Context, emailNormalized string) (User, bool, error)
	IssueResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time, sendEmail func() error) error
	ResetPassword(ctx context.Context, tokenHash string, newPasswordHash string) error
}

type EmailSender interface {
	SendPasswordReset(ctx context.Context, toEmail string, resetURL string) error
}

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

type PasswordResetToken struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	TokenHash    string
	ExpiresAt    time.Time
	UsedAt       *time.Time
	SupersededAt *time.Time
}

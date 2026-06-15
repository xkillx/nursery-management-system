package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ChildInfo struct {
	FirstName   string
	MiddleName  string
	LastName    string
	DateOfBirth time.Time
	StartDate   time.Time
	Notes       string
}

type ChildCreationResult struct {
	ID         uuid.UUID
	FirstName  string
	MiddleName *string
	LastName   *string
	StartDate  time.Time
}

type ChildCreator interface {
	CreateChild(ctx context.Context, tx pgx.Tx, child ChildInfo, tenantID, branchID uuid.UUID) (ChildCreationResult, error)
}

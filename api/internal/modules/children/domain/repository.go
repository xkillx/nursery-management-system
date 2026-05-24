package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Tx is a transaction interface matching pgx.Tx for dependency injection.
type Tx = pgx.Tx

type StatusFilter string

const (
	StatusActive   StatusFilter = "active"
	StatusInactive StatusFilter = "inactive"
	StatusAll      StatusFilter = "all"
)

type ChildCorrectionInfo struct {
	ID        uuid.UUID
	StartDate time.Time
	EndDate   *time.Time
}

type Repository interface {
	List(ctx context.Context, tenantID, branchID uuid.UUID, filter StatusFilter, limit, offset int) ([]Child, error)
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (Child, bool, error)
	Create(ctx context.Context, child Child, notes string, tenantID, branchID uuid.UUID) error
	Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error)
	MarkInactive(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (Child, bool, error)
	ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (bool, error)
	ListAttendance(ctx context.Context, tenantID, branchID uuid.UUID, localDate time.Time) ([]AttendanceChild, error)
	GetChildForCorrection(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (ChildCorrectionInfo, bool, error)
}

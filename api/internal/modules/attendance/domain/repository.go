package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CorrectionParams struct {
	SessionID  *uuid.UUID
	ChildID    *uuid.UUID
	CheckInAt  time.Time
	CheckOutAt time.Time
	ReasonCode string
	ReasonNote string
}

type Repository interface {
	CreateOpenSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, occurredAt time.Time, localDate time.Time, userID, membershipID uuid.UUID, requestID string) (Session, error)
	GetOpenSessionForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (Session, bool, error)
	CompleteSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, session Session, occurredAt time.Time, localDate time.Time, userID, membershipID uuid.UUID, requestID string) (Session, error)
	GetSessionForCorrection(ctx context.Context, tx pgx.Tx, tenantID, branchID, sessionID uuid.UUID) (Session, bool, error)
	CreateCorrectedSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, params CorrectionParams, checkInLocalDate, checkOutLocalDate time.Time, occurredAt time.Time, userID, membershipID uuid.UUID, requestID string) (Session, error)
	CorrectSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, session Session, params CorrectionParams, checkInLocalDate, checkOutLocalDate time.Time, occurredAt time.Time, userID, membershipID uuid.UUID, requestID string) (Session, error)
	HasOverlappingSession(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, excludeSessionID *uuid.UUID, checkInAt, checkOutAt time.Time) (bool, error)
}

type ChildCorrectionChecker interface {
	GetChildForCorrection(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (ChildCorrectionInfo, bool, error)
}

type ChildCorrectionInfo struct {
	ID        uuid.UUID
	StartDate time.Time
	EndDate   *time.Time
}

type ChildEnrollmentChecker interface {
	CheckEnrollmentForAttendance(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) error
}

package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository interface {
	CreateOpenSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, occurredAt time.Time, localDate time.Time, userID, membershipID uuid.UUID, requestID string) (Session, error)
	GetOpenSessionForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (Session, bool, error)
	CompleteSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, session Session, occurredAt time.Time, localDate time.Time, userID, membershipID uuid.UUID, requestID string) (Session, error)
}

type ChildEnrollmentChecker interface {
	CheckEnrollmentForAttendance(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) error
}

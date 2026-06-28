package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Tx = any

type Repository interface {
	Create(ctx context.Context, tx Tx, marker AbsenceMarker) (AbsenceMarker, error)
	FindActiveByChildDate(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) (AbsenceMarker, bool, error)
	GetByID(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (AbsenceMarker, bool, error)
	Clear(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID, clearedAt time.Time, clearedByUserID, clearedByMembershipID uuid.UUID) (AbsenceMarker, bool, error)
	HasAttendanceForChildDate(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) (bool, error)
}

type ChildEnrollmentChecker interface {
	CheckEnrollmentForAttendance(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) error
}

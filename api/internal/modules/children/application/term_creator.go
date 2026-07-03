package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/platform/tenant"
)

type EnrollmentTermCreator interface {
	CreateEnrollmentTerm(ctx context.Context, tx pgx.Tx, actor tenant.ActorContext, childID uuid.UUID, termStartDate time.Time, bookingPatternID uuid.UUID) (uuid.UUID, error)
}

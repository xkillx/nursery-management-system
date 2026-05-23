package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	postgresguardian "nursery-management-system/api/internal/modules/guardians/infrastructure/postgres"
	postgreschild "nursery-management-system/api/internal/modules/children/infrastructure/postgres"
	postgresparent "nursery-management-system/api/internal/modules/parentmappings/infrastructure/postgres"
	"nursery-management-system/api/internal/modules/parentmappings/domain"
	attendancedomain "nursery-management-system/api/internal/modules/attendance/domain"
	childdomain "nursery-management-system/api/internal/modules/children/domain"
)

type guardianCheckerAdapter struct {
	repo *postgresguardian.GuardianRepository
}

func (a *guardianCheckerAdapter) IsActive(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID) (bool, bool, error) {
	return a.repo.GetActive(ctx, tx, tenantID, branchID, guardianID)
}

type childCheckerAdapter struct {
	repo *postgreschild.ChildRepository
}

func (a *childCheckerAdapter) ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (bool, error) {
	return a.repo.ExistsInScope(ctx, tx, tenantID, branchID, childID)
}

type membershipCheckerAdapter struct {
	repo *postgresparent.ParentMappingRepository
}

func (a *membershipCheckerAdapter) GetForScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, membershipID uuid.UUID) (domain.MembershipInfo, bool, error) {
	return a.repo.GetMembershipForScope(ctx, tx, tenantID, branchID, membershipID)
}

type childEnrollmentCheckerAdapter struct {
	repo *postgreschild.ChildRepository
}

func (a *childEnrollmentCheckerAdapter) CheckEnrollmentForAttendance(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) error {
	child, found, err := a.repo.GetForAttendanceCheck(ctx, tx, tenantID, branchID, childID)
	if err != nil {
		return fmt.Errorf("check child enrollment: %w", err)
	}
	if !found {
		return attendancedomain.ErrChildNotFound
	}
	if !child.IsActive {
		return attendancedomain.ErrChildNotFound
	}
	if !child.EnrollmentComplete() {
		return attendancedomain.ErrChildEnrollmentIncomplete
	}
	if localDate.Before(child.StartDate) {
		return attendancedomain.ErrChildEnrollmentIncomplete
	}
	if child.EndDate != nil && localDate.After(*child.EndDate) {
		return attendancedomain.ErrChildEnrollmentIncomplete
	}
	return nil
}

// Ensure adapter satisfies the interface at compile time.
var _ childdomain.Repository = (*postgreschild.ChildRepository)(nil)

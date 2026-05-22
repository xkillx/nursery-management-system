package bootstrap

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	postgresguardian "nursery-management-system/api/internal/modules/guardians/infrastructure/postgres"
	postgreschild "nursery-management-system/api/internal/modules/children/infrastructure/postgres"
	postgresparent "nursery-management-system/api/internal/modules/parentmappings/infrastructure/postgres"
	"nursery-management-system/api/internal/modules/parentmappings/domain"
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

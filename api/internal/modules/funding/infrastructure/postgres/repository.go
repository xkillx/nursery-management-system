package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/funding/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Get(ctx context.Context, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) (domain.FundingProfile, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.FundingProfileGet(ctx, sqlc.FundingProfileGetParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		ChildID:      uuidToPgtype(childID),
		BillingMonth: timeToPgtypeDate(billingMonth),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.FundingProfile{}, false, nil
	}
	if err != nil {
		return domain.FundingProfile{}, false, fmt.Errorf("get funding profile: %w", err)
	}
	return mapProfile(row), true, nil
}

func (r *Repository) GetForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) (domain.FundingProfile, bool, error) {
	q := sqlc.New(tx)
	row, err := q.FundingProfileGetForUpdate(ctx, sqlc.FundingProfileGetForUpdateParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		ChildID:      uuidToPgtype(childID),
		BillingMonth: timeToPgtypeDate(billingMonth),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.FundingProfile{}, false, nil
	}
	if err != nil {
		return domain.FundingProfile{}, false, fmt.Errorf("get funding profile for update: %w", err)
	}
	return mapProfile(row), true, nil
}

func (r *Repository) Create(ctx context.Context, tx domain.Tx, profile domain.FundingProfile) (domain.FundingProfile, error) {
	q := sqlc.New(tx)
	row, err := q.FundingProfileCreate(ctx, sqlc.FundingProfileCreateParams{
		ID:                     uuidToPgtype(profile.ID),
		TenantID:               uuidToPgtype(profile.TenantID),
		BranchID:               uuidToPgtype(profile.BranchID),
		ChildID:                uuidToPgtype(profile.ChildID),
		BillingMonth:           timeToPgtypeDate(profile.BillingMonth),
		FundedAllowanceMinutes: int32(profile.FundedAllowanceMinutes),
	})
	if err != nil {
		return domain.FundingProfile{}, fmt.Errorf("create funding profile: %w", err)
	}
	return mapProfile(row), nil
}

func (r *Repository) UpdateAllowance(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID, billingMonth time.Time, minutes int) (domain.FundingProfile, error) {
	q := sqlc.New(tx)
	row, err := q.FundingProfileUpdateAllowance(ctx, sqlc.FundingProfileUpdateAllowanceParams{
		FundedAllowanceMinutes: int32(minutes),
		TenantID:               uuidToPgtype(tenantID),
		BranchID:               uuidToPgtype(branchID),
		ChildID:                uuidToPgtype(childID),
		BillingMonth:           timeToPgtypeDate(billingMonth),
	})
	if err != nil {
		return domain.FundingProfile{}, fmt.Errorf("update funding profile allowance: %w", err)
	}
	return mapProfile(row), nil
}

func (r *Repository) GetChildEnrollmentForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID) (domain.ChildEnrollment, bool, error) {
	q := sqlc.New(tx)
	row, err := q.FundingChildEnrollmentGetForUpdate(ctx, sqlc.FundingChildEnrollmentGetForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ChildEnrollment{}, false, nil
	}
	if err != nil {
		return domain.ChildEnrollment{}, false, fmt.Errorf("get child enrollment for update: %w", err)
	}
	return domain.ChildEnrollment{
		ID:        pgtypeUUIDToUUID(row.ID),
		StartDate: pgtypeDateToTime(row.StartDate),
		EndDate:   pgtypeDateToTimePtr(row.EndDate),
	}, true, nil
}

func mapProfile(row sqlc.FundingProfile) domain.FundingProfile {
	return domain.FundingProfile{
		ID:                     pgtypeUUIDToUUID(row.ID),
		TenantID:               pgtypeUUIDToUUID(row.TenantID),
		BranchID:               pgtypeUUIDToUUID(row.BranchID),
		ChildID:                pgtypeUUIDToUUID(row.ChildID),
		BillingMonth:           pgtypeDateToTime(row.BillingMonth),
		FundedAllowanceMinutes: int(row.FundedAllowanceMinutes),
		CreatedAt:              pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:              pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func timeToPgtypeDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func pgtypeDateToTime(d pgtype.Date) time.Time {
	return d.Time
}

func pgtypeDateToTimePtr(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

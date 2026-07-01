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

	"nursery-management-system/api/internal/modules/siteprofile/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type SiteProfileRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *SiteProfileRepository {
	return &SiteProfileRepository{pool: pool}
}

func (r *SiteProfileRepository) GetByBranch(ctx context.Context, tenantID, branchID uuid.UUID) (domain.SiteProfile, error) {
	q := sqlc.New(r.pool)
	row, err := q.SiteProfileGetByBranch(ctx, sqlc.SiteProfileGetByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
	})
	if isNoRows(err) {
		return domain.SiteProfile{}, domainerrors.NotFound("site_profile", "Site profile not found.")
	}
	if err != nil {
		return domain.SiteProfile{}, fmt.Errorf("query site profile by branch: %w", err)
	}
	return mapSiteProfile(row), nil
}

func (r *SiteProfileRepository) Upsert(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, profile domain.SiteProfile) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.SiteProfileUpsert(ctx, sqlc.SiteProfileUpsertParams{
		ID:              uuidToPgtype(profile.ID),
		TenantID:        uuidToPgtype(tenantID),
		BranchID:        uuidToPgtype(branchID),
		NurseryName:     profile.NurseryName,
		Description:     profile.Description,
		Phone:           profile.Phone,
		Email:           profile.Email,
		Website:         profile.Website,
		AddressStreet:   profile.AddressStreet,
		AddressCity:     profile.AddressCity,
		AddressPostcode: profile.AddressPostcode,
	})
}

func mapSiteProfile(row sqlc.SiteProfile) domain.SiteProfile {
	return domain.SiteProfile{
		ID:              pgtypeUUIDToUUID(row.ID),
		TenantID:        pgtypeUUIDToUUID(row.TenantID),
		BranchID:        pgtypeUUIDToUUID(row.BranchID),
		NurseryName:     row.NurseryName,
		Description:     row.Description,
		Phone:           row.Phone,
		Email:           row.Email,
		Website:         row.Website,
		AddressStreet:   row.AddressStreet,
		AddressCity:     row.AddressCity,
		AddressPostcode: row.AddressPostcode,
		CreatedAt:       pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	if !u.Valid {
		return uuid.Nil
	}
	return uuid.UUID(u.Bytes)
}

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

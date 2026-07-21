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

type FundingRecordRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewFundingRecordRepository(pool *pgxpool.Pool) *FundingRecordRepositoryImpl {
	return &FundingRecordRepositoryImpl{pool: pool}
}

func (r *FundingRecordRepositoryImpl) GetFundingRecord(ctx context.Context, tenantID, branchID, childID uuid.UUID) (domain.FundingRecord, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildFundingRecordGetByChild(ctx, sqlc.ChildFundingRecordGetByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.FundingRecord{}, false, nil
	}
	if err != nil {
		return domain.FundingRecord{}, false, fmt.Errorf("get funding record: %w", err)
	}
	return mapFundingRecordFromGetRow(row), true, nil
}

func (r *FundingRecordRepositoryImpl) UpsertFundingRecord(ctx context.Context, tx domain.Tx, record domain.FundingRecord) (domain.FundingRecord, error) {
	var q *sqlc.Queries
	if tx != nil {
		q = sqlc.New(tx.(pgx.Tx))
	} else {
		q = sqlc.New(r.pool)
	}

	var startDate, endDate pgtype.Date
	if record.FundingStartDate != nil {
		startDate = timeToPgtypeDate(*record.FundingStartDate)
	}
	if record.FundingEndDate != nil {
		endDate = timeToPgtypeDate(*record.FundingEndDate)
	}

	row, err := q.ChildFundingRecordUpsert(ctx, sqlc.ChildFundingRecordUpsertParams{
		ID:                       uuidToPgtype(record.ID),
		TenantID:                 uuidToPgtype(record.TenantID),
		BranchID:                 uuidToPgtype(record.BranchID),
		ChildID:                  uuidToPgtype(record.ChildID),
		FundingEnabled:           record.FundingEnabled,
		FundingType:              string(record.FundingType),
		FundingModel:             string(record.FundingModel),
		Column8:                  ptrToNumeric(record.FundedHoursPerWeek),
		Column9:                  startDate,
		Column10:                 endDate,
		Column11:                 record.EligibilityCode,
		EligibilityCodeValidated: record.EligibilityCodeValidated,
		EvidenceReceived:         record.EvidenceReceived,
		BenefitsStatus:           "unknown",
	})
	if err != nil {
		return domain.FundingRecord{}, fmt.Errorf("upsert funding record: %w", err)
	}
	return mapFundingRecordFromUpsertRow(row), nil
}

func mapFundingRecordFromGetRow(row sqlc.ChildFundingRecordGetByChildRow) domain.FundingRecord {
	var startDate, endDate *time.Time
	if row.FundingStartDate.Valid {
		t := row.FundingStartDate.Time
		startDate = &t
	}
	if row.FundingEndDate.Valid {
		t := row.FundingEndDate.Time
		endDate = &t
	}

	var fundedHoursPerWeek *float64
	if row.FundedHoursPerWeek.Valid {
		f, _ := row.FundedHoursPerWeek.Float64Value()
		v := f.Float64
		fundedHoursPerWeek = &v
	}

	var eligibilityCode *string
	if row.EligibilityCode.Valid {
		s := row.EligibilityCode.String
		eligibilityCode = &s
	}

	return domain.FundingRecord{
		ID:                       pgtypeUUIDToUUID(row.ID),
		TenantID:                 pgtypeUUIDToUUID(row.TenantID),
		BranchID:                 pgtypeUUIDToUUID(row.BranchID),
		ChildID:                  pgtypeUUIDToUUID(row.ChildID),
		FundingEnabled:           row.FundingEnabled,
		FundingType:              domain.FundingType(row.FundingType),
		FundingModel:             domain.FundingModel(row.FundingModel),
		FundedHoursPerWeek:       fundedHoursPerWeek,
		FundingStartDate:         startDate,
		FundingEndDate:           endDate,
		EligibilityCode:          eligibilityCode,
		EligibilityCodeValidated: row.EligibilityCodeValidated,
		EvidenceReceived:         row.EvidenceReceived,
		CreatedAt:                row.CreatedAt.Time,
		UpdatedAt:                row.UpdatedAt.Time,
	}
}

func mapFundingRecordFromUpsertRow(row sqlc.ChildFundingRecordUpsertRow) domain.FundingRecord {
	var startDate, endDate *time.Time
	if row.FundingStartDate.Valid {
		t := row.FundingStartDate.Time
		startDate = &t
	}
	if row.FundingEndDate.Valid {
		t := row.FundingEndDate.Time
		endDate = &t
	}

	var fundedHoursPerWeek *float64
	if row.FundedHoursPerWeek.Valid {
		f, _ := row.FundedHoursPerWeek.Float64Value()
		v := f.Float64
		fundedHoursPerWeek = &v
	}

	var eligibilityCode *string
	if row.EligibilityCode.Valid {
		s := row.EligibilityCode.String
		eligibilityCode = &s
	}

	return domain.FundingRecord{
		ID:                       pgtypeUUIDToUUID(row.ID),
		TenantID:                 pgtypeUUIDToUUID(row.TenantID),
		BranchID:                 pgtypeUUIDToUUID(row.BranchID),
		ChildID:                  pgtypeUUIDToUUID(row.ChildID),
		FundingEnabled:           row.FundingEnabled,
		FundingType:              domain.FundingType(row.FundingType),
		FundingModel:             domain.FundingModel(row.FundingModel),
		FundedHoursPerWeek:       fundedHoursPerWeek,
		FundingStartDate:         startDate,
		FundingEndDate:           endDate,
		EligibilityCode:          eligibilityCode,
		EligibilityCodeValidated: row.EligibilityCodeValidated,
		EvidenceReceived:         row.EvidenceReceived,
		CreatedAt:                row.CreatedAt.Time,
		UpdatedAt:                row.UpdatedAt.Time,
	}
}

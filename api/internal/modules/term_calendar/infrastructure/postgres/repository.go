package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/term_calendar/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type AcademicTermRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *AcademicTermRepository {
	return &AcademicTermRepository{pool: pool}
}

func (r *AcademicTermRepository) ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) ([]domain.AcademicTerm, error) {
	q := sqlc.New(r.pool)
	rows, err := q.AcademicTermsListByBranch(ctx, sqlc.AcademicTermsListByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  !includeArchived,
	})
	if err != nil {
		return nil, fmt.Errorf("query academic terms list: %w", err)
	}
	out := make([]domain.AcademicTerm, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapAcademicTerm(row))
	}
	return out, nil
}

func (r *AcademicTermRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.AcademicTerm, error) {
	q := sqlc.New(r.pool)
	row, err := q.AcademicTermsGetByID(ctx, sqlc.AcademicTermsGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if isNoRows(err) {
		return domain.AcademicTerm{}, domainerrors.NotFound("academic_term", "Academic term not found.")
	}
	if err != nil {
		return domain.AcademicTerm{}, fmt.Errorf("query academic term by id: %w", err)
	}
	return mapAcademicTerm(row), nil
}

func (r *AcademicTermRepository) Create(ctx context.Context, term domain.AcademicTerm) error {
	q := sqlc.New(r.pool)
	return q.AcademicTermsCreate(ctx, sqlc.AcademicTermsCreateParams{
		ID:        uuidToPgtype(term.ID),
		TenantID:  uuidToPgtype(term.TenantID),
		BranchID:  uuidToPgtype(term.BranchID),
		Name:      term.Name,
		Kind:      term.Kind,
		StartDate: timeToPgtypeDate(term.StartDate),
		EndDate:   timeToPgtypeDate(term.EndDate),
	})
}

func (r *AcademicTermRepository) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}

	params := sqlc.AcademicTermsUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	}

	if v, ok := fields["name"]; ok {
		params.SetName = true
		params.Name = v.(string)
	}
	if v, ok := fields["kind"]; ok {
		params.SetKind = true
		params.Kind = v.(string)
	}
	if v, ok := fields["start_date"]; ok {
		params.SetStartDate = true
		params.StartDate = timeToPgtypeDate(v.(time.Time))
	}
	if v, ok := fields["end_date"]; ok {
		params.SetEndDate = true
		params.EndDate = timeToPgtypeDate(v.(time.Time))
	}

	q := sqlc.New(r.pool)
	return q.AcademicTermsUpdate(ctx, params)
}

func (r *AcademicTermRepository) Archive(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.AcademicTermsArchive(ctx, sqlc.AcademicTermsArchiveParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func (r *AcademicTermRepository) ActiveNameExists(ctx context.Context, tenantID, branchID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	q := sqlc.New(r.pool)
	params := sqlc.AcademicTermsCheckActiveNameExistsParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Name:     name,
		Column4:  pgtype.UUID{Valid: false},
	}
	if excludeID != nil {
		params.Column4 = uuidToPgtype(*excludeID)
	}
	return q.AcademicTermsCheckActiveNameExists(ctx, params)
}

func (r *AcademicTermRepository) GetByIDForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) (domain.AcademicTerm, error) {
	q := sqlc.New(tx.(pgx.Tx))
	row, err := q.AcademicTermsGetByIDForUpdate(ctx, sqlc.AcademicTermsGetByIDForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if isNoRows(err) {
		return domain.AcademicTerm{}, domainerrors.NotFound("academic_term", "Academic term not found.")
	}
	if err != nil {
		return domain.AcademicTerm{}, fmt.Errorf("query academic term for update: %w", err)
	}
	return mapAcademicTerm(row), nil
}

func (r *AcademicTermRepository) ListActiveDateRanges(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) ([]domain.TermDateRange, error) {
	q := sqlc.New(r.pool)
	rows, err := q.AcademicTermsListActiveDateRanges(ctx, sqlc.AcademicTermsListActiveDateRangesParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		EndDate:   timeToPgtypeDate(from),
		StartDate: timeToPgtypeDate(to),
	})
	if err != nil {
		return nil, fmt.Errorf("query active date ranges: %w", err)
	}
	out := make([]domain.TermDateRange, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.TermDateRange{
			StartDate: pgtypeDateToTime(row.StartDate),
			EndDate:   pgtypeDateToTime(row.EndDate),
		})
	}
	return out, nil
}

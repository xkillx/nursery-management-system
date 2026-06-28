package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/term/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type TermRepository struct {
	pool *pgxpool.Pool
}

func NewTermRepository(pool *pgxpool.Pool) *TermRepository {
	return &TermRepository{pool: pool}
}

func (r *TermRepository) q() *sqlc.Queries               { return sqlc.New(r.pool) }
func (r *TermRepository) qTx(tx domain.Tx) *sqlc.Queries { return sqlc.New(tx.(pgx.Tx)) }

func (r *TermRepository) Insert(ctx context.Context, tx domain.Tx, t *domain.Term) (*domain.Term, error) {
	row, err := r.qTx(tx).TermInsert(ctx, sqlc.TermInsertParams{
		ID:                    pgtypeUUID(t.ID),
		TenantID:              pgtypeUUID(t.TenantID),
		BranchID:              pgtypeUUID(t.BranchID),
		ChildID:               pgtypeUUID(t.ChildID),
		TermStartDate:         pgtypeDate(t.TermStartDate),
		TermEndDate:           pgtypeDate(t.TermEndDate),
		BookingPatternID:      pgtypeUUID(t.BookingPatternID),
		SiteHourlyRateMinor:   int32(t.SiteHourlyRateMinor),
		Status:                string(t.Status),
		CreatedByMembershipID: pgtypeUUID(t.CreatedByMembershipID),
	})
	if err != nil {
		return nil, fmt.Errorf("insert term: %w", err)
	}
	return mapTermRow(row), nil
}

func (r *TermRepository) Terminate(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID, terminatedAt time.Time, reasonCode, reasonNote string) (int64, error) {
	rows, err := r.qTx(tx).TermTerminate(ctx, sqlc.TermTerminateParams{
		TenantID:              pgtypeUUID(tenantID),
		BranchID:              pgtypeUUID(branchID),
		ID:                    pgtypeUUID(id),
		TerminatedAt:          pgtypeTimestamptz(terminatedAt),
		TerminationReasonCode: pgtypeText(reasonCode),
		TerminationReasonNote: pgtypeText(reasonNote),
	})
	if err != nil {
		return 0, fmt.Errorf("terminate term: %w", err)
	}
	return rows, nil
}

func (r *TermRepository) UpdateStatus(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID, status domain.TermStatus) (int64, error) {
	rows, err := r.qTx(tx).TermUpdateStatus(ctx, sqlc.TermUpdateStatusParams{
		TenantID: pgtypeUUID(tenantID),
		BranchID: pgtypeUUID(branchID),
		ID:       pgtypeUUID(id),
		Status:   string(status),
	})
	if err != nil {
		return 0, fmt.Errorf("update term status: %w", err)
	}
	return rows, nil
}

func (r *TermRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (*domain.Term, bool, error) {
	row, err := r.q().TermGetByID(ctx, sqlc.TermGetByIDParams{
		TenantID: pgtypeUUID(tenantID),
		BranchID: pgtypeUUID(branchID),
		ID:       pgtypeUUID(id),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get term by id: %w", err)
	}
	return mapTermRow(row), true, nil
}

func (r *TermRepository) GetByIDInTx(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) (*domain.Term, bool, error) {
	row, err := r.qTx(tx).TermGetByID(ctx, sqlc.TermGetByIDParams{
		TenantID: pgtypeUUID(tenantID),
		BranchID: pgtypeUUID(branchID),
		ID:       pgtypeUUID(id),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get term by id in tx: %w", err)
	}
	return mapTermRow(row), true, nil
}

func (r *TermRepository) GetActiveForChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.Term, bool, error) {
	row, err := r.q().TermGetActiveForChild(ctx, sqlc.TermGetActiveForChildParams{
		TenantID: pgtypeUUID(tenantID),
		BranchID: pgtypeUUID(branchID),
		ChildID:  pgtypeUUID(childID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get active term: %w", err)
	}
	return mapTermRow(row), true, nil
}

func (r *TermRepository) GetActiveForChildInTx(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID) (*domain.Term, bool, error) {
	row, err := r.qTx(tx).TermGetActiveForChild(ctx, sqlc.TermGetActiveForChildParams{
		TenantID: pgtypeUUID(tenantID),
		BranchID: pgtypeUUID(branchID),
		ChildID:  pgtypeUUID(childID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get active term in tx: %w", err)
	}
	return mapTermRow(row), true, nil
}

func (r *TermRepository) ListByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]domain.Term, error) {
	rows, err := r.q().TermListByChild(ctx, sqlc.TermListByChildParams{
		TenantID: pgtypeUUID(tenantID),
		BranchID: pgtypeUUID(branchID),
		ChildID:  pgtypeUUID(childID),
	})
	if err != nil {
		return nil, fmt.Errorf("list terms by child: %w", err)
	}
	out := make([]domain.Term, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapTermRow(row))
	}
	return out, nil
}

func (r *TermRepository) ListActiveByBranch(ctx context.Context, tenantID, branchID uuid.UUID) ([]domain.Term, error) {
	rows, err := r.q().TermListActiveByBranch(ctx, sqlc.TermListActiveByBranchParams{
		TenantID: pgtypeUUID(tenantID),
		BranchID: pgtypeUUID(branchID),
	})
	if err != nil {
		return nil, fmt.Errorf("list active terms by branch: %w", err)
	}
	out := make([]domain.Term, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapTermRow(row))
	}
	return out, nil
}

func (r *TermRepository) ListExpiringWithin(ctx context.Context, tenantID, branchID uuid.UUID, maxTermEndDate time.Time) ([]domain.Term, error) {
	rows, err := r.q().TermListExpiringWithin(ctx, sqlc.TermListExpiringWithinParams{
		TenantID:    pgtypeUUID(tenantID),
		BranchID:    pgtypeUUID(branchID),
		TermEndDate: pgtypeDate(maxTermEndDate),
	})
	if err != nil {
		return nil, fmt.Errorf("list expiring terms: %w", err)
	}
	out := make([]domain.Term, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapTermRow(row))
	}
	return out, nil
}

func (r *TermRepository) ListEndingOnOrBefore(ctx context.Context, tenantID, branchID uuid.UUID, endDate time.Time) ([]domain.Term, error) {
	rows, err := r.q().TermListEndingOnOrBefore(ctx, sqlc.TermListEndingOnOrBeforeParams{
		TenantID:    pgtypeUUID(tenantID),
		BranchID:    pgtypeUUID(branchID),
		TermEndDate: pgtypeDate(endDate),
	})
	if err != nil {
		return nil, fmt.Errorf("list ending terms: %w", err)
	}
	out := make([]domain.Term, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapTermRow(row))
	}
	return out, nil
}

func (r *TermRepository) ListActiveInBillingMonth(ctx context.Context, tenantID, branchID uuid.UUID, monthStart, monthEnd time.Time) ([]domain.Term, error) {
	rows, err := r.q().TermListActiveInBillingMonth(ctx, sqlc.TermListActiveInBillingMonthParams{
		TenantID:      pgtypeUUID(tenantID),
		BranchID:      pgtypeUUID(branchID),
		TermStartDate: pgtypeDate(monthStart),
		TermEndDate:   pgtypeDate(monthEnd),
	})
	if err != nil {
		return nil, fmt.Errorf("list active terms in billing month: %w", err)
	}
	out := make([]domain.Term, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapTermRow(row))
	}
	return out, nil
}

func (r *TermRepository) ListActiveForChildUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID) ([]domain.Term, error) {
	rows, err := r.qTx(tx).TermListForChildUpdateLock(ctx, sqlc.TermListForChildUpdateLockParams{
		TenantID: pgtypeUUID(tenantID),
		BranchID: pgtypeUUID(branchID),
		ChildID:  pgtypeUUID(childID),
	})
	if err != nil {
		return nil, fmt.Errorf("list active terms for update: %w", err)
	}
	out := make([]domain.Term, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapTermRow(row))
	}
	return out, nil
}

func (r *TermRepository) SetChildCurrentTermID(ctx context.Context, tx domain.Tx, tenantID, branchID, childID, termID uuid.UUID) error {
	err := r.qTx(tx).ChildSetCurrentTermID(ctx, sqlc.ChildSetCurrentTermIDParams{
		TenantID:      pgtypeUUID(tenantID),
		BranchID:      pgtypeUUID(branchID),
		ID:            pgtypeUUID(childID),
		CurrentTermID: pgtypeUUID(termID),
	})
	if err != nil {
		return fmt.Errorf("set child current term: %w", err)
	}
	return nil
}

func (r *TermRepository) ClearChildCurrentTermID(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID) error {
	err := r.qTx(tx).ChildClearCurrentTermID(ctx, sqlc.ChildClearCurrentTermIDParams{
		TenantID: pgtypeUUID(tenantID),
		BranchID: pgtypeUUID(branchID),
		ID:       pgtypeUUID(childID),
	})
	if err != nil {
		return fmt.Errorf("clear child current term: %w", err)
	}
	return nil
}

func mapTermRow(row sqlc.Term) *domain.Term {
	return &domain.Term{
		ID:                    pgtypeUUIDToUUID(row.ID),
		TenantID:              pgtypeUUIDToUUID(row.TenantID),
		BranchID:              pgtypeUUIDToUUID(row.BranchID),
		ChildID:               pgtypeUUIDToUUID(row.ChildID),
		TermStartDate:         pgtypeDateToTime(row.TermStartDate),
		TermEndDate:           pgtypeDateToTime(row.TermEndDate),
		BookingPatternID:      pgtypeUUIDToUUID(row.BookingPatternID),
		SiteHourlyRateMinor:   int(row.SiteHourlyRateMinor),
		Status:                domain.TermStatus(row.Status),
		TerminationReasonCode: pgtypeTextToStrPtr(row.TerminationReasonCode),
		TerminationReasonNote: pgtypeTextToStrPtr(row.TerminationReasonNote),
		TerminatedAt:          pgtypeTimestamptzToTimePtr(row.TerminatedAt),
		CreatedAt:             pgtypeTimestamptzToTime(row.CreatedAt),
		CreatedByMembershipID: pgtypeUUIDToUUID(row.CreatedByMembershipID),
		UpdatedAt:             pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

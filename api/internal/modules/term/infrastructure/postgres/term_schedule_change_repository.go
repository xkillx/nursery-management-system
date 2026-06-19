package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/term/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type ScheduleChangeRepository struct {
	pool *pgxpool.Pool
}

func NewScheduleChangeRepository(pool *pgxpool.Pool) *ScheduleChangeRepository {
	return &ScheduleChangeRepository{pool: pool}
}

func (r *ScheduleChangeRepository) q() *sqlc.Queries            { return sqlc.New(r.pool) }
func (r *ScheduleChangeRepository) qTx(tx pgx.Tx) *sqlc.Queries { return sqlc.New(tx) }

func (r *ScheduleChangeRepository) Insert(ctx context.Context, tx pgx.Tx, c *domain.TermScheduleChange) (*domain.TermScheduleChange, error) {
	row, err := r.qTx(tx).TermScheduleChangeInsert(ctx, sqlc.TermScheduleChangeInsertParams{
		ID:                       pgtypeUUID(c.ID),
		TenantID:                 pgtypeUUID(c.TenantID),
		BranchID:                 pgtypeUUID(c.BranchID),
		TermID:                   pgtypeUUID(c.TermID),
		PreviousBookingPatternID: pgtypeUUID(c.PreviousBookingPatternID),
		NewBookingPatternID:      pgtypeUUID(c.NewBookingPatternID),
		ChangeKind:               string(c.ChangeKind),
		EffectiveFrom:            pgtypeDate(c.EffectiveFrom),
		RequestID:                c.RequestID,
	})
	if err != nil {
		return nil, fmt.Errorf("insert schedule change: %w", err)
	}
	return mapScheduleChangeRow(row), nil
}

func (r *ScheduleChangeRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (*domain.TermScheduleChange, bool, error) {
	row, err := r.q().TermScheduleChangeGetByID(ctx, sqlc.TermScheduleChangeGetByIDParams{
		TenantID: pgtypeUUID(tenantID),
		BranchID: pgtypeUUID(branchID),
		ID:       pgtypeUUID(id),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get schedule change: %w", err)
	}
	return mapScheduleChangeRow(row), true, nil
}

func (r *ScheduleChangeRepository) ListByTerm(ctx context.Context, tenantID, branchID, termID uuid.UUID) ([]domain.TermScheduleChange, error) {
	rows, err := r.q().TermScheduleChangeListByTerm(ctx, sqlc.TermScheduleChangeListByTermParams{
		TenantID: pgtypeUUID(tenantID),
		BranchID: pgtypeUUID(branchID),
		TermID:   pgtypeUUID(termID),
	})
	if err != nil {
		return nil, fmt.Errorf("list schedule changes: %w", err)
	}
	out := make([]domain.TermScheduleChange, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapScheduleChangeRow(row))
	}
	return out, nil
}

func (r *ScheduleChangeRepository) Approve(ctx context.Context, tx pgx.Tx, tenantID, branchID, id, approverMembershipID uuid.UUID) (int64, error) {
	rows, err := r.qTx(tx).TermScheduleChangeApprove(ctx, sqlc.TermScheduleChangeApproveParams{
		TenantID:               pgtypeUUID(tenantID),
		BranchID:               pgtypeUUID(branchID),
		ID:                     pgtypeUUID(id),
		ApprovedByMembershipID: pgtypeUUID(approverMembershipID),
	})
	if err != nil {
		return 0, fmt.Errorf("approve schedule change: %w", err)
	}
	return rows, nil
}

func (r *ScheduleChangeRepository) Reject(ctx context.Context, tx pgx.Tx, tenantID, branchID, id, approverMembershipID uuid.UUID) (int64, error) {
	rows, err := r.qTx(tx).TermScheduleChangeReject(ctx, sqlc.TermScheduleChangeRejectParams{
		TenantID:               pgtypeUUID(tenantID),
		BranchID:               pgtypeUUID(branchID),
		ID:                     pgtypeUUID(id),
		ApprovedByMembershipID: pgtypeUUID(approverMembershipID),
	})
	if err != nil {
		return 0, fmt.Errorf("reject schedule change: %w", err)
	}
	return rows, nil
}

func mapScheduleChangeRow(row sqlc.TermScheduleChange) *domain.TermScheduleChange {
	out := &domain.TermScheduleChange{
		ID:                       pgtypeUUIDToUUID(row.ID),
		TenantID:                 pgtypeUUIDToUUID(row.TenantID),
		BranchID:                 pgtypeUUIDToUUID(row.BranchID),
		TermID:                   pgtypeUUIDToUUID(row.TermID),
		PreviousBookingPatternID: pgtypeUUIDToUUID(row.PreviousBookingPatternID),
		NewBookingPatternID:      pgtypeUUIDToUUID(row.NewBookingPatternID),
		ChangeKind:               domain.ScheduleChangeKind(row.ChangeKind),
		RequestedAt:              pgtypeTimestamptzToTime(row.RequestedAt),
		EffectiveFrom:            pgtypeDateToTime(row.EffectiveFrom),
		ApprovedByMembershipID:   pgtypeUUIDToUUIDPtr(row.ApprovedByMembershipID),
		RejectedAt:               pgtypeTimestamptzToTimePtr(row.RejectedAt),
		RequestID:                row.RequestID,
	}
	if row.ApprovalDecision.Valid {
		d := domain.ScheduleChangeDecision(row.ApprovalDecision.String)
		out.ApprovalDecision = &d
	}
	return out
}

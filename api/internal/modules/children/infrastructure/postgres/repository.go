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

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type ChildRepository struct {
	pool *pgxpool.Pool
}

func NewChildRepository(pool *pgxpool.Pool) *ChildRepository {
	return &ChildRepository{pool: pool}
}

func (r *ChildRepository) List(ctx context.Context, tenantID, branchID uuid.UUID, filter domain.StatusFilter, limit, offset int) ([]domain.Child, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ChildrenList(ctx, sqlc.ChildrenListParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		Limit:        int32(limit),
		Offset:       int32(offset),
		StatusFilter: string(filter),
	})
	if err != nil {
		return nil, fmt.Errorf("query children: %w", err)
	}
	out := make([]domain.Child, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapChildRow(row.ID, row.FullName, row.DateOfBirth, row.StartDate, row.EndDate,
			row.CoreHourlyRateMinor, row.Notes, row.IsActive, row.LeftAt, row.LeftReasonCode,
			row.LeftReasonNote, row.HasGuardianLink, row.CreatedAt, row.UpdatedAt))
	}
	return out, nil
}

func (r *ChildRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.Child, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.ChildrenGetByID(ctx, sqlc.ChildrenGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Child{}, false, nil
	}
	if err != nil {
		return domain.Child{}, false, fmt.Errorf("query child by id: %w", err)
	}
	return mapChildRow(row.ID, row.FullName, row.DateOfBirth, row.StartDate, row.EndDate,
		row.CoreHourlyRateMinor, row.Notes, row.IsActive, row.LeftAt, row.LeftReasonCode,
		row.LeftReasonNote, row.HasGuardianLink, row.CreatedAt, row.UpdatedAt), true, nil
}

func (r *ChildRepository) Create(ctx context.Context, child domain.Child, notes string, tenantID, branchID uuid.UUID) error {
	q := sqlc.New(r.pool)
	return q.ChildrenCreate(ctx, sqlc.ChildrenCreateParams{
		ID:                  uuidToPgtype(child.ID),
		TenantID:            uuidToPgtype(tenantID),
		BranchID:            uuidToPgtype(branchID),
		FullName:            child.FullName,
		DateOfBirth:         timeToPgtypeDate(child.DateOfBirth),
		StartDate:           timeToPgtypeDate(child.StartDate),
		EndDate:             timeToPgtypeDatePtr(child.EndDate),
		CoreHourlyRateMinor: int32(child.CoreHourlyRateMinor),
		Column9:             notes,
	})
}

func (r *ChildRepository) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}

	params := sqlc.ChildrenUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	}

	if v, ok := fields["full_name"]; ok {
		params.SetFullName = int32(1)
		params.FullName = v.(string)
	}
	if v, ok := fields["date_of_birth"]; ok {
		params.SetDateOfBirth = int32(1)
		params.DateOfBirth = timeToPgtypeDate(v.(time.Time))
	}
	if v, ok := fields["start_date"]; ok {
		params.SetStartDate = int32(1)
		params.StartDate = timeToPgtypeDate(v.(time.Time))
	}
	if v, ok := fields["end_date"]; ok {
		params.SetEndDate = int32(1)
		if t, ok := v.(time.Time); ok {
			params.EndDate = timeToPgtypeDate(t)
		}
	}
	if v, ok := fields["core_hourly_rate_minor"]; ok {
		params.SetCoreHourlyRateMinor = int32(1)
		params.CoreHourlyRateMinor = int32(v.(int))
	}
	if v, ok := fields["notes"]; ok {
		params.SetNotes = int32(1)
		params.Notes = v.(string)
	}

	q := sqlc.New(r.pool)
	ct, err := q.ChildrenUpdate(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("update child: %w", err)
	}
	return ct, nil
}

func (r *ChildRepository) MarkInactive(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error {
	q := sqlc.New(tx)
	return q.ChildrenMarkInactive(ctx, sqlc.ChildrenMarkInactiveParams{
		LeftReasonCode: reasonCode,
		Column2:        reasonNote,
		TenantID:       uuidToPgtype(tenantID),
		BranchID:       uuidToPgtype(branchID),
		ID:             uuidToPgtype(id),
	})
}

func (r *ChildRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.Child, bool, error) {
	q := sqlc.New(tx)
	row, err := q.ChildrenGetByIDForUpdate(ctx, sqlc.ChildrenGetByIDForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Child{}, false, nil
	}
	if err != nil {
		return domain.Child{}, false, fmt.Errorf("query child for update: %w", err)
	}
	return mapChildRow(row.ID, row.FullName, row.DateOfBirth, row.StartDate, row.EndDate,
		row.CoreHourlyRateMinor, row.Notes, row.IsActive, row.LeftAt, row.LeftReasonCode,
		row.LeftReasonNote, row.HasGuardianLink, row.CreatedAt, row.UpdatedAt), true, nil
}

func (r *ChildRepository) ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (bool, error) {
	q := sqlc.New(tx)
	exists, err := q.ChildrenExistsInScope(ctx, sqlc.ChildrenExistsInScopeParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if err != nil {
		return false, fmt.Errorf("check child exists in scope: %w", err)
	}
	return exists, nil
}

func (r *ChildRepository) ListAttendance(ctx context.Context, tenantID, branchID uuid.UUID, localDate time.Time) ([]domain.AttendanceChild, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ChildrenListAttendance(ctx, sqlc.ChildrenListAttendanceParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		LocalDate: timeToPgtypeDate(localDate),
	})
	if err != nil {
		return nil, fmt.Errorf("query attendance children: %w", err)
	}
	out := make([]domain.AttendanceChild, 0, len(rows))
	for _, row := range rows {
		var hasIncomplete bool
		if v, ok := row.HasIncompleteSession.(bool); ok {
			hasIncomplete = v
		}
		out = append(out, domain.AttendanceChild{
			ID:                   pgtypeUUIDToUUID(row.ID),
			FullName:             row.FullName,
			EnrollmentComplete:   row.EnrollmentComplete.Bool,
			AttendanceState:      row.AttendanceState,
			OpenSessionID:        pgtypeUUIDToUUIDPtr(row.OpenSessionID),
			CheckedInAt:          pgtypeTimestamptzToTimePtr(row.CheckedInAt),
			HasIncompleteSession: hasIncomplete,
			AbsenceMarkerID:      pgtypeUUIDToUUIDPtr(row.AbsenceMarkerID),
			AbsenceMarkedAt:      pgtypeTimestamptzToTimePtr(row.AbsenceMarkedAt),
		})
	}
	return out, nil
}

func (r *ChildRepository) GetForAttendanceCheck(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (domain.Child, bool, error) {
	q := sqlc.New(tx)
	row, err := q.ChildrenGetByID(ctx, sqlc.ChildrenGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Child{}, false, nil
	}
	if err != nil {
		return domain.Child{}, false, fmt.Errorf("get child for attendance check: %w", err)
	}
	return mapChildRow(row.ID, row.FullName, row.DateOfBirth, row.StartDate, row.EndDate,
		row.CoreHourlyRateMinor, row.Notes, row.IsActive, row.LeftAt, row.LeftReasonCode,
		row.LeftReasonNote, row.HasGuardianLink, row.CreatedAt, row.UpdatedAt), true, nil
}

func (r *ChildRepository) GetChildForCorrection(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (domain.ChildCorrectionInfo, bool, error) {
	q := sqlc.New(tx)
	row, err := q.ChildrenGetForCorrection(ctx, sqlc.ChildrenGetForCorrectionParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ChildCorrectionInfo{}, false, nil
	}
	if err != nil {
		return domain.ChildCorrectionInfo{}, false, fmt.Errorf("get child for correction: %w", err)
	}
	return domain.ChildCorrectionInfo{
		ID:        pgtypeUUIDToUUID(row.ID),
		StartDate: pgtypeDateToTime(row.StartDate),
		EndDate:   pgtypeDateToTimePtr(row.EndDate),
	}, true, nil
}

func mapChildRow(
	id pgtype.UUID, fullName string, dateOfBirth, startDate, endDate pgtype.Date,
	coreHourlyRateMinor int32, notes pgtype.Text, isActive bool, leftAt pgtype.Timestamptz,
	leftReasonCode interface{}, leftReasonNote pgtype.Text, hasGuardianLink bool,
	createdAt, updatedAt pgtype.Timestamptz,
) domain.Child {
	return domain.Child{
		ID:                  pgtypeUUIDToUUID(id),
		FullName:            fullName,
		DateOfBirth:         pgtypeDateToTime(dateOfBirth),
		StartDate:           pgtypeDateToTime(startDate),
		EndDate:             pgtypeDateToTimePtr(endDate),
		CoreHourlyRateMinor: int(coreHourlyRateMinor),
		Notes:               pgtypeTextToStringPtr(notes),
		IsActive:            isActive,
		LeftAt:              pgtypeTimestamptzToTimePtr(leftAt),
		LeftReasonCode:      ifaceToStringPtr(leftReasonCode),
		LeftReasonNote:      pgtypeTextToStringPtr(leftReasonNote),
		HasGuardianLink:     hasGuardianLink,
		CreatedAt:           pgtypeTimestamptzToTime(createdAt),
		UpdatedAt:           pgtypeTimestamptzToTime(updatedAt),
	}
}

func ifaceToStringPtr(v interface{}) *string {
	if v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}
	return &s
}

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func pgtypeUUIDToUUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

func timeToPgtypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func pgtypeTimestamptzToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func timeToPgtypeDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func timeToPgtypeDatePtr(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *t, Valid: true}
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

func pgtypeTextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

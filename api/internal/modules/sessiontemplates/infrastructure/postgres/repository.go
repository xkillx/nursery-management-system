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

	"nursery-management-system/api/internal/modules/sessiontemplates/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type SessionTemplateRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *SessionTemplateRepository {
	return &SessionTemplateRepository{pool: pool}
}

func (r *SessionTemplateRepository) ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) ([]domain.SessionTemplate, error) {
	q := sqlc.New(r.pool)
	rows, err := q.SessionTemplatesListByBranch(ctx, sqlc.SessionTemplatesListByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  !includeArchived,
	})
	if err != nil {
		return nil, fmt.Errorf("query session templates list: %w", err)
	}
	out := make([]domain.SessionTemplate, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSessionTemplate(row))
	}
	return out, nil
}

func (r *SessionTemplateRepository) ListByBranchPaginated(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool, limit, offset int) ([]domain.SessionTemplate, error) {
	q := sqlc.New(r.pool)
	rows, err := q.SessionTemplatesListByBranchPaginated(ctx, sqlc.SessionTemplatesListByBranchPaginatedParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  !includeArchived,
		Limit:    pgtype.Int4{Int32: int32(limit), Valid: true},
		Offset:   pgtype.Int4{Int32: int32(offset), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("query session templates list paginated: %w", err)
	}
	out := make([]domain.SessionTemplate, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSessionTemplate(row))
	}
	return out, nil
}

func (r *SessionTemplateRepository) CountByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) (int, error) {
	q := sqlc.New(r.pool)
	count, err := q.SessionTemplatesCountByBranch(ctx, sqlc.SessionTemplatesCountByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  !includeArchived,
	})
	if err != nil {
		return 0, fmt.Errorf("query session templates count: %w", err)
	}
	return int(count), nil
}

func (r *SessionTemplateRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.SessionTemplate, error) {
	q := sqlc.New(r.pool)
	row, err := q.SessionTemplatesGetByID(ctx, sqlc.SessionTemplatesGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SessionTemplate{}, domainerrors.NotFound("session_template", "Session template not found.")
	}
	if err != nil {
		return domain.SessionTemplate{}, fmt.Errorf("query session template by id: %w", err)
	}
	return mapSessionTemplate(row), nil
}

func (r *SessionTemplateRepository) Create(ctx context.Context, t domain.SessionTemplate) error {
	q := sqlc.New(r.pool)
	var desc interface{}
	if t.Description != nil {
		desc = *t.Description
	}
	return q.SessionTemplatesCreate(ctx, sqlc.SessionTemplatesCreateParams{
		ID:       uuidToPgtype(t.ID),
		TenantID: uuidToPgtype(t.TenantID),
		BranchID: uuidToPgtype(t.BranchID),
		Name:     t.Name,
		Column5:  desc,
	})
}

func (r *SessionTemplateRepository) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}

	params := sqlc.SessionTemplatesUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	}

	if v, ok := fields["name"]; ok {
		params.SetName = int32(1)
		params.Name = v.(string)
	}
	if v, ok := fields["description"]; ok {
		params.SetDescription = int32(1)
		if d, ok := v.(*string); ok {
			if d == nil {
				params.Description = interface{}(nil)
			} else {
				params.Description = *d
			}
		}
	}

	q := sqlc.New(r.pool)
	return q.SessionTemplatesUpdate(ctx, params)
}

func (r *SessionTemplateRepository) Archive(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.SessionTemplatesArchive(ctx, sqlc.SessionTemplatesArchiveParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func (r *SessionTemplateRepository) Reactivate(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.SessionTemplatesReactivate(ctx, sqlc.SessionTemplatesReactivateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func (r *SessionTemplateRepository) ActiveNameExists(ctx context.Context, tenantID, branchID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	q := sqlc.New(r.pool)
	params := sqlc.SessionTemplatesCheckActiveNameExistsParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Name:     name,
		Column4:  pgtype.UUID{Valid: false},
	}
	if excludeID != nil {
		params.Column4 = uuidToPgtype(*excludeID)
	}
	return q.SessionTemplatesCheckActiveNameExists(ctx, params)
}

func (r *SessionTemplateRepository) Exists(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) (bool, error) {
	q := sqlc.New(tx.(pgx.Tx))
	return q.SessionTemplatesExists(ctx, sqlc.SessionTemplatesExistsParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func (r *SessionTemplateRepository) GetByIDForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) (domain.SessionTemplate, error) {
	q := sqlc.New(tx.(pgx.Tx))
	row, err := q.SessionTemplatesGetByIDForUpdate(ctx, sqlc.SessionTemplatesGetByIDForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SessionTemplate{}, domainerrors.NotFound("session_template", "Session template not found.")
	}
	if err != nil {
		return domain.SessionTemplate{}, fmt.Errorf("query session template for update: %w", err)
	}
	return mapSessionTemplate(row), nil
}

func (r *SessionTemplateRepository) InsertEntry(ctx context.Context, tx domain.Tx, entry domain.SessionTemplateEntry) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.SessionTemplateEntriesInsert(ctx, sqlc.SessionTemplateEntriesInsertParams{
		ID:            uuidToPgtype(entry.ID),
		TenantID:      uuidToPgtype(entry.TenantID),
		BranchID:      uuidToPgtype(entry.BranchID),
		TemplateID:    uuidToPgtype(entry.TemplateID),
		DayOfWeek:     int32(entry.DayOfWeek),
		SessionTypeID: uuidToPgtype(entry.SessionTypeID),
	})
}

func (r *SessionTemplateRepository) DeleteEntriesByTemplate(ctx context.Context, tx domain.Tx, tenantID, branchID, templateID uuid.UUID) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.SessionTemplateEntriesDeleteByTemplate(ctx, sqlc.SessionTemplateEntriesDeleteByTemplateParams{
		TenantID:   uuidToPgtype(tenantID),
		BranchID:   uuidToPgtype(branchID),
		TemplateID: uuidToPgtype(templateID),
	})
}

func (r *SessionTemplateRepository) EntriesListByTemplate(ctx context.Context, tenantID, branchID, templateID uuid.UUID) ([]domain.SessionTemplateEntry, error) {
	q := sqlc.New(r.pool)
	rows, err := q.SessionTemplateEntriesListByTemplate(ctx, sqlc.SessionTemplateEntriesListByTemplateParams{
		TenantID:   uuidToPgtype(tenantID),
		BranchID:   uuidToPgtype(branchID),
		TemplateID: uuidToPgtype(templateID),
	})
	if err != nil {
		return nil, fmt.Errorf("list session template entries: %w", err)
	}
	out := make([]domain.SessionTemplateEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSessionTemplateEntryRow(row))
	}
	return out, nil
}

func (r *SessionTemplateRepository) EntriesListByTemplateTx(ctx context.Context, tx domain.Tx, tenantID, branchID, templateID uuid.UUID) ([]domain.SessionTemplateEntry, error) {
	q := sqlc.New(tx.(pgx.Tx))
	rows, err := q.SessionTemplateEntriesListByTemplate(ctx, sqlc.SessionTemplateEntriesListByTemplateParams{
		TenantID:   uuidToPgtype(tenantID),
		BranchID:   uuidToPgtype(branchID),
		TemplateID: uuidToPgtype(templateID),
	})
	if err != nil {
		return nil, fmt.Errorf("list session template entries (tx): %w", err)
	}
	out := make([]domain.SessionTemplateEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSessionTemplateEntryRow(row))
	}
	return out, nil
}

func mapSessionTemplate(row sqlc.SessionTemplate) domain.SessionTemplate {
	return domain.SessionTemplate{
		ID:          pgtypeUUIDToUUID(row.ID),
		TenantID:    pgtypeUUIDToUUID(row.TenantID),
		BranchID:    pgtypeUUIDToUUID(row.BranchID),
		Name:        row.Name,
		Description: pgtypeTextToStringPtr(row.Description),
		IsActive:    row.IsActive,
		CreatedAt:   pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:   pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func mapSessionTemplateEntryRow(row sqlc.SessionTemplateEntriesListByTemplateRow) domain.SessionTemplateEntry {
	return domain.SessionTemplateEntry{
		ID:            pgtypeUUIDToUUID(row.ID),
		TenantID:      pgtypeUUIDToUUID(row.TenantID),
		BranchID:      pgtypeUUIDToUUID(row.BranchID),
		TemplateID:    pgtypeUUIDToUUID(row.TemplateID),
		DayOfWeek:     int(row.DayOfWeek),
		SessionTypeID: pgtypeUUIDToUUID(row.SessionTypeID),
		SessionType: &domain.EntrySessionType{
			ID:           pgtypeUUIDToUUID(row.SessionTypeID),
			Name:         row.SessionTypeName,
			StartMinutes: pgtypeTimeToMinutes(row.SessionTypeStartTime),
			EndMinutes:   pgtypeTimeToMinutes(row.SessionTypeEndTime),
			IsActive:     row.SessionTypeIsActive,
		},
		CreatedAt: pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt: pgtypeTimestamptzToTime(row.UpdatedAt),
	}
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

func pgtypeTextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func pgtypeTimeToMinutes(t pgtype.Time) int {
	if !t.Valid {
		return 0
	}
	return int(t.Microseconds / 60 / 1_000_000)
}

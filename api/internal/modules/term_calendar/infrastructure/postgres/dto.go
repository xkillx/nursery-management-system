package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"nursery-management-system/api/internal/modules/term_calendar/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

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

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func mapAcademicTerm(row sqlc.AcademicTerm) domain.AcademicTerm {
	return domain.AcademicTerm{
		ID:        pgtypeUUIDToUUID(row.ID),
		TenantID:  pgtypeUUIDToUUID(row.TenantID),
		BranchID:  pgtypeUUIDToUUID(row.BranchID),
		Name:      row.Name,
		Kind:      row.Kind,
		StartDate: pgtypeDateToTime(row.StartDate),
		EndDate:   pgtypeDateToTime(row.EndDate),
		IsActive:  row.IsActive,
		CreatedAt: pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt: pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

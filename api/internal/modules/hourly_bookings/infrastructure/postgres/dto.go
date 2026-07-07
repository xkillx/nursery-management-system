package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"nursery-management-system/api/internal/modules/hourly_bookings/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func pgtypeUUIDToPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	v := uuid.UUID(u.Bytes)
	return &v
}

func ptrToPgtypeUUID(p *uuid.UUID) pgtype.UUID {
	if p == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: [16]byte(*p), Valid: true}
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

func mapHourlyBooking(row sqlc.HourlyBooking) domain.HourlyBooking {
	return domain.HourlyBooking{
		ID:                   pgtypeUUIDToUUID(row.ID),
		TenantID:             pgtypeUUIDToUUID(row.TenantID),
		BranchID:             pgtypeUUIDToUUID(row.BranchID),
		ChildID:              pgtypeUUIDToUUID(row.ChildID),
		CalendarDate:         pgtypeDateToTime(row.CalendarDate),
		StartTimeMinutes:     int(row.StartTimeMinutes),
		DurationMinutes:      int(row.DurationMinutes),
		SessionTypeID:        pgtypeUUIDToPtr(row.SessionTypeID),
		BookedByMembershipID: pgtypeUUIDToUUID(row.BookedByMembershipID),
		Status:               row.Status,
		CreatedAt:            pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:            pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

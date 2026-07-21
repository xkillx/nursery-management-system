package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"nursery-management-system/api/internal/modules/bookings/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func pgtypeUUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	v := uuid.UUID(u.Bytes)
	return &v
}

func timeToPgtypeDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func pgtypeDateToTime(d pgtype.Date) time.Time {
	return d.Time
}

func pgtypeDatePtr(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func pgtypeTextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func timeToPgtypeDatePtr(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func stringToPgtypeText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func float64ToPgtypeNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	n := pgtype.Numeric{}
	_ = n.Scan(fmt.Sprintf("%g", *f))
	return n
}

type bookingRow struct {
	ID                   pgtype.UUID
	TenantID             pgtype.UUID
	BranchID             pgtype.UUID
	ChildID              pgtype.UUID
	SessionTemplateID    pgtype.UUID
	DaysOfWeek           []int32
	EffectiveStartDate   pgtype.Date
	EffectiveEndDate     pgtype.Date
	FundingType          pgtype.Text
	FundingHoursPerWeek  pgtype.Numeric
	LaReference          pgtype.Text
	SessionEntries       []byte
	Status               string
	BookedByMembershipID pgtype.UUID
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
}

func mapBooking(row bookingRow) domain.Booking {
	var sessionEntries []domain.SessionEntry
	if len(row.SessionEntries) > 0 {
		_ = json.Unmarshal(row.SessionEntries, &sessionEntries)
	}

	var fundingHours *float64
	if row.FundingHoursPerWeek.Valid {
		f, _ := row.FundingHoursPerWeek.Float64Value()
		v := f.Float64
		fundingHours = &v
	}

	return domain.Booking{
		ID:                   pgtypeUUIDToUUID(row.ID),
		TenantID:             pgtypeUUIDToUUID(row.TenantID),
		BranchID:             pgtypeUUIDToUUID(row.BranchID),
		ChildID:              pgtypeUUIDToUUID(row.ChildID),
		SessionTemplateID:    pgtypeUUIDPtr(row.SessionTemplateID),
		DaysOfWeek:           row.DaysOfWeek,
		EffectiveStartDate:   pgtypeDateToTime(row.EffectiveStartDate),
		EffectiveEndDate:     pgtypeDatePtr(row.EffectiveEndDate),
		FundingType:          pgtypeTextPtr(row.FundingType),
		FundingHoursPerWeek:  fundingHours,
		LaReference:          pgtypeTextPtr(row.LaReference),
		SessionEntries:       sessionEntries,
		Status:               row.Status,
		BookedByMembershipID: pgtypeUUIDToUUID(row.BookedByMembershipID),
		CreatedAt:            pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:            pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func bookingsGetByIDRowToBookingRow(r sqlc.BookingsGetByIDRow) bookingRow {
	return bookingRow{
		ID: r.ID, TenantID: r.TenantID, BranchID: r.BranchID, ChildID: r.ChildID,
		SessionTemplateID: r.SessionTemplateID, DaysOfWeek: r.DaysOfWeek,
		EffectiveStartDate: r.EffectiveStartDate, EffectiveEndDate: r.EffectiveEndDate,
		FundingType: r.FundingType, FundingHoursPerWeek: r.FundingHoursPerWeek,
		LaReference: r.LaReference, SessionEntries: r.SessionEntries,
		Status: r.Status, BookedByMembershipID: r.BookedByMembershipID,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func bookingsGetByIDForUpdateRowToBookingRow(r sqlc.BookingsGetByIDForUpdateRow) bookingRow {
	return bookingRow{
		ID: r.ID, TenantID: r.TenantID, BranchID: r.BranchID, ChildID: r.ChildID,
		SessionTemplateID: r.SessionTemplateID, DaysOfWeek: r.DaysOfWeek,
		EffectiveStartDate: r.EffectiveStartDate, EffectiveEndDate: r.EffectiveEndDate,
		FundingType: r.FundingType, FundingHoursPerWeek: r.FundingHoursPerWeek,
		LaReference: r.LaReference, SessionEntries: r.SessionEntries,
		Status: r.Status, BookedByMembershipID: r.BookedByMembershipID,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func bookingsListByBranchPaginatedRowToBookingRow(r sqlc.BookingsListByBranchPaginatedRow) bookingRow {
	return bookingRow{
		ID: r.ID, TenantID: r.TenantID, BranchID: r.BranchID, ChildID: r.ChildID,
		SessionTemplateID: r.SessionTemplateID, DaysOfWeek: r.DaysOfWeek,
		EffectiveStartDate: r.EffectiveStartDate, EffectiveEndDate: r.EffectiveEndDate,
		FundingType: r.FundingType, FundingHoursPerWeek: r.FundingHoursPerWeek,
		LaReference: r.LaReference, SessionEntries: r.SessionEntries,
		Status: r.Status, BookedByMembershipID: r.BookedByMembershipID,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func bookingsListByChildAndDateRangeRowToBookingRow(r sqlc.BookingsListByChildAndDateRangeRow) bookingRow {
	return bookingRow{
		ID: r.ID, TenantID: r.TenantID, BranchID: r.BranchID, ChildID: r.ChildID,
		SessionTemplateID: r.SessionTemplateID, DaysOfWeek: r.DaysOfWeek,
		EffectiveStartDate: r.EffectiveStartDate, EffectiveEndDate: r.EffectiveEndDate,
		FundingType: r.FundingType, FundingHoursPerWeek: r.FundingHoursPerWeek,
		LaReference: r.LaReference, SessionEntries: r.SessionEntries,
		Status: r.Status, BookedByMembershipID: r.BookedByMembershipID,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func mapUnifiedBookingRow(row sqlc.BookingsUnifiedListByBranchRow) domain.UnifiedBookingRow {
	return domain.UnifiedBookingRow{
		BookingType:       row.BookingType,
		ID:                pgtypeUUIDToUUID(row.ID),
		TenantID:          pgtypeUUIDToUUID(row.TenantID),
		BranchID:          pgtypeUUIDToUUID(row.BranchID),
		ChildID:           pgtypeUUIDToUUID(row.ChildID),
		StartDate:         pgtypeDateToTime(row.StartDate),
		EndDate:           pgtypeDatePtr(row.EndDate),
		SessionTemplateID: pgtypeUUIDPtr(row.SessionTemplateID),
		Status:            row.Status,
		CreatedAt:         pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:         pgtypeTimestamptzToTime(row.UpdatedAt),
		ChildFirstName:    row.ChildFirstName,
		ChildLastName:     row.ChildLastName.String,
	}
}

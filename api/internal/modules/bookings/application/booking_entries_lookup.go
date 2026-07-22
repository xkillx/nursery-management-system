package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	billingdomain "nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

// BookingEntriesLookupAdapter implements billing.BookingEntriesLookup by
// querying the bookings table for the child's active recurring booking and
// expanding session_entries JSONB with session type details.
type BookingEntriesLookupAdapter struct {
	pool *pgxpool.Pool
}

func NewBookingEntriesLookupAdapter(pool *pgxpool.Pool) *BookingEntriesLookupAdapter {
	return &BookingEntriesLookupAdapter{pool: pool}
}

func (a *BookingEntriesLookupAdapter) GetEntriesForChildInMonth(ctx context.Context, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) ([]billingdomain.BookedPatternEntry, error) {
	q := sqlc.New(a.pool)

	monthStart := billingMonth.UTC()
	if monthStart.Day() != 1 {
		return nil, fmt.Errorf("billing_month must be the 1st of a month (got day=%d)", monthStart.Day())
	}
	monthEnd := monthStart.AddDate(0, 1, -1)

	rows, err := q.BookingEntriesForChildInMonth(ctx, sqlc.BookingEntriesForChildInMonthParams{
		TenantID:   pgtype.UUID{Bytes: [16]byte(tenantID), Valid: true},
		BranchID:   pgtype.UUID{Bytes: [16]byte(branchID), Valid: true},
		ChildID:    pgtype.UUID{Bytes: [16]byte(childID), Valid: true},
		MonthStart: pgtype.Date{Time: monthStart, Valid: true},
		MonthEnd:   pgtype.Date{Time: monthEnd, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("query booking entries for child: %w", err)
	}

	entries := make([]billingdomain.BookedPatternEntry, 0, len(rows))
	for _, row := range rows {
		startMin := timeOfDayToMinutesBooking(row.SessionTypeStartTime)
		endMin := timeOfDayToMinutesBooking(row.SessionTypeEndTime)
		if endMin <= startMin {
			continue
		}
		entries = append(entries, billingdomain.BookedPatternEntry{
			DayOfWeek: int(row.DayOfWeek),
			SessionType: billingdomain.BookedSessionType{
				ID:              row.SessionTypeID.String(),
				Name:            row.SessionTypeName,
				StartMinutes:    startMin,
				EndMinutes:      endMin,
				DurationMinutes: endMin - startMin,
			},
		})
	}
	return entries, nil
}

// timeOfDayToMinutesBooking converts a pgtype.Time (microseconds since midnight)
// to minutes since midnight, rounded down.
func timeOfDayToMinutesBooking(t pgtype.Time) int {
	return int(t.Microseconds / (60 * 1_000_000))
}

// Compile-time check that BookingEntriesLookupAdapter implements the interface.
var _ billingdomain.BookingEntriesLookup = (*BookingEntriesLookupAdapter)(nil)

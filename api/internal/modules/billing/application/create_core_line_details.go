package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

// coreSessionsComputer recomputes the pattern-derived BookedSessions for a
// manual draft's core childcare line, mirroring GenerateTermInvoices (R10).
// It is used by both CreateDraftInvoice and CreateAndIssueInvoiceFromForm.
type coreSessionsComputer struct {
	repo                 domain.BillingRepository
	bookingEntriesLookup domain.BookingEntriesLookup
	termDateLookup       domain.TermDateLookup
	closureDateLookup    domain.ClosureDateLookup
	holidayPeriodLookup  domain.HolidayPeriodLookup
}

// computeCoreLineDetails returns the enriched core-line details JSON to
// persist on a core_childcare line, or nil when the breakdown should collapse
// (no BookedSessions to persist). The entered line must match the
// pattern-derived values — if the manager edited the core line's quantity,
// unit or line amount, the breakdown collapses (R10, mirroring U4).
func (c *coreSessionsComputer) computeCoreLineDetails(ctx context.Context, tx pgx.Tx, actor tenant.ActorContext, childID uuid.UUID, billingMonth time.Time, entered DraftInvoiceLineInput) ([]byte, error) {
	bookings, err := c.repo.ListActiveBookingsForGeneration(ctx, tx, actor.TenantID, actor.BranchID, billingMonth)
	if err != nil {
		return nil, fmt.Errorf("list active bookings: %w", err)
	}
	var bookingRow *domain.BillableChildRow
	for i := range bookings {
		if bookings[i].ChildID == childID {
			bookingRow = &bookings[i]
			break
		}
	}
	if bookingRow == nil {
		// No pattern-derived baseline for this child: persist nothing so the
		// fallback aggregate row renders (R14).
		return nil, nil
	}

	entries, err := c.bookingEntriesLookup.GetEntriesForChildInMonth(ctx, actor.TenantID, actor.BranchID, childID, billingMonth)
	if err != nil {
		return nil, fmt.Errorf("lookup booking entries: %w", err)
	}

	var termDates []domain.TermDateRange
	if bookingRow.TermTimeOnly && c.termDateLookup != nil {
		termDates, err = c.termDateLookup.GetTermDateRangesForBranchAndMonth(ctx, actor.TenantID, actor.BranchID, billingMonth)
		if err != nil {
			return nil, fmt.Errorf("lookup term dates: %w", err)
		}
	}
	var closureDates []time.Time
	if c.closureDateLookup != nil {
		closureDates, err = c.closureDateLookup.GetClosureDatesForBranchAndMonth(ctx, actor.TenantID, actor.BranchID, billingMonth)
		if err != nil {
			return nil, fmt.Errorf("lookup closure dates: %w", err)
		}
	}
	var holidayPeriods []domain.HolidayPeriodDateRange
	if bookingRow.TermTimeOnly && c.holidayPeriodLookup != nil {
		holidayPeriods, err = c.holidayPeriodLookup.GetHolidayPeriodsForBranchAndMonth(ctx, actor.TenantID, actor.BranchID, billingMonth)
		if err != nil {
			return nil, fmt.Errorf("lookup holiday periods: %w", err)
		}
	}

	calc, calcErr := domain.CalculateBookedCoreMinutesInMonth(
		"", entries, billingMonth, bookingRow.SiteHourlyRateMinor, termDates, closureDates, holidayPeriods,
	)
	if calcErr != nil {
		return nil, fmt.Errorf("calculate booked minutes: %w", calcErr)
	}

	// If the entered core-line values diverge from the pattern-derived
	// values, the breakdown collapses (no BookedSessions persisted).
	if entered.QuantityMinutes != calc.TotalMinutes ||
		entered.UnitAmountMinor != bookingRow.SiteHourlyRateMinor ||
		entered.LineAmountMinor != calc.Subtotal.Minor() {
		return nil, nil
	}

	if len(calc.Sessions) == 0 {
		return nil, nil
	}

	coreLineDetails := domain.CoreLineDetails{
		BookedCoreMinutes: calc.TotalMinutes,
		BookedSessions:    enrichBookedSessions(calc),
		BookedPerEntry:    calc.PerEntry,
	}
	detailsJSON, jsonErr := json.Marshal(coreLineDetails)
	if jsonErr != nil {
		return nil, fmt.Errorf("marshal core line details: %w", jsonErr)
	}
	return detailsJSON, nil
}

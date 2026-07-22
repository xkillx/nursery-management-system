package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TxManager is the interface satisfied by *transaction.Manager. The
// application layer depends on this interface so it can be mocked in tests.
type TxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

// SessionTypeInfo is a minimal projection of a session type used to validate
// entries. It is consumed via the SessionTypeLookup interface.
type SessionTypeInfo struct {
	ID           uuid.UUID
	Name         string
	StartMinutes int
	EndMinutes   int
	IsActive     bool
}

// SessionTypeLookup is implemented by an adapter in the bootstrap layer so the
// children module does not import the sessiontypes module directly.
type SessionTypeLookup interface {
	GetActiveInScope(ctx context.Context, tenantID, branchID, sessionTypeID uuid.UUID) (SessionTypeInfo, bool, error)
}

// ── Clock injection for "today" semantics ─────────────────────────────────

// TodayFunc returns today's date in Europe/London as a time.Time at 00:00 local.
type TodayFunc func() time.Time

// LondonTodayDate returns the current London local date (00:00 local) at the
// given UTC instant. The clock injection is at the use-case layer; this is the
// default implementation that uses time.Now().
func LondonTodayDate(clock func() time.Time) time.Time {
	utc := clock().UTC()
	// Europe/London is loaded by attendance application; we re-implement the
	// conversion here to avoid a cross-module import in the same package tree.
	// The actual timezone is loaded by the caller; this helper is only used
	// when a clock is injected.
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		// Fall back to UTC if London tz data is missing.
		return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	}
	t := utc.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
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

// ── Inputs ────────────────────────────────────────────────────────────────

type BookingPatternEntryInput struct {
	DayOfWeek     int
	SessionTypeID uuid.UUID
}

type CreateBookingPatternInput struct {
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	Entries       []BookingPatternEntryInput
	TermTimeOnly  bool
}

type UpdateBookingPatternInput struct {
	EffectiveFrom *time.Time
	Entries       *[]BookingPatternEntryInput
	TermTimeOnly  *bool
}

// ── Use cases ─────────────────────────────────────────────────────────────

type ListBookingPatterns struct {
	repo domain.Repository
}

func NewListBookingPatterns(repo domain.Repository) *ListBookingPatterns {
	return &ListBookingPatterns{repo: repo}
}

func (uc *ListBookingPatterns) Execute(ctx context.Context, actor tenant.ActorContext, childID string) ([]domain.BookingPattern, error) {
	cid, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	_, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, cid)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("check child exists: %w", err))
	}
	if !found {
		return nil, domainerrors.NotFound("child", "Resource not found.")
	}
	patterns, err := uc.repo.ListByChild(ctx, actor.TenantID, actor.BranchID, cid)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("list child booking patterns: %w", err))
	}
	return patterns, nil
}

type GetBookingPattern struct {
	repo domain.Repository
}

func NewGetBookingPattern(repo domain.Repository) *GetBookingPattern {
	return &GetBookingPattern{repo: repo}
}

func (uc *GetBookingPattern) Execute(ctx context.Context, actor tenant.ActorContext, childID, patternID string) (*domain.BookingPattern, error) {
	cid, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	pid, err := parseUUID(patternID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "pattern_id")
	}
	_, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, cid)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("check child exists: %w", err))
	}
	if !found {
		return nil, domainerrors.NotFound("child", "Resource not found.")
	}
	pattern, found, err := uc.repo.GetPatternByID(ctx, actor.TenantID, actor.BranchID, pid)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get child booking pattern: %w", err))
	}
	if !found || pattern.ChildID != cid {
		return nil, domainerrors.NotFound("booking_pattern", "Resource not found.")
	}
	return pattern, nil
}

type GetCurrentBookingPattern struct {
	repo  domain.Repository
	clock TodayFunc
}

func NewGetCurrentBookingPattern(repo domain.Repository, clock TodayFunc) *GetCurrentBookingPattern {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &GetCurrentBookingPattern{repo: repo, clock: clock}
}

func (uc *GetCurrentBookingPattern) Execute(ctx context.Context, actor tenant.ActorContext, childID, date string) (*domain.BookingPattern, error) {
	cid, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	var targetDate time.Time
	if date == "" {
		targetDate = LondonTodayDate(uc.clock)
	} else {
		parsed, perr := time.Parse("2006-01-02", date)
		if perr != nil {
			return nil, domainerrors.Validation("Invalid request payload.", "date")
		}
		targetDate = parsed
	}

	_, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, cid)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("check child exists: %w", err))
	}
	if !found {
		return nil, domainerrors.NotFound("child", "Resource not found.")
	}

	pattern, found, err := uc.repo.GetActiveForDate(ctx, actor.TenantID, actor.BranchID, cid, targetDate)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get active booking pattern: %w", err))
	}
	if !found {
		return nil, domainerrors.NotFound("booking_pattern", "Resource not found.")
	}
	return pattern, nil
}

type CreateBookingPattern struct {
	repo          domain.Repository
	audit         *audit.Writer
	txm           TxManager
	sessionLookup SessionTypeLookup
	clock         TodayFunc
}

func NewCreateBookingPattern(repo domain.Repository, auditWriter *audit.Writer, txm TxManager, lookup SessionTypeLookup, clock TodayFunc) *CreateBookingPattern {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &CreateBookingPattern{repo: repo, audit: auditWriter, txm: txm, sessionLookup: lookup, clock: clock}
}

func (uc *CreateBookingPattern) Execute(ctx context.Context, actor tenant.ActorContext, childID string, in CreateBookingPatternInput) (*domain.BookingPattern, error) {
	cid, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	if len(in.Entries) == 0 {
		return nil, domainerrors.Validation("Invalid request payload.", "entries")
	}

	resolved, err := resolveBookingPatternEntries(ctx, uc.sessionLookup, actor, in.Entries)
	if err != nil {
		return nil, err
	}

	var result *domain.BookingPattern
	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		exists, eerr := uc.repo.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, cid)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("check child exists: %w", eerr))
		}
		if !exists {
			return domainerrors.NotFound("child", "Resource not found.")
		}

		var terr error
		result, terr = createBookingPatternInTx(ctx, tx, uc.repo, uc.audit, actor, cid, in.EffectiveFrom, in.EffectiveTo, resolved, in.TermTimeOnly, true, uc.clock)
		return terr
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type UpdateBookingPattern struct {
	repo          domain.Repository
	audit         *audit.Writer
	txm           TxManager
	sessionLookup SessionTypeLookup
	clock         TodayFunc
}

func NewUpdateBookingPattern(repo domain.Repository, auditWriter *audit.Writer, txm TxManager, lookup SessionTypeLookup, clock TodayFunc) *UpdateBookingPattern {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &UpdateBookingPattern{repo: repo, audit: auditWriter, txm: txm, sessionLookup: lookup, clock: clock}
}

func (uc *UpdateBookingPattern) Execute(ctx context.Context, actor tenant.ActorContext, childID, patternID string, in UpdateBookingPatternInput) (*domain.BookingPattern, error) {
	cid, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	pid, err := parseUUID(patternID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "pattern_id")
	}

	today := LondonTodayDate(uc.clock)

	var result *domain.BookingPattern
	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		exists, eerr := uc.repo.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, cid)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("check child exists: %w", eerr))
		}
		if !exists {
			return domainerrors.NotFound("child", "Resource not found.")
		}

		pattern, pfound, perr := uc.repo.GetPatternByID(ctx, actor.TenantID, actor.BranchID, pid)
		if perr != nil {
			return domainerrors.Internal(fmt.Errorf("get booking pattern: %w", perr))
		}
		if !pfound || pattern.ChildID != cid {
			return domainerrors.NotFound("booking_pattern", "Resource not found.")
		}

		// Only the open pattern with effective_from >= today is editable.
		if !pattern.IsCurrent {
			return domainerrors.New("booking_pattern_not_editable", "Resource not editable.", "pattern_id")
		}
		if pattern.EffectiveFrom.Before(today) {
			return domainerrors.New("booking_pattern_not_editable", "Resource not editable.", "pattern_id")
		}

		newFrom := pattern.EffectiveFrom
		if in.EffectiveFrom != nil {
			if in.EffectiveFrom.Before(today) {
				return domainerrors.New("booking_pattern_backdated", "Invalid request payload.", "effective_from")
			}
			newFrom = *in.EffectiveFrom
		}

		// Validate + resolve entries if provided.
		var resolved []domain.BookingPatternEntry
		if in.Entries != nil {
			if len(*in.Entries) == 0 {
				return domainerrors.Validation("Invalid request payload.", "entries")
			}
			seen := make(map[BookingPatternEntryInput]struct{}, len(*in.Entries))
			seenDays := make(map[int]struct{}, len(*in.Entries))
			resolved = make([]domain.BookingPatternEntry, 0, len(*in.Entries))
			for _, e := range *in.Entries {
				if e.DayOfWeek < 1 || e.DayOfWeek > 5 {
					return domainerrors.Validation("Invalid request payload.", "day_of_week")
				}
				if e.SessionTypeID == uuid.Nil {
					return domainerrors.Validation("Invalid request payload.", "session_type_id")
				}
				if _, dup := seenDays[e.DayOfWeek]; dup {
					return domainerrors.New("booking_pattern_duplicate_day", "Invalid request payload.", "entries")
				}
				seenDays[e.DayOfWeek] = struct{}{}
				key := BookingPatternEntryInput{DayOfWeek: e.DayOfWeek, SessionTypeID: e.SessionTypeID}
				if _, dup := seen[key]; dup {
					return domainerrors.New("booking_pattern_duplicate_entry", "Invalid request payload.", "entries")
				}
				seen[key] = struct{}{}
				info, found, lerr := uc.sessionLookup.GetActiveInScope(ctx, actor.TenantID, actor.BranchID, e.SessionTypeID)
				if lerr != nil {
					return domainerrors.Internal(fmt.Errorf("lookup session type: %w", lerr))
				}
				if !found {
					return domainerrors.Forbidden("session_type_not_in_branch", "Invalid request payload.")
				}
				if !info.IsActive {
					return domainerrors.New("session_type_archived", "Invalid request payload.", "session_type_id")
				}
				resolved = append(resolved, domain.BookingPatternEntry{
					ID:        uid.NewUUID(),
					DayOfWeek: e.DayOfWeek,
					SessionType: &domain.EntrySessionType{
						ID:           info.ID,
						Name:         info.Name,
						StartMinutes: info.StartMinutes,
						EndMinutes:   info.EndMinutes,
						IsActive:     info.IsActive,
					},
				})
			}
		}

		// If effective_from changed, re-close the previous pattern adjacently.
		changed := newFrom.After(pattern.EffectiveFrom)
		if changed {
			prevPattern, pfound, perr := uc.repo.GetPreviousClosedByChild(ctx, tx, actor.TenantID, actor.BranchID, cid)
			if perr != nil {
				return domainerrors.Internal(fmt.Errorf("get previous closed pattern: %w", perr))
			}
			if pfound {
				// New effective_from - 1 must be >= previous.effective_from.
				closeTo := newFrom.AddDate(0, 0, -1)
				if closeTo.Before(prevPattern.EffectiveFrom) {
					return domainerrors.New("booking_pattern_overlap", "Invalid request payload.", "effective_from")
				}
				// Close the previous pattern; since we are already inside the
				// transaction and the current pattern is open, the previous is
				// closed by ID.
				if err := uc.repo.ClosePatternByID(ctx, tx, actor.TenantID, actor.BranchID, prevPattern.ID, closeTo); err != nil {
					return domainerrors.Internal(fmt.Errorf("re-close previous pattern: %w", err))
				}
			}
			// Update current pattern's effective_from.
			if err := uc.repo.UpdateEffectiveFrom(ctx, tx, actor.TenantID, actor.BranchID, pattern.ID, newFrom); err != nil {
				return domainerrors.Internal(fmt.Errorf("update pattern effective_from: %w", err))
			}
		}

		// Replace entries if provided.
		if resolved != nil {
			if err := uc.repo.ReplaceEntries(ctx, tx, actor.TenantID, actor.BranchID, pattern.ID, resolved); err != nil {
				return domainerrors.Internal(fmt.Errorf("replace entries: %w", err))
			}
		}

		if in.TermTimeOnly != nil {
			if err := uc.repo.UpdateTermTimeOnly(ctx, tx, actor.TenantID, actor.BranchID, pattern.ID, *in.TermTimeOnly); err != nil {
				return domainerrors.Internal(fmt.Errorf("update term_time_only: %w", err))
			}
		}

		// Re-load pattern with joined entries for response.
		updated, ufound, uerr := uc.repo.GetPatternByID(ctx, actor.TenantID, actor.BranchID, pattern.ID)
		if uerr != nil {
			return domainerrors.Internal(fmt.Errorf("reload booking pattern: %w", uerr))
		}
		if !ufound {
			return domainerrors.NotFound("booking_pattern", "Resource not found.")
		}

		if uc.audit != nil {
			if aerr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: "child_booking_pattern_updated",
				EntityType: "child",
				EntityID:   cid,
				Details: map[string]any{
					"pattern_id":     updated.ID.String(),
					"effective_from": updated.EffectiveFrom.Format("2006-01-02"),
					"entry_count":    len(updated.Entries),
				},
			}); aerr != nil {
				return domainerrors.Internal(fmt.Errorf("audit child_booking_pattern_updated: %w", aerr))
			}
		}
		result = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

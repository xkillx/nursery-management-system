package application

import (
	"context"
	"errors"
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

// BookingPatternInput is the input for creating a booking pattern embedded
// within a larger creation flow (e.g. create child with full profile).
type BookingPatternInput struct {
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	Entries       []BookingPatternEntryInput
	TermTimeOnly  bool
}

// resolveBookingPatternEntries validates, deduplicates, and resolves session
// types for a set of booking pattern entries.
func resolveBookingPatternEntries(ctx context.Context, lookup SessionTypeLookup, actor tenant.ActorContext, entries []BookingPatternEntryInput) ([]domain.BookingPatternEntry, error) {
	seen := make(map[BookingPatternEntryInput]struct{}, len(entries))
	seenDays := make(map[int]struct{}, len(entries))
	resolved := make([]domain.BookingPatternEntry, 0, len(entries))
	for _, e := range entries {
		if e.DayOfWeek < 1 || e.DayOfWeek > 7 {
			return nil, domainerrors.Validation("Invalid request payload.", "day_of_week")
		}
		if e.SessionTypeID == uuid.Nil {
			return nil, domainerrors.Validation("Invalid request payload.", "session_type_id")
		}
		if _, dup := seenDays[e.DayOfWeek]; dup {
			return nil, domainerrors.New("booking_pattern_duplicate_day", "Invalid request payload.", "entries")
		}
		seenDays[e.DayOfWeek] = struct{}{}
		key := BookingPatternEntryInput{DayOfWeek: e.DayOfWeek, SessionTypeID: e.SessionTypeID}
		if _, dup := seen[key]; dup {
			return nil, domainerrors.New("booking_pattern_duplicate_entry", "Invalid request payload.", "entries")
		}
		seen[key] = struct{}{}

		info, found, err := lookup.GetActiveInScope(ctx, actor.TenantID, actor.BranchID, e.SessionTypeID)
		if err != nil {
			var de *domainerrors.DomainError
			if errors.As(err, &de) {
				return nil, de
			}
			return nil, domainerrors.Internal(fmt.Errorf("lookup session type: %w", err))
		}
		if !found {
			return nil, domainerrors.Forbidden("session_type_not_in_branch", "Invalid request payload.")
		}
		if !info.IsActive {
			return nil, domainerrors.New("session_type_archived", "Invalid request payload.", "session_type_id")
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
	return resolved, nil
}

// createBookingPatternInTx performs the transactional logic for creating a
// booking pattern. It checks effectiveFrom is not backdated, closes any open
// pattern adjacently, inserts the new pattern, and optionally writes an audit
// entry. The caller must have already verified the child exists in scope.
func createBookingPatternInTx(ctx context.Context, tx pgx.Tx, repo domain.Repository, auditWriter *audit.Writer, actor tenant.ActorContext, childID uuid.UUID, effectiveFrom time.Time, effectiveTo *time.Time, entries []domain.BookingPatternEntry, termTimeOnly bool, writeAudit bool, clock TodayFunc) (*domain.BookingPattern, error) {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	today := LondonTodayDate(clock)
	if effectiveFrom.Before(today) {
		return nil, domainerrors.New("booking_pattern_backdated", "Invalid request payload.", "effective_from")
	}
	if effectiveTo != nil && effectiveTo.Before(effectiveFrom) {
		return nil, domainerrors.New("booking_pattern_effective_to_before_from", "Invalid request payload.", "effective_to")
	}

	openPattern, ofound, oerr := repo.GetCurrentOpenByChild(ctx, tx, actor.TenantID, actor.BranchID, childID)
	if oerr != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get current open pattern: %w", oerr))
	}
	if ofound {
		if !effectiveFrom.After(openPattern.EffectiveFrom) {
			return nil, domainerrors.New("booking_pattern_overlap", "Invalid request payload.", "effective_from")
		}
		closeTo := effectiveFrom.AddDate(0, 0, -1)
		if err := repo.CloseCurrentPattern(ctx, tx, actor.TenantID, actor.BranchID, childID, closeTo); err != nil {
			return nil, domainerrors.Internal(fmt.Errorf("close previous pattern: %w", err))
		}
	}

	pattern := &domain.BookingPattern{
		ID:            uid.NewUUID(),
		TenantID:      actor.TenantID,
		BranchID:      actor.BranchID,
		ChildID:       childID,
		EffectiveFrom: effectiveFrom,
		EffectiveTo:   effectiveTo,
		TermTimeOnly:  termTimeOnly,
	}
	saved, serr := repo.InsertPattern(ctx, tx, pattern, entries)
	if serr != nil {
		return nil, domainerrors.Internal(fmt.Errorf("insert booking pattern: %w", serr))
	}

	if writeAudit && auditWriter != nil {
		details := map[string]any{
			"pattern_id":     saved.ID.String(),
			"effective_from": saved.EffectiveFrom.Format("2006-01-02"),
			"entry_count":    len(saved.Entries),
		}
		if saved.EffectiveTo != nil {
			details["effective_to"] = saved.EffectiveTo.Format("2006-01-02")
		}
		if aerr := auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_booking_pattern_created",
			EntityType: "child",
			EntityID:   childID,
			Details:    details,
		}); aerr != nil {
			return nil, domainerrors.Internal(fmt.Errorf("audit child_booking_pattern_created: %w", aerr))
		}
	}
	return saved, nil
}

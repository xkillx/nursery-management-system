package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TermStatus is the lifecycle state of a Term. Derived from dates and lifecycle actions.
type TermStatus string

const (
	TermStatusPreTerm        TermStatus = "pre_term"
	TermStatusActive         TermStatus = "active"
	TermStatusPendingRenewal TermStatus = "pending_renewal"
	TermStatusEnded          TermStatus = "ended"
	TermStatusTerminated     TermStatus = "terminated"
)

// ScheduleChangeKind classifies an in-term booking pattern change.
type ScheduleChangeKind string

const (
	ScheduleChangeDecrease ScheduleChangeKind = "decrease"
	ScheduleChangeIncrease ScheduleChangeKind = "increase"
)

// ScheduleChangeDecision is the manager's decision on a pending increase.
type ScheduleChangeDecision string

const (
	ScheduleChangeApproved ScheduleChangeDecision = "approved"
	ScheduleChangeRejected ScheduleChangeDecision = "rejected"
)

// Term is a 12-month commercial commitment between a nursery site and a parent for one child.
type Term struct {
	ID                    uuid.UUID
	TenantID              uuid.UUID
	BranchID              uuid.UUID
	ChildID               uuid.UUID
	TermStartDate         time.Time
	TermEndDate           time.Time
	BookingPatternID      uuid.UUID
	SiteHourlyRateMinor   int
	Status                TermStatus
	TerminationReasonCode *string
	TerminationReasonNote *string
	TerminatedAt          *time.Time
	CreatedAt             time.Time
	CreatedByMembershipID uuid.UUID
	UpdatedAt             time.Time
}

// TermScheduleChange is the audit record for a single in-term booking pattern change.
type TermScheduleChange struct {
	ID                       uuid.UUID
	TenantID                 uuid.UUID
	BranchID                 uuid.UUID
	TermID                   uuid.UUID
	PreviousBookingPatternID uuid.UUID
	NewBookingPatternID      uuid.UUID
	ChangeKind               ScheduleChangeKind
	RequestedAt              time.Time
	EffectiveFrom            time.Time
	ApprovedByMembershipID   *uuid.UUID
	ApprovalDecision         *ScheduleChangeDecision
	RejectedAt               *time.Time
	RequestID                string
}

// NewTerm constructs a Term with the canonical 12-month end date derived from start.
func NewTerm(
	id, tenantID, branchID, childID uuid.UUID,
	termStartDate time.Time,
	bookingPatternID uuid.UUID,
	siteHourlyRateMinor int,
	createdByMembershipID uuid.UUID,
) (*Term, error) {
	if id == uuid.Nil || tenantID == uuid.Nil || branchID == uuid.Nil || childID == uuid.Nil {
		return nil, fmt.Errorf("term: id/tenant/branch/child are required")
	}
	if bookingPatternID == uuid.Nil {
		return nil, fmt.Errorf("term: booking_pattern_id is required")
	}
	if createdByMembershipID == uuid.Nil {
		return nil, fmt.Errorf("term: created_by_membership_id is required")
	}
	if siteHourlyRateMinor < 0 {
		return nil, fmt.Errorf("term: site_hourly_rate_minor must be >= 0")
	}
	if err := ValidateTermStartDate(termStartDate); err != nil {
		return nil, err
	}
	termEnd := TermEndDateFor(termStartDate)
	status := initialStatus(termStartDate)
	return &Term{
		ID:                    id,
		TenantID:              tenantID,
		BranchID:              branchID,
		ChildID:               childID,
		TermStartDate:         termStartDate,
		TermEndDate:           termEnd,
		BookingPatternID:      bookingPatternID,
		SiteHourlyRateMinor:   siteHourlyRateMinor,
		Status:                status,
		CreatedByMembershipID: createdByMembershipID,
	}, nil
}

// TermEndDateFor returns term_start_date + 12 months - 1 day.
func TermEndDateFor(termStart time.Time) time.Time {
	return termStart.AddDate(0, 12, 0).AddDate(0, 0, -1)
}

// ValidateTermStartDate enforces that the term starts on the first of a calendar month (UTC date).
func ValidateTermStartDate(t time.Time) error {
	t = t.UTC()
	if t.Day() != 1 {
		return fmt.Errorf("term: term_start_date must be the 1st of a calendar month (got day=%d)", t.Day())
	}
	if t.Hour() != 0 || t.Minute() != 0 || t.Second() != 0 || t.Nanosecond() != 0 {
		return fmt.Errorf("term: term_start_date must be at 00:00:00.000")
	}
	return nil
}

// ValidateEffectiveFrom enforces that the schedule change effective_from is the 1st of a calendar month.
func ValidateEffectiveFrom(t time.Time) error {
	t = t.UTC()
	if t.Day() != 1 {
		return fmt.Errorf("term: effective_from must be the 1st of a calendar month (got day=%d)", t.Day())
	}
	if t.Hour() != 0 || t.Minute() != 0 || t.Second() != 0 || t.Nanosecond() != 0 {
		return fmt.Errorf("term: effective_from must be at 00:00:00.000")
	}
	return nil
}

func initialStatus(start time.Time) TermStatus {
	today := todayUTC()
	if start.After(today) {
		return TermStatusPreTerm
	}
	return TermStatusActive
}

// TodayFunc returns the current UTC time. Replaced in tests for deterministic behaviour.
type TodayFunc func() time.Time

func todayUTC() time.Time {
	return time.Now().UTC().Truncate(24 * time.Hour)
}

// DeriveStatus returns the live status based on calendar dates and the current status.
// Used by daily scheduler jobs; does not change terminated/ended terms.
func (t *Term) DeriveStatus(today time.Time) TermStatus {
	switch t.Status {
	case TermStatusTerminated, TermStatusEnded:
		return t.Status
	}
	if today.Before(t.TermStartDate) {
		return TermStatusPreTerm
	}
	if today.After(t.TermEndDate) {
		return TermStatusEnded
	}
	return TermStatusActive
}

// PendingRenewalThresholdDays is the lookahead window for marking a Term pending_renewal.
const PendingRenewalThresholdDays = 30

// ShouldBePendingRenewal returns true if today is within the renewal window and status is active.
func (t *Term) ShouldBePendingRenewal(today time.Time) bool {
	if t.Status != TermStatusActive {
		return false
	}
	days := int(t.TermEndDate.Sub(today).Hours() / 24)
	return days <= PendingRenewalThresholdDays
}

// ShouldBeEnded returns true if today is on/after term_end_date + 1 day (post-term grace).
func (t *Term) ShouldBeEnded(today time.Time) bool {
	if t.Status == TermStatusTerminated || t.Status == TermStatusEnded {
		return false
	}
	return !today.Before(t.TermEndDate.AddDate(0, 0, 1))
}

// Errors
var (
	ErrTermNotFound                  = errors.New("term not found")
	ErrTermAlreadyExists             = errors.New("an active term already exists for this child")
	ErrInvalidStartDate              = errors.New("term_start_date must be the 1st of a calendar month")
	ErrInvalidEffectiveFrom          = errors.New("effective_from must be the 1st of a calendar month")
	ErrInvalidRate                   = errors.New("site_hourly_rate_minor must be >= 0")
	ErrTermNotActive                 = errors.New("term is not active")
	ErrScheduleChangeNotFound        = errors.New("schedule change not found")
	ErrScheduleChangeAlreadyDecided  = errors.New("schedule change already decided")
	ErrScheduleChangeEffectiveInPast = errors.New("schedule change effective_from is in the past")
	ErrDecreaseAutoApproved          = errors.New("decrease schedule changes are auto-approved; use mark_applied flow")
)

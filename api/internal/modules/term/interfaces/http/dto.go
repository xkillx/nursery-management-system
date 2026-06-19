package httpterm

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/term/application"
	"nursery-management-system/api/internal/modules/term/domain"
)

type createTermRequest struct {
	TermStartDate    string `json:"term_start_date"` // YYYY-MM-01
	BookingPatternID string `json:"booking_pattern_id"`
}

type termResponse struct {
	ID                    string  `json:"id"`
	ChildID               string  `json:"child_id"`
	TermStartDate         string  `json:"term_start_date"`
	TermEndDate           string  `json:"term_end_date"`
	BookingPatternID      string  `json:"booking_pattern_id"`
	SiteHourlyRateMinor   int     `json:"site_hourly_rate_minor"`
	Status                string  `json:"status"`
	TerminationReasonCode *string `json:"termination_reason_code,omitempty"`
	TerminationReasonNote *string `json:"termination_reason_note,omitempty"`
	TerminatedAt          *string `json:"terminated_at,omitempty"`
	CreatedAt             string  `json:"created_at"`
}

type termListResponse struct {
	Terms []termResponse `json:"terms"`
}

func toTermResponse(t *domain.Term) termResponse {
	resp := termResponse{
		ID:                  t.ID.String(),
		ChildID:             t.ChildID.String(),
		TermStartDate:       t.TermStartDate.Format("2006-01-02"),
		TermEndDate:         t.TermEndDate.Format("2006-01-02"),
		BookingPatternID:    t.BookingPatternID.String(),
		SiteHourlyRateMinor: t.SiteHourlyRateMinor,
		Status:              string(t.Status),
		CreatedAt:           t.CreatedAt.Format(time.RFC3339),
	}
	if t.TerminationReasonCode != nil {
		s := *t.TerminationReasonCode
		resp.TerminationReasonCode = &s
	}
	if t.TerminationReasonNote != nil {
		s := *t.TerminationReasonNote
		resp.TerminationReasonNote = &s
	}
	if t.TerminatedAt != nil {
		s := t.TerminatedAt.Format(time.RFC3339)
		resp.TerminatedAt = &s
	}
	return resp
}

func toTermListResponse(terms []domain.Term) termListResponse {
	out := make([]termResponse, 0, len(terms))
	for i := range terms {
		out = append(out, toTermResponse(&terms[i]))
	}
	return termListResponse{Terms: out}
}

type requestScheduleChangeRequest struct {
	NewBookingPatternID string `json:"new_booking_pattern_id"`
	EffectiveFrom       string `json:"effective_from"` // YYYY-MM-01
	ChangeKind          string `json:"change_kind"`    // "decrease" | "increase"
}

type scheduleChangeResponse struct {
	ID                       string  `json:"id"`
	TermID                   string  `json:"term_id"`
	PreviousBookingPatternID string  `json:"previous_booking_pattern_id"`
	NewBookingPatternID      string  `json:"new_booking_pattern_id"`
	ChangeKind               string  `json:"change_kind"`
	RequestedAt              string  `json:"requested_at"`
	EffectiveFrom            string  `json:"effective_from"`
	ApprovedByMembershipID   *string `json:"approved_by_membership_id,omitempty"`
	ApprovalDecision         *string `json:"approval_decision,omitempty"`
	RejectedAt               *string `json:"rejected_at,omitempty"`
}

func toScheduleChangeResponse(c *domain.TermScheduleChange) scheduleChangeResponse {
	resp := scheduleChangeResponse{
		ID:                       c.ID.String(),
		TermID:                   c.TermID.String(),
		PreviousBookingPatternID: c.PreviousBookingPatternID.String(),
		NewBookingPatternID:      c.NewBookingPatternID.String(),
		ChangeKind:               string(c.ChangeKind),
		RequestedAt:              c.RequestedAt.Format(time.RFC3339),
		EffectiveFrom:            c.EffectiveFrom.Format("2006-01-02"),
	}
	if c.ApprovedByMembershipID != nil {
		s := c.ApprovedByMembershipID.String()
		resp.ApprovedByMembershipID = &s
	}
	if c.ApprovalDecision != nil {
		s := string(*c.ApprovalDecision)
		resp.ApprovalDecision = &s
	}
	if c.RejectedAt != nil {
		s := c.RejectedAt.Format(time.RFC3339)
		resp.RejectedAt = &s
	}
	return resp
}

type terminateTermRequest struct {
	ReasonCode     string `json:"reason_code"`
	ReasonNote     string `json:"reason_note"`
	EffectiveMonth string `json:"effective_month"` // YYYY-MM-01; defaults to next month
}

func parseDateYYYYMMDD(s, field string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, &validationError{Field: field, Message: "is required"}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, &validationError{Field: field, Message: "invalid date format (YYYY-MM-DD)"}
	}
	return t, nil
}

func parseUUIDRaw(s, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(s))
	if err != nil {
		return uuid.Nil, &validationError{Field: field, Message: "invalid uuid"}
	}
	return id, nil
}

type validationError struct {
	Field   string
	Message string
}

func (e *validationError) Error() string {
	return e.Field + ": " + e.Message
}

func toCreateTermInput(childID uuid.UUID, req createTermRequest) (application.CreateTermInput, error) {
	startDate, err := parseDateYYYYMMDD(req.TermStartDate, "term_start_date")
	if err != nil {
		return application.CreateTermInput{}, err
	}
	patternID, err := parseUUIDRaw(req.BookingPatternID, "booking_pattern_id")
	if err != nil {
		return application.CreateTermInput{}, err
	}
	return application.CreateTermInput{
		ChildID:          childID,
		TermStartDate:    startDate,
		BookingPatternID: patternID,
	}, nil
}

func toRequestScheduleChangeInput(termID uuid.UUID, req requestScheduleChangeRequest) (application.RequestScheduleChangeInput, error) {
	eff, err := parseDateYYYYMMDD(req.EffectiveFrom, "effective_from")
	if err != nil {
		return application.RequestScheduleChangeInput{}, err
	}
	newPatternID, err := parseUUIDRaw(req.NewBookingPatternID, "new_booking_pattern_id")
	if err != nil {
		return application.RequestScheduleChangeInput{}, err
	}
	kind := strings.ToLower(strings.TrimSpace(req.ChangeKind))
	if kind != "decrease" && kind != "increase" {
		return application.RequestScheduleChangeInput{}, &validationError{Field: "change_kind", Message: "must be 'decrease' or 'increase'"}
	}
	return application.RequestScheduleChangeInput{
		TermID:              termID,
		NewBookingPatternID: newPatternID,
		EffectiveFrom:       eff,
		ChangeKind:          domain.ScheduleChangeKind(kind),
	}, nil
}

func toTerminateTermInput(termID uuid.UUID, req terminateTermRequest) (application.TerminateTermInput, error) {
	in := application.TerminateTermInput{TermID: termID, ReasonCode: strings.TrimSpace(req.ReasonCode), ReasonNote: strings.TrimSpace(req.ReasonNote)}
	if !strings.EqualFold(strings.TrimSpace(req.EffectiveMonth), "") {
		t, err := parseDateYYYYMMDD(req.EffectiveMonth, "effective_month")
		if err != nil {
			return in, err
		}
		in.EffectiveMonth = t
	}
	return in, nil
}

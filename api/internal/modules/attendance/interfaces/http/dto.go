package httpattendance

import (
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/attendance/application"
	"nursery-management-system/api/internal/modules/attendance/domain"
)

type checkInRequest struct {
	ChildID string `json:"child_id" binding:"required"`
}

type checkOutRequest struct {
	ChildID string `json:"child_id" binding:"required"`
}

type correctionRequest struct {
	SessionID  *string `json:"session_id"`
	ChildID    *string `json:"child_id"`
	CheckInAt  string  `json:"check_in_at" binding:"required"`
	CheckOutAt string  `json:"check_out_at" binding:"required"`
	ReasonCode string  `json:"reason_code" binding:"required"`
	ReasonNote string  `json:"reason_note"`
}

type sessionResponse struct {
	ID                string  `json:"id"`
	ChildID           string  `json:"child_id"`
	Status            string  `json:"status"`
	CheckInAt         string  `json:"check_in_at"`
	CheckOutAt        *string `json:"check_out_at"`
	CheckInLocalDate  string  `json:"check_in_local_date"`
	CheckOutLocalDate *string `json:"check_out_local_date"`
	DurationMinutes   *int    `json:"duration_minutes"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type invoiceWarningResponse struct {
	BillingMonth  string `json:"billing_month"`
	InvoiceID     string `json:"invoice_id"`
	InvoiceNumber string `json:"invoice_number"`
	Status        string `json:"status"`
}

type correctionSessionContextResponse struct {
	ChildID           string                `json:"child_id"`
	SelectedLocalDate string                `json:"selected_local_date"`
	InvoiceWarning    *invoiceWarningResponse `json:"invoice_warning"`
	Items             []sessionResponse     `json:"items"`
}

type correctionHistoryResponse struct {
	Session sessionResponse               `json:"session"`
	Items   []correctionHistoryEventResponse `json:"items"`
}

type correctionHistoryEventResponse struct {
	ID                    string  `json:"id"`
	EventType             string  `json:"event_type"`
	OccurredAt            string  `json:"occurred_at"`
	LocalDate             string  `json:"local_date"`
	RecordedByUserID      string  `json:"recorded_by_user_id"`
	RecordedByMembershipID string  `json:"recorded_by_membership_id"`
	RecordedByLabel       *string `json:"recorded_by_label"`
	ReasonCode            *string `json:"reason_code"`
	ReasonNote            *string `json:"reason_note"`
	PreviousCheckInAt     *string `json:"previous_check_in_at"`
	PreviousCheckOutAt    *string `json:"previous_check_out_at"`
	CorrectedCheckInAt    *string `json:"corrected_check_in_at"`
	CorrectedCheckOutAt   *string `json:"corrected_check_out_at"`
	CreatedByCorrection   bool    `json:"created_by_correction"`
}

func toSessionResponse(s domain.Session) sessionResponse {
	resp := sessionResponse{
		ID:               s.ID.String(),
		ChildID:          s.ChildID.String(),
		Status:           string(s.Status),
		CheckInAt:        s.CheckInAt.UTC().Format(time.RFC3339),
		CheckInLocalDate: s.CheckInLocalDate.Format("2006-01-02"),
		CreatedAt:        s.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        s.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if s.CheckOutAt != nil {
		v := s.CheckOutAt.UTC().Format(time.RFC3339)
		resp.CheckOutAt = &v
	}
	if s.CheckOutLocalDate != nil {
		v := s.CheckOutLocalDate.Format("2006-01-02")
		resp.CheckOutLocalDate = &v
	}
	if s.DurationMinutes != nil {
		resp.DurationMinutes = s.DurationMinutes
	}
	return resp
}

func toCorrectionSessionContextResponse(ctx domain.CorrectionSessionContext) correctionSessionContextResponse {
	resp := correctionSessionContextResponse{
		ChildID:           ctx.ChildID.String(),
		SelectedLocalDate: ctx.SelectedLocalDate.Format("2006-01-02"),
		Items:             make([]sessionResponse, 0, len(ctx.Sessions)),
	}
	if ctx.InvoiceWarning != nil {
		resp.InvoiceWarning = &invoiceWarningResponse{
			BillingMonth:  ctx.InvoiceWarning.BillingMonth,
			InvoiceID:     ctx.InvoiceWarning.InvoiceID.String(),
			InvoiceNumber: ctx.InvoiceWarning.InvoiceNumber,
			Status:        ctx.InvoiceWarning.Status,
		}
	}
	for _, s := range ctx.Sessions {
		resp.Items = append(resp.Items, toSessionResponse(s))
	}
	return resp
}

func toCorrectionHistoryResponse(result application.CorrectionHistoryResult) correctionHistoryResponse {
	items := make([]correctionHistoryEventResponse, 0, len(result.Events))
	for _, evt := range result.Events {
		items = append(items, toHistoryEventResponse(evt))
	}
	return correctionHistoryResponse{
		Session: toSessionResponse(result.Session),
		Items:   items,
	}
}

func toHistoryEventResponse(evt domain.CorrectionHistoryEvent) correctionHistoryEventResponse {
	resp := correctionHistoryEventResponse{
		ID:                     evt.ID.String(),
		EventType:              string(evt.EventType),
		OccurredAt:             evt.OccurredAt.UTC().Format(time.RFC3339),
		LocalDate:              evt.LocalDate.Format("2006-01-02"),
		RecordedByUserID:       evt.RecordedByUserID.String(),
		RecordedByMembershipID: evt.RecordedByMembershipID.String(),
		CreatedByCorrection:    evt.CreatedByCorrection,
	}
	if evt.RecordedByLabel != nil {
		resp.RecordedByLabel = evt.RecordedByLabel
	}
	if evt.ReasonCode != nil {
		resp.ReasonCode = evt.ReasonCode
	}
	if evt.ReasonNote != nil {
		resp.ReasonNote = evt.ReasonNote
	}
	if evt.PreviousCheckInAt != nil {
		v := evt.PreviousCheckInAt.UTC().Format(time.RFC3339)
		resp.PreviousCheckInAt = &v
	}
	if evt.PreviousCheckOutAt != nil {
		v := evt.PreviousCheckOutAt.UTC().Format(time.RFC3339)
		resp.PreviousCheckOutAt = &v
	}
	if evt.CorrectedCheckInAt != nil {
		v := evt.CorrectedCheckInAt.UTC().Format(time.RFC3339)
		resp.CorrectedCheckInAt = &v
	}
	if evt.CorrectedCheckOutAt != nil {
		v := evt.CorrectedCheckOutAt.UTC().Format(time.RFC3339)
		resp.CorrectedCheckOutAt = &v
	}
	return resp
}

func parseChildID(raw string) (uuid.UUID, error) {
	return uuid.Parse(raw)
}

func parseCorrectionRequest(req correctionRequest) (domain.CorrectionParams, error) {
	var params domain.CorrectionParams

	if req.SessionID != nil && *req.SessionID != "" {
		id, err := uuid.Parse(*req.SessionID)
		if err != nil {
			return domain.CorrectionParams{}, err
		}
		params.SessionID = &id
	}

	if req.ChildID != nil && *req.ChildID != "" {
		id, err := uuid.Parse(*req.ChildID)
		if err != nil {
			return domain.CorrectionParams{}, err
		}
		params.ChildID = &id
	}

	checkInAt, err := time.Parse(time.RFC3339, req.CheckInAt)
	if err != nil {
		return domain.CorrectionParams{}, err
	}
	params.CheckInAt = checkInAt.UTC()

	checkOutAt, err := time.Parse(time.RFC3339, req.CheckOutAt)
	if err != nil {
		return domain.CorrectionParams{}, err
	}
	params.CheckOutAt = checkOutAt.UTC()

	params.ReasonCode = req.ReasonCode
	params.ReasonNote = req.ReasonNote

	return params, nil
}

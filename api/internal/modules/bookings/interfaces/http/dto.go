package httpbookings

import (
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/bookings/application"
	"nursery-management-system/api/internal/modules/bookings/domain"
)

// ── Response DTOs ──────────────────────────────────────────────────────────

type sessionEntryResponse struct {
	DayOfWeek     int32  `json:"day_of_week"`
	SessionTypeID string `json:"session_type_id"`
}

type bookingResponse struct {
	ID                   string                 `json:"id"`
	ChildID              string                 `json:"child_id"`
	SessionTemplateID    *string                `json:"session_template_id,omitempty"`
	RoomID               string                 `json:"room_id"`
	DaysOfWeek           []int32                `json:"days_of_week"`
	EffectiveStartDate   string                 `json:"effective_start_date"`
	EffectiveEndDate     *string                `json:"effective_end_date,omitempty"`
	FundingType          *string                `json:"funding_type,omitempty"`
	FundingHoursPerWeek  *float64               `json:"funding_hours_per_week,omitempty"`
	LaReference          *string                `json:"la_reference,omitempty"`
	SessionEntries       []sessionEntryResponse `json:"session_entries,omitempty"`
	Status               string                 `json:"status"`
	BookedByMembershipID string                 `json:"booked_by_membership_id"`
	CreatedAt            string                 `json:"created_at"`
	UpdatedAt            string                 `json:"updated_at"`
}

type unifiedBookingResponse struct {
	BookingType       string  `json:"booking_type"`
	ID                string  `json:"id"`
	ChildID           string  `json:"child_id"`
	StartDate         string  `json:"start_date"`
	EndDate           *string `json:"end_date,omitempty"`
	RoomID            *string `json:"room_id,omitempty"`
	SessionTemplateID *string `json:"session_template_id,omitempty"`
	Status            string  `json:"status"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	ChildFirstName    string  `json:"child_first_name"`
	ChildLastName     string  `json:"child_last_name"`
	RoomName          *string `json:"room_name,omitempty"`
}

type roomCapacityEntryResponse struct {
	Date        string `json:"date"`
	RoomID      string `json:"room_id"`
	RoomName    string `json:"room_name"`
	Capacity    int    `json:"capacity"`
	BookedCount int    `json:"booked_count"`
}

// ── Request DTOs ───────────────────────────────────────────────────────────

type sessionEntryRequest struct {
	DayOfWeek     int32  `json:"day_of_week" binding:"required"`
	SessionTypeID string `json:"session_type_id" binding:"required"`
}

type createBookingRequest struct {
	ChildID             string                `json:"child_id" binding:"required"`
	SessionTemplateID   string                `json:"session_template_id"`
	RoomID              string                `json:"room_id" binding:"required"`
	DaysOfWeek          []int32               `json:"days_of_week"`
	EffectiveStartDate  string                `json:"effective_start_date" binding:"required"`
	EffectiveEndDate    *string               `json:"effective_end_date"`
	FundingType         *string               `json:"funding_type"`
	FundingHoursPerWeek *float64              `json:"funding_hours_per_week"`
	LaReference         *string               `json:"la_reference"`
	SessionEntries      []sessionEntryRequest `json:"session_entries"`
}

type updateBookingRequest struct {
	RoomID              *string  `json:"room_id"`
	DaysOfWeek          []int32  `json:"days_of_week"`
	EffectiveStartDate  *string  `json:"effective_start_date"`
	EffectiveEndDate    *string  `json:"effective_end_date"`
	FundingType         *string  `json:"funding_type"`
	FundingHoursPerWeek *float64 `json:"funding_hours_per_week"`
	LaReference         *string  `json:"la_reference"`
}

type cloneBookingRequest struct {
	ChildID *string `json:"child_id"`
}

// ── Mappers ────────────────────────────────────────────────────────────────

func toBookingResponse(b domain.Booking) bookingResponse {
	var endDate *string
	if b.EffectiveEndDate != nil {
		s := b.EffectiveEndDate.UTC().Format("2006-01-02")
		endDate = &s
	}

	var sessionTemplateID *string
	if b.SessionTemplateID != nil {
		s := b.SessionTemplateID.String()
		sessionTemplateID = &s
	}

	var sessionEntries []sessionEntryResponse
	if len(b.SessionEntries) > 0 {
		sessionEntries = make([]sessionEntryResponse, 0, len(b.SessionEntries))
		for _, e := range b.SessionEntries {
			sessionEntries = append(sessionEntries, sessionEntryResponse{
				DayOfWeek:     e.DayOfWeek,
				SessionTypeID: e.SessionTypeID.String(),
			})
		}
	}

	return bookingResponse{
		ID:                   b.ID.String(),
		ChildID:              b.ChildID.String(),
		SessionTemplateID:    sessionTemplateID,
		RoomID:               b.RoomID.String(),
		DaysOfWeek:           b.DaysOfWeek,
		EffectiveStartDate:   b.EffectiveStartDate.UTC().Format("2006-01-02"),
		EffectiveEndDate:     endDate,
		FundingType:          b.FundingType,
		FundingHoursPerWeek:  b.FundingHoursPerWeek,
		LaReference:          b.LaReference,
		SessionEntries:       sessionEntries,
		Status:               b.Status,
		BookedByMembershipID: b.BookedByMembershipID.String(),
		CreatedAt:            b.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:            b.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toUnifiedBookingListResponse(items []domain.UnifiedBookingRow) []unifiedBookingResponse {
	out := make([]unifiedBookingResponse, 0, len(items))
	for _, b := range items {
		var endDate *string
		if b.EndDate != nil {
			s := b.EndDate.UTC().Format("2006-01-02")
			endDate = &s
		}
		var roomID *string
		if b.RoomID != nil {
			s := b.RoomID.String()
			roomID = &s
		}
		var sessionTemplateID *string
		if b.SessionTemplateID != nil {
			s := b.SessionTemplateID.String()
			sessionTemplateID = &s
		}

		out = append(out, unifiedBookingResponse{
			BookingType:       b.BookingType,
			ID:                b.ID.String(),
			ChildID:           b.ChildID.String(),
			StartDate:         b.StartDate.UTC().Format("2006-01-02"),
			EndDate:           endDate,
			RoomID:            roomID,
			SessionTemplateID: sessionTemplateID,
			Status:            b.Status,
			CreatedAt:         b.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:         b.UpdatedAt.UTC().Format(time.RFC3339),
			ChildFirstName:    b.ChildFirstName,
			ChildLastName:     b.ChildLastName,
			RoomName:          b.RoomName,
		})
	}
	return out
}

func toRoomCapacityListResponse(entries []application.RoomCapacityEntry) []roomCapacityEntryResponse {
	out := make([]roomCapacityEntryResponse, 0, len(entries))
	for _, e := range entries {
		out = append(out, roomCapacityEntryResponse{
			Date:        e.Date.UTC().Format("2006-01-02"),
			RoomID:      e.RoomID.String(),
			RoomName:    e.RoomName,
			Capacity:    e.Capacity,
			BookedCount: e.BookedCount,
		})
	}
	return out
}

func parseCreateRequest(req createBookingRequest) (application.CreateBookingParams, error) {
	childID, err := uuid.Parse(req.ChildID)
	if err != nil {
		return application.CreateBookingParams{}, err
	}

	var sessionTemplateID *uuid.UUID
	if req.SessionTemplateID != "" {
		id, err := uuid.Parse(req.SessionTemplateID)
		if err != nil {
			return application.CreateBookingParams{}, err
		}
		sessionTemplateID = &id
	}

	roomID, err := uuid.Parse(req.RoomID)
	if err != nil {
		return application.CreateBookingParams{}, err
	}
	startDate, err := time.Parse("2006-01-02", req.EffectiveStartDate)
	if err != nil {
		return application.CreateBookingParams{}, err
	}
	var endDate *time.Time
	if req.EffectiveEndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EffectiveEndDate)
		if err != nil {
			return application.CreateBookingParams{}, err
		}
		endDate = &t
	}

	var sessionEntries []domain.SessionEntry
	if len(req.SessionEntries) > 0 {
		sessionEntries = make([]domain.SessionEntry, 0, len(req.SessionEntries))
		for _, e := range req.SessionEntries {
			typeID, err := uuid.Parse(e.SessionTypeID)
			if err != nil {
				return application.CreateBookingParams{}, err
			}
			sessionEntries = append(sessionEntries, domain.SessionEntry{
				DayOfWeek:     e.DayOfWeek,
				SessionTypeID: typeID,
			})
		}
	}

	return application.CreateBookingParams{
		ChildID:             childID,
		SessionTemplateID:   sessionTemplateID,
		RoomID:              roomID,
		DaysOfWeek:          req.DaysOfWeek,
		EffectiveStartDate:  startDate,
		EffectiveEndDate:    endDate,
		FundingType:         req.FundingType,
		FundingHoursPerWeek: req.FundingHoursPerWeek,
		LaReference:         req.LaReference,
		SessionEntries:      sessionEntries,
	}, nil
}

func parseUpdateRequest(req updateBookingRequest) (application.UpdateBookingParams, error) {
	var params application.UpdateBookingParams

	if req.RoomID != nil {
		id, err := uuid.Parse(*req.RoomID)
		if err != nil {
			return application.UpdateBookingParams{}, err
		}
		params.RoomID = &id
	}
	if req.DaysOfWeek != nil {
		params.DaysOfWeek = req.DaysOfWeek
	}
	if req.EffectiveStartDate != nil {
		t, err := time.Parse("2006-01-02", *req.EffectiveStartDate)
		if err != nil {
			return application.UpdateBookingParams{}, err
		}
		params.EffectiveStartDate = &t
	}
	if req.EffectiveEndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EffectiveEndDate)
		if err != nil {
			return application.UpdateBookingParams{}, err
		}
		params.EffectiveEndDate = &t
	}
	if req.FundingType != nil {
		params.FundingType = req.FundingType
	}
	if req.FundingHoursPerWeek != nil {
		params.FundingHoursPerWeek = req.FundingHoursPerWeek
	}
	if req.LaReference != nil {
		params.LaReference = req.LaReference
	}

	return params, nil
}

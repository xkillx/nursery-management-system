package httpsessiontemplates

import (
	"fmt"
	"time"

	"nursery-management-system/api/internal/modules/sessiontemplates/domain"
)

type entrySessionTypeResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	IsActive  bool   `json:"is_active"`
}

type sessionTemplateEntryResponse struct {
	ID          string                `json:"id"`
	DayOfWeek   int                   `json:"day_of_week"`
	SessionType *entrySessionTypeResponse `json:"session_type"`
}

type sessionTemplateResponse struct {
	ID          string                        `json:"id"`
	BranchID    string                        `json:"branch_id"`
	Name        string                        `json:"name"`
	Description *string                       `json:"description,omitempty"`
	IsActive    bool                          `json:"is_active"`
	CreatedAt   string                        `json:"created_at"`
	UpdatedAt   string                        `json:"updated_at"`
	Entries     []sessionTemplateEntryResponse `json:"entries"`
}

type createSessionTemplateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description *string `json:"description"`
	Entries     []struct {
		DayOfWeek     int    `json:"day_of_week" binding:"required"`
		SessionTypeID string `json:"session_type_id" binding:"required"`
	} `json:"entries" binding:"required"`
}

type updateSessionTemplateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Entries     *[]struct {
		DayOfWeek     int    `json:"day_of_week" binding:"required"`
		SessionTypeID string `json:"session_type_id" binding:"required"`
	} `json:"entries"`
}

func toEntrySessionTypeResponse(s *domain.EntrySessionType) *entrySessionTypeResponse {
	if s == nil {
		return nil
	}
	return &entrySessionTypeResponse{
		ID:        s.ID.String(),
		Name:      s.Name,
		StartTime: minutesToHHMM(s.StartMinutes),
		EndTime:   minutesToHHMM(s.EndMinutes),
		IsActive:  s.IsActive,
	}
}

func toSessionTemplateResponse(t domain.SessionTemplate) sessionTemplateResponse {
	entries := make([]sessionTemplateEntryResponse, 0, len(t.Entries))
	for _, e := range t.Entries {
		entries = append(entries, sessionTemplateEntryResponse{
			ID:          e.ID.String(),
			DayOfWeek:   e.DayOfWeek,
			SessionType: toEntrySessionTypeResponse(e.SessionType),
		})
	}
	return sessionTemplateResponse{
		ID:          t.ID.String(),
		BranchID:    t.BranchID.String(),
		Name:        t.Name,
		Description: t.Description,
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.UTC().Format(time.RFC3339),
		Entries:     entries,
	}
}

func toSessionTemplateListResponse(items []domain.SessionTemplate) []sessionTemplateResponse {
	out := make([]sessionTemplateResponse, 0, len(items))
	for _, t := range items {
		// List responses don't hydrate entries (saves a query per template).
		t.Entries = nil
		out = append(out, toSessionTemplateResponse(t))
	}
	return out
}

func minutesToHHMM(m int) string {
	hh := m / 60
	mm := m % 60
	return fmt.Sprintf("%02d:%02d", hh, mm)
}

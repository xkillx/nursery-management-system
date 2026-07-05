package httptermcalendar

import (
	"time"

	"nursery-management-system/api/internal/modules/term_calendar/domain"
)

type academicTermResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type createTermRequest struct {
	Name      string `json:"name" binding:"required"`
	Kind      string `json:"kind" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

type updateTermRequest struct {
	Name      *string `json:"name"`
	Kind      *string `json:"kind"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

func toTermResponse(t domain.AcademicTerm) academicTermResponse {
	return academicTermResponse{
		ID:        t.ID.String(),
		Name:      t.Name,
		Kind:      t.Kind,
		StartDate: t.StartDate.UTC().Format("2006-01-02"),
		EndDate:   t.EndDate.UTC().Format("2006-01-02"),
		IsActive:  t.IsActive,
		CreatedAt: t.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toTermListResponse(items []domain.AcademicTerm) []academicTermResponse {
	out := make([]academicTermResponse, 0, len(items))
	for _, t := range items {
		out = append(out, toTermResponse(t))
	}
	return out
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

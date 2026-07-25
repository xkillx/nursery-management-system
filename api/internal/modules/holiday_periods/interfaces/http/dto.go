package httphp

import (
	"time"

	"nursery-management-system/api/internal/modules/holiday_periods/domain"
)

type holidayPeriodResponse struct {
	ID        string `json:"id"`
	BranchID  string `json:"branch_id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type createHolidayPeriodRequest struct {
	Name      string `json:"name" binding:"required"`
	Type      string `json:"type" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

type updateHolidayPeriodRequest struct {
	Name      *string `json:"name"`
	Type      *string `json:"type"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
}

func toHolidayPeriodResponse(hp domain.HolidayPeriod) holidayPeriodResponse {
	return holidayPeriodResponse{
		ID:        hp.ID.String(),
		BranchID:  hp.BranchID.String(),
		Name:      hp.Name,
		StartDate: hp.StartDate.UTC().Format("2006-01-02"),
		EndDate:   hp.EndDate.UTC().Format("2006-01-02"),
		Type:      string(hp.Type),
		CreatedAt: hp.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: hp.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toHolidayPeriodListResponse(items []domain.HolidayPeriod) []holidayPeriodResponse {
	out := make([]holidayPeriodResponse, 0, len(items))
	for _, hp := range items {
		out = append(out, toHolidayPeriodResponse(hp))
	}
	return out
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

package httpclosure

import (
	"time"

	"nursery-management-system/api/internal/modules/branch_closures/domain"
)

type closureDayResponse struct {
	ID        string  `json:"id"`
	BranchID  string  `json:"branch_id"`
	Date      string  `json:"date"`
	Reason    *string `json:"reason"`
	CreatedAt string  `json:"created_at"`
}

type createClosureDayRequest struct {
	Date   string  `json:"date" binding:"required"`
	Reason *string `json:"reason"`
}

func toClosureDayResponse(c domain.BranchClosureDay) closureDayResponse {
	return closureDayResponse{
		ID:        c.ID.String(),
		BranchID:  c.BranchID.String(),
		Date:      c.Date.UTC().Format("2006-01-02"),
		Reason:    c.Reason,
		CreatedAt: c.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toClosureDayListResponse(items []domain.BranchClosureDay) []closureDayResponse {
	out := make([]closureDayResponse, 0, len(items))
	for _, c := range items {
		out = append(out, toClosureDayResponse(c))
	}
	return out
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

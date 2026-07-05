package domain

import (
	"time"

	"github.com/google/uuid"
)

type AcademicTerm struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	BranchID  uuid.UUID
	Name      string
	Kind      string
	StartDate time.Time
	EndDate   time.Time
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ValidTermKind(k string) bool {
	switch k {
	case "autumn", "spring", "summer":
		return true
	}
	return false
}

type TermDateRange struct {
	StartDate time.Time
	EndDate   time.Time
}

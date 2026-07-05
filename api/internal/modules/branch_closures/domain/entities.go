package domain

import (
	"time"

	"github.com/google/uuid"
)

type BranchClosureDay struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	BranchID  uuid.UUID
	Date      time.Time
	Reason    *string
	CreatedAt time.Time
}

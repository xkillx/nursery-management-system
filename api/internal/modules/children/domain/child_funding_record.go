package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChildFundingRecord struct {
	ID                          uuid.UUID
	TenantID                    uuid.UUID
	BranchID                    uuid.UUID
	ChildID                     uuid.UUID

	BenefitsContributeToFees YesNoUnknown
	WorkingTaxCredit         YesNoUnknown
	CollegeUniPaidToParent   YesNoUnknown
	CollegeUniPaidToNursery  YesNoUnknown
	Funding3yoTermTime       YesNoUnknown
	Funding2yoTermTime       YesNoUnknown
	FundingSupportNotes      *string
	FundingSupportReviewed   bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

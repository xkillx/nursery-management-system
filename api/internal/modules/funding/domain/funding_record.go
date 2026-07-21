package domain

import (
	"time"

	"github.com/google/uuid"
)

type FundingType string

const (
	FundingTypeNone                FundingType = "none"
	FundingTypeUniversal15         FundingType = "universal_15"
	FundingTypeWorkingParent       FundingType = "working_parent"
	FundingTypeWorkingParentUnder3 FundingType = "working_parent_under_3"
	FundingTypeDisadvantaged2yo    FundingType = "disadvantaged_2yo"
	FundingTypeUnknown             FundingType = "unknown"
)

var validFundingTypes = map[FundingType]bool{
	FundingTypeNone: true, FundingTypeUniversal15: true, FundingTypeWorkingParent: true,
	FundingTypeWorkingParentUnder3: true, FundingTypeDisadvantaged2yo: true, FundingTypeUnknown: true,
}

func (ft FundingType) Valid() bool {
	return validFundingTypes[ft]
}

type FundingModel string

const (
	FundingModelTermTimeOnly FundingModel = "term_time_only"
	FundingModelStretched    FundingModel = "stretched"
	FundingModelUnknown      FundingModel = "unknown"
)

var validFundingModels = map[FundingModel]bool{
	FundingModelTermTimeOnly: true, FundingModelStretched: true, FundingModelUnknown: true,
}

func (fm FundingModel) Valid() bool {
	return validFundingModels[fm]
}

type FundingRecord struct {
	ID                       uuid.UUID
	TenantID                 uuid.UUID
	BranchID                 uuid.UUID
	ChildID                  uuid.UUID
	FundingEnabled           bool
	FundingType              FundingType
	FundingModel             FundingModel
	FundedHoursPerWeek       *float64
	FundingStartDate         *time.Time
	FundingEndDate           *time.Time
	EligibilityCode          *string
	EligibilityCodeValidated bool
	EvidenceReceived         bool
	EYPP                     bool
	DAF                      bool
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

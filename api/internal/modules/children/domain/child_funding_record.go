package domain

import (
	"time"

	"github.com/google/uuid"
)

type FundingType string

const (
	FundingTypeNone        FundingType = "none"
	FundingTypeFifteenHours FundingType = "fifteen_hours"
	FundingTypeThirtyHours FundingType = "thirty_hours"
	FundingTypeTwoYearOld  FundingType = "two_year_old"
	FundingTypeCustom      FundingType = "custom"
	FundingTypeUnknown     FundingType = "unknown"
)

var validFundingTypes = map[FundingType]bool{
	FundingTypeNone: true, FundingTypeFifteenHours: true, FundingTypeThirtyHours: true,
	FundingTypeTwoYearOld: true, FundingTypeCustom: true, FundingTypeUnknown: true,
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

type BenefitsStatus string

const (
	BenefitsStatusNo      BenefitsStatus = "no"
	BenefitsStatusYes     BenefitsStatus = "yes"
	BenefitsStatusUnknown BenefitsStatus = "unknown"
)

var validBenefitsStatuses = map[BenefitsStatus]bool{
	BenefitsStatusNo: true, BenefitsStatusYes: true, BenefitsStatusUnknown: true,
}

func (bs BenefitsStatus) Valid() bool {
	return validBenefitsStatuses[bs]
}

type ChildFundingRecord struct {
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
	BenefitsStatus           BenefitsStatus
	BenefitNotes             *string
	ManagerNotes             *string
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

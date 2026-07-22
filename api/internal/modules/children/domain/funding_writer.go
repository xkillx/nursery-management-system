package domain

import (
	"context"

	"github.com/google/uuid"
)

// ChildFundingRecordInput is the lightweight data carrier for funding writes
// during child registration. The adapter converts string enum values to the
// funding module's typed enums.
type ChildFundingRecordInput struct {
	FundingEnabled           bool
	FundingType              string
	FundingModel             string
	FundedHoursPerWeek       *float64
	FundingStartDate         *string
	FundingEndDate           *string
	EligibilityCode          *string
	EligibilityCodeValidated bool
	EvidenceReceived         bool
	BenefitsStatus           string
	Benefits                 []string
	OtherBenefitName         *string
	BenefitNotes             *string
	ManagerNotes             *string
}

// ChildFundingWriter abstracts funding writes during child registration.
// The funding module provides the concrete implementation via a bootstrap adapter,
// so the children module never imports the funding module directly.
type ChildFundingWriter interface {
	SaveFunding(ctx context.Context, tx any, tenantID, branchID, childID uuid.UUID, input *ChildFundingRecordInput) error
}

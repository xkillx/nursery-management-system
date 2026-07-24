package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ChildFundingData holds the funding record for a child, including benefit fields
// stored in the child_funding_records table.
type ChildFundingData struct {
	ID                       uuid.UUID
	ChildID                  uuid.UUID
	FundingEnabled           bool
	FundingType              string
	FundingModel             string
	FundedHoursPerWeek       *float64
	FundingStartDate         *time.Time
	FundingEndDate           *time.Time
	EligibilityCode          *string
	EligibilityCodeValidated bool
	EvidenceReceived         bool
	BenefitsStatus           string
	BenefitNotes             *string
	ManagerNotes             *string
	OtherBenefitName         *string
	BenefitUniversalCredit   bool
	BenefitIncomeSupport     bool
	BenefitJobseekers        bool
	BenefitESAIncomeRelated  bool
	BenefitChildTaxCredit    bool
	BenefitOtherSupport      bool
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

// ChildFundingReader abstracts funding reads for the children module.
// The funding module provides the concrete implementation via a bootstrap adapter,
// so the children module never imports the funding module directly.
type ChildFundingReader interface {
	GetFundingForChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildFundingData, bool, error)
}

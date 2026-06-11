package domain

import (
	"time"

	"github.com/google/uuid"
)

type OfficeCheckStatus string

const (
	OfficeCheckStatusUnknown       OfficeCheckStatus = "unknown"
	OfficeCheckStatusComplete      OfficeCheckStatus = "complete"
	OfficeCheckStatusMissing       OfficeCheckStatus = "missing"
	OfficeCheckStatusNotApplicable OfficeCheckStatus = "not_applicable"
)

type TermTimeOnlyStatus string

const (
	TermTimeOnlyStatusUnknown       TermTimeOnlyStatus = "unknown"
	TermTimeOnlyStatusYes           TermTimeOnlyStatus = "yes"
	TermTimeOnlyStatusNo            TermTimeOnlyStatus = "no"
	TermTimeOnlyStatusNotApplicable TermTimeOnlyStatus = "not_applicable"
)

type OfficeUseChecklist struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	BranchID  uuid.UUID
	ChildID   uuid.UUID

	DepositStatus       OfficeCheckStatus
	DepositPaidDate     *time.Time

	ApplicationDateStatus OfficeCheckStatus
	ApplicationDate       *time.Time

	StartDateStatus OfficeCheckStatus
	DateLeft        *time.Time

	SessionsDaysRequestedStatus OfficeCheckStatus
	SessionsDaysRequested       *string

	TermTimeOnlySpaceStatus TermTimeOnlyStatus

	ContractStatus OfficeCheckStatus
	ContractDate   *time.Time

	HandbookStatus OfficeCheckStatus
	HandbookDate   *time.Time

	RedBookStatus          OfficeCheckStatus
	RedBookCheckedDate     *time.Time

	BirthCertificatePassportStatus      OfficeCheckStatus
	BirthCertificatePassportCheckedDate *time.Time

	ProofOfAddressStatus        OfficeCheckStatus
	ProofOfAddressCheckedDate   *time.Time

	Notes *string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type OfficeChildSummary struct {
	ID          uuid.UUID
	FullName    string
	DateOfBirth time.Time
	StartDate   *time.Time
	EndDate     *time.Time
}

type OfficeUseChecklistWithChild struct {
	Checklist       *OfficeUseChecklist
	Child           OfficeChildSummary
	ChecklistExists bool
}

func DefaultOfficeUseChecklist() *OfficeUseChecklist {
	return &OfficeUseChecklist{
		ID:                            uuid.Nil,
		DepositStatus:                 OfficeCheckStatusUnknown,
		ApplicationDateStatus:         OfficeCheckStatusUnknown,
		StartDateStatus:               OfficeCheckStatusUnknown,
		SessionsDaysRequestedStatus:   OfficeCheckStatusUnknown,
		TermTimeOnlySpaceStatus:       TermTimeOnlyStatusUnknown,
		ContractStatus:                OfficeCheckStatusUnknown,
		HandbookStatus:                OfficeCheckStatusUnknown,
		RedBookStatus:                 OfficeCheckStatusUnknown,
		BirthCertificatePassportStatus: OfficeCheckStatusUnknown,
		ProofOfAddressStatus:          OfficeCheckStatusUnknown,
	}
}

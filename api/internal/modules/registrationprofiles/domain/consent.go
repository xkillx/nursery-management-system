package domain

import (
	"time"

	"github.com/google/uuid"
)

type ConsentSource string

const (
	ConsentSourcePaperForm ConsentSource = "paper_form"
)

type ConsentRecord struct {
	ID      uuid.UUID
	TenantID uuid.UUID
	BranchID uuid.UUID
	ChildID  uuid.UUID
	Version  int
	Source   ConsentSource

	SignerName       string
	SignedDate       time.Time
	PaperFormOnFile  bool

	UrgentMedicalTreatment         bool
	UrgentMedicalTreatmentExceptions *string
	Plasters                       bool
	SafeguardingReportingAcknowledgement bool
	InformationSharingConsent      bool
	AreaSENCOLiaison               bool
	HealthVisitorLiaison           bool
	TransitionDocuments            bool
	LocalOutings                   bool
	FacePainting                   bool
	ParentSuppliedSunCream         bool
	ParentSuppliedNappyCream       bool
	DevelopmentProfilePhotos       bool
	NurseryDisplayBoards           bool
	PromotionalLiterature          bool
	NurseryWebsite                 bool
	StaffStudentCoursework         bool
	SocialMedia                    bool
	SocialMediaChannelNotes        *string

	NotesExceptions *string

	EnteredByUserID       uuid.UUID
	EnteredByMembershipID uuid.UUID

	CreatedAt time.Time
}

type ConsentWithCompleteness struct {
	Current      *ConsentRecord   `json:"current"`
	History      []ConsentRecord  `json:"history"`
	Completeness ConsentCompleteness `json:"completeness"`
}

type ConsentCompleteness struct {
	IsComplete      bool     `json:"is_complete"`
	MissingDecisions []string `json:"missing_decisions,omitempty"`
}

func ComputeConsentCompleteness(record *ConsentRecord) ConsentCompleteness {
	if record == nil {
		return ConsentCompleteness{IsComplete: false, MissingDecisions: []string{"no_consent_record"}}
	}
	missing := make([]string, 0)

	if !record.SafeguardingReportingAcknowledgement {
		missing = append(missing, "safeguarding_reporting_acknowledgement")
	}

	if !record.InformationSharingConsent {
		missing = append(missing, "information_sharing_consent")
	}

	if record.SignerName == "" {
		missing = append(missing, "signer_name")
	}

	return ConsentCompleteness{
		IsComplete:      len(missing) == 0,
		MissingDecisions: missing,
	}
}

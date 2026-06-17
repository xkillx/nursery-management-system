package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChildConsent struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	BranchID uuid.UUID
	ChildID  uuid.UUID

	UrgentMedicalTreatment               bool
	UrgentMedicalTreatmentExceptions     *string
	Plasters                             bool
	SafeguardingReportingAcknowledgement bool
	InformationSharingConsent            bool
	GDPRDataProcessingConsent            bool
	AreaSENCOLiaison                     bool
	HealthVisitorLiaison                 bool
	TransitionDocuments                  bool
	LocalOutings                         bool
	FacePainting                         bool
	ParentSuppliedSunCream               bool
	ParentSuppliedNappyCream             bool
	DevelopmentProfilePhotos             bool
	NurseryDisplayBoards                 bool
	PromotionalLiterature                bool
	NurseryWebsite                       bool
	StaffStudentCoursework               bool
	SocialMedia                          bool
	SocialMediaChannelNotes              *string

	NotesExceptions *string

	SignerName      string
	SignedDate      time.Time
	PaperFormOnFile bool

	EnteredByUserID       uuid.UUID
	EnteredByMembershipID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}

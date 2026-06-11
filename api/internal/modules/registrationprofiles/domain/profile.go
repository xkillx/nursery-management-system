package domain

import (
	"time"

	"github.com/google/uuid"
)

type YesNoUnknown string

const (
	YesNoUnknownUnknown YesNoUnknown = "unknown"
	YesNoUnknownNo      YesNoUnknown = "no"
	YesNoUnknownYes     YesNoUnknown = "yes"
)

type ImmunisationStatus string

const (
	ImmunisationUnknown       ImmunisationStatus = "unknown"
	ImmunisationUpToDate      ImmunisationStatus = "up_to_date"
	ImmunisationRefused       ImmunisationStatus = "refused"
	ImmunisationPartial       ImmunisationStatus = "partial"
	ImmunisationNotRecorded   ImmunisationStatus = "not_recorded"
)

type ContactType string

const (
	ContactTypeParentCarer         ContactType = "parent_carer"
	ContactTypeEmergencyContact    ContactType = "emergency_contact"
	ContactTypeAuthorisedCollector ContactType = "authorised_collector"
)

type ProfessionalReferral struct {
	Type             string  `json:"type"`
	ReferredDate     *string `json:"referred_date,omitempty"`
	ReferredBy       *string `json:"referred_by,omitempty"`
	WaitingListStatus string `json:"waiting_list_status"`
	Notes            *string `json:"notes,omitempty"`
}

type Profile struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	BranchID  uuid.UUID
	ChildID   uuid.UUID

	Sex                    *string
	Religion               *string
	EthnicOrigin           *string
	FirstLanguage          *string
	OtherLanguages         []string
	HomeAddress            map[string]any
	HomePostcode           *string
	HomeTelephone          *string

	DisabilityStatus       YesNoUnknown
	DisabilityNotes        *string
	AccessRequirements     *string

	MedicalConditionsStatus       YesNoUnknown
	MedicalConditionsNotes        *string
	PrescribedMedicationStatus    YesNoUnknown
	MedicationNotes               *string
	ImmunisationStatus            ImmunisationStatus
	ImmunisationCountry           *string
	IllnessDiagnosisHistory       *string
	DietaryRequirementsStatus     YesNoUnknown
	DietaryRequirementsNotes      *string
	DietarySideEffects            *string

	DoctorName            *string
	DoctorAddress         *string
	DoctorPhone           *string
	HealthVisitorName     *string
	HealthVisitorAddress  *string
	HealthVisitorPhone    *string

	SocialServicesStatus          YesNoUnknown
	SocialServicesNotes           *string
	SocialWorkerContactDetails    *string
	ConcernWalking                YesNoUnknown
	ConcernSpeechLanguage         YesNoUnknown
	ConcernHearing                YesNoUnknown
	ConcernSight                  YesNoUnknown
	ConcernEmotionalWellbeing     YesNoUnknown
	ConcernBehaviour              YesNoUnknown
	ProfessionalReferrals         []ProfessionalReferral

	ParentalResponsibilityNotes  *string

	Over18CollectionAcknowledged bool
	CollectionPasswordHash       *string
	CollectionPasswordUpdatedAt         *time.Time
	CollectionPasswordUpdatedByUserID   *uuid.UUID
	CollectionPasswordUpdatedByMembershipID *uuid.UUID

	BenefitsContributeToFees     YesNoUnknown
	WorkingTaxCredit             YesNoUnknown
	CollegeUniPaidToParent       YesNoUnknown
	CollegeUniPaidToNursery      YesNoUnknown
	Funding3yoTermTime           YesNoUnknown
	Funding2yoTermTime           YesNoUnknown
	FundingSupportNotes          *string

	RoutineCareNotes        *string

	GDPRDeclaredByName      *string
	GDPRDeclaredAt          *time.Time
	GDPRDeclarationDate     *time.Time

	DemographicsHomeReviewed       bool
	MedicalDietaryReviewed         bool
	HealthContactsReviewed         bool
	SocialDevelopmentReviewed      bool
	ParentResponsibilityReviewed   bool
	EmergencyCollectionReviewed    bool
	FundingSupportReviewed         bool
	RoutineCareReviewed            bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ContactEntry struct {
	ID                       uuid.UUID
	TenantID                 uuid.UUID
	BranchID                 uuid.UUID
	ProfileID                uuid.UUID
	ChildID                  uuid.UUID
	ContactType              ContactType
	SortOrder                int
	FullName                 string
	RelationshipToChild      *string
	Address                  map[string]any
	Telephone                *string
	Email                    *string
	WorkAddress              map[string]any
	HasParentalResponsibility *bool
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type ChildSummary struct {
	ID          uuid.UUID
	FullName    string
	DateOfBirth time.Time
}

type ProfileWithChild struct {
	Profile        *Profile
	Child          ChildSummary
	Contacts       []ContactEntry
	ProfileExists  bool
}

type CollectionPasswordMetadata struct {
	IsSet            bool
	UpdatedAt        *time.Time
	UpdatedByUserID  *uuid.UUID
	UpdatedByMembershipID *uuid.UUID
}

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
	ImmunisationUnknown     ImmunisationStatus = "unknown"
	ImmunisationUpToDate    ImmunisationStatus = "up_to_date"
	ImmunisationRefused     ImmunisationStatus = "refused"
	ImmunisationPartial     ImmunisationStatus = "partial"
	ImmunisationNotRecorded ImmunisationStatus = "not_recorded"
)

type ChildProfile struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	BranchID    uuid.UUID
	ChildID     uuid.UUID
	Sex         *string
	Religion    *string
	EthnicOrigin *string
	FirstLanguage *string
	OtherLanguages *string
	HomeAddress  map[string]any
	HomePostcode *string
	HomeTelephone *string

	DisabilityStatus   YesNoUnknown
	DisabilityNotes    *string
	AccessRequirements *string

	RoutineCareNotes *string

	GDPRDeclaredByName  *string
	GDPRDeclaredAt      *time.Time
	GDPRDeclarationDate *time.Time

	RegistrationDate *time.Time

	DemographicsHomeReviewed     bool
	MedicalDietaryReviewed       bool
	HealthContactsReviewed       bool
	SocialDevelopmentReviewed    bool
	ParentResponsibilityReviewed bool
	EmergencyCollectionReviewed  bool
	RoutineCareReviewed          bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

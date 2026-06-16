package domain

import (
	"time"

	"github.com/google/uuid"
)

type CompleteRegistrationInput struct {
	Child              ChildRegistrationInfo
	Profile            ProfileSectionsInput
	Consents           ConsentInput
	CollectionPassword string
}

type ChildRegistrationInfo struct {
	FirstName     string
	MiddleName    string
	LastName      string
	DateOfBirth   string
	StartDate     string
	Notes         string
	PrimaryRoomID *string
}

type ProfileSectionsInput struct {
	DemographicsHome       *DemographicsHomeInput
	MedicalDietary         *MedicalDietaryInput
	HealthContacts         *HealthContactsInput
	SocialDevelopment      *SocialDevelopmentInput
	ParentCarers           []ContactEntryInput
	EmergencyContacts      []ContactEntryInput
	AuthorisedCollectors   []ContactEntryInput
	Collection             *CollectionInput
	FundingSupport         *FundingSupportInput
	RoutineCare            *RoutineCareInput
	GDPRDeclaration        *GDPRDeclarationInput
	PaperFormCompletedDate *string
}

type DemographicsHomeInput struct {
	Sex                      *string
	Religion                 *string
	EthnicOrigin             *string
	FirstLanguage            *string
	OtherLanguages           *string
	HomeAddress              map[string]any
	HomePostcode             *string
	HomeTelephone            *string
	DisabilityStatus         *string
	DisabilityNotes          *string
	AccessRequirements       *string
	DemographicsHomeReviewed bool
}

type MedicalDietaryInput struct {
	MedicalConditionsStatus    *string
	MedicalConditionsNotes     *string
	PrescribedMedicationStatus *string
	MedicationNotes            *string
	ImmunisationStatus         *string
	ImmunisationCountry        *string
	IllnessDiagnosisHistory    *string
	DietaryRequirementsStatus  *string
	DietaryRequirementsNotes   *string
	DietarySideEffects         *string
	MedicalDietaryReviewed     bool
}

type HealthContactsInput struct {
	DoctorName             *string
	DoctorAddress          *string
	DoctorPhone            *string
	HealthVisitorName      *string
	HealthVisitorAddress   *string
	HealthVisitorPhone     *string
	HealthContactsReviewed bool
}

type SocialDevelopmentInput struct {
	SocialServicesStatus      *string
	SocialServicesNotes       *string
	SocialWorkerName          *string
	SocialWorkerPhone         *string
	SocialWorkerEmail         *string
	ConcernWalking            *string
	ConcernSpeechLanguage     *string
	ConcernHearing            *string
	ConcernSight              *string
	ConcernEmotionalWellbeing *string
	ConcernBehaviour          *string
	ProfessionalReferrals     []ProfessionalReferralInput
	SocialDevelopmentReviewed bool
}

type ProfessionalReferralInput struct {
	Type              string
	ReferredDate      *string
	ReferredBy        *string
	WaitingListStatus string
	Notes             *string
}

type ContactEntryInput struct {
	FullName                  string
	RelationshipToChild       *string
	Address                   map[string]any
	Telephone                 *string
	Email                     *string
	WorkAddress               map[string]any
	HasParentalResponsibility *bool
}

type CollectionInput struct {
	Over18CollectionAcknowledged bool
	EmergencyCollectionReviewed  bool
}

type FundingSupportInput struct {
	BenefitsContributeToFees *string
	WorkingTaxCredit         *string
	CollegeUniPaidToParent   *string
	CollegeUniPaidToNursery  *string
	Funding3yoTermTime       *string
	Funding2yoTermTime       *string
	FundingSupportNotes      *string
	FundingSupportReviewed   bool
}

type RoutineCareInput struct {
	RoutineCareNotes    *string
	RoutineCareReviewed bool
}

type GDPRDeclarationInput struct {
	GDPRDeclaredByName  *string
	GDPRDeclarationDate *string
}

type ConsentInput struct {
	PaperFormOnFile bool

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
}

func (c *CompleteRegistrationInput) ToProfile(tenantID, branchID, childID uuid.UUID) *Profile {
	p := &Profile{
		ID:       uuid.Nil,
		TenantID: tenantID,
		BranchID: branchID,
		ChildID:  childID,
	}

	if pi := c.Profile.DemographicsHome; pi != nil {
		p.Sex = pi.Sex
		p.Religion = pi.Religion
		p.EthnicOrigin = pi.EthnicOrigin
		p.FirstLanguage = pi.FirstLanguage
		p.OtherLanguages = pi.OtherLanguages
		p.HomeAddress = pi.HomeAddress
		p.HomePostcode = pi.HomePostcode
		p.HomeTelephone = pi.HomeTelephone
		p.DisabilityStatus = parseYesNoUnknown(pi.DisabilityStatus)
		p.DisabilityNotes = pi.DisabilityNotes
		p.AccessRequirements = pi.AccessRequirements
		p.DemographicsHomeReviewed = pi.DemographicsHomeReviewed
	}

	if pi := c.Profile.MedicalDietary; pi != nil {
		p.MedicalConditionsStatus = parseYesNoUnknown(pi.MedicalConditionsStatus)
		p.MedicalConditionsNotes = pi.MedicalConditionsNotes
		p.PrescribedMedicationStatus = parseYesNoUnknown(pi.PrescribedMedicationStatus)
		p.MedicationNotes = pi.MedicationNotes
		p.ImmunisationStatus = parseImmunisationStatus(pi.ImmunisationStatus)
		p.ImmunisationCountry = pi.ImmunisationCountry
		p.IllnessDiagnosisHistory = pi.IllnessDiagnosisHistory
		p.DietaryRequirementsStatus = parseYesNoUnknown(pi.DietaryRequirementsStatus)
		p.DietaryRequirementsNotes = pi.DietaryRequirementsNotes
		p.DietarySideEffects = pi.DietarySideEffects
		p.MedicalDietaryReviewed = pi.MedicalDietaryReviewed
	}

	if pi := c.Profile.HealthContacts; pi != nil {
		p.DoctorName = pi.DoctorName
		p.DoctorAddress = pi.DoctorAddress
		p.DoctorPhone = pi.DoctorPhone
		p.HealthVisitorName = pi.HealthVisitorName
		p.HealthVisitorAddress = pi.HealthVisitorAddress
		p.HealthVisitorPhone = pi.HealthVisitorPhone
		p.HealthContactsReviewed = pi.HealthContactsReviewed
	}

	if pi := c.Profile.SocialDevelopment; pi != nil {
		p.SocialServicesStatus = parseYesNoUnknown(pi.SocialServicesStatus)
		p.SocialServicesNotes = pi.SocialServicesNotes
		p.SocialWorkerName = pi.SocialWorkerName
		p.SocialWorkerPhone = pi.SocialWorkerPhone
		p.SocialWorkerEmail = pi.SocialWorkerEmail
		p.ConcernWalking = parseYesNoUnknown(pi.ConcernWalking)
		p.ConcernSpeechLanguage = parseYesNoUnknown(pi.ConcernSpeechLanguage)
		p.ConcernHearing = parseYesNoUnknown(pi.ConcernHearing)
		p.ConcernSight = parseYesNoUnknown(pi.ConcernSight)
		p.ConcernEmotionalWellbeing = parseYesNoUnknown(pi.ConcernEmotionalWellbeing)
		p.ConcernBehaviour = parseYesNoUnknown(pi.ConcernBehaviour)
		p.SocialDevelopmentReviewed = pi.SocialDevelopmentReviewed
	}

	if pi := c.Profile.Collection; pi != nil {
		p.Over18CollectionAcknowledged = pi.Over18CollectionAcknowledged
		p.EmergencyCollectionReviewed = pi.EmergencyCollectionReviewed
	}

	if pi := c.Profile.FundingSupport; pi != nil {
		p.BenefitsContributeToFees = parseYesNoUnknown(pi.BenefitsContributeToFees)
		p.WorkingTaxCredit = parseYesNoUnknown(pi.WorkingTaxCredit)
		p.CollegeUniPaidToParent = parseYesNoUnknown(pi.CollegeUniPaidToParent)
		p.CollegeUniPaidToNursery = parseYesNoUnknown(pi.CollegeUniPaidToNursery)
		p.Funding3yoTermTime = parseYesNoUnknown(pi.Funding3yoTermTime)
		p.Funding2yoTermTime = parseYesNoUnknown(pi.Funding2yoTermTime)
		p.FundingSupportNotes = pi.FundingSupportNotes
		p.FundingSupportReviewed = pi.FundingSupportReviewed
	}

	if pi := c.Profile.RoutineCare; pi != nil {
		p.RoutineCareNotes = pi.RoutineCareNotes
		p.RoutineCareReviewed = pi.RoutineCareReviewed
	}

	if pi := c.Profile.GDPRDeclaration; pi != nil {
		p.GDPRDeclaredByName = pi.GDPRDeclaredByName
	}

	if c.Profile.PaperFormCompletedDate != nil && *c.Profile.PaperFormCompletedDate != "" {
		t, err := time.Parse("2006-01-02", *c.Profile.PaperFormCompletedDate)
		if err == nil {
			p.PaperFormCompletedDate = &t
		}
	}

	return p
}

func parseYesNoUnknown(s *string) YesNoUnknown {
	if s == nil || *s == "" {
		return YesNoUnknownUnknown
	}
	switch *s {
	case "yes":
		return YesNoUnknownYes
	case "no":
		return YesNoUnknownNo
	default:
		return YesNoUnknownUnknown
	}
}

func parseImmunisationStatus(s *string) ImmunisationStatus {
	if s == nil || *s == "" {
		return ImmunisationUnknown
	}
	switch *s {
	case "up_to_date":
		return ImmunisationUpToDate
	case "refused":
		return ImmunisationRefused
	case "partial":
		return ImmunisationPartial
	default:
		return ImmunisationUnknown
	}
}

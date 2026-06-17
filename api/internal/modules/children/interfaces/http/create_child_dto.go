package httpchild

import (
	"fmt"

	"nursery-management-system/api/internal/modules/children/application"
	"nursery-management-system/api/internal/modules/children/domain"
)

type createChildRequest struct {
	Child             childIdentityPayload   `json:"child"`
	Profile           *childProfilePayload   `json:"profile"`
	Health            *childHealthPayload    `json:"health"`
	Safeguarding      *childSafeguardingPayload `json:"safeguarding"`
	Contacts          *childContactsPayload  `json:"contacts"`
	Consent           *childConsentPayload   `json:"consent"`
	Funding           *childFundingPayload   `json:"funding"`
	CollectionSettings *collectionSettingsPayload `json:"collection_settings"`
	Room              *roomAssignmentPayload `json:"room"`
}

type childIdentityPayload struct {
	FirstName   string  `json:"first_name"`
	MiddleName  *string `json:"middle_name"`
	LastName    *string `json:"last_name"`
	DateOfBirth string  `json:"date_of_birth"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Notes       *string `json:"notes"`
}

type childProfilePayload struct {
	Sex                       *string                `json:"sex"`
	Religion                  *string                `json:"religion"`
	EthnicOrigin              *string                `json:"ethnic_origin"`
	FirstLanguage             *string                `json:"first_language"`
	OtherLanguages            *string                `json:"other_languages"`
	HomeAddress               map[string]any         `json:"home_address"`
	HomePostcode              *string                `json:"home_postcode"`
	HomeTelephone             *string                `json:"home_telephone"`
	DisabilityStatus          string                 `json:"disability_status"`
	DisabilityNotes           *string                `json:"disability_notes"`
	AccessRequirements        *string                `json:"access_requirements"`
	RoutineCareNotes          *string                `json:"routine_care_notes"`
	GDPRDeclaredByName        *string                `json:"gdpr_declared_by_name"`
	GDPRDeclaredAt            *string                `json:"gdpr_declared_at"`
	GDPRDeclarationDate       *string                `json:"gdpr_declaration_date"`
	RegistrationDate          *string                `json:"registration_date"`
	DemographicsHomeReviewed  bool                   `json:"demographics_home_reviewed"`
	MedicalDietaryReviewed    bool                   `json:"medical_dietary_reviewed"`
	HealthContactsReviewed    bool                   `json:"health_contacts_reviewed"`
	SocialDevelopmentReviewed bool                   `json:"social_development_reviewed"`
	ParentResponsibilityReviewed bool                 `json:"parent_responsibility_reviewed"`
	EmergencyCollectionReviewed bool                  `json:"emergency_collection_reviewed"`
	RoutineCareReviewed       bool                   `json:"routine_care_reviewed"`
}

type childHealthPayload struct {
	MedicalConditionsStatus    string  `json:"medical_conditions_status"`
	MedicalConditionsNotes     *string `json:"medical_conditions_notes"`
	PrescribedMedicationStatus string  `json:"prescribed_medication_status"`
	MedicationNotes            *string `json:"medication_notes"`
	ImmunisationStatus         string  `json:"immunisation_status"`
	ImmunisationCountry        *string `json:"immunisation_country"`
	IllnessDiagnosisHistory    *string `json:"illness_diagnosis_history"`
	DietaryRequirementsStatus  string  `json:"dietary_requirements_status"`
	DietaryRequirementsNotes   *string `json:"dietary_requirements_notes"`
	DietarySideEffects         *string `json:"dietary_side_effects"`
	DoctorName                 *string `json:"doctor_name"`
	DoctorAddress              *string `json:"doctor_address"`
	DoctorPhone                *string `json:"doctor_phone"`
	HealthVisitorName          *string `json:"health_visitor_name"`
	HealthVisitorAddress       *string `json:"health_visitor_address"`
	HealthVisitorPhone         *string `json:"health_visitor_phone"`
}

type childSafeguardingPayload struct {
	SocialServicesStatus      string  `json:"social_services_status"`
	SocialServicesNotes       *string `json:"social_services_notes"`
	SocialWorkerName          *string `json:"social_worker_name"`
	SocialWorkerPhone         *string `json:"social_worker_phone"`
	SocialWorkerEmail         *string `json:"social_worker_email"`
	ConcernWalking            string  `json:"concern_walking"`
	ConcernSpeechLanguage     string  `json:"concern_speech_language"`
	ConcernHearing            string  `json:"concern_hearing"`
	ConcernSight              string  `json:"concern_sight"`
	ConcernEmotionalWellbeing string  `json:"concern_emotional_wellbeing"`
	ConcernBehaviour          string  `json:"concern_behaviour"`
	ProfessionalReferrals     []domain.ProfessionalReferral `json:"professional_referrals"`
	RestrictedNotes           *string `json:"restricted_notes"`
}

type childContactsPayload struct {
	ParentCarers         []contactPayload `json:"parent_carers"`
	EmergencyContacts    []contactPayload `json:"emergency_contacts"`
	AuthorisedCollectors []contactPayload `json:"authorised_collectors"`
}

type childConsentPayload struct {
	UrgentMedicalTreatment               bool    `json:"urgent_medical_treatment"`
	UrgentMedicalTreatmentExceptions     *string `json:"urgent_medical_treatment_exceptions"`
	Plasters                             bool    `json:"plasters"`
	SafeguardingReportingAcknowledgement bool    `json:"safeguarding_reporting_acknowledgement"`
	InformationSharingConsent            bool    `json:"information_sharing_consent"`
	GDPRDataProcessingConsent            bool    `json:"gdpr_data_processing_consent"`
	AreaSENCOLiaison                     bool    `json:"area_senco_liaison"`
	HealthVisitorLiaison                 bool    `json:"health_visitor_liaison"`
	TransitionDocuments                  bool    `json:"transition_documents"`
	LocalOutings                         bool    `json:"local_outings"`
	FacePainting                         bool    `json:"face_painting"`
	ParentSuppliedSunCream               bool    `json:"parent_supplied_sun_cream"`
	ParentSuppliedNappyCream             bool    `json:"parent_supplied_nappy_cream"`
	DevelopmentProfilePhotos             bool    `json:"development_profile_photos"`
	NurseryDisplayBoards                 bool    `json:"nursery_display_boards"`
	PromotionalLiterature                bool    `json:"promotional_literature"`
	NurseryWebsite                       bool    `json:"nursery_website"`
	StaffStudentCoursework               bool    `json:"staff_student_coursework"`
	SocialMedia                          bool    `json:"social_media"`
	SocialMediaChannelNotes              *string `json:"social_media_channel_notes"`
	NotesExceptions                      *string `json:"notes_exceptions"`
	SignerName                           string  `json:"signer_name"`
	SignedDate                           string  `json:"signed_date"`
	PaperFormOnFile                      bool    `json:"paper_form_on_file"`
}

type childFundingPayload struct {
	BenefitsContributeToFees string  `json:"benefits_contribute_to_fees"`
	WorkingTaxCredit         string  `json:"working_tax_credit"`
	CollegeUniPaidToParent   string  `json:"college_uni_paid_to_parent"`
	CollegeUniPaidToNursery  string  `json:"college_uni_paid_to_nursery"`
	Funding3yoTermTime       string  `json:"funding_3yo_term_time"`
	Funding2yoTermTime       string  `json:"funding_2yo_term_time"`
	FundingSupportNotes      *string `json:"funding_support_notes"`
	FundingSupportReviewed   bool    `json:"funding_support_reviewed"`
}

type collectionSettingsPayload struct {
	Over18CollectionAcknowledged bool    `json:"over_18_collection_acknowledged"`
	Password                     *string `json:"password"`
}

type roomAssignmentPayload struct {
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
}

func mapCreateChildRequest(req createChildRequest) (application.CreateChildFullInput, error) {
	in := application.CreateChildFullInput{
		Child: application.CreateChildIdentityInput{
			FirstName: req.Child.FirstName,
			DateOfBirth: req.Child.DateOfBirth,
			StartDate:   req.Child.StartDate,
		},
	}
	if req.Child.MiddleName != nil {
		in.Child.MiddleName = *req.Child.MiddleName
	}
	if req.Child.LastName != nil {
		in.Child.LastName = *req.Child.LastName
	}
	if req.Child.EndDate != nil {
		in.Child.EndDate = *req.Child.EndDate
	}
	if req.Child.Notes != nil {
		in.Child.Notes = *req.Child.Notes
	}

	if req.Profile != nil {
		in.Profile = mapChildProfilePayloadToInput(req.Profile)
	}
	if req.Health != nil {
		in.Health = mapChildHealthPayloadToInput(req.Health)
	}
	if req.Safeguarding != nil {
		in.Safeguarding = mapChildSafeguardingPayloadToInput(req.Safeguarding)
	}
	if req.Contacts != nil {
		in.Contacts = mapChildContactsPayloadToInput(req.Contacts)
	}
	if req.Consent != nil {
		in.Consent = mapChildConsentPayloadToInput(req.Consent)
	}
	if req.Funding != nil {
		in.Funding = mapChildFundingPayloadToInput(req.Funding)
	}
	if req.CollectionSettings != nil {
		in.CollectionSettings = &application.ChildCollectionSettingsInput{
			Over18CollectionAcknowledged: req.CollectionSettings.Over18CollectionAcknowledged,
		}
		if req.CollectionSettings.Password != nil {
			in.CollectionSettings.Password = *req.CollectionSettings.Password
		}
	}
	if req.Room != nil {
		in.Room = &application.ChildRoomAssignmentInput{
			RoomID:    req.Room.RoomID,
			StartDate: req.Room.StartDate,
		}
	}
	return in, nil
}

func mapChildProfilePayloadToInput(p *childProfilePayload) *application.ChildProfileInput {
	return &application.ChildProfileInput{
		Sex: p.Sex, Religion: p.Religion, EthnicOrigin: p.EthnicOrigin,
		FirstLanguage: p.FirstLanguage, OtherLanguages: p.OtherLanguages,
		HomeAddress: p.HomeAddress, HomePostcode: p.HomePostcode, HomeTelephone: p.HomeTelephone,
		DisabilityStatus: p.DisabilityStatus, DisabilityNotes: p.DisabilityNotes, AccessRequirements: p.AccessRequirements,
		RoutineCareNotes: p.RoutineCareNotes,
		GDPRDeclaredByName: p.GDPRDeclaredByName, GDPRDeclaredAt: p.GDPRDeclaredAt, GDPRDeclarationDate: p.GDPRDeclarationDate,
		RegistrationDate: p.RegistrationDate,
		DemographicsHomeReviewed: p.DemographicsHomeReviewed, MedicalDietaryReviewed: p.MedicalDietaryReviewed,
		HealthContactsReviewed: p.HealthContactsReviewed, SocialDevelopmentReviewed: p.SocialDevelopmentReviewed,
		ParentResponsibilityReviewed: p.ParentResponsibilityReviewed, EmergencyCollectionReviewed: p.EmergencyCollectionReviewed,
		RoutineCareReviewed: p.RoutineCareReviewed,
	}
}

func mapChildHealthPayloadToInput(p *childHealthPayload) *application.ChildHealthProfileInput {
	return &application.ChildHealthProfileInput{
		MedicalConditionsStatus: p.MedicalConditionsStatus, MedicalConditionsNotes: p.MedicalConditionsNotes,
		PrescribedMedicationStatus: p.PrescribedMedicationStatus, MedicationNotes: p.MedicationNotes,
		ImmunisationStatus: p.ImmunisationStatus, ImmunisationCountry: p.ImmunisationCountry,
		IllnessDiagnosisHistory: p.IllnessDiagnosisHistory,
		DietaryRequirementsStatus: p.DietaryRequirementsStatus, DietaryRequirementsNotes: p.DietaryRequirementsNotes,
		DietarySideEffects: p.DietarySideEffects,
		DoctorName: p.DoctorName, DoctorAddress: p.DoctorAddress, DoctorPhone: p.DoctorPhone,
		HealthVisitorName: p.HealthVisitorName, HealthVisitorAddress: p.HealthVisitorAddress, HealthVisitorPhone: p.HealthVisitorPhone,
	}
}

func mapChildSafeguardingPayloadToInput(p *childSafeguardingPayload) *application.ChildSafeguardingProfileInput {
	return &application.ChildSafeguardingProfileInput{
		SocialServicesStatus: p.SocialServicesStatus, SocialServicesNotes: p.SocialServicesNotes,
		SocialWorkerName: p.SocialWorkerName, SocialWorkerPhone: p.SocialWorkerPhone, SocialWorkerEmail: p.SocialWorkerEmail,
		ConcernWalking: p.ConcernWalking, ConcernSpeechLanguage: p.ConcernSpeechLanguage,
		ConcernHearing: p.ConcernHearing, ConcernSight: p.ConcernSight,
		ConcernEmotionalWellbeing: p.ConcernEmotionalWellbeing, ConcernBehaviour: p.ConcernBehaviour,
		ProfessionalReferrals: p.ProfessionalReferrals, RestrictedNotes: p.RestrictedNotes,
	}
}

func mapChildContactsPayloadToInput(p *childContactsPayload) []application.ChildContactInput {
	out := make([]application.ChildContactInput, 0, len(p.ParentCarers)+len(p.EmergencyContacts)+len(p.AuthorisedCollectors))
	for _, c := range p.ParentCarers {
		out = append(out, application.ChildContactInput{
			ContactType: domain.ContactTypeParentCarer, FullName: c.FullName,
			RelationshipToChild: c.RelationshipToChild, Address: c.Address,
			Telephone: c.Telephone, Email: c.Email, WorkAddress: c.WorkAddress,
			HasParentalResponsibility: c.HasParentalResponsibility,
		})
	}
	for _, c := range p.EmergencyContacts {
		out = append(out, application.ChildContactInput{
			ContactType: domain.ContactTypeEmergencyContact, FullName: c.FullName,
			RelationshipToChild: c.RelationshipToChild, Address: c.Address,
			Telephone: c.Telephone, Email: c.Email, WorkAddress: c.WorkAddress,
			HasParentalResponsibility: c.HasParentalResponsibility,
		})
	}
	for _, c := range p.AuthorisedCollectors {
		out = append(out, application.ChildContactInput{
			ContactType: domain.ContactTypeAuthorisedCollector, FullName: c.FullName,
			RelationshipToChild: c.RelationshipToChild, Address: c.Address,
			Telephone: c.Telephone, Email: c.Email, WorkAddress: c.WorkAddress,
			HasParentalResponsibility: c.HasParentalResponsibility,
		})
	}
	return out
}

func mapChildConsentPayloadToInput(p *childConsentPayload) *application.ChildConsentInput {
	return &application.ChildConsentInput{
		UrgentMedicalTreatment: p.UrgentMedicalTreatment, UrgentMedicalTreatmentExceptions: p.UrgentMedicalTreatmentExceptions,
		Plasters: p.Plasters, SafeguardingReportingAcknowledgement: p.SafeguardingReportingAcknowledgement,
		InformationSharingConsent: p.InformationSharingConsent, GDPRDataProcessingConsent: p.GDPRDataProcessingConsent,
		AreaSENCOLiaison: p.AreaSENCOLiaison, HealthVisitorLiaison: p.HealthVisitorLiaison,
		TransitionDocuments: p.TransitionDocuments, LocalOutings: p.LocalOutings, FacePainting: p.FacePainting,
		ParentSuppliedSunCream: p.ParentSuppliedSunCream, ParentSuppliedNappyCream: p.ParentSuppliedNappyCream,
		DevelopmentProfilePhotos: p.DevelopmentProfilePhotos, NurseryDisplayBoards: p.NurseryDisplayBoards,
		PromotionalLiterature: p.PromotionalLiterature, NurseryWebsite: p.NurseryWebsite,
		StaffStudentCoursework: p.StaffStudentCoursework, SocialMedia: p.SocialMedia,
		SocialMediaChannelNotes: p.SocialMediaChannelNotes, NotesExceptions: p.NotesExceptions,
		SignerName: p.SignerName, SignedDate: p.SignedDate, PaperFormOnFile: p.PaperFormOnFile,
	}
}

func mapChildFundingPayloadToInput(p *childFundingPayload) *application.ChildFundingRecordInput {
	return &application.ChildFundingRecordInput{
		BenefitsContributeToFees: p.BenefitsContributeToFees, WorkingTaxCredit: p.WorkingTaxCredit,
		CollegeUniPaidToParent: p.CollegeUniPaidToParent, CollegeUniPaidToNursery: p.CollegeUniPaidToNursery,
		Funding3yoTermTime: p.Funding3yoTermTime, Funding2yoTermTime: p.Funding2yoTermTime,
		FundingSupportNotes: p.FundingSupportNotes, FundingSupportReviewed: p.FundingSupportReviewed,
	}
}

type childCreationResponse struct {
	ID                 string   `json:"id"`
	FirstName          string   `json:"first_name"`
	MiddleName         *string  `json:"middle_name"`
	LastName           *string  `json:"last_name"`
	StartDate          string   `json:"start_date"`
	CreatedSubRecords  []string `json:"created_sub_records"`
}

func toChildCreationResponse(r *application.ChildCreationResult) childCreationResponse {
	return childCreationResponse{
		ID:                r.ChildID.String(),
		FirstName:         r.FirstName,
		MiddleName:        r.MiddleName,
		LastName:          r.LastName,
		StartDate:         r.StartDate,
		CreatedSubRecords: r.CreatedSubRecords,
	}
}

var _ = fmt.Sprint

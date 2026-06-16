package httpregistrationprofile

type submitCompleteRegistrationRequest struct {
	Child               childWritePayload           `json:"child"`
	RegistrationProfile *registrationProfilePayload `json:"registration_profile,omitempty"`
	Consents            *consentPayload             `json:"consents,omitempty"`
	CollectionPassword  *string                     `json:"collection_password,omitempty"`
}

type childWritePayload struct {
	FirstName     string  `json:"first_name"`
	MiddleName    *string `json:"middle_name"`
	LastName      *string `json:"last_name"`
	DateOfBirth   string  `json:"date_of_birth"`
	StartDate     string  `json:"start_date"`
	Notes         string  `json:"notes,omitempty"`
	PrimaryRoomID *string `json:"primary_room_id,omitempty"`
}

type registrationProfilePayload struct {
	DemographicsHome        *demographicsHomePayload  `json:"demographics_home,omitempty"`
	MedicalDietary          *medicalDietaryPayload    `json:"medical_dietary,omitempty"`
	HealthContacts          *healthContactsPayload    `json:"health_contacts,omitempty"`
	SocialDevelopment       *socialDevelopmentPayload `json:"social_development,omitempty"`
	ParentCarers            []contactEntryPayload     `json:"parent_carers,omitempty"`
	EmergencyContacts       []contactEntryPayload     `json:"emergency_contacts,omitempty"`
	AuthorisedCollectors    []contactEntryPayload     `json:"authorised_collectors,omitempty"`
	Collection              *collectionPayload        `json:"collection,omitempty"`
	FundingSupport          *fundingSupportPayload    `json:"funding_support,omitempty"`
	RoutineCare             *routineCarePayload       `json:"routine_care,omitempty"`
	GDPRDeclaration         *gdprDeclarationPayload   `json:"gdpr_declaration,omitempty"`
	RegistrationDate        *string                   `json:"registration_date,omitempty"`
}

type demographicsHomePayload struct {
	Sex                      *string `json:"sex,omitempty"`
	Religion                 *string `json:"religion,omitempty"`
	EthnicOrigin             *string `json:"ethnic_origin,omitempty"`
	FirstLanguage            *string `json:"first_language,omitempty"`
	OtherLanguages           *string `json:"other_languages,omitempty"`
	HomeAddress              any     `json:"home_address,omitempty"`
	HomePostcode             *string `json:"home_postcode,omitempty"`
	HomeTelephone            *string `json:"home_telephone,omitempty"`
	DisabilityStatus         *string `json:"disability_status,omitempty"`
	DisabilityNotes          *string `json:"disability_notes,omitempty"`
	AccessRequirements       *string `json:"access_requirements,omitempty"`
	DemographicsHomeReviewed bool    `json:"demographics_home_reviewed"`
}

type medicalDietaryPayload struct {
	MedicalConditionsStatus    *string `json:"medical_conditions_status,omitempty"`
	MedicalConditionsNotes     *string `json:"medical_conditions_notes,omitempty"`
	PrescribedMedicationStatus *string `json:"prescribed_medication_status,omitempty"`
	MedicationNotes            *string `json:"medication_notes,omitempty"`
	ImmunisationStatus         *string `json:"immunisation_status,omitempty"`
	ImmunisationCountry        *string `json:"immunisation_country,omitempty"`
	IllnessDiagnosisHistory    *string `json:"illness_diagnosis_history,omitempty"`
	DietaryRequirementsStatus  *string `json:"dietary_requirements_status,omitempty"`
	DietaryRequirementsNotes   *string `json:"dietary_requirements_notes,omitempty"`
	DietarySideEffects         *string `json:"dietary_side_effects,omitempty"`
	MedicalDietaryReviewed     bool    `json:"medical_dietary_reviewed"`
}

type healthContactsPayload struct {
	DoctorName             *string `json:"doctor_name,omitempty"`
	DoctorAddress          *string `json:"doctor_address,omitempty"`
	DoctorPhone            *string `json:"doctor_phone,omitempty"`
	HealthVisitorName      *string `json:"health_visitor_name,omitempty"`
	HealthVisitorAddress   *string `json:"health_visitor_address,omitempty"`
	HealthVisitorPhone     *string `json:"health_visitor_phone,omitempty"`
	HealthContactsReviewed bool    `json:"health_contacts_reviewed"`
}

type socialDevelopmentPayload struct {
	SocialServicesStatus      *string `json:"social_services_status,omitempty"`
	SocialServicesNotes       *string `json:"social_services_notes,omitempty"`
	SocialWorkerName          *string `json:"social_worker_name,omitempty"`
	SocialWorkerPhone         *string `json:"social_worker_phone,omitempty"`
	SocialWorkerEmail         *string `json:"social_worker_email,omitempty"`
	ConcernWalking            *string `json:"concern_walking,omitempty"`
	ConcernSpeechLanguage     *string `json:"concern_speech_language,omitempty"`
	ConcernHearing            *string `json:"concern_hearing,omitempty"`
	ConcernSight              *string `json:"concern_sight,omitempty"`
	ConcernEmotionalWellbeing *string `json:"concern_emotional_wellbeing,omitempty"`
	ConcernBehaviour          *string `json:"concern_behaviour,omitempty"`
	SocialDevelopmentReviewed bool    `json:"social_development_reviewed"`
}

type contactEntryPayload struct {
	FullName                  string  `json:"full_name"`
	RelationshipToChild       *string `json:"relationship_to_child,omitempty"`
	Address                   any     `json:"address,omitempty"`
	Telephone                 *string `json:"telephone,omitempty"`
	Email                     *string `json:"email,omitempty"`
	WorkAddress               any     `json:"work_address,omitempty"`
	HasParentalResponsibility *bool   `json:"has_parental_responsibility,omitempty"`
}

type collectionPayload struct {
	Over18CollectionAcknowledged bool `json:"over18_collection_acknowledged"`
	EmergencyCollectionReviewed  bool `json:"emergency_collection_reviewed"`
}

type fundingSupportPayload struct {
	BenefitsContributeToFees *string `json:"benefits_contribute_to_fees,omitempty"`
	WorkingTaxCredit         *string `json:"working_tax_credit,omitempty"`
	CollegeUniPaidToParent   *string `json:"college_uni_paid_to_parent,omitempty"`
	CollegeUniPaidToNursery  *string `json:"college_uni_paid_to_nursery,omitempty"`
	Funding3yoTermTime       *string `json:"funding_3yo_term_time,omitempty"`
	Funding2yoTermTime       *string `json:"funding_2yo_term_time,omitempty"`
	FundingSupportNotes      *string `json:"funding_support_notes,omitempty"`
	FundingSupportReviewed   bool    `json:"funding_support_reviewed"`
}

type routineCarePayload struct {
	RoutineCareNotes    *string `json:"routine_care_notes,omitempty"`
	RoutineCareReviewed bool    `json:"routine_care_reviewed"`
}

type gdprDeclarationPayload struct {
	GDPRDeclaredByName  *string `json:"gdpr_declared_by_name,omitempty"`
	GDPRDeclarationDate *string `json:"gdpr_declaration_date,omitempty"`
}

type consentPayload struct {
	UrgentMedicalTreatment               bool    `json:"urgent_medical_treatment"`
	UrgentMedicalTreatmentExceptions     *string `json:"urgent_medical_treatment_exceptions,omitempty"`
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
	SocialMediaChannelNotes              *string `json:"social_media_channel_notes,omitempty"`

	NotesExceptions *string `json:"notes_exceptions,omitempty"`
}

type childRecordResponse struct {
	ID         string  `json:"id"`
	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name"`
	LastName   *string `json:"last_name"`
	StartDate  string  `json:"start_date"`
}

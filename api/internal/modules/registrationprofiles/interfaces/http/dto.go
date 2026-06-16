package httpregistrationprofile

import (
	"time"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
)

type childSummaryResponse struct {
	ID          string  `json:"id"`
	FirstName   string  `json:"first_name"`
	MiddleName  *string `json:"middle_name"`
	LastName    *string `json:"last_name"`
	DateOfBirth string  `json:"date_of_birth"`
}

type profileResponse struct {
	ID        string `json:"id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type registrationProfileResponse struct {
	Child                  childSummaryResponse       `json:"child"`
	ProfileExists          bool                       `json:"profile_exists"`
	Profile                *profileResponse           `json:"profile,omitempty"`
	DemographicsHome       *demographicsHomeResponse  `json:"demographics_home,omitempty"`
	MedicalDietary         *medicalDietaryResponse    `json:"medical_dietary,omitempty"`
	HealthContacts         *healthContactsResponse    `json:"health_contacts,omitempty"`
	SocialDevelopment      *socialDevelopmentResponse `json:"social_development,omitempty"`
	ParentCarers           []contactEntryResponse     `json:"parent_carers,omitempty"`
	EmergencyContacts      []contactEntryResponse     `json:"emergency_contacts,omitempty"`
	AuthorisedCollectors   []contactEntryResponse     `json:"authorised_collectors,omitempty"`
	Collection             *collectionResponse        `json:"collection,omitempty"`
	FundingSupport         *fundingSupportResponse    `json:"funding_support,omitempty"`
	RoutineCare            *routineCareResponse       `json:"routine_care,omitempty"`
	GDPRDeclaration        *gdprDeclarationResponse   `json:"gdpr_declaration,omitempty"`
	RegistrationDate       *string                    `json:"registration_date,omitempty"`
	Completeness           *completenessResponse      `json:"completeness"`
}

type demographicsHomeResponse struct {
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

type medicalDietaryResponse struct {
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

type healthContactsResponse struct {
	DoctorName             *string `json:"doctor_name,omitempty"`
	DoctorAddress          *string `json:"doctor_address,omitempty"`
	DoctorPhone            *string `json:"doctor_phone,omitempty"`
	HealthVisitorName      *string `json:"health_visitor_name,omitempty"`
	HealthVisitorAddress   *string `json:"health_visitor_address,omitempty"`
	HealthVisitorPhone     *string `json:"health_visitor_phone,omitempty"`
	HealthContactsReviewed bool    `json:"health_contacts_reviewed"`
}

type socialDevelopmentResponse struct {
	SocialServicesStatus      *string                       `json:"social_services_status,omitempty"`
	SocialServicesNotes       *string                       `json:"social_services_notes,omitempty"`
	SocialWorkerName          *string                       `json:"social_worker_name,omitempty"`
	SocialWorkerPhone         *string                       `json:"social_worker_phone,omitempty"`
	SocialWorkerEmail         *string                       `json:"social_worker_email,omitempty"`
	ConcernWalking            *string                       `json:"concern_walking,omitempty"`
	ConcernSpeechLanguage     *string                       `json:"concern_speech_language,omitempty"`
	ConcernHearing            *string                       `json:"concern_hearing,omitempty"`
	ConcernSight              *string                       `json:"concern_sight,omitempty"`
	ConcernEmotionalWellbeing *string                       `json:"concern_emotional_wellbeing,omitempty"`
	ConcernBehaviour          *string                       `json:"concern_behaviour,omitempty"`
	ProfessionalReferrals     []domain.ProfessionalReferral `json:"professional_referrals,omitempty"`
	SocialDevelopmentReviewed bool                          `json:"social_development_reviewed"`
}

type contactEntryResponse struct {
	FullName                  string  `json:"full_name"`
	RelationshipToChild       *string `json:"relationship_to_child,omitempty"`
	Address                   any     `json:"address,omitempty"`
	Telephone                 *string `json:"telephone,omitempty"`
	Email                     *string `json:"email,omitempty"`
	WorkAddress               any     `json:"work_address,omitempty"`
	HasParentalResponsibility *bool   `json:"has_parental_responsibility,omitempty"`
}

type collectionResponse struct {
	IsSet                        bool    `json:"is_set"`
	UpdatedAt                    *string `json:"last_updated_at,omitempty"`
	UpdatedByUserID              *string `json:"last_updated_by_user_id,omitempty"`
	UpdatedByMembershipID        *string `json:"last_updated_by_membership_id,omitempty"`
	Over18CollectionAcknowledged bool    `json:"over18_collection_acknowledged"`
	EmergencyCollectionReviewed  bool    `json:"emergency_collection_reviewed"`
}

type fundingSupportResponse struct {
	BenefitsContributeToFees *string `json:"benefits_contribute_to_fees,omitempty"`
	WorkingTaxCredit         *string `json:"working_tax_credit,omitempty"`
	CollegeUniPaidToParent   *string `json:"college_uni_paid_to_parent,omitempty"`
	CollegeUniPaidToNursery  *string `json:"college_uni_paid_to_nursery,omitempty"`
	Funding3yoTermTime       *string `json:"funding_3yo_term_time,omitempty"`
	Funding2yoTermTime       *string `json:"funding_2yo_term_time,omitempty"`
	FundingSupportNotes      *string `json:"funding_support_notes,omitempty"`
	FundingSupportReviewed   bool    `json:"funding_support_reviewed"`
}

type routineCareResponse struct {
	RoutineCareNotes    *string `json:"routine_care_notes,omitempty"`
	RoutineCareReviewed bool    `json:"routine_care_reviewed"`
}

type gdprDeclarationResponse struct {
	GDPRDeclaredByName  *string `json:"gdpr_declared_by_name,omitempty"`
	GDPRDeclaredAt      *string `json:"gdpr_declared_at,omitempty"`
	GDPRDeclarationDate *string `json:"gdpr_declaration_date,omitempty"`
}

type completenessResponse struct {
	IsComplete      bool                          `json:"is_complete"`
	MissingSections []domain.SectionCode          `json:"missing_sections,omitempty"`
	Sections        []completenessSectionResponse `json:"sections"`
}

type completenessSectionResponse struct {
	Code          domain.SectionCode `json:"code"`
	Status        string             `json:"status"`
	MissingFields []string           `json:"missing_fields,omitempty"`
}

type collectionPasswordRequest struct {
	Password string `json:"password"`
}

func toRegistrationProfileResponse(pwc domain.ProfileWithChild, comp domain.Completeness) *registrationProfileResponse {
	resp := &registrationProfileResponse{
		Child: childSummaryResponse{
			ID:          pwc.Child.ID.String(),
			FirstName:   pwc.Child.FirstName,
			MiddleName:  pwc.Child.MiddleName,
			LastName:    pwc.Child.LastName,
			DateOfBirth: pwc.Child.DateOfBirth.Format("2006-01-02"),
		},
		ProfileExists: pwc.ProfileExists,
		Completeness:  toCompletenessResponse(comp),
	}

	if !pwc.ProfileExists || pwc.Profile == nil {
		return resp
	}

	p := pwc.Profile
	resp.Profile = &profileResponse{
		ID:        p.ID.String(),
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}

	resp.DemographicsHome = &demographicsHomeResponse{
		Sex:                      p.Sex,
		Religion:                 p.Religion,
		EthnicOrigin:             p.EthnicOrigin,
		FirstLanguage:            p.FirstLanguage,
		OtherLanguages:           p.OtherLanguages,
		HomeAddress:              p.HomeAddress,
		HomePostcode:             p.HomePostcode,
		HomeTelephone:            p.HomeTelephone,
		DisabilityStatus:         statusPtr(string(p.DisabilityStatus)),
		DisabilityNotes:          p.DisabilityNotes,
		AccessRequirements:       p.AccessRequirements,
		DemographicsHomeReviewed: p.DemographicsHomeReviewed,
	}

	resp.MedicalDietary = &medicalDietaryResponse{
		MedicalConditionsStatus:    statusPtr(string(p.MedicalConditionsStatus)),
		MedicalConditionsNotes:     p.MedicalConditionsNotes,
		PrescribedMedicationStatus: statusPtr(string(p.PrescribedMedicationStatus)),
		MedicationNotes:            p.MedicationNotes,
		ImmunisationStatus:         statusPtr(string(p.ImmunisationStatus)),
		ImmunisationCountry:        p.ImmunisationCountry,
		IllnessDiagnosisHistory:    p.IllnessDiagnosisHistory,
		DietaryRequirementsStatus:  statusPtr(string(p.DietaryRequirementsStatus)),
		DietaryRequirementsNotes:   p.DietaryRequirementsNotes,
		DietarySideEffects:         p.DietarySideEffects,
		MedicalDietaryReviewed:     p.MedicalDietaryReviewed,
	}

	resp.HealthContacts = &healthContactsResponse{
		DoctorName:             p.DoctorName,
		DoctorAddress:          p.DoctorAddress,
		DoctorPhone:            p.DoctorPhone,
		HealthVisitorName:      p.HealthVisitorName,
		HealthVisitorAddress:   p.HealthVisitorAddress,
		HealthVisitorPhone:     p.HealthVisitorPhone,
		HealthContactsReviewed: p.HealthContactsReviewed,
	}

	resp.SocialDevelopment = &socialDevelopmentResponse{
		SocialServicesStatus:      statusPtr(string(p.SocialServicesStatus)),
		SocialServicesNotes:       p.SocialServicesNotes,
		SocialWorkerName:          p.SocialWorkerName,
		SocialWorkerPhone:         p.SocialWorkerPhone,
		SocialWorkerEmail:         p.SocialWorkerEmail,
		ConcernWalking:            statusPtr(string(p.ConcernWalking)),
		ConcernSpeechLanguage:     statusPtr(string(p.ConcernSpeechLanguage)),
		ConcernHearing:            statusPtr(string(p.ConcernHearing)),
		ConcernSight:              statusPtr(string(p.ConcernSight)),
		ConcernEmotionalWellbeing: statusPtr(string(p.ConcernEmotionalWellbeing)),
		ConcernBehaviour:          statusPtr(string(p.ConcernBehaviour)),
		ProfessionalReferrals:     p.ProfessionalReferrals,
		SocialDevelopmentReviewed: p.SocialDevelopmentReviewed,
	}

	for _, c := range pwc.Contacts {
		entry := contactEntryResponse{
			FullName:                  c.FullName,
			RelationshipToChild:       c.RelationshipToChild,
			Address:                   c.Address,
			Telephone:                 c.Telephone,
			Email:                     c.Email,
			WorkAddress:               c.WorkAddress,
			HasParentalResponsibility: c.HasParentalResponsibility,
		}
		switch c.ContactType {
		case domain.ContactTypeParentCarer:
			resp.ParentCarers = append(resp.ParentCarers, entry)
		case domain.ContactTypeEmergencyContact:
			resp.EmergencyContacts = append(resp.EmergencyContacts, entry)
		case domain.ContactTypeAuthorisedCollector:
			resp.AuthorisedCollectors = append(resp.AuthorisedCollectors, entry)
		}
	}

	var updatedAtStr *string
	if p.CollectionPasswordUpdatedAt != nil {
		s := p.CollectionPasswordUpdatedAt.Format(time.RFC3339)
		updatedAtStr = &s
	}
	var updatedByUserIDStr *string
	if p.CollectionPasswordUpdatedByUserID != nil {
		s := p.CollectionPasswordUpdatedByUserID.String()
		updatedByUserIDStr = &s
	}
	var updatedByMembershipIDStr *string
	if p.CollectionPasswordUpdatedByMembershipID != nil {
		s := p.CollectionPasswordUpdatedByMembershipID.String()
		updatedByMembershipIDStr = &s
	}

	resp.Collection = &collectionResponse{
		IsSet:                        p.CollectionPasswordHash != nil,
		UpdatedAt:                    updatedAtStr,
		UpdatedByUserID:              updatedByUserIDStr,
		UpdatedByMembershipID:        updatedByMembershipIDStr,
		Over18CollectionAcknowledged: p.Over18CollectionAcknowledged,
		EmergencyCollectionReviewed:  p.EmergencyCollectionReviewed,
	}

	resp.FundingSupport = &fundingSupportResponse{
		BenefitsContributeToFees: statusPtr(string(p.BenefitsContributeToFees)),
		WorkingTaxCredit:         statusPtr(string(p.WorkingTaxCredit)),
		CollegeUniPaidToParent:   statusPtr(string(p.CollegeUniPaidToParent)),
		CollegeUniPaidToNursery:  statusPtr(string(p.CollegeUniPaidToNursery)),
		Funding3yoTermTime:       statusPtr(string(p.Funding3yoTermTime)),
		Funding2yoTermTime:       statusPtr(string(p.Funding2yoTermTime)),
		FundingSupportNotes:      p.FundingSupportNotes,
		FundingSupportReviewed:   p.FundingSupportReviewed,
	}

	resp.RoutineCare = &routineCareResponse{
		RoutineCareNotes:    p.RoutineCareNotes,
		RoutineCareReviewed: p.RoutineCareReviewed,
	}

	if p.GDPRDeclaredByName != nil || p.GDPRDeclaredAt != nil || p.GDPRDeclarationDate != nil {
		resp.GDPRDeclaration = &gdprDeclarationResponse{
			GDPRDeclaredByName: p.GDPRDeclaredByName,
		}
		if p.GDPRDeclaredAt != nil {
			s := p.GDPRDeclaredAt.Format(time.RFC3339)
			resp.GDPRDeclaration.GDPRDeclaredAt = &s
		}
		if p.GDPRDeclarationDate != nil {
			s := p.GDPRDeclarationDate.Format("2006-01-02")
			resp.GDPRDeclaration.GDPRDeclarationDate = &s
		}
	}

	if p.RegistrationDate != nil {
		s := p.RegistrationDate.Format("2006-01-02")
		resp.RegistrationDate = &s
	}

	return resp
}

func toCompletenessResponse(c domain.Completeness) *completenessResponse {
	sections := make([]completenessSectionResponse, len(c.Sections))
	for i, s := range c.Sections {
		sections[i] = completenessSectionResponse{
			Code:          s.Code,
			Status:        string(s.Status),
			MissingFields: s.MissingFields,
		}
	}
	return &completenessResponse{
		IsComplete:      c.IsComplete,
		MissingSections: c.MissingSections,
		Sections:        sections,
	}
}

func statusPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

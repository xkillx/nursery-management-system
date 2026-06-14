package domain

type SectionCode string

const (
	SectionDemographicsHome     SectionCode = "child_demographics_home"
	SectionMedicalDietary       SectionCode = "medical_dietary"
	SectionHealthContacts       SectionCode = "health_contacts"
	SectionSocialDevelopment    SectionCode = "social_development"
	SectionParentResponsibility SectionCode = "parent_responsibility"
	SectionEmergencyCollection  SectionCode = "emergency_collection"
	SectionFundingSupport       SectionCode = "funding_support"
	SectionRoutineCare          SectionCode = "routine_care"
	SectionGDPRDeclaration      SectionCode = "gdpr_declaration"
)

type CompletenessStatus string

const (
	StatusComplete   CompletenessStatus = "complete"
	StatusIncomplete CompletenessStatus = "incomplete"
)

type CompletenessSection struct {
	Code          SectionCode        `json:"code"`
	Status        CompletenessStatus `json:"status"`
	MissingFields []string           `json:"missing_fields,omitempty"`
}

type Completeness struct {
	IsComplete     bool                 `json:"is_complete"`
	MissingSections []SectionCode       `json:"missing_sections,omitempty"`
	Sections       []CompletenessSection `json:"sections"`
}

type completenessInput struct {
	profile           *Profile
	contacts          []ContactEntry
	parentCarers      int
	emergencyContacts int
	authorisedCollectors int
	passwordIsSet     bool
}

func ComputeCompleteness(p *Profile, contacts []ContactEntry, passwordIsSet bool) Completeness {
	in := completenessInput{
		profile:             p,
		contacts:            contacts,
		passwordIsSet:       passwordIsSet,
	}

	for _, c := range contacts {
		switch c.ContactType {
		case ContactTypeParentCarer:
			in.parentCarers++
		case ContactTypeEmergencyContact:
			in.emergencyContacts++
		case ContactTypeAuthorisedCollector:
			in.authorisedCollectors++
		}
	}

	sections := []CompletenessSection{
		computeDemographicsHome(in),
		computeMedicalDietary(in),
		computeHealthContacts(in),
		computeSocialDevelopment(in),
		computeParentResponsibility(in),
		computeEmergencyCollection(in),
		computeFundingSupport(in),
		computeRoutineCare(in),
		computeGDPRDeclaration(in),
	}

	missing := make([]SectionCode, 0)
	for _, s := range sections {
		if s.Status == StatusIncomplete {
			missing = append(missing, s.Code)
		}
	}

	return Completeness{
		IsComplete:      len(missing) == 0,
		MissingSections: missing,
		Sections:        sections,
	}
}

func computeDemographicsHome(in completenessInput) CompletenessSection {
	s := CompletenessSection{Code: SectionDemographicsHome}
	if in.profile.DemographicsHomeReviewed {
		s.Status = StatusComplete
	} else {
		s.Status = StatusIncomplete
		s.MissingFields = []string{"review_required"}
	}
	return s
}

func computeMedicalDietary(in completenessInput) CompletenessSection {
	s := CompletenessSection{Code: SectionMedicalDietary}
	missing := make([]string, 0)

	if !in.profile.MedicalDietaryReviewed {
		missing = append(missing, "review_required")
	}
	if in.profile.MedicalConditionsStatus == YesNoUnknownUnknown {
		missing = append(missing, "medical_conditions_status_unknown")
	}
	if in.profile.PrescribedMedicationStatus == YesNoUnknownUnknown {
		missing = append(missing, "prescribed_medication_status_unknown")
	}
	if in.profile.DietaryRequirementsStatus == YesNoUnknownUnknown {
		missing = append(missing, "dietary_requirements_status_unknown")
	}
	if in.profile.ImmunisationStatus == ImmunisationUnknown {
		missing = append(missing, "immunisation_status_unknown")
	}

	if in.profile.MedicalConditionsStatus == YesNoUnknownYes && (in.profile.MedicalConditionsNotes == nil || *in.profile.MedicalConditionsNotes == "") {
		missing = append(missing, "medical_conditions_notes_required")
	}
	if in.profile.PrescribedMedicationStatus == YesNoUnknownYes && (in.profile.MedicationNotes == nil || *in.profile.MedicationNotes == "") {
		missing = append(missing, "medication_notes_required")
	}
	if in.profile.DietaryRequirementsStatus == YesNoUnknownYes && (in.profile.DietaryRequirementsNotes == nil || *in.profile.DietaryRequirementsNotes == "") {
		missing = append(missing, "dietary_requirements_notes_required")
	}

	if len(missing) == 0 {
		s.Status = StatusComplete
	} else {
		s.Status = StatusIncomplete
		s.MissingFields = missing
	}
	return s
}

func computeHealthContacts(in completenessInput) CompletenessSection {
	s := CompletenessSection{Code: SectionHealthContacts}
	if in.profile.HealthContactsReviewed {
		s.Status = StatusComplete
	} else {
		s.Status = StatusIncomplete
		s.MissingFields = []string{"review_required"}
	}
	return s
}

func computeSocialDevelopment(in completenessInput) CompletenessSection {
	s := CompletenessSection{Code: SectionSocialDevelopment}
	missing := make([]string, 0)

	if !in.profile.SocialDevelopmentReviewed {
		missing = append(missing, "review_required")
	}
	if in.profile.SocialServicesStatus == YesNoUnknownUnknown {
		missing = append(missing, "social_services_status_unknown")
	}
	if in.profile.SocialServicesStatus == YesNoUnknownYes {
		hasWorkerContact := (in.profile.SocialWorkerName != nil && *in.profile.SocialWorkerName != "") ||
			(in.profile.SocialWorkerPhone != nil && *in.profile.SocialWorkerPhone != "") ||
			(in.profile.SocialWorkerEmail != nil && *in.profile.SocialWorkerEmail != "")
		if (in.profile.SocialServicesNotes == nil || *in.profile.SocialServicesNotes == "") &&
			!hasWorkerContact {
			missing = append(missing, "social_services_notes_or_worker_required")
		}
	}
	if in.profile.ConcernWalking == YesNoUnknownUnknown {
		missing = append(missing, "concern_walking_unknown")
	}
	if in.profile.ConcernSpeechLanguage == YesNoUnknownUnknown {
		missing = append(missing, "concern_speech_language_unknown")
	}
	if in.profile.ConcernHearing == YesNoUnknownUnknown {
		missing = append(missing, "concern_hearing_unknown")
	}
	if in.profile.ConcernSight == YesNoUnknownUnknown {
		missing = append(missing, "concern_sight_unknown")
	}
	if in.profile.ConcernEmotionalWellbeing == YesNoUnknownUnknown {
		missing = append(missing, "concern_emotional_wellbeing_unknown")
	}
	if in.profile.ConcernBehaviour == YesNoUnknownUnknown {
		missing = append(missing, "concern_behaviour_unknown")
	}

	if len(missing) == 0 {
		s.Status = StatusComplete
	} else {
		s.Status = StatusIncomplete
		s.MissingFields = missing
	}
	return s
}

func computeParentResponsibility(in completenessInput) CompletenessSection {
	s := CompletenessSection{Code: SectionParentResponsibility}
	missing := make([]string, 0)

	if !in.profile.ParentResponsibilityReviewed {
		missing = append(missing, "review_required")
	}
	if in.parentCarers == 0 && (in.profile.ParentalResponsibilityNotes == nil || *in.profile.ParentalResponsibilityNotes == "") {
		missing = append(missing, "parent_carer_or_notes_required")
	}

	if len(missing) == 0 {
		s.Status = StatusComplete
	} else {
		s.Status = StatusIncomplete
		s.MissingFields = missing
	}
	return s
}

func computeEmergencyCollection(in completenessInput) CompletenessSection {
	s := CompletenessSection{Code: SectionEmergencyCollection}
	missing := make([]string, 0)

	if !in.profile.EmergencyCollectionReviewed {
		missing = append(missing, "review_required")
	}
	if in.emergencyContacts == 0 {
		missing = append(missing, "emergency_contact_missing")
	}
	if in.authorisedCollectors == 0 {
		missing = append(missing, "authorised_collector_missing")
	}
	if !in.profile.Over18CollectionAcknowledged {
		missing = append(missing, "over18_collection_acknowledgement_required")
	}
	if !in.passwordIsSet {
		missing = append(missing, "collection_password_missing")
	}

	if len(missing) == 0 {
		s.Status = StatusComplete
	} else {
		s.Status = StatusIncomplete
		s.MissingFields = missing
	}
	return s
}

func computeFundingSupport(in completenessInput) CompletenessSection {
	s := CompletenessSection{Code: SectionFundingSupport}
	missing := make([]string, 0)

	if !in.profile.FundingSupportReviewed {
		missing = append(missing, "review_required")
	}
	if in.profile.BenefitsContributeToFees == YesNoUnknownUnknown {
		missing = append(missing, "benefits_contribute_to_fees_unknown")
	}
	if in.profile.WorkingTaxCredit == YesNoUnknownUnknown {
		missing = append(missing, "working_tax_credit_unknown")
	}
	if in.profile.CollegeUniPaidToParent == YesNoUnknownUnknown {
		missing = append(missing, "college_uni_paid_to_parent_unknown")
	}
	if in.profile.CollegeUniPaidToNursery == YesNoUnknownUnknown {
		missing = append(missing, "college_uni_paid_to_nursery_unknown")
	}
	if in.profile.Funding3yoTermTime == YesNoUnknownUnknown {
		missing = append(missing, "funding_3yo_term_time_unknown")
	}
	if in.profile.Funding2yoTermTime == YesNoUnknownUnknown {
		missing = append(missing, "funding_2yo_term_time_unknown")
	}

	if in.profile.FundingSupportNotes != nil && *in.profile.FundingSupportNotes != "" {
		missing = nil
	}

	if len(missing) == 0 {
		s.Status = StatusComplete
	} else {
		s.Status = StatusIncomplete
		s.MissingFields = missing
	}
	return s
}

func computeRoutineCare(in completenessInput) CompletenessSection {
	s := CompletenessSection{Code: SectionRoutineCare}
	if in.profile.RoutineCareReviewed {
		s.Status = StatusComplete
	} else {
		s.Status = StatusIncomplete
		s.MissingFields = []string{"review_required"}
	}
	return s
}

func computeGDPRDeclaration(in completenessInput) CompletenessSection {
	s := CompletenessSection{Code: SectionGDPRDeclaration}
	missing := make([]string, 0)

	if in.profile.GDPRDeclaredByName == nil || *in.profile.GDPRDeclaredByName == "" {
		missing = append(missing, "gdpr_declared_by_name_missing")
	}
	if in.profile.GDPRDeclaredAt == nil {
		missing = append(missing, "gdpr_declared_at_missing")
	}
	if in.profile.GDPRDeclarationDate == nil {
		missing = append(missing, "gdpr_declaration_date_missing")
	}

	if len(missing) == 0 {
		s.Status = StatusComplete
	} else {
		s.Status = StatusIncomplete
		s.MissingFields = missing
	}
	return s
}

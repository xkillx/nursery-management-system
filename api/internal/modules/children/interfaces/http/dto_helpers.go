package httpchild

import (
	"time"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/children/application"
	"nursery-management-system/api/internal/modules/children/domain"
)

func timeStringPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

func dateStringPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

type childContactsRequest struct {
	ParentCarers         []contactPayload `json:"parent_carers"`
	EmergencyContacts    []contactPayload `json:"emergency_contacts"`
	AuthorisedCollectors []contactPayload `json:"authorised_collectors"`
}

type contactPayload struct {
	FullName                  string         `json:"full_name"`
	RelationshipToChild       *string        `json:"relationship_to_child"`
	Address                   map[string]any `json:"address"`
	Telephone                 *string        `json:"telephone"`
	Email                     *string        `json:"email"`
	WorkAddress               map[string]any `json:"work_address"`
	HasParentalResponsibility *bool          `json:"has_parental_responsibility"`
}

type contactResponse struct {
	ID                        string         `json:"id"`
	ContactType               string         `json:"contact_type"`
	SortOrder                 int            `json:"sort_order"`
	FullName                  string         `json:"full_name"`
	RelationshipToChild       *string        `json:"relationship_to_child,omitempty"`
	Address                   map[string]any `json:"address"`
	Telephone                 *string        `json:"telephone,omitempty"`
	Email                     *string        `json:"email,omitempty"`
	WorkAddress               map[string]any `json:"work_address"`
	HasParentalResponsibility *bool          `json:"has_parental_responsibility"`
	CreatedAt                 string         `json:"created_at"`
	UpdatedAt                 string         `json:"updated_at"`
}

func mapChildContactsRequest(req childContactsRequest) []application.ChildContactInput {
	out := make([]application.ChildContactInput, 0, len(req.ParentCarers)+len(req.EmergencyContacts)+len(req.AuthorisedCollectors))
	for _, p := range req.ParentCarers {
		out = append(out, application.ChildContactInput{
			ContactType: domain.ContactTypeParentCarer, FullName: p.FullName,
			RelationshipToChild: p.RelationshipToChild, Address: p.Address,
			Telephone: p.Telephone, Email: p.Email, WorkAddress: p.WorkAddress,
			HasParentalResponsibility: p.HasParentalResponsibility,
		})
	}
	for _, p := range req.EmergencyContacts {
		out = append(out, application.ChildContactInput{
			ContactType: domain.ContactTypeEmergencyContact, FullName: p.FullName,
			RelationshipToChild: p.RelationshipToChild, Address: p.Address,
			Telephone: p.Telephone, Email: p.Email, WorkAddress: p.WorkAddress,
			HasParentalResponsibility: p.HasParentalResponsibility,
		})
	}
	for _, p := range req.AuthorisedCollectors {
		out = append(out, application.ChildContactInput{
			ContactType: domain.ContactTypeAuthorisedCollector, FullName: p.FullName,
			RelationshipToChild: p.RelationshipToChild, Address: p.Address,
			Telephone: p.Telephone, Email: p.Email, WorkAddress: p.WorkAddress,
			HasParentalResponsibility: p.HasParentalResponsibility,
		})
	}
	return out
}

func toChildContactsResponse(contacts []domain.ChildContact) gin.H {
	parentCarers := make([]contactResponse, 0)
	emergencyContacts := make([]contactResponse, 0)
	authorisedCollectors := make([]contactResponse, 0)
	for _, c := range contacts {
		cr := contactResponse{
			ID:                        c.ID.String(),
			ContactType:               string(c.ContactType),
			SortOrder:                 c.SortOrder,
			FullName:                  c.FullName,
			RelationshipToChild:       c.RelationshipToChild,
			Address:                   c.Address,
			Telephone:                 c.Telephone,
			Email:                     c.Email,
			WorkAddress:               c.WorkAddress,
			HasParentalResponsibility: c.HasParentalResponsibility,
			CreatedAt:                 c.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:                 c.UpdatedAt.UTC().Format(time.RFC3339),
		}
		switch c.ContactType {
		case domain.ContactTypeParentCarer:
			parentCarers = append(parentCarers, cr)
		case domain.ContactTypeEmergencyContact:
			emergencyContacts = append(emergencyContacts, cr)
		case domain.ContactTypeAuthorisedCollector:
			authorisedCollectors = append(authorisedCollectors, cr)
		}
	}
	return gin.H{
		"parent_carers":         parentCarers,
		"emergency_contacts":    emergencyContacts,
		"authorised_collectors": authorisedCollectors,
	}
}

type childHealthRequest struct {
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

type childHealthResponse struct {
	ID                         string  `json:"id"`
	ChildID                    string  `json:"child_id"`
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
	CreatedAt                  string  `json:"created_at"`
	UpdatedAt                  string  `json:"updated_at"`
}

func mapChildHealthRequest(req childHealthRequest) *application.ChildHealthProfileInput {
	return &application.ChildHealthProfileInput{
		MedicalConditionsStatus:    req.MedicalConditionsStatus,
		MedicalConditionsNotes:     req.MedicalConditionsNotes,
		PrescribedMedicationStatus: req.PrescribedMedicationStatus,
		MedicationNotes:            req.MedicationNotes,
		ImmunisationStatus:         req.ImmunisationStatus,
		ImmunisationCountry:        req.ImmunisationCountry,
		IllnessDiagnosisHistory:    req.IllnessDiagnosisHistory,
		DietaryRequirementsStatus:  req.DietaryRequirementsStatus,
		DietaryRequirementsNotes:   req.DietaryRequirementsNotes,
		DietarySideEffects:         req.DietarySideEffects,
		DoctorName:                 req.DoctorName,
		DoctorAddress:              req.DoctorAddress,
		DoctorPhone:                req.DoctorPhone,
		HealthVisitorName:          req.HealthVisitorName,
		HealthVisitorAddress:       req.HealthVisitorAddress,
		HealthVisitorPhone:         req.HealthVisitorPhone,
	}
}

func toChildHealthResponse(p *domain.ChildHealthProfile) gin.H {
	if p == nil {
		return gin.H{"health": nil}
	}
	return gin.H{"health": childHealthResponse{
		ID: p.ID.String(), ChildID: p.ChildID.String(),
		MedicalConditionsStatus:    string(p.MedicalConditionsStatus),
		MedicalConditionsNotes:     p.MedicalConditionsNotes,
		PrescribedMedicationStatus: string(p.PrescribedMedicationStatus),
		MedicationNotes:            p.MedicationNotes,
		ImmunisationStatus:         string(p.ImmunisationStatus),
		ImmunisationCountry:        p.ImmunisationCountry,
		IllnessDiagnosisHistory:    p.IllnessDiagnosisHistory,
		DietaryRequirementsStatus:  string(p.DietaryRequirementsStatus),
		DietaryRequirementsNotes:   p.DietaryRequirementsNotes,
		DietarySideEffects:         p.DietarySideEffects,
		DoctorName:                 p.DoctorName, DoctorAddress: p.DoctorAddress, DoctorPhone: p.DoctorPhone,
		HealthVisitorName: p.HealthVisitorName, HealthVisitorAddress: p.HealthVisitorAddress, HealthVisitorPhone: p.HealthVisitorPhone,
		CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.UTC().Format(time.RFC3339),
	}}
}

type childSafeguardingRequest struct {
	SocialServicesStatus      string                        `json:"social_services_status"`
	SocialServicesNotes       *string                       `json:"social_services_notes"`
	SocialWorkerName          *string                       `json:"social_worker_name"`
	SocialWorkerPhone         *string                       `json:"social_worker_phone"`
	SocialWorkerEmail         *string                       `json:"social_worker_email"`
	ConcernWalking            string                        `json:"concern_walking"`
	ConcernSpeechLanguage     string                        `json:"concern_speech_language"`
	ConcernHearing            string                        `json:"concern_hearing"`
	ConcernSight              string                        `json:"concern_sight"`
	ConcernEmotionalWellbeing string                        `json:"concern_emotional_wellbeing"`
	ConcernBehaviour          string                        `json:"concern_behaviour"`
	ProfessionalReferrals     []domain.ProfessionalReferral `json:"professional_referrals"`
	RestrictedNotes           *string                       `json:"restricted_notes"`
}

type childSafeguardingResponse struct {
	ID                        string                        `json:"id"`
	ChildID                   string                        `json:"child_id"`
	SocialServicesStatus      string                        `json:"social_services_status"`
	SocialServicesNotes       *string                       `json:"social_services_notes"`
	SocialWorkerName          *string                       `json:"social_worker_name"`
	SocialWorkerPhone         *string                       `json:"social_worker_phone"`
	SocialWorkerEmail         *string                       `json:"social_worker_email"`
	ConcernWalking            string                        `json:"concern_walking"`
	ConcernSpeechLanguage     string                        `json:"concern_speech_language"`
	ConcernHearing            string                        `json:"concern_hearing"`
	ConcernSight              string                        `json:"concern_sight"`
	ConcernEmotionalWellbeing string                        `json:"concern_emotional_wellbeing"`
	ConcernBehaviour          string                        `json:"concern_behaviour"`
	ProfessionalReferrals     []domain.ProfessionalReferral `json:"professional_referrals"`
	RestrictedNotes           *string                       `json:"restricted_notes"`
	CreatedAt                 string                        `json:"created_at"`
	UpdatedAt                 string                        `json:"updated_at"`
}

func mapChildSafeguardingRequest(req childSafeguardingRequest) *application.ChildSafeguardingProfileInput {
	return &application.ChildSafeguardingProfileInput{
		SocialServicesStatus:      req.SocialServicesStatus,
		SocialServicesNotes:       req.SocialServicesNotes,
		SocialWorkerName:          req.SocialWorkerName,
		SocialWorkerPhone:         req.SocialWorkerPhone,
		SocialWorkerEmail:         req.SocialWorkerEmail,
		ConcernWalking:            req.ConcernWalking,
		ConcernSpeechLanguage:     req.ConcernSpeechLanguage,
		ConcernHearing:            req.ConcernHearing,
		ConcernSight:              req.ConcernSight,
		ConcernEmotionalWellbeing: req.ConcernEmotionalWellbeing,
		ConcernBehaviour:          req.ConcernBehaviour,
		ProfessionalReferrals:     req.ProfessionalReferrals,
		RestrictedNotes:           req.RestrictedNotes,
	}
}

func toChildSafeguardingResponse(p *domain.ChildSafeguardingProfile) gin.H {
	if p == nil {
		return gin.H{"safeguarding": nil}
	}
	return gin.H{"safeguarding": childSafeguardingResponse{
		ID: p.ID.String(), ChildID: p.ChildID.String(),
		SocialServicesStatus:      string(p.SocialServicesStatus),
		SocialServicesNotes:       p.SocialServicesNotes,
		SocialWorkerName:          p.SocialWorkerName,
		SocialWorkerPhone:         p.SocialWorkerPhone,
		SocialWorkerEmail:         p.SocialWorkerEmail,
		ConcernWalking:            string(p.ConcernWalking),
		ConcernSpeechLanguage:     string(p.ConcernSpeechLanguage),
		ConcernHearing:            string(p.ConcernHearing),
		ConcernSight:              string(p.ConcernSight),
		ConcernEmotionalWellbeing: string(p.ConcernEmotionalWellbeing),
		ConcernBehaviour:          string(p.ConcernBehaviour),
		ProfessionalReferrals:     p.ProfessionalReferrals,
		RestrictedNotes:           p.RestrictedNotes,
		CreatedAt:                 p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                 p.UpdatedAt.UTC().Format(time.RFC3339),
	}}
}

type childConsentRequest struct {
	UrgentMedicalTreatment               bool    `json:"urgent_medical_treatment"`
	UrgentMedicalTreatmentExceptions     *string `json:"urgent_medical_treatment_exceptions"`
	Plasters                             bool    `json:"plasters"`
	SafeguardingReportingAcknowledgement bool    `json:"safeguarding_reporting_acknowledgement"`
	InformationSharingConsent            bool    `json:"information_sharing_consent"`
	GDPRDataProcessingConsent            bool    `json:"gdpr_data_processing_consent"`
	InformationTruthfulnessDeclaration   bool    `json:"information_truthfulness_declaration"`
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

type childConsentResponse struct {
	ID                                   string  `json:"id"`
	ChildID                              string  `json:"child_id"`
	UrgentMedicalTreatment               bool    `json:"urgent_medical_treatment"`
	UrgentMedicalTreatmentExceptions     *string `json:"urgent_medical_treatment_exceptions"`
	Plasters                             bool    `json:"plasters"`
	SafeguardingReportingAcknowledgement bool    `json:"safeguarding_reporting_acknowledgement"`
	InformationSharingConsent            bool    `json:"information_sharing_consent"`
	GDPRDataProcessingConsent            bool    `json:"gdpr_data_processing_consent"`
	InformationTruthfulnessDeclaration   bool    `json:"information_truthfulness_declaration"`
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
	CreatedAt                            string  `json:"created_at"`
	UpdatedAt                            string  `json:"updated_at"`
}

func mapChildConsentRequest(req childConsentRequest) *application.ChildConsentInput {
	return &application.ChildConsentInput{
		UrgentMedicalTreatment:               req.UrgentMedicalTreatment,
		UrgentMedicalTreatmentExceptions:     req.UrgentMedicalTreatmentExceptions,
		Plasters:                             req.Plasters,
		SafeguardingReportingAcknowledgement: req.SafeguardingReportingAcknowledgement,
		InformationSharingConsent:            req.InformationSharingConsent,
		GDPRDataProcessingConsent:            req.GDPRDataProcessingConsent,
		InformationTruthfulnessDeclaration:   req.InformationTruthfulnessDeclaration,
		AreaSENCOLiaison:                     req.AreaSENCOLiaison,
		HealthVisitorLiaison:                 req.HealthVisitorLiaison,
		TransitionDocuments:                  req.TransitionDocuments,
		LocalOutings:                         req.LocalOutings,
		FacePainting:                         req.FacePainting,
		ParentSuppliedSunCream:               req.ParentSuppliedSunCream,
		ParentSuppliedNappyCream:             req.ParentSuppliedNappyCream,
		DevelopmentProfilePhotos:             req.DevelopmentProfilePhotos,
		NurseryDisplayBoards:                 req.NurseryDisplayBoards,
		PromotionalLiterature:                req.PromotionalLiterature,
		NurseryWebsite:                       req.NurseryWebsite,
		StaffStudentCoursework:               req.StaffStudentCoursework,
		SocialMedia:                          req.SocialMedia,
		SocialMediaChannelNotes:              req.SocialMediaChannelNotes,
		NotesExceptions:                      req.NotesExceptions,
		SignerName:                           req.SignerName,
		SignedDate:                           req.SignedDate,
		PaperFormOnFile:                      req.PaperFormOnFile,
	}
}

func toChildConsentResponse(p *domain.ChildConsent) childConsentResponse {
	return childConsentResponse{
		ID: p.ID.String(), ChildID: p.ChildID.String(),
		UrgentMedicalTreatment:               p.UrgentMedicalTreatment,
		UrgentMedicalTreatmentExceptions:     p.UrgentMedicalTreatmentExceptions,
		Plasters:                             p.Plasters,
		SafeguardingReportingAcknowledgement: p.SafeguardingReportingAcknowledgement,
		InformationSharingConsent:            p.InformationSharingConsent,
		GDPRDataProcessingConsent:            p.GDPRDataProcessingConsent,
		InformationTruthfulnessDeclaration:   p.InformationTruthfulnessDeclaration,
		AreaSENCOLiaison:                     p.AreaSENCOLiaison,
		HealthVisitorLiaison:                 p.HealthVisitorLiaison,
		TransitionDocuments:                  p.TransitionDocuments,
		LocalOutings:                         p.LocalOutings,
		FacePainting:                         p.FacePainting,
		ParentSuppliedSunCream:               p.ParentSuppliedSunCream,
		ParentSuppliedNappyCream:             p.ParentSuppliedNappyCream,
		DevelopmentProfilePhotos:             p.DevelopmentProfilePhotos,
		NurseryDisplayBoards:                 p.NurseryDisplayBoards,
		PromotionalLiterature:                p.PromotionalLiterature,
		NurseryWebsite:                       p.NurseryWebsite,
		StaffStudentCoursework:               p.StaffStudentCoursework,
		SocialMedia:                          p.SocialMedia,
		SocialMediaChannelNotes:              p.SocialMediaChannelNotes,
		NotesExceptions:                      p.NotesExceptions,
		SignerName:                           p.SignerName,
		SignedDate:                           p.SignedDate.Format("2006-01-02"),
		PaperFormOnFile:                      p.PaperFormOnFile,
		CreatedAt:                            p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                            p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

type childFundingRequest struct {
	FundingEnabled           bool     `json:"funding_enabled"`
	FundingType              string   `json:"funding_type"`
	FundingModel             string   `json:"funding_model"`
	FundedHoursPerWeek       *float64 `json:"funded_hours_per_week"`
	FundingStartDate         *string  `json:"funding_start_date"`
	FundingEndDate           *string  `json:"funding_end_date"`
	EligibilityCode          *string  `json:"eligibility_code"`
	EligibilityCodeValidated bool     `json:"eligibility_code_validated"`
	EvidenceReceived         bool     `json:"evidence_received"`
	BenefitsStatus           string   `json:"benefits_status"`
	Benefits                 []string `json:"benefits"`
	OtherBenefitName         *string  `json:"other_benefit_name"`
	BenefitNotes             *string  `json:"benefit_notes"`
	ManagerNotes             *string  `json:"manager_notes"`
}

type childFundingResponse struct {
	ID                       string   `json:"id"`
	ChildID                  string   `json:"child_id"`
	FundingEnabled           bool     `json:"funding_enabled"`
	FundingType              string   `json:"funding_type"`
	FundingModel             string   `json:"funding_model"`
	FundedHoursPerWeek       *float64 `json:"funded_hours_per_week"`
	FundingStartDate         *string  `json:"funding_start_date"`
	FundingEndDate           *string  `json:"funding_end_date"`
	EligibilityCode          *string  `json:"eligibility_code"`
	EligibilityCodeValidated bool     `json:"eligibility_code_validated"`
	EvidenceReceived         bool     `json:"evidence_received"`
	BenefitsStatus           string   `json:"benefits_status"`
	Benefits                 []string `json:"benefits"`
	OtherBenefitName         *string  `json:"other_benefit_name"`
	BenefitNotes             *string  `json:"benefit_notes"`
	ManagerNotes             *string  `json:"manager_notes"`
	CreatedAt                string   `json:"created_at"`
	UpdatedAt                string   `json:"updated_at"`
}

func mapChildFundingRequest(req childFundingRequest) *application.ChildFundingRecordInput {
	return &application.ChildFundingRecordInput{
		FundingEnabled:           req.FundingEnabled,
		FundingType:              req.FundingType,
		FundingModel:             req.FundingModel,
		FundedHoursPerWeek:       req.FundedHoursPerWeek,
		FundingStartDate:         req.FundingStartDate,
		FundingEndDate:           req.FundingEndDate,
		EligibilityCode:          req.EligibilityCode,
		EligibilityCodeValidated: req.EligibilityCodeValidated,
		EvidenceReceived:         req.EvidenceReceived,
		BenefitsStatus:           req.BenefitsStatus,
		Benefits:                 req.Benefits,
		OtherBenefitName:         req.OtherBenefitName,
		BenefitNotes:             req.BenefitNotes,
		ManagerNotes:             req.ManagerNotes,
	}
}

func toChildFundingResponse(p *domain.ChildFundingRecord) childFundingResponse {
	var startDate, endDate *string
	if p.FundingStartDate != nil {
		s := p.FundingStartDate.Format("2006-01-02")
		startDate = &s
	}
	if p.FundingEndDate != nil {
		s := p.FundingEndDate.Format("2006-01-02")
		endDate = &s
	}
	benefits := make([]string, 0, 6)
	if p.BenefitUniversalCredit {
		benefits = append(benefits, "universal_credit")
	}
	if p.BenefitIncomeSupport {
		benefits = append(benefits, "income_support")
	}
	if p.BenefitJobseekersAllowance {
		benefits = append(benefits, "jobseekers_allowance")
	}
	if p.BenefitESAIncomeRelated {
		benefits = append(benefits, "esa_income_related")
	}
	if p.BenefitChildTaxCredit {
		benefits = append(benefits, "child_tax_credit")
	}
	if p.BenefitOtherSupport {
		benefits = append(benefits, "other_support")
	}
	return childFundingResponse{
		ID:                       p.ID.String(),
		ChildID:                  p.ChildID.String(),
		FundingEnabled:           p.FundingEnabled,
		FundingType:              string(p.FundingType),
		FundingModel:             string(p.FundingModel),
		FundedHoursPerWeek:       p.FundedHoursPerWeek,
		FundingStartDate:         startDate,
		FundingEndDate:           endDate,
		EligibilityCode:          p.EligibilityCode,
		EligibilityCodeValidated: p.EligibilityCodeValidated,
		EvidenceReceived:         p.EvidenceReceived,
		BenefitsStatus:           string(p.BenefitsStatus),
		Benefits:                 benefits,
		OtherBenefitName:         p.OtherBenefitName,
		BenefitNotes:             p.BenefitNotes,
		ManagerNotes:             p.ManagerNotes,
		CreatedAt:                p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

type childCollectionSettingsRequest struct {
	Password                     *string `json:"password"`
	PasswordHint                 *string `json:"password_hint"`
	Over18CollectionAcknowledged *bool   `json:"over_18_collection_acknowledged"`
}

type childCollectionSettingsResponse struct {
	ID                                      string  `json:"id"`
	ChildID                                 string  `json:"child_id"`
	Over18CollectionAcknowledged            bool    `json:"over_18_collection_acknowledged"`
	CollectionPasswordSet                   bool    `json:"collection_password_set"`
	CollectionPassword                      string  `json:"collection_password"`
	CollectionPasswordHint                  string  `json:"collection_password_hint"`
	CollectionPasswordUpdatedAt             *string `json:"collection_password_updated_at,omitempty"`
	CollectionPasswordUpdatedByUserID       *string `json:"collection_password_updated_by_user_id,omitempty"`
	CollectionPasswordUpdatedByMembershipID *string `json:"collection_password_updated_by_membership_id,omitempty"`
	CreatedAt                               string  `json:"created_at"`
	UpdatedAt                               string  `json:"updated_at"`
}

func toChildCollectionSettingsResponse(p *domain.ChildCollectionSetting) childCollectionSettingsResponse {
	resp := childCollectionSettingsResponse{
		ID: p.ID.String(), ChildID: p.ChildID.String(),
		Over18CollectionAcknowledged:            p.Over18CollectionAcknowledged,
		CollectionPasswordSet:                   p.CollectionPassword != "",
		CollectionPassword:                      p.CollectionPassword,
		CollectionPasswordHint:                  p.CollectionPasswordHint,
		CollectionPasswordUpdatedAt:             timeStringPtr(p.CollectionPasswordUpdatedAt),
		CollectionPasswordUpdatedByUserID:       uuidStringPtr(p.CollectionPasswordUpdatedByUserID),
		CollectionPasswordUpdatedByMembershipID: uuidStringPtr(p.CollectionPasswordUpdatedByMembershipID),
		CreatedAt:                               p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:                               p.UpdatedAt.UTC().Format(time.RFC3339),
	}
	return resp
}

func uuidStringPtr(u interface{ String() string }) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}

type roomAssignmentRequest struct {
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
}

type roomAssignmentResponse struct {
	ID        string  `json:"id"`
	ChildID   string  `json:"child_id"`
	RoomID    string  `json:"room_id"`
	StartDate string  `json:"start_date"`
	EndDate   *string `json:"end_date,omitempty"`
	IsCurrent bool    `json:"is_current"`
	CreatedAt string  `json:"created_at"`
}

func toRoomAssignmentResponse(a domain.ChildRoomAssignment) roomAssignmentResponse {
	return roomAssignmentResponse{
		ID:        a.ID.String(),
		ChildID:   a.ChildID.String(),
		RoomID:    a.RoomID.String(),
		StartDate: a.StartDate.Format("2006-01-02"),
		EndDate:   dateStringPtr(a.EndDate),
		IsCurrent: a.IsCurrent,
		CreatedAt: a.CreatedAt.UTC().Format(time.RFC3339),
	}
}

type childBillingProfileRequest struct {
	BillingBasis    string  `json:"billing_basis"`
	CustomRateMinor *int    `json:"custom_rate_minor"`
	EffectiveFrom   *string `json:"effective_from"`
}

type childBillingProfileResponse struct {
	ID              string `json:"id"`
	ChildID         string `json:"child_id"`
	BillingBasis    string `json:"billing_basis"`
	CustomRateMinor *int   `json:"custom_rate_minor"`
	EffectiveFrom   string `json:"effective_from"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func toChildBillingProfileResponse(p *domain.ChildBillingProfile) childBillingProfileResponse {
	return childBillingProfileResponse{
		ID: p.ID.String(), ChildID: p.ChildID.String(),
		BillingBasis:    string(p.BillingBasis),
		CustomRateMinor: p.CustomRateMinor,
		EffectiveFrom:   p.EffectiveFrom.Format("2006-01-02"),
		CreatedAt:       p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

type childLeavingRecordResponse struct {
	ID         string  `json:"id"`
	ChildID    string  `json:"child_id"`
	LeftAt     string  `json:"left_at"`
	ReasonCode string  `json:"reason_code"`
	ReasonNote *string `json:"reason_note"`
	CreatedAt  string  `json:"created_at"`
}

func toChildLeavingRecordResponse(p *domain.ChildLeavingRecord) childLeavingRecordResponse {
	return childLeavingRecordResponse{
		ID: p.ID.String(), ChildID: p.ChildID.String(),
		LeftAt:     p.LeftAt.UTC().Format(time.RFC3339),
		ReasonCode: p.ReasonCode,
		ReasonNote: p.ReasonNote,
		CreatedAt:  p.CreatedAt.UTC().Format(time.RFC3339),
	}
}

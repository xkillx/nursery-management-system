package application

import (
	"strings"
	"time"

	"github.com/google/uuid"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/uid"
)

type PatchSection struct {
	DemographicsHome     *DemographicsHomePatch  `json:"demographics_home,omitempty"`
	MedicalDietary       *MedicalDietaryPatch    `json:"medical_dietary,omitempty"`
	HealthContacts       *HealthContactsPatch    `json:"health_contacts,omitempty"`
	SocialDevelopment    *SocialDevelopmentPatch `json:"social_development,omitempty"`
	ParentCarers         *[]ContactEntryPatch    `json:"parent_carers,omitempty"`
	EmergencyContacts    *[]ContactEntryPatch    `json:"emergency_contacts,omitempty"`
	AuthorisedCollectors *[]ContactEntryPatch    `json:"authorised_collectors,omitempty"`
	Collection           *CollectionPatch        `json:"collection,omitempty"`
	FundingSupport       *FundingSupportPatch    `json:"funding_support,omitempty"`
	RoutineCare          *RoutineCarePatch       `json:"routine_care,omitempty"`
	GDPRDeclaration      *GDPRDeclarationPatch   `json:"gdpr_declaration,omitempty"`
}

type DemographicsHomePatch struct {
	Sex                      *string         `json:"sex,omitempty"`
	Religion                 *string         `json:"religion,omitempty"`
	EthnicOrigin             *string         `json:"ethnic_origin,omitempty"`
	FirstLanguage            *string         `json:"first_language,omitempty"`
	OtherLanguages           *string         `json:"other_languages,omitempty"`
	HomeAddress              *map[string]any `json:"home_address,omitempty"`
	HomePostcode             *string         `json:"home_postcode,omitempty"`
	HomeTelephone            *string         `json:"home_telephone,omitempty"`
	DisabilityStatus         *string         `json:"disability_status,omitempty"`
	DisabilityNotes          *string         `json:"disability_notes,omitempty"`
	AccessRequirements       *string         `json:"access_requirements,omitempty"`
	DemographicsHomeReviewed *bool           `json:"demographics_home_reviewed,omitempty"`
}

type MedicalDietaryPatch struct {
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
	MedicalDietaryReviewed     *bool   `json:"medical_dietary_reviewed,omitempty"`
}

type HealthContactsPatch struct {
	DoctorName             *string `json:"doctor_name,omitempty"`
	DoctorAddress          *string `json:"doctor_address,omitempty"`
	DoctorPhone            *string `json:"doctor_phone,omitempty"`
	HealthVisitorName      *string `json:"health_visitor_name,omitempty"`
	HealthVisitorAddress   *string `json:"health_visitor_address,omitempty"`
	HealthVisitorPhone     *string `json:"health_visitor_phone,omitempty"`
	HealthContactsReviewed *bool   `json:"health_contacts_reviewed,omitempty"`
}

type SocialDevelopmentPatch struct {
	SocialServicesStatus      *string                        `json:"social_services_status,omitempty"`
	SocialServicesNotes       *string                        `json:"social_services_notes,omitempty"`
	SocialWorkerName          *string                        `json:"social_worker_name,omitempty"`
	SocialWorkerPhone         *string                        `json:"social_worker_phone,omitempty"`
	SocialWorkerEmail         *string                        `json:"social_worker_email,omitempty"`
	ConcernWalking            *string                        `json:"concern_walking,omitempty"`
	ConcernSpeechLanguage     *string                        `json:"concern_speech_language,omitempty"`
	ConcernHearing            *string                        `json:"concern_hearing,omitempty"`
	ConcernSight              *string                        `json:"concern_sight,omitempty"`
	ConcernEmotionalWellbeing *string                        `json:"concern_emotional_wellbeing,omitempty"`
	ConcernBehaviour          *string                        `json:"concern_behaviour,omitempty"`
	ProfessionalReferrals     *[]domain.ProfessionalReferral `json:"professional_referrals,omitempty"`
	SocialDevelopmentReviewed *bool                          `json:"social_development_reviewed,omitempty"`
}

type ContactEntryPatch struct {
	FullName                  string          `json:"full_name"`
	RelationshipToChild       *string         `json:"relationship_to_child,omitempty"`
	Address                   *map[string]any `json:"address,omitempty"`
	Telephone                 *string         `json:"telephone,omitempty"`
	Email                     *string         `json:"email,omitempty"`
	WorkAddress               *map[string]any `json:"work_address,omitempty"`
	HasParentalResponsibility *bool           `json:"has_parental_responsibility,omitempty"`
}

type CollectionPatch struct {
	Over18CollectionAcknowledged *bool `json:"over18_collection_acknowledged,omitempty"`
	EmergencyCollectionReviewed  *bool `json:"emergency_collection_reviewed,omitempty"`
}

type FundingSupportPatch struct {
	BenefitsContributeToFees *string `json:"benefits_contribute_to_fees,omitempty"`
	WorkingTaxCredit         *string `json:"working_tax_credit,omitempty"`
	CollegeUniPaidToParent   *string `json:"college_uni_paid_to_parent,omitempty"`
	CollegeUniPaidToNursery  *string `json:"college_uni_paid_to_nursery,omitempty"`
	Funding3yoTermTime       *string `json:"funding_3yo_term_time,omitempty"`
	Funding2yoTermTime       *string `json:"funding_2yo_term_time,omitempty"`
	FundingSupportNotes      *string `json:"funding_support_notes,omitempty"`
	FundingSupportReviewed   *bool   `json:"funding_support_reviewed,omitempty"`
}

type RoutineCarePatch struct {
	RoutineCareNotes    *string `json:"routine_care_notes,omitempty"`
	RoutineCareReviewed *bool   `json:"routine_care_reviewed,omitempty"`
}

type GDPRDeclarationPatch struct {
	GDPRDeclaredByName  *string `json:"gdpr_declared_by_name,omitempty"`
	GDPRDeclarationDate *string `json:"gdpr_declaration_date,omitempty"`
}

func MergePatch(p *domain.Profile, patch PatchSection) ([]domain.SectionCode, error) {
	changed := make([]domain.SectionCode, 0)

	if patch.DemographicsHome != nil {
		if err := applyDemographicsHome(p, *patch.DemographicsHome); err != nil {
			return nil, err
		}
		changed = append(changed, domain.SectionDemographicsHome)
	}
	if patch.MedicalDietary != nil {
		if err := applyMedicalDietary(p, *patch.MedicalDietary); err != nil {
			return nil, err
		}
		changed = append(changed, domain.SectionMedicalDietary)
	}
	if patch.HealthContacts != nil {
		applyHealthContacts(p, *patch.HealthContacts)
		changed = append(changed, domain.SectionHealthContacts)
	}
	if patch.SocialDevelopment != nil {
		if err := applySocialDevelopment(p, *patch.SocialDevelopment); err != nil {
			return nil, err
		}
		changed = append(changed, domain.SectionSocialDevelopment)
	}
	if patch.FundingSupport != nil {
		if err := applyFundingSupport(p, *patch.FundingSupport); err != nil {
			return nil, err
		}
		changed = append(changed, domain.SectionFundingSupport)
	}
	if patch.RoutineCare != nil {
		applyRoutineCare(p, *patch.RoutineCare)
		changed = append(changed, domain.SectionRoutineCare)
	}
	if patch.GDPRDeclaration != nil {
		if err := applyGDPRDeclaration(p, *patch.GDPRDeclaration); err != nil {
			return nil, err
		}
		changed = append(changed, domain.SectionGDPRDeclaration)
	}
	if patch.Collection != nil {
		applyCollection(p, *patch.Collection)
		changed = append(changed, domain.SectionEmergencyCollection)
	}
	if patch.ParentCarers != nil || patch.EmergencyContacts != nil || patch.AuthorisedCollectors != nil {
		changed = append(changed, domain.SectionParentResponsibility)
		changed = append(changed, domain.SectionEmergencyCollection)
	}

	return changed, nil
}

func applyDemographicsHome(p *domain.Profile, patch DemographicsHomePatch) error {
	if patch.Sex != nil {
		v := strings.TrimSpace(*patch.Sex)
		if v == "" {
			p.Sex = nil
		} else {
			p.Sex = &v
		}
	}
	if patch.Religion != nil {
		v := strings.TrimSpace(*patch.Religion)
		if v == "" {
			p.Religion = nil
		} else {
			p.Religion = &v
		}
	}
	if patch.EthnicOrigin != nil {
		v := strings.TrimSpace(*patch.EthnicOrigin)
		if v == "" {
			p.EthnicOrigin = nil
		} else {
			p.EthnicOrigin = &v
		}
	}
	if patch.FirstLanguage != nil {
		v := strings.TrimSpace(*patch.FirstLanguage)
		if v == "" {
			p.FirstLanguage = nil
		} else {
			p.FirstLanguage = &v
		}
	}
	if patch.OtherLanguages != nil {
		v := strings.TrimSpace(*patch.OtherLanguages)
		if v == "" {
			p.OtherLanguages = nil
		} else {
			p.OtherLanguages = &v
		}
	}
	if patch.HomeAddress != nil {
		p.HomeAddress = *patch.HomeAddress
	}
	if patch.HomePostcode != nil {
		v := strings.TrimSpace(*patch.HomePostcode)
		if v == "" {
			p.HomePostcode = nil
		} else {
			p.HomePostcode = &v
		}
	}
	if patch.HomeTelephone != nil {
		v := strings.TrimSpace(*patch.HomeTelephone)
		if v == "" {
			p.HomeTelephone = nil
		} else {
			p.HomeTelephone = &v
		}
	}
	if patch.DisabilityStatus != nil {
		v, err := parseYesNoUnknown(*patch.DisabilityStatus)
		if err != nil {
			return err
		}
		p.DisabilityStatus = v
	}
	if patch.DisabilityNotes != nil {
		v := strings.TrimSpace(*patch.DisabilityNotes)
		if v == "" {
			p.DisabilityNotes = nil
		} else {
			p.DisabilityNotes = &v
		}
	}
	if patch.AccessRequirements != nil {
		v := strings.TrimSpace(*patch.AccessRequirements)
		if v == "" {
			p.AccessRequirements = nil
		} else {
			p.AccessRequirements = &v
		}
	}
	if patch.DemographicsHomeReviewed != nil {
		p.DemographicsHomeReviewed = *patch.DemographicsHomeReviewed
	}
	return nil
}

func applyMedicalDietary(p *domain.Profile, patch MedicalDietaryPatch) error {
	if patch.MedicalConditionsStatus != nil {
		v, err := parseYesNoUnknown(*patch.MedicalConditionsStatus)
		if err != nil {
			return err
		}
		p.MedicalConditionsStatus = v
	}
	if patch.MedicalConditionsNotes != nil {
		v := strings.TrimSpace(*patch.MedicalConditionsNotes)
		if v == "" {
			p.MedicalConditionsNotes = nil
		} else {
			p.MedicalConditionsNotes = &v
		}
	}
	if patch.PrescribedMedicationStatus != nil {
		v, err := parseYesNoUnknown(*patch.PrescribedMedicationStatus)
		if err != nil {
			return err
		}
		p.PrescribedMedicationStatus = v
	}
	if patch.MedicationNotes != nil {
		v := strings.TrimSpace(*patch.MedicationNotes)
		if v == "" {
			p.MedicationNotes = nil
		} else {
			p.MedicationNotes = &v
		}
	}
	if patch.ImmunisationStatus != nil {
		v, err := parseImmunisationStatus(*patch.ImmunisationStatus)
		if err != nil {
			return err
		}
		p.ImmunisationStatus = v
	}
	if patch.ImmunisationCountry != nil {
		v := strings.TrimSpace(*patch.ImmunisationCountry)
		if v == "" {
			p.ImmunisationCountry = nil
		} else {
			p.ImmunisationCountry = &v
		}
	}
	if patch.IllnessDiagnosisHistory != nil {
		v := strings.TrimSpace(*patch.IllnessDiagnosisHistory)
		if v == "" {
			p.IllnessDiagnosisHistory = nil
		} else {
			p.IllnessDiagnosisHistory = &v
		}
	}
	if patch.DietaryRequirementsStatus != nil {
		v, err := parseYesNoUnknown(*patch.DietaryRequirementsStatus)
		if err != nil {
			return err
		}
		p.DietaryRequirementsStatus = v
	}
	if patch.DietaryRequirementsNotes != nil {
		v := strings.TrimSpace(*patch.DietaryRequirementsNotes)
		if v == "" {
			p.DietaryRequirementsNotes = nil
		} else {
			p.DietaryRequirementsNotes = &v
		}
	}
	if patch.DietarySideEffects != nil {
		v := strings.TrimSpace(*patch.DietarySideEffects)
		if v == "" {
			p.DietarySideEffects = nil
		} else {
			p.DietarySideEffects = &v
		}
	}
	if patch.MedicalDietaryReviewed != nil {
		p.MedicalDietaryReviewed = *patch.MedicalDietaryReviewed
	}
	return nil
}

func applyHealthContacts(p *domain.Profile, patch HealthContactsPatch) {
	if patch.DoctorName != nil {
		v := strings.TrimSpace(*patch.DoctorName)
		if v == "" {
			p.DoctorName = nil
		} else {
			p.DoctorName = &v
		}
	}
	if patch.DoctorAddress != nil {
		v := strings.TrimSpace(*patch.DoctorAddress)
		if v == "" {
			p.DoctorAddress = nil
		} else {
			p.DoctorAddress = &v
		}
	}
	if patch.DoctorPhone != nil {
		v := strings.TrimSpace(*patch.DoctorPhone)
		if v == "" {
			p.DoctorPhone = nil
		} else {
			p.DoctorPhone = &v
		}
	}
	if patch.HealthVisitorName != nil {
		v := strings.TrimSpace(*patch.HealthVisitorName)
		if v == "" {
			p.HealthVisitorName = nil
		} else {
			p.HealthVisitorName = &v
		}
	}
	if patch.HealthVisitorAddress != nil {
		v := strings.TrimSpace(*patch.HealthVisitorAddress)
		if v == "" {
			p.HealthVisitorAddress = nil
		} else {
			p.HealthVisitorAddress = &v
		}
	}
	if patch.HealthVisitorPhone != nil {
		v := strings.TrimSpace(*patch.HealthVisitorPhone)
		if v == "" {
			p.HealthVisitorPhone = nil
		} else {
			p.HealthVisitorPhone = &v
		}
	}
	if patch.HealthContactsReviewed != nil {
		p.HealthContactsReviewed = *patch.HealthContactsReviewed
	}
}

func applySocialDevelopment(p *domain.Profile, patch SocialDevelopmentPatch) error {
	if patch.SocialServicesStatus != nil {
		v, err := parseYesNoUnknown(*patch.SocialServicesStatus)
		if err != nil {
			return err
		}
		p.SocialServicesStatus = v
	}
	if patch.SocialServicesNotes != nil {
		v := strings.TrimSpace(*patch.SocialServicesNotes)
		if v == "" {
			p.SocialServicesNotes = nil
		} else {
			p.SocialServicesNotes = &v
		}
	}
	if patch.SocialWorkerName != nil {
		v := strings.TrimSpace(*patch.SocialWorkerName)
		if v == "" {
			p.SocialWorkerName = nil
		} else {
			p.SocialWorkerName = &v
		}
	}
	if patch.SocialWorkerPhone != nil {
		v := strings.TrimSpace(*patch.SocialWorkerPhone)
		if v == "" {
			p.SocialWorkerPhone = nil
		} else {
			p.SocialWorkerPhone = &v
		}
	}
	if patch.SocialWorkerEmail != nil {
		v := strings.TrimSpace(*patch.SocialWorkerEmail)
		if v == "" {
			p.SocialWorkerEmail = nil
		} else {
			p.SocialWorkerEmail = &v
		}
	}
	if patch.ConcernWalking != nil {
		v, err := parseYesNoUnknown(*patch.ConcernWalking)
		if err != nil {
			return err
		}
		p.ConcernWalking = v
	}
	if patch.ConcernSpeechLanguage != nil {
		v, err := parseYesNoUnknown(*patch.ConcernSpeechLanguage)
		if err != nil {
			return err
		}
		p.ConcernSpeechLanguage = v
	}
	if patch.ConcernHearing != nil {
		v, err := parseYesNoUnknown(*patch.ConcernHearing)
		if err != nil {
			return err
		}
		p.ConcernHearing = v
	}
	if patch.ConcernSight != nil {
		v, err := parseYesNoUnknown(*patch.ConcernSight)
		if err != nil {
			return err
		}
		p.ConcernSight = v
	}
	if patch.ConcernEmotionalWellbeing != nil {
		v, err := parseYesNoUnknown(*patch.ConcernEmotionalWellbeing)
		if err != nil {
			return err
		}
		p.ConcernEmotionalWellbeing = v
	}
	if patch.ConcernBehaviour != nil {
		v, err := parseYesNoUnknown(*patch.ConcernBehaviour)
		if err != nil {
			return err
		}
		p.ConcernBehaviour = v
	}
	if patch.ProfessionalReferrals != nil {
		for _, r := range *patch.ProfessionalReferrals {
			if r.Type == "other" && (r.Notes == nil || strings.TrimSpace(*r.Notes) == "") {
				return domainerrors.Validation("Invalid request payload.", "professional_referrals")
			}
			if r.WaitingListStatus != "unknown" && r.WaitingListStatus != "no" && r.WaitingListStatus != "yes" {
				return domainerrors.Validation("Invalid request payload.", "professional_referrals.waiting_list_status")
			}
		}
		p.ProfessionalReferrals = *patch.ProfessionalReferrals
	}
	if patch.SocialDevelopmentReviewed != nil {
		p.SocialDevelopmentReviewed = *patch.SocialDevelopmentReviewed
	}
	return nil
}

func applyFundingSupport(p *domain.Profile, patch FundingSupportPatch) error {
	if patch.BenefitsContributeToFees != nil {
		v, err := parseYesNoUnknown(*patch.BenefitsContributeToFees)
		if err != nil {
			return err
		}
		p.BenefitsContributeToFees = v
	}
	if patch.WorkingTaxCredit != nil {
		v, err := parseYesNoUnknown(*patch.WorkingTaxCredit)
		if err != nil {
			return err
		}
		p.WorkingTaxCredit = v
	}
	if patch.CollegeUniPaidToParent != nil {
		v, err := parseYesNoUnknown(*patch.CollegeUniPaidToParent)
		if err != nil {
			return err
		}
		p.CollegeUniPaidToParent = v
	}
	if patch.CollegeUniPaidToNursery != nil {
		v, err := parseYesNoUnknown(*patch.CollegeUniPaidToNursery)
		if err != nil {
			return err
		}
		p.CollegeUniPaidToNursery = v
	}
	if patch.Funding3yoTermTime != nil {
		v, err := parseYesNoUnknown(*patch.Funding3yoTermTime)
		if err != nil {
			return err
		}
		p.Funding3yoTermTime = v
	}
	if patch.Funding2yoTermTime != nil {
		v, err := parseYesNoUnknown(*patch.Funding2yoTermTime)
		if err != nil {
			return err
		}
		p.Funding2yoTermTime = v
	}
	if patch.FundingSupportNotes != nil {
		v := strings.TrimSpace(*patch.FundingSupportNotes)
		if v == "" {
			p.FundingSupportNotes = nil
		} else {
			p.FundingSupportNotes = &v
		}
	}
	if patch.FundingSupportReviewed != nil {
		p.FundingSupportReviewed = *patch.FundingSupportReviewed
	}
	return nil
}

func applyRoutineCare(p *domain.Profile, patch RoutineCarePatch) {
	if patch.RoutineCareNotes != nil {
		v := strings.TrimSpace(*patch.RoutineCareNotes)
		if v == "" {
			p.RoutineCareNotes = nil
		} else {
			p.RoutineCareNotes = &v
		}
	}
	if patch.RoutineCareReviewed != nil {
		p.RoutineCareReviewed = *patch.RoutineCareReviewed
	}
}

func applyGDPRDeclaration(p *domain.Profile, patch GDPRDeclarationPatch) error {
	if patch.GDPRDeclaredByName != nil {
		v := strings.TrimSpace(*patch.GDPRDeclaredByName)
		if v == "" {
			p.GDPRDeclaredByName = nil
		} else {
			p.GDPRDeclaredByName = &v
		}
	}
	if patch.GDPRDeclarationDate != nil {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*patch.GDPRDeclarationDate))
		if err != nil {
			return domainerrors.Validation("Invalid request payload.", "gdpr_declaration_date")
		}
		p.GDPRDeclarationDate = &t
	}
	return nil
}

func applyCollection(p *domain.Profile, patch CollectionPatch) {
	if patch.Over18CollectionAcknowledged != nil {
		p.Over18CollectionAcknowledged = *patch.Over18CollectionAcknowledged
	}
	if patch.EmergencyCollectionReviewed != nil {
		p.EmergencyCollectionReviewed = *patch.EmergencyCollectionReviewed
	}
}

func parseYesNoUnknown(v string) (domain.YesNoUnknown, error) {
	switch v {
	case "unknown":
		return domain.YesNoUnknownUnknown, nil
	case "no":
		return domain.YesNoUnknownNo, nil
	case "yes":
		return domain.YesNoUnknownYes, nil
	default:
		return "", domainerrors.Validation("Invalid request payload.", "yes_no_unknown")
	}
}

func parseImmunisationStatus(v string) (domain.ImmunisationStatus, error) {
	switch v {
	case "unknown":
		return domain.ImmunisationUnknown, nil
	case "up_to_date":
		return domain.ImmunisationUpToDate, nil
	case "refused":
		return domain.ImmunisationRefused, nil
	case "partial":
		return domain.ImmunisationPartial, nil
	case "not_recorded":
		return domain.ImmunisationNotRecorded, nil
	default:
		return "", domainerrors.Validation("Invalid request payload.", "immunisation_status")
	}
}

func SubmittedContactTypes(patch PatchSection) []domain.ContactType {
	types := make([]domain.ContactType, 0)
	if patch.ParentCarers != nil {
		types = append(types, domain.ContactTypeParentCarer)
	}
	if patch.EmergencyContacts != nil {
		types = append(types, domain.ContactTypeEmergencyContact)
	}
	if patch.AuthorisedCollectors != nil {
		types = append(types, domain.ContactTypeAuthorisedCollector)
	}
	return types
}

func BuildContactEntries(contactType domain.ContactType, patches []ContactEntryPatch, profileID, tenantID, branchID, childID uuid.UUID) []domain.ContactEntry {
	entries := make([]domain.ContactEntry, 0, len(patches))
	for i, p := range patches {
		addr := map[string]any{}
		if p.Address != nil {
			addr = *p.Address
		}
		workAddr := map[string]any{}
		if p.WorkAddress != nil {
			workAddr = *p.WorkAddress
		}
		entries = append(entries, domain.ContactEntry{
			ID:                        uid.NewUUID(),
			TenantID:                  tenantID,
			BranchID:                  branchID,
			ProfileID:                 profileID,
			ChildID:                   childID,
			ContactType:               contactType,
			SortOrder:                 i,
			FullName:                  strings.TrimSpace(p.FullName),
			RelationshipToChild:       p.RelationshipToChild,
			Address:                   addr,
			Telephone:                 p.Telephone,
			Email:                     p.Email,
			WorkAddress:               workAddr,
			HasParentalResponsibility: p.HasParentalResponsibility,
		})
	}
	return entries
}

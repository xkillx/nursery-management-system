package application

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"nursery-management-system/api/internal/modules/children/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

func bcryptGeneratePassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func buildChildProfileFromInput(tenantID, branchID, childID uuid.UUID, in *ChildProfileInput) *domain.ChildProfile {
	p := &domain.ChildProfile{
		ID:                          uuid.New(),
		TenantID:                    tenantID,
		BranchID:                    branchID,
		ChildID:                     childID,
		Sex:                         in.Sex,
		Religion:                    in.Religion,
		EthnicOrigin:                in.EthnicOrigin,
		FirstLanguage:               in.FirstLanguage,
		OtherLanguages:              in.OtherLanguages,
		HomePostcode:                in.HomePostcode,
		HomeTelephone:               in.HomeTelephone,
		DisabilityStatus:            domain.YesNoUnknown(in.DisabilityStatus),
		DisabilityNotes:             in.DisabilityNotes,
		AccessRequirements:          in.AccessRequirements,
		RoutineCareNotes:            in.RoutineCareNotes,
		GDPRDeclaredByName:          in.GDPRDeclaredByName,
		DemographicsHomeReviewed:    in.DemographicsHomeReviewed,
		MedicalDietaryReviewed:      in.MedicalDietaryReviewed,
		HealthContactsReviewed:      in.HealthContactsReviewed,
		SocialDevelopmentReviewed:   in.SocialDevelopmentReviewed,
		ParentResponsibilityReviewed: in.ParentResponsibilityReviewed,
		EmergencyCollectionReviewed: in.EmergencyCollectionReviewed,
		RoutineCareReviewed:         in.RoutineCareReviewed,
	}
	if in.HomeAddress != nil {
		p.HomeAddress = in.HomeAddress
	} else {
		p.HomeAddress = map[string]any{}
	}
	if s := trimEmptyToNil(in.GDPRDeclaredAt); s != nil {
		if t, err := time.Parse(time.RFC3339, *s); err == nil {
			p.GDPRDeclaredAt = &t
		}
	}
	if s := trimEmptyToNil(in.GDPRDeclarationDate); s != nil {
		if t, err := time.Parse("2006-01-02", *s); err == nil {
			p.GDPRDeclarationDate = &t
		}
	}
	if s := trimEmptyToNil(in.RegistrationDate); s != nil {
		if t, err := time.Parse("2006-01-02", *s); err == nil {
			p.RegistrationDate = &t
		}
	}
	return p
}

func buildChildHealthFromInput(tenantID, branchID, childID uuid.UUID, in *ChildHealthProfileInput) *domain.ChildHealthProfile {
	return &domain.ChildHealthProfile{
		ID:                         uuid.New(),
		TenantID:                   tenantID,
		BranchID:                   branchID,
		ChildID:                    childID,
		MedicalConditionsStatus:    domain.YesNoUnknown(in.MedicalConditionsStatus),
		MedicalConditionsNotes:     in.MedicalConditionsNotes,
		PrescribedMedicationStatus: domain.YesNoUnknown(in.PrescribedMedicationStatus),
		MedicationNotes:            in.MedicationNotes,
		ImmunisationStatus:         domain.ImmunisationStatus(in.ImmunisationStatus),
		ImmunisationCountry:        in.ImmunisationCountry,
		IllnessDiagnosisHistory:    in.IllnessDiagnosisHistory,
		DietaryRequirementsStatus:  domain.YesNoUnknown(in.DietaryRequirementsStatus),
		DietaryRequirementsNotes:   in.DietaryRequirementsNotes,
		DietarySideEffects:         in.DietarySideEffects,
		DoctorName:                 in.DoctorName,
		DoctorAddress:              in.DoctorAddress,
		DoctorPhone:                in.DoctorPhone,
		HealthVisitorName:          in.HealthVisitorName,
		HealthVisitorAddress:       in.HealthVisitorAddress,
		HealthVisitorPhone:         in.HealthVisitorPhone,
	}
}

func buildChildSafeguardingFromInput(tenantID, branchID, childID uuid.UUID, in *ChildSafeguardingProfileInput) *domain.ChildSafeguardingProfile {
	p := &domain.ChildSafeguardingProfile{
		ID:                          uuid.New(),
		TenantID:                    tenantID,
		BranchID:                    branchID,
		ChildID:                     childID,
		SocialServicesStatus:        domain.YesNoUnknown(in.SocialServicesStatus),
		SocialServicesNotes:         in.SocialServicesNotes,
		SocialWorkerName:            in.SocialWorkerName,
		SocialWorkerPhone:           in.SocialWorkerPhone,
		SocialWorkerEmail:           in.SocialWorkerEmail,
		ConcernWalking:              domain.YesNoUnknown(in.ConcernWalking),
		ConcernSpeechLanguage:       domain.YesNoUnknown(in.ConcernSpeechLanguage),
		ConcernHearing:              domain.YesNoUnknown(in.ConcernHearing),
		ConcernSight:                domain.YesNoUnknown(in.ConcernSight),
		ConcernEmotionalWellbeing:   domain.YesNoUnknown(in.ConcernEmotionalWellbeing),
		ConcernBehaviour:            domain.YesNoUnknown(in.ConcernBehaviour),
		RestrictedNotes:             in.RestrictedNotes,
	}
	if in.ProfessionalReferrals != nil {
		p.ProfessionalReferrals = in.ProfessionalReferrals
	} else {
		p.ProfessionalReferrals = []domain.ProfessionalReferral{}
	}
	return p
}

func buildChildContactEntries(tenantID, branchID, childID uuid.UUID, inputs []ChildContactInput) []domain.ChildContact {
	entries := make([]domain.ChildContact, 0, len(inputs))
	for i, src := range inputs {
		e := domain.ChildContact{
			ID:                        uuid.New(),
			TenantID:                  tenantID,
			BranchID:                  branchID,
			ChildID:                   childID,
			ContactType:               src.ContactType,
			SortOrder:                 i,
			FullName:                  src.FullName,
			RelationshipToChild:       src.RelationshipToChild,
			Telephone:                 src.Telephone,
			Email:                     src.Email,
			HasParentalResponsibility: src.HasParentalResponsibility,
		}
		if src.Address != nil {
			e.Address = src.Address
		} else {
			e.Address = map[string]any{}
		}
		if src.WorkAddress != nil {
			e.WorkAddress = src.WorkAddress
		} else {
			e.WorkAddress = map[string]any{}
		}
		entries = append(entries, e)
	}
	return entries
}

func buildChildFundingFromInput(tenantID, branchID, childID uuid.UUID, in *ChildFundingRecordInput) *domain.ChildFundingRecord {
	return &domain.ChildFundingRecord{
		ID:                          uuid.New(),
		TenantID:                    tenantID,
		BranchID:                    branchID,
		ChildID:                     childID,
		BenefitsContributeToFees:    domain.YesNoUnknown(in.BenefitsContributeToFees),
		WorkingTaxCredit:            domain.YesNoUnknown(in.WorkingTaxCredit),
		CollegeUniPaidToParent:      domain.YesNoUnknown(in.CollegeUniPaidToParent),
		CollegeUniPaidToNursery:     domain.YesNoUnknown(in.CollegeUniPaidToNursery),
		Funding3yoTermTime:          domain.YesNoUnknown(in.Funding3yoTermTime),
		Funding2yoTermTime:          domain.YesNoUnknown(in.Funding2yoTermTime),
		FundingSupportNotes:         in.FundingSupportNotes,
		FundingSupportReviewed:      in.FundingSupportReviewed,
	}
}

func trimEmptyToNil(s *string) *string {
	if s == nil {
		return nil
	}
	if strings.TrimSpace(*s) == "" {
		return nil
	}
	return s
}

// ValidateReasonCode checks a reason code and note for lifecycle actions.
func ValidateReasonCode(code, note string) error {
	code = strings.TrimSpace(code)
	note = strings.TrimSpace(note)

	if code == "" {
		return domainerrors.New("child_lifecycle_reason_required", "Invalid request payload.", "reason_code")
	}
	if _, ok := domain.ValidReasonCodes[domain.ReasonCode(code)]; !ok {
		return domainerrors.New("lifecycle_reason_invalid", "Invalid request payload.", "reason_code")
	}
	if len(note) > maxReasonNoteLen {
		return domainerrors.Validation("Invalid request payload.", "reason_note")
	}
	if code == string(domain.ReasonOther) && note == "" {
		return domainerrors.New("reason_note_required_for_other", "Invalid request payload.", "reason_note")
	}
	return nil
}
